package team_invite

import (
	"context"
	"errors"
	"testing"

	"task-service/internal/domain"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

type mockTeamInviteStore struct {
	*MockteamReader
	*MockteamMemberInviter
}

func TestHandleInvitesMemberAndCreatesOutboxEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	teamReader := NewMockteamReader(ctrl)
	memberInviter := NewMockteamMemberInviter(ctrl)
	users := NewMockinviteUserFinder(ctrl)
	uc := New(mockTeamInviteStore{MockteamReader: teamReader, MockteamMemberInviter: memberInviter}, users)

	teamReader.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleAdmin, nil)
	users.EXPECT().FindByEmail(gomock.Any(), "member@example.com").Return(&models.User{ID: 20, Email: "member@example.com"}, nil)
	teamReader.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&models.Team{ID: 1, Name: "Backend"}, nil)
	memberInviter.EXPECT().
		AddMemberWithOutboxEvent(gomock.Any(), int64(1), int64(20), models.RoleMember, gomock.AssignableToTypeOf(&models.OutboxEvent{})).
		DoAndReturn(func(_ context.Context, _, _ int64, _ models.Role, event *models.OutboxEvent) error {
			if event.Type != models.OutboxEventTypeTeamInviteEmail {
				t.Fatalf("unexpected event type: %#v", event)
			}
			payload, err := event.TeamInviteEmailPayload()
			if err != nil {
				t.Fatalf("decode event payload: %v", err)
			}
			if payload.Email != "member@example.com" || payload.TeamName != "Backend" {
				t.Fatalf("unexpected event payload: %#v", payload)
			}
			return nil
		})

	if err := uc.Invite(context.Background(), 10, 1, Input{Email: "member@example.com"}); err != nil {
		t.Fatalf("handle failed: %v", err)
	}
}

func TestHandleRejectsMemberInviter(t *testing.T) {
	ctrl := gomock.NewController(t)
	teamReader := NewMockteamReader(ctrl)
	memberInviter := NewMockteamMemberInviter(ctrl)
	users := NewMockinviteUserFinder(ctrl)
	uc := New(mockTeamInviteStore{MockteamReader: teamReader, MockteamMemberInviter: memberInviter}, users)

	teamReader.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleMember, nil)

	err := uc.Invite(context.Background(), 10, 1, Input{Email: "member@example.com"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestHandleRejectsOwnerRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	teamReader := NewMockteamReader(ctrl)
	memberInviter := NewMockteamMemberInviter(ctrl)
	users := NewMockinviteUserFinder(ctrl)
	uc := New(mockTeamInviteStore{MockteamReader: teamReader, MockteamMemberInviter: memberInviter}, users)

	teamReader.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleAdmin, nil)

	err := uc.Invite(context.Background(), 10, 1, Input{Email: "member@example.com", Role: models.RoleOwner})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestHandleReturnsInviterRoleLookupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	teamReader := NewMockteamReader(ctrl)
	memberInviter := NewMockteamMemberInviter(ctrl)
	users := NewMockinviteUserFinder(ctrl)
	uc := New(mockTeamInviteStore{MockteamReader: teamReader, MockteamMemberInviter: memberInviter}, users)

	teamReader.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.Role(""), domain.ErrNotFound)

	err := uc.Invite(context.Background(), 10, 1, Input{Email: "member@example.com"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestHandleReturnsUserLookupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	teamReader := NewMockteamReader(ctrl)
	memberInviter := NewMockteamMemberInviter(ctrl)
	users := NewMockinviteUserFinder(ctrl)
	uc := New(mockTeamInviteStore{MockteamReader: teamReader, MockteamMemberInviter: memberInviter}, users)

	teamReader.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleAdmin, nil)
	users.EXPECT().FindByEmail(gomock.Any(), "missing@example.com").Return(nil, domain.ErrNotFound)

	err := uc.Invite(context.Background(), 10, 1, Input{Email: "missing@example.com"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestHandleReturnsTeamLookupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	teamReader := NewMockteamReader(ctrl)
	memberInviter := NewMockteamMemberInviter(ctrl)
	users := NewMockinviteUserFinder(ctrl)
	uc := New(mockTeamInviteStore{MockteamReader: teamReader, MockteamMemberInviter: memberInviter}, users)

	teamReader.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleAdmin, nil)
	users.EXPECT().FindByEmail(gomock.Any(), "member@example.com").Return(&models.User{ID: 20, Email: "member@example.com"}, nil)
	teamReader.EXPECT().FindByID(gomock.Any(), int64(1)).Return(nil, domain.ErrNotFound)

	err := uc.Invite(context.Background(), 10, 1, Input{Email: "member@example.com"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestHandleReturnsAddMemberError(t *testing.T) {
	ctrl := gomock.NewController(t)
	teamReader := NewMockteamReader(ctrl)
	memberInviter := NewMockteamMemberInviter(ctrl)
	users := NewMockinviteUserFinder(ctrl)
	uc := New(mockTeamInviteStore{MockteamReader: teamReader, MockteamMemberInviter: memberInviter}, users)

	teamReader.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleAdmin, nil)
	users.EXPECT().FindByEmail(gomock.Any(), "member@example.com").Return(&models.User{ID: 20, Email: "member@example.com"}, nil)
	teamReader.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&models.Team{ID: 1, Name: "Backend"}, nil)
	memberInviter.EXPECT().
		AddMemberWithOutboxEvent(gomock.Any(), int64(1), int64(20), models.RoleAdmin, gomock.AssignableToTypeOf(&models.OutboxEvent{})).
		Return(domain.ErrConflict)

	err := uc.Invite(context.Background(), 10, 1, Input{Email: "member@example.com", Role: models.RoleAdmin})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}
