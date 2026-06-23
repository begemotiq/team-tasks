//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package outbox_dispatch

import (
	"context"
	"time"

	"task-service/internal/domain/models"
)

type eventStore interface {
	ClaimPending(ctx context.Context, limit int) ([]models.OutboxEvent, error)
	MarkProcessed(ctx context.Context, id int64, claimToken string) error
	MarkFailed(ctx context.Context, id int64, claimToken string, retryAt time.Time, message string) error
	MarkDeadLetter(ctx context.Context, id int64, claimToken string, message string) error
}

type inviteSender interface {
	SendInvite(ctx context.Context, toEmail string, teamName string) error
}
