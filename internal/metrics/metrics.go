package metrics

import (
	"errors"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"task-service/internal/domain"
)

const (
	cacheResultHit     = "hit"
	cacheResultMiss    = "miss"
	cacheResultSuccess = "success"
	cacheResultError   = "error"

	outboxResultClaimed    = "claimed"
	outboxResultProcessed  = "processed"
	outboxResultRetried    = "retried"
	outboxResultDeadLetter = "dead_letter"
	runResultSuccess       = "success"
	runResultError         = "error"
)

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_service_http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "task_service_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	cacheRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_service_cache_requests_total",
		Help: "Total number of cache lookup requests.",
	}, []string{"cache", "operation", "result"})
	cacheOperations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_service_cache_operations_total",
		Help: "Total number of cache write and invalidation operations.",
	}, []string{"cache", "operation", "result"})

	dbErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_service_db_errors_total",
		Help: "Total number of unexpected database operation errors.",
	}, []string{"repository", "operation"})

	outboxEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_service_outbox_events_total",
		Help: "Total number of outbox events by processing result.",
	}, []string{"result"})
	outboxDispatchRuns = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_service_outbox_dispatch_runs_total",
		Help: "Total number of outbox dispatch runs.",
	}, []string{"result"})
	outboxDispatchErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_service_outbox_dispatch_errors_total",
		Help: "Total number of outbox dispatch errors by stage.",
	}, []string{"stage"})
	outboxCleanupRuns = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_service_outbox_cleanup_runs_total",
		Help: "Total number of outbox cleanup runs.",
	}, []string{"result"})
	outboxCleanupDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "task_service_outbox_cleanup_deleted_total",
		Help: "Total number of processed outbox events deleted by cleanup.",
	})
)

func ObserveHTTPRequest(method, path string, status int, duration time.Duration) {
	statusCode := strconv.Itoa(status)
	httpRequests.WithLabelValues(method, path, statusCode).Inc()
	httpDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

func RecordCacheLookup(cache, operation string, hit bool, err error) {
	result := cacheResultMiss
	if err != nil {
		result = cacheResultError
	} else if hit {
		result = cacheResultHit
	}
	cacheRequests.WithLabelValues(cache, operation, result).Inc()
}

func RecordCacheOperation(cache, operation string, err error) {
	result := cacheResultSuccess
	if err != nil {
		result = cacheResultError
	}
	cacheOperations.WithLabelValues(cache, operation, result).Inc()
}

func RecordDBError(repository, operation string, err error) {
	if err == nil || isExpectedDomainError(err) {
		return
	}
	dbErrors.WithLabelValues(repository, operation).Inc()
}

func RecordOutboxDispatch(claimed, processed, retried, deadLettered int, errorStage string, err error) {
	addOutboxEvents(outboxResultClaimed, claimed)
	addOutboxEvents(outboxResultProcessed, processed)
	addOutboxEvents(outboxResultRetried, retried)
	addOutboxEvents(outboxResultDeadLetter, deadLettered)

	if err != nil {
		outboxDispatchRuns.WithLabelValues(runResultError).Inc()
		if errorStage == "" {
			errorStage = "unknown"
		}
		outboxDispatchErrors.WithLabelValues(errorStage).Inc()
		return
	}
	outboxDispatchRuns.WithLabelValues(runResultSuccess).Inc()
}

func RecordOutboxCleanup(deleted int64, err error) {
	if err != nil {
		outboxCleanupRuns.WithLabelValues(runResultError).Inc()
		return
	}
	outboxCleanupRuns.WithLabelValues(runResultSuccess).Inc()
	if deleted > 0 {
		outboxCleanupDeleted.Add(float64(deleted))
	}
}

func addOutboxEvents(result string, count int) {
	if count <= 0 {
		return
	}
	outboxEvents.WithLabelValues(result).Add(float64(count))
}

func isExpectedDomainError(err error) bool {
	return errors.Is(err, domain.ErrNotFound) ||
		errors.Is(err, domain.ErrConflict) ||
		errors.Is(err, domain.ErrUnauthorized) ||
		errors.Is(err, domain.ErrForbidden) ||
		errors.Is(err, domain.ErrPayloadTooLarge) ||
		errors.Is(err, domain.ErrInvalidInput)
}
