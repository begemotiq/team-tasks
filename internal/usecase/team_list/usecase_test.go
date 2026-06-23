package team_list

import (
	"context"
	"testing"

	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleReturnsUserTeams(t *testing.T) {
	ctrl := gomock.NewController(t)
	teams := NewMockteamLister(ctrl)
	uc := New(teams)
	expected := []models.Team{{ID: 1, Name: "Backend"}}

	teams.EXPECT().ListByUser(gomock.Any(), int64(10)).Return(expected, nil)

	result, err := uc.List(context.Background(), 10)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if len(result) != 1 || result[0].ID != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
}
