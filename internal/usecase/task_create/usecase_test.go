package task_create

import (
	"context"
	"errors"
	"testing"

	"task-service/internal/domain"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleCreatesTaskWritesHistoryAndInvalidatesCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskCreator(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	assigneeID := int64(2)

	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleMember, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), assigneeID).Return(models.RoleMember, nil)
	tasks.EXPECT().
		CreateWithHistory(
			gomock.Any(),
			gomock.AssignableToTypeOf(&models.Task{}),
			gomock.AssignableToTypeOf(&models.TaskHistory{}),
		).
		DoAndReturn(func(_ context.Context, task *models.Task, history *models.TaskHistory) error {
			if task.Title != "Implement API" || task.TeamID != 1 || task.CreatedBy != 10 || task.AssigneeID == nil || *task.AssigneeID != assigneeID {
				t.Fatalf("unexpected task before create: %#v", task)
			}
			if history.TaskID != 0 || history.ChangedBy != 10 || history.Field != "created" {
				t.Fatalf("unexpected history before create: %#v", history)
			}
			task.ID = 100
			history.TaskID = task.ID
			return nil
		})
	cache.EXPECT().DeleteTeamTasks(gomock.Any(), int64(1)).Return(nil)

	task, err := uc.Create(context.Background(), 10, Input{
		Title:      "Implement API",
		Status:     models.TaskStatusTodo,
		TeamID:     1,
		AssigneeID: &assigneeID,
	})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if task.ID != 100 {
		t.Fatalf("unexpected task: %#v", task)
	}
}

func TestHandleRejectsAssigneeOutsideTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskCreator(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	assigneeID := int64(2)

	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleMember, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), assigneeID).Return(models.Role(""), domain.ErrNotFound)

	_, err := uc.Create(context.Background(), 10, Input{Title: "Implement API", TeamID: 1, AssigneeID: &assigneeID})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestHandleReturnsExternalErrorWhenCacheInvalidationFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskCreator(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)

	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleMember, nil)
	tasks.EXPECT().
		CreateWithHistory(gomock.Any(), gomock.AssignableToTypeOf(&models.Task{}), gomock.AssignableToTypeOf(&models.TaskHistory{})).
		DoAndReturn(func(_ context.Context, task *models.Task, _ *models.TaskHistory) error {
			task.ID = 100
			return nil
		})
	cache.EXPECT().DeleteTeamTasks(gomock.Any(), int64(1)).Return(errors.New("redis unavailable"))

	_, err := uc.Create(context.Background(), 10, Input{
		Title:  "Implement API",
		Status: models.TaskStatusTodo,
		TeamID: 1,
	})
	if !errors.Is(err, domain.ErrExternal) {
		t.Fatalf("expected external error, got %v", err)
	}
}

func TestHandleReturnsCreatorMembershipError(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskCreator(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)

	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.Role(""), domain.ErrNotFound)

	_, err := uc.Create(context.Background(), 10, Input{Title: "Implement API", TeamID: 1})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestHandleReturnsTaskCreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskCreator(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)

	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleMember, nil)
	tasks.EXPECT().
		CreateWithHistory(gomock.Any(), gomock.AssignableToTypeOf(&models.Task{}), gomock.AssignableToTypeOf(&models.TaskHistory{})).
		Return(domain.ErrInvalidInput)

	_, err := uc.Create(context.Background(), 10, Input{Title: "Implement API", TeamID: 1})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}
