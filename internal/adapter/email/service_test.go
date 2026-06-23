package email

import (
	"errors"
	"testing"
	"time"

	"task-service/internal/domain"
)

func TestCircuitBreakerOpensAfterThreshold(t *testing.T) {
	breaker := NewCircuitBreaker(2, time.Minute)
	fail := errors.New("email failed")
	calls := 0

	for i := 0; i < 2; i++ {
		err := breaker.Execute(func() error {
			calls++
			return fail
		})
		if !errors.Is(err, fail) {
			t.Fatalf("expected original failure, got %v", err)
		}
	}

	err := breaker.Execute(func() error {
		calls++
		return nil
	})
	if !errors.Is(err, domain.ErrExternal) {
		t.Fatalf("expected circuit-open external error, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("open circuit should not call dependency, calls=%d", calls)
	}
}
