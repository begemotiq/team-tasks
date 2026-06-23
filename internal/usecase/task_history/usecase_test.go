package task_history

import (
	"context"
	"errors"
	"testing"

	"task-service/internal/domain"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleChecksMembershipAndReturnsHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskHistoryReader(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	uc := New(tasks, teams)
	task := &models.Task{ID: 100, TeamID: 1}
	history := []models.TaskHistory{{ID: 1, TaskID: 100, Field: "created"}}

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(task, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleMember, nil)
	tasks.EXPECT().History(gomock.Any(), int64(100)).Return(history, nil)

	result, err := uc.GetHistory(context.Background(), 10, 100)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if len(result) != 1 || result[0].Field != "created" {
		t.Fatalf("unexpected history: %#v", result)
	}
}

func TestHandleRejectsNonMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskHistoryReader(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	uc := New(tasks, teams)

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(&models.Task{ID: 100, TeamID: 1}, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.Role(""), domain.ErrNotFound)

	_, err := uc.GetHistory(context.Background(), 10, 100)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestHandleReturnsTaskLookupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskHistoryReader(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	uc := New(tasks, teams)

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(nil, domain.ErrNotFound)

	_, err := uc.GetHistory(context.Background(), 10, 100)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestHandleReturnsHistoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskHistoryReader(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	uc := New(tasks, teams)

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(&models.Task{ID: 100, TeamID: 1}, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleMember, nil)
	tasks.EXPECT().History(gomock.Any(), int64(100)).Return(nil, domain.ErrNotFound)

	_, err := uc.GetHistory(context.Background(), 10, 100)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
