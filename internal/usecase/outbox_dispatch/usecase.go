package outbox_dispatch

import (
	"context"
	"errors"
	"fmt"
	"time"

	"task-service/internal/domain/models"
)

const defaultMaxAttempts = 5

var errPermanent = errors.New("permanent outbox error")

type Result struct {
	Claimed      int
	Processed    int
	Retried      int
	DeadLettered int
	ErrorStage   string
}

type UseCase struct {
	events      eventStore
	email       inviteSender
	retryDelay  time.Duration
	maxAttempts int
}

func New(events eventStore, email inviteSender, retryDelay time.Duration, maxAttempts int) *UseCase {
	if retryDelay <= 0 {
		retryDelay = time.Minute
	}
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxAttempts
	}
	return &UseCase{events: events, email: email, retryDelay: retryDelay, maxAttempts: maxAttempts}
}

func (uc *UseCase) Dispatch(ctx context.Context, limit int) (Result, error) {
	result := Result{}
	events, err := uc.events.ClaimPending(ctx, limit)
	if err != nil {
		result.ErrorStage = "claim"
		return result, err
	}
	result.Claimed = len(events)

	for _, event := range events {
		if err := uc.dispatch(ctx, event); err != nil {
			if errors.Is(err, errPermanent) || event.Attempts+1 >= uc.maxAttempts {
				if markErr := uc.events.MarkDeadLetter(ctx, event.ID, event.ClaimToken, err.Error()); markErr != nil {
					result.ErrorStage = "mark_dead_letter"
					return result, markErr
				}
				result.DeadLettered++
				continue
			}
			retryAt := time.Now().Add(uc.retryDelay)
			if markErr := uc.events.MarkFailed(ctx, event.ID, event.ClaimToken, retryAt, err.Error()); markErr != nil {
				result.ErrorStage = "mark_failed"
				return result, markErr
			}
			result.Retried++
			continue
		}
		if err := uc.events.MarkProcessed(ctx, event.ID, event.ClaimToken); err != nil {
			result.ErrorStage = "mark_processed"
			return result, err
		}
		result.Processed++
	}
	return result, nil
}

func (uc *UseCase) dispatch(ctx context.Context, event models.OutboxEvent) error {
	switch event.Type {
	case models.OutboxEventTypeTeamInviteEmail:
		payload, err := event.TeamInviteEmailPayload()
		if err != nil {
			return permanentError(fmt.Errorf("invalid team invite payload: %w", err))
		}
		return uc.email.SendInvite(ctx, payload.Email, payload.TeamName)
	default:
		return permanentError(fmt.Errorf("unsupported outbox event type %q", event.Type))
	}
}

func permanentError(err error) error {
	return fmt.Errorf("%w: %v", errPermanent, err)
}
