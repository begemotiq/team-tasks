package outbox_cleanup

import (
	"context"
	"time"
)

const defaultRetention = 21 * 24 * time.Hour

type UseCase struct {
	events    eventCleaner
	retention time.Duration
}

func New(events eventCleaner, retention time.Duration) *UseCase {
	if retention <= 0 {
		retention = defaultRetention
	}
	return &UseCase{events: events, retention: retention}
}

func (uc *UseCase) Cleanup(ctx context.Context) (int64, error) {
	return uc.events.DeleteProcessedBefore(ctx, time.Now().Add(-uc.retention))
}
