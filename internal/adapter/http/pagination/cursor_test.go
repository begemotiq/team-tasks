package pagination_test

import (
	"errors"
	"testing"
	"time"

	"task-service/internal/adapter/http/pagination"
	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

func TestTaskCursorRoundTrip(t *testing.T) {
	cursor := &models.TaskCursor{
		CreatedAt: time.Date(2026, 6, 22, 10, 30, 0, 123, time.UTC),
		ID:        100,
	}

	token := pagination.EncodeTaskCursor(cursor)
	if token == "" {
		t.Fatal("expected encoded cursor")
	}

	decoded, err := pagination.DecodeTaskCursor(token)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if decoded.ID != cursor.ID || !decoded.CreatedAt.Equal(cursor.CreatedAt) {
		t.Fatalf("unexpected cursor: %#v", decoded)
	}
}

func TestDecodeTaskCursorRejectsInvalidValue(t *testing.T) {
	if _, err := pagination.DecodeTaskCursor("bad"); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}
