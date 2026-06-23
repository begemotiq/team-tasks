package outbox_cleanup

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

func TestHandleDeletesProcessedEventsBeforeRetention(t *testing.T) {
	ctrl := gomock.NewController(t)
	events := NewMockeventCleaner(ctrl)
	uc := New(events, 21*24*time.Hour)

	events.EXPECT().
		DeleteProcessedBefore(gomock.Any(), gomock.AssignableToTypeOf(time.Time{})).
		DoAndReturn(func(_ context.Context, before time.Time) (int64, error) {
			if time.Since(before) < 20*24*time.Hour {
				t.Fatalf("unexpected cleanup cutoff: %v", before)
			}
			return 3, nil
		})

	deleted, err := uc.Cleanup(context.Background())
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if deleted != 3 {
		t.Fatalf("unexpected deleted count: %d", deleted)
	}
}
