package team_create

import (
	"context"
	"errors"
	"testing"

	"task-service/internal/domain"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleCreatesTeamWithOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	teams := NewMockteamCreator(ctrl)
	uc := New(teams)

	teams.EXPECT().
		CreateWithOwner(gomock.Any(), gomock.AssignableToTypeOf(&models.Team{}), int64(10)).
		DoAndReturn(func(_ context.Context, team *models.Team, ownerID int64) error {
			if team.Name != "Backend" || team.CreatedBy != ownerID {
				t.Fatalf("unexpected team before create: %#v", team)
			}
			team.ID = 1
			return nil
		})

	team, err := uc.Create(context.Background(), 10, Input{Name: "Backend"})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if team.ID != 1 || team.Name != "Backend" {
		t.Fatalf("unexpected team: %#v", team)
	}
}

func TestHandleReturnsCreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	teams := NewMockteamCreator(ctrl)
	uc := New(teams)

	teams.EXPECT().
		CreateWithOwner(gomock.Any(), gomock.AssignableToTypeOf(&models.Team{}), int64(10)).
		Return(domain.ErrConflict)

	_, err := uc.Create(context.Background(), 10, Input{Name: "Backend"})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}
