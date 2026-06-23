//go:build integration

package integration

import (
	"errors"
	"testing"
	"time"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

func TestMySQLOutboxRepository(t *testing.T) {
	fixture := newFixture(t)
	owner := fixture.user("owner")
	member := fixture.user("member")
	deadLetterMember := fixture.user("dead-letter-member")
	team := fixture.team("backend", owner)

	event, err := models.NewTeamInviteEmailEvent(member.Email, team.Name)
	if err != nil {
		t.Fatal(err)
	}
	if err := fixture.repos.teams.AddMemberWithOutboxEvent(fixture.ctx, team.ID, member.ID, models.RoleMember, event); err != nil {
		t.Fatal(err)
	}
	if event.ID == 0 {
		t.Fatal("outbox event id is empty")
	}

	events, err := fixture.repos.outbox.ClaimPending(fixture.ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].ID != event.ID || events[0].Type != models.OutboxEventTypeTeamInviteEmail {
		t.Fatalf("unexpected claimed events: %#v", events)
	}
	firstClaimToken := events[0].ClaimToken
	if firstClaimToken == "" {
		t.Fatal("claimed event token is empty")
	}
	payload, err := events[0].TeamInviteEmailPayload()
	if err != nil {
		t.Fatal(err)
	}
	if payload.Email != member.Email || payload.TeamName != team.Name {
		t.Fatalf("unexpected outbox payload: %#v", payload)
	}

	if err := fixture.repos.outbox.MarkFailed(fixture.ctx, event.ID, firstClaimToken, time.Now().Add(-time.Second), "email is down"); err != nil {
		t.Fatal(err)
	}
	events, err = fixture.repos.outbox.ClaimPending(fixture.ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].ID != event.ID || events[0].Attempts != 1 {
		t.Fatalf("unexpected retried event: %#v", events)
	}
	secondClaimToken := events[0].ClaimToken
	if secondClaimToken == "" || secondClaimToken == firstClaimToken {
		t.Fatalf("unexpected second claim token %q after first %q", secondClaimToken, firstClaimToken)
	}

	err = fixture.repos.outbox.MarkProcessed(fixture.ctx, event.ID, firstClaimToken)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected stale claim token to reject processed mark, got %v", err)
	}
	if err := fixture.repos.outbox.MarkProcessed(fixture.ctx, event.ID, secondClaimToken); err != nil {
		t.Fatal(err)
	}
	err = fixture.repos.outbox.MarkFailed(fixture.ctx, event.ID, secondClaimToken, time.Now().Add(-time.Second), "late worker")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected processed event to reject stale failure mark, got %v", err)
	}
	events, err = fixture.repos.outbox.ClaimPending(fixture.ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("processed event was claimed again: %#v", events)
	}

	deadLetterEvent, err := models.NewTeamInviteEmailEvent(deadLetterMember.Email, team.Name)
	if err != nil {
		t.Fatal(err)
	}
	if err := fixture.repos.teams.AddMemberWithOutboxEvent(fixture.ctx, team.ID, deadLetterMember.ID, models.RoleMember, deadLetterEvent); err != nil {
		t.Fatal(err)
	}
	events, err = fixture.repos.outbox.ClaimPending(fixture.ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].ID != deadLetterEvent.ID {
		t.Fatalf("unexpected dead-letter candidate: %#v", events)
	}
	deadLetterClaimToken := events[0].ClaimToken
	if deadLetterClaimToken == "" {
		t.Fatal("dead-letter candidate token is empty")
	}
	if err := fixture.repos.outbox.MarkDeadLetter(fixture.ctx, deadLetterEvent.ID, deadLetterClaimToken, "invalid payload"); err != nil {
		t.Fatal(err)
	}
	err = fixture.repos.outbox.MarkProcessed(fixture.ctx, deadLetterEvent.ID, deadLetterClaimToken)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected dead-letter event to reject stale processed mark, got %v", err)
	}
	events, err = fixture.repos.outbox.ClaimPending(fixture.ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("dead-letter event was claimed again: %#v", events)
	}

	deleted, err := fixture.repos.outbox.DeleteProcessedBefore(fixture.ctx, time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Fatalf("expected one processed event to be deleted, got %d", deleted)
	}
}

func TestMySQLOutboxRepositoryRollsBackMemberWhenEventInsertFails(t *testing.T) {
	fixture := newFixture(t)
	owner := fixture.user("owner")
	member := fixture.user("member")
	team := fixture.team("backend", owner)

	event := &models.OutboxEvent{
		Type:    models.OutboxEventTypeTeamInviteEmail,
		Payload: []byte("{bad-json"),
		Status:  models.OutboxEventStatusPending,
	}

	err := fixture.repos.teams.AddMemberWithOutboxEvent(fixture.ctx, team.ID, member.ID, models.RoleMember, event)
	if err == nil {
		t.Fatal("expected invalid JSON payload to fail")
	}

	_, err = fixture.repos.teams.GetMemberRole(fixture.ctx, team.ID, member.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected team member insert to be rolled back, got %v", err)
	}
}
