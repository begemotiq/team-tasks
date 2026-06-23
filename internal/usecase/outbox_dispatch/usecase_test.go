package outbox_dispatch

import (
	"context"
	"errors"
	"testing"
	"time"

	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleSendsInviteAndMarksProcessed(t *testing.T) {
	ctrl := gomock.NewController(t)
	events := NewMockeventStore(ctrl)
	email := NewMockinviteSender(ctrl)
	uc := New(events, email, time.Minute, 3)
	event, err := models.NewTeamInviteEmailEvent("member@example.com", "Backend")
	if err != nil {
		t.Fatal(err)
	}
	event.ID = 100
	event.ClaimToken = "claim-token"

	events.EXPECT().ClaimPending(gomock.Any(), 10).Return([]models.OutboxEvent{*event}, nil)
	email.EXPECT().SendInvite(gomock.Any(), "member@example.com", "Backend").Return(nil)
	events.EXPECT().MarkProcessed(gomock.Any(), int64(100), "claim-token").Return(nil)

	result, err := uc.Dispatch(context.Background(), 10)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if result.Claimed != 1 || result.Processed != 1 {
		t.Fatalf("unexpected dispatch result: %+v", result)
	}
}

func TestHandleMarksFailedWhenInviteSendFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	events := NewMockeventStore(ctrl)
	email := NewMockinviteSender(ctrl)
	uc := New(events, email, time.Minute, 3)
	event, err := models.NewTeamInviteEmailEvent("member@example.com", "Backend")
	if err != nil {
		t.Fatal(err)
	}
	event.ID = 100
	event.ClaimToken = "claim-token"
	sendErr := errors.New("email is down")

	events.EXPECT().ClaimPending(gomock.Any(), 10).Return([]models.OutboxEvent{*event}, nil)
	email.EXPECT().SendInvite(gomock.Any(), "member@example.com", "Backend").Return(sendErr)
	events.EXPECT().
		MarkFailed(gomock.Any(), int64(100), "claim-token", gomock.AssignableToTypeOf(time.Time{}), "email is down").
		Return(nil)

	result, err := uc.Dispatch(context.Background(), 10)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if result.Claimed != 1 || result.Retried != 1 {
		t.Fatalf("unexpected dispatch result: %+v", result)
	}
}

func TestHandleMarksDeadLetterWhenMaxAttemptsReached(t *testing.T) {
	ctrl := gomock.NewController(t)
	events := NewMockeventStore(ctrl)
	email := NewMockinviteSender(ctrl)
	uc := New(events, email, time.Minute, 3)
	event, err := models.NewTeamInviteEmailEvent("member@example.com", "Backend")
	if err != nil {
		t.Fatal(err)
	}
	event.ID = 100
	event.ClaimToken = "claim-token"
	event.Attempts = 2
	sendErr := errors.New("email is down")

	events.EXPECT().ClaimPending(gomock.Any(), 10).Return([]models.OutboxEvent{*event}, nil)
	email.EXPECT().SendInvite(gomock.Any(), "member@example.com", "Backend").Return(sendErr)
	events.EXPECT().
		MarkDeadLetter(gomock.Any(), int64(100), "claim-token", "email is down").
		Return(nil)

	result, err := uc.Dispatch(context.Background(), 10)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if result.Claimed != 1 || result.DeadLettered != 1 {
		t.Fatalf("unexpected dispatch result: %+v", result)
	}
}

func TestHandleMarksDeadLetterForInvalidPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	events := NewMockeventStore(ctrl)
	email := NewMockinviteSender(ctrl)
	uc := New(events, email, time.Minute, 3)
	event := models.OutboxEvent{
		ID:         100,
		Type:       models.OutboxEventTypeTeamInviteEmail,
		Payload:    []byte("{bad-json"),
		ClaimToken: "claim-token",
	}

	events.EXPECT().ClaimPending(gomock.Any(), 10).Return([]models.OutboxEvent{event}, nil)
	events.EXPECT().
		MarkDeadLetter(gomock.Any(), int64(100), "claim-token", gomock.AssignableToTypeOf("")).
		Return(nil)

	result, err := uc.Dispatch(context.Background(), 10)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if result.Claimed != 1 || result.DeadLettered != 1 {
		t.Fatalf("unexpected dispatch result: %+v", result)
	}
}

func TestHandleMarksDeadLetterForUnsupportedEventType(t *testing.T) {
	ctrl := gomock.NewController(t)
	events := NewMockeventStore(ctrl)
	email := NewMockinviteSender(ctrl)
	uc := New(events, email, time.Minute, 3)
	event := models.OutboxEvent{
		ID:         100,
		Type:       models.OutboxEventType("unknown"),
		ClaimToken: "claim-token",
	}

	events.EXPECT().ClaimPending(gomock.Any(), 10).Return([]models.OutboxEvent{event}, nil)
	events.EXPECT().
		MarkDeadLetter(gomock.Any(), int64(100), "claim-token", gomock.AssignableToTypeOf("")).
		Return(nil)

	result, err := uc.Dispatch(context.Background(), 10)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if result.Claimed != 1 || result.DeadLettered != 1 {
		t.Fatalf("unexpected dispatch result: %+v", result)
	}
}

func TestHandleReturnsClaimError(t *testing.T) {
	ctrl := gomock.NewController(t)
	events := NewMockeventStore(ctrl)
	email := NewMockinviteSender(ctrl)
	uc := New(events, email, time.Minute, 3)
	claimErr := errors.New("claim failed")

	events.EXPECT().ClaimPending(gomock.Any(), 10).Return(nil, claimErr)

	result, err := uc.Dispatch(context.Background(), 10)
	if !errors.Is(err, claimErr) {
		t.Fatalf("expected claim error, got %v", err)
	}
	if result.ErrorStage != "claim" {
		t.Fatalf("unexpected error stage: %+v", result)
	}
}

func TestHandleReturnsMarkProcessedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	events := NewMockeventStore(ctrl)
	email := NewMockinviteSender(ctrl)
	uc := New(events, email, time.Minute, 3)
	event, err := models.NewTeamInviteEmailEvent("member@example.com", "Backend")
	if err != nil {
		t.Fatal(err)
	}
	event.ID = 100
	event.ClaimToken = "claim-token"
	markErr := errors.New("mark processed failed")

	events.EXPECT().ClaimPending(gomock.Any(), 10).Return([]models.OutboxEvent{*event}, nil)
	email.EXPECT().SendInvite(gomock.Any(), "member@example.com", "Backend").Return(nil)
	events.EXPECT().MarkProcessed(gomock.Any(), int64(100), "claim-token").Return(markErr)

	result, err := uc.Dispatch(context.Background(), 10)
	if !errors.Is(err, markErr) {
		t.Fatalf("expected mark error, got %v", err)
	}
	if result.ErrorStage != "mark_processed" {
		t.Fatalf("unexpected error stage: %+v", result)
	}
}

func TestHandleReturnsMarkFailedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	events := NewMockeventStore(ctrl)
	email := NewMockinviteSender(ctrl)
	uc := New(events, email, time.Minute, 3)
	event, err := models.NewTeamInviteEmailEvent("member@example.com", "Backend")
	if err != nil {
		t.Fatal(err)
	}
	event.ID = 100
	event.ClaimToken = "claim-token"
	markErr := errors.New("mark failed failed")

	events.EXPECT().ClaimPending(gomock.Any(), 10).Return([]models.OutboxEvent{*event}, nil)
	email.EXPECT().SendInvite(gomock.Any(), "member@example.com", "Backend").Return(errors.New("email is down"))
	events.EXPECT().
		MarkFailed(gomock.Any(), int64(100), "claim-token", gomock.AssignableToTypeOf(time.Time{}), "email is down").
		Return(markErr)

	result, err := uc.Dispatch(context.Background(), 10)
	if !errors.Is(err, markErr) {
		t.Fatalf("expected mark error, got %v", err)
	}
	if result.ErrorStage != "mark_failed" {
		t.Fatalf("unexpected error stage: %+v", result)
	}
}

func TestHandleReturnsMarkDeadLetterError(t *testing.T) {
	ctrl := gomock.NewController(t)
	events := NewMockeventStore(ctrl)
	email := NewMockinviteSender(ctrl)
	uc := New(events, email, time.Minute, 3)
	event := models.OutboxEvent{ID: 100, Type: models.OutboxEventType("unknown"), ClaimToken: "claim-token"}
	markErr := errors.New("mark dead letter failed")

	events.EXPECT().ClaimPending(gomock.Any(), 10).Return([]models.OutboxEvent{event}, nil)
	events.EXPECT().
		MarkDeadLetter(gomock.Any(), int64(100), "claim-token", gomock.AssignableToTypeOf("")).
		Return(markErr)

	result, err := uc.Dispatch(context.Background(), 10)
	if !errors.Is(err, markErr) {
		t.Fatalf("expected mark error, got %v", err)
	}
	if result.ErrorStage != "mark_dead_letter" {
		t.Fatalf("unexpected error stage: %+v", result)
	}
}
