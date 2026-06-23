package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"task-service/internal/config"
	"task-service/internal/domain"
)

type Service struct {
	endpoint string
	client   *http.Client
	breaker  *CircuitBreaker
}

func NewService(cfg config.EmailConfig) *Service {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}
	return &Service{
		endpoint: cfg.Endpoint,
		client:   &http.Client{Timeout: timeout},
		breaker:  NewCircuitBreaker(cfg.FailureThreshold, cfg.OpenTimeout),
	}
}

func (s *Service) SendInvite(ctx context.Context, toEmail string, teamName string) error {
	return s.breaker.Execute(func() error {
		if s.endpoint == "" {
			return nil
		}
		payload, err := json.Marshal(map[string]string{
			"email": toEmail,
			"team":  teamName,
		})
		if err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := s.client.Do(req)
		if err != nil {
			return fmt.Errorf("%w: %v", domain.ErrExternal, err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return fmt.Errorf("%w: email service returned %d", domain.ErrExternal, resp.StatusCode)
		}
		return nil
	})
}

type CircuitBreaker struct {
	mu        sync.Mutex
	failures  int
	threshold int
	openedAt  time.Time
	openFor   time.Duration
}

func NewCircuitBreaker(threshold int, openFor time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 3
	}
	if openFor <= 0 {
		openFor = 30 * time.Second
	}
	return &CircuitBreaker{threshold: threshold, openFor: openFor}
}

func (b *CircuitBreaker) Execute(fn func() error) error {
	if !b.allow() {
		return fmt.Errorf("%w: circuit breaker is open", domain.ErrExternal)
	}
	err := fn()
	b.record(err)
	return err
}

func (b *CircuitBreaker) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.openedAt.IsZero() {
		return true
	}
	if time.Since(b.openedAt) >= b.openFor {
		return true
	}
	return false
}

func (b *CircuitBreaker) record(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err == nil {
		b.failures = 0
		b.openedAt = time.Time{}
		return
	}
	b.failures++
	if b.failures >= b.threshold {
		b.openedAt = time.Now()
	}
}
