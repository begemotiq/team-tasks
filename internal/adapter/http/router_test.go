package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"task-service/internal/domain"
)

type rejectingTokens struct{}

func (rejectingTokens) ParseToken(_ string) (int64, error) {
	return 0, domain.ErrUnauthorized
}

func TestProtectedRoutesRateLimitBadJWT(t *testing.T) {
	router := NewRouter(Dependencies{
		Tokens:            rejectingTokens{},
		RequestsPerMinute: 1,
	})

	first := httptest.NewRequest(http.MethodPost, "/api/v1/teams", nil)
	first.RemoteAddr = "203.0.113.10:1000"
	first.Header.Set("Authorization", "Bearer bad-token")
	firstRecorder := httptest.NewRecorder()
	router.ServeHTTP(firstRecorder, first)
	if firstRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected first request to reach auth and return 401, got %d", firstRecorder.Code)
	}

	second := httptest.NewRequest(http.MethodPost, "/api/v1/teams", nil)
	second.RemoteAddr = "203.0.113.10:1001"
	second.Header.Set("Authorization", "Bearer bad-token")
	secondRecorder := httptest.NewRecorder()
	router.ServeHTTP(secondRecorder, second)
	if secondRecorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second bad JWT request to be rate limited, got %d", secondRecorder.Code)
	}
}
