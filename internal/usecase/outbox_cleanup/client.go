//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package outbox_cleanup

import (
	"context"
	"time"
)

type eventCleaner interface {
	DeleteProcessedBefore(ctx context.Context, before time.Time) (int64, error)
}
