package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"task-service/internal/adapter/http/requestctx"
)

func TestClientIPIgnoresXForwardedFor(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/v1/teams", nil)
	request.RemoteAddr = "203.0.113.10:1234"
	request.Header.Set("X-Forwarded-For", "198.51.100.50")

	if got := clientIP(request); got != "203.0.113.10" {
		t.Fatalf("expected RemoteAddr IP, got %q", got)
	}
}

func TestRateLimiterCleansExpiredBuckets(t *testing.T) {
	now := time.Date(2026, 6, 23, 10, 0, 0, 0, time.UTC)
	limiter := NewRateLimiter(10)
	limiter.now = func() time.Time { return now }

	if !limiter.Allow("ip:203.0.113.10") {
		t.Fatal("first request must be allowed")
	}

	now = now.Add(2 * time.Minute)
	if !limiter.Allow("ip:203.0.113.20") {
		t.Fatal("second key must be allowed")
	}

	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	if _, ok := limiter.buckets["ip:203.0.113.10"]; ok {
		t.Fatalf("expired bucket was not cleaned: %#v", limiter.buckets)
	}
	if _, ok := limiter.buckets["ip:203.0.113.20"]; !ok {
		t.Fatalf("new bucket is missing: %#v", limiter.buckets)
	}
}

func TestRequestIDMiddlewareGeneratesRequestID(t *testing.T) {
	var contextRequestID string
	handler := requestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		contextRequestID, ok = requestctx.RequestIDFromContext(r.Context())
		if !ok {
			t.Fatal("request id was not stored in context")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	responseRequestID := recorder.Header().Get(requestIDHeader)
	if responseRequestID == "" {
		t.Fatal("request id response header is empty")
	}
	if contextRequestID != responseRequestID {
		t.Fatalf("context request id %q does not match response header %q", contextRequestID, responseRequestID)
	}
}

func TestRequestIDMiddlewarePreservesIncomingRequestID(t *testing.T) {
	const incomingRequestID = "test-request-id"
	handler := requestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, ok := requestctx.RequestIDFromContext(r.Context())
		if !ok || requestID != incomingRequestID {
			t.Fatalf("unexpected request id in context: %q", requestID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)
	request.Header.Set(requestIDHeader, incomingRequestID)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if got := recorder.Header().Get(requestIDHeader); got != incomingRequestID {
		t.Fatalf("expected incoming request id %q, got %q", incomingRequestID, got)
	}
}

func TestRouteLabelUsesStableLabelForUnknownNotFoundRoute(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/not-found/123", nil)

	if got := routeLabel(request, http.StatusNotFound); got != "not_found" {
		t.Fatalf("expected not_found route label, got %q", got)
	}
}
