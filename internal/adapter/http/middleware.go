package http

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"net"
	stdhttp "net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"task-service/internal/adapter/http/requestctx"
	"task-service/internal/adapter/http/response"
	"task-service/internal/domain"
	"task-service/internal/metrics"
	"task-service/internal/usecase"
)

const requestIDHeader = "X-Request-ID"

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func requestIDMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		requestID := strings.TrimSpace(r.Header.Get(requestIDHeader))
		if !validRequestID(requestID) {
			requestID = newRequestID()
		}
		w.Header().Set(requestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(requestctx.WithRequestID(r.Context(), requestID)))
	})
}

func authMiddleware(tokens usecase.TokenParser) func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				response.Error(w, domain.ErrUnauthorized)
				return
			}
			userID, err := tokens.ParseToken(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
			if err != nil {
				response.Error(w, err)
				return
			}
			next.ServeHTTP(w, r.WithContext(requestctx.WithUserID(r.Context(), userID)))
		})
	}
}

func rateLimitMiddleware(limiter *RateLimiter) func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			key := "ip:" + clientIP(r)
			if userID, ok := requestctx.UserIDFromContext(r.Context()); ok {
				key = "user:" + strconv.FormatInt(userID, 10)
			}
			if !limiter.Allow(key) {
				w.Header().Set("Retry-After", "60")
				response.JSON(w, stdhttp.StatusTooManyRequests, response.NewError("rate limit exceeded"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func recoverMiddleware(logger *slog.Logger) func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.ErrorContext(r.Context(), "panic recovered",
						"request_id", requestID(r),
						"method", r.Method,
						"path", r.URL.Path,
						"panic", recovered,
						"stack", string(debug.Stack()),
					)
					response.JSON(w, stdhttp.StatusInternalServerError, response.NewError("internal server error"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func requestLogMiddleware(logger *slog.Logger) func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: stdhttp.StatusOK}
			next.ServeHTTP(recorder, r)

			route := r.URL.Path
			if chiRoute := routePattern(r); chiRoute != "" {
				route = chiRoute
			}
			logger.InfoContext(r.Context(), "http request completed",
				"request_id", requestID(r),
				"method", r.Method,
				"path", r.URL.Path,
				"route", route,
				"status", recorder.status,
				"bytes", recorder.bytes,
				"duration_ms", float64(time.Since(start).Microseconds())/1000,
				"client_ip", clientIP(r),
				"user_agent", r.UserAgent(),
			)
		})
	}
}

type statusRecorder struct {
	stdhttp.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(stdhttp.StatusOK)
	}
	written, err := r.ResponseWriter.Write(body)
	r.bytes += written
	return written, err
}

func metricsMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: stdhttp.StatusOK}
		next.ServeHTTP(recorder, r)
		route := routeLabel(r, recorder.status)
		metrics.ObserveHTTPRequest(r.Method, route, recorder.status, time.Since(start))
	})
}

func routeLabel(r *stdhttp.Request, status int) string {
	if chiRoute := routePattern(r); chiRoute != "" {
		return chiRoute
	}
	if status == stdhttp.StatusNotFound {
		return "not_found"
	}
	return "unknown"
}

func routePattern(r *stdhttp.Request) string {
	routeCtx := chi.RouteContext(r.Context())
	if routeCtx == nil {
		return ""
	}
	return routeCtx.RoutePattern()
}

func requestID(r *stdhttp.Request) string {
	requestID, _ := requestctx.RequestIDFromContext(r.Context())
	return requestID
}

func validRequestID(requestID string) bool {
	if requestID == "" || len(requestID) > 128 {
		return false
	}
	for _, char := range requestID {
		if char <= ' ' || char == 127 {
			return false
		}
	}
	return true
}

func newRequestID() string {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(data[:])
}

func clientIP(r *stdhttp.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}

type RateLimiter struct {
	mu              sync.Mutex
	limit           int
	window          time.Duration
	cleanupInterval time.Duration
	lastCleanup     time.Time
	now             func() time.Time
	buckets         map[string]*rateBucket
}

type rateBucket struct {
	count int
	reset time.Time
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 100
	}
	return &RateLimiter{
		limit:           requestsPerMinute,
		window:          time.Minute,
		cleanupInterval: time.Minute,
		now:             time.Now,
		buckets:         make(map[string]*rateBucket),
	}
}

func (l *RateLimiter) Allow(key string) bool {
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.lastCleanup.IsZero() {
		l.lastCleanup = now
	}
	if !now.Before(l.lastCleanup.Add(l.cleanupInterval)) {
		l.cleanupExpiredLocked(now)
		l.lastCleanup = now
	}

	bucket, ok := l.buckets[key]
	if !ok || !now.Before(bucket.reset) {
		l.buckets[key] = &rateBucket{count: 1, reset: now.Add(l.window)}
		return true
	}
	if bucket.count >= l.limit {
		return false
	}
	bucket.count++
	return true
}

func (l *RateLimiter) cleanupExpiredLocked(now time.Time) {
	for key, bucket := range l.buckets {
		if !now.Before(bucket.reset) {
			delete(l.buckets, key)
		}
	}
}
