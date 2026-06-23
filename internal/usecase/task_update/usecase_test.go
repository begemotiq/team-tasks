package task_update

import (
	"context"
	"errors"
	"testing"
	"time"

	"task-service/internal/domain"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleUpdatesStatusWritesHistoryAndInvalidatesCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	status := models.TaskStatusDone
	current := &models.Task{ID: 100, Title: "Implement API", Status: models.TaskStatusTodo, TeamID: 1, CreatedBy: 20}

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(current, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleOwner, nil)
	tasks.EXPECT().
		UpdateWithHistory(gomock.Any(), gomock.AssignableToTypeOf(&models.Task{}), gomock.Any()).
		DoAndReturn(func(_ context.Context, task *models.Task, history []models.TaskHistory) error {
			if task.Status != models.TaskStatusDone {
				t.Fatalf("status was not updated: %#v", task)
			}
			if len(history) != 1 || history[0].Field != "status" || history[0].OldValue != "todo" || history[0].NewValue != "done" {
				t.Fatalf("unexpected history: %#v", history)
			}
			return nil
		})
	cache.EXPECT().DeleteTeamTasks(gomock.Any(), int64(1)).Return(nil)

	updated, err := uc.Update(context.Background(), 10, 100, Input{Status: &status})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if updated.Status != models.TaskStatusDone {
		t.Fatalf("unexpected task: %#v", updated)
	}
}

func TestHandleReturnsCurrentTaskWhenNoChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	current := &models.Task{ID: 100, Title: "Implement API", Status: models.TaskStatusTodo, TeamID: 1, CreatedBy: 10}

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(current, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleMember, nil)

	updated, err := uc.Update(context.Background(), 10, 100, Input{})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if updated != current {
		t.Fatalf("expected current task pointer, got %#v", updated)
	}
}

func TestHandleReturnsTaskLookupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	lookupErr := errors.New("task lookup failed")

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(nil, lookupErr)

	_, err := uc.Update(context.Background(), 10, 100, Input{})
	if !errors.Is(err, lookupErr) {
		t.Fatalf("expected lookup error, got %v", err)
	}
}

func TestHandleReturnsActorRoleLookupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	roleErr := errors.New("role lookup failed")
	current := &models.Task{ID: 100, Title: "Implement API", Status: models.TaskStatusTodo, TeamID: 1, CreatedBy: 10}

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(current, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.Role(""), roleErr)

	_, err := uc.Update(context.Background(), 10, 100, Input{})
	if !errors.Is(err, roleErr) {
		t.Fatalf("expected role error, got %v", err)
	}
}

func TestHandleRejectsUnrelatedMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	title := "Updated"
	current := &models.Task{ID: 100, Title: "Implement API", Status: models.TaskStatusTodo, TeamID: 1, CreatedBy: 20}

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(current, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleMember, nil)

	_, err := uc.Update(context.Background(), 10, 100, Input{Title: &title})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestHandleRejectsAssigneeOutsideTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	assigneeID := int64(99)
	current := &models.Task{ID: 100, Title: "Implement API", Status: models.TaskStatusTodo, TeamID: 1, CreatedBy: 10}

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(current, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleOwner, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), assigneeID).Return(models.Role(""), domain.ErrNotFound)

	_, err := uc.Update(context.Background(), 10, 100, Input{AssigneeID: Optional[int64]{Set: true, Valid: true, Value: assigneeID}})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestHandleReturnsAssigneeMembershipError(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	assigneeID := int64(99)
	lookupErr := errors.New("membership lookup failed")
	current := &models.Task{ID: 100, Title: "Implement API", Status: models.TaskStatusTodo, TeamID: 1, CreatedBy: 10}

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(current, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleOwner, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), assigneeID).Return(models.Role(""), lookupErr)

	_, err := uc.Update(context.Background(), 10, 100, Input{AssigneeID: Optional[int64]{Set: true, Valid: true, Value: assigneeID}})
	if !errors.Is(err, lookupErr) {
		t.Fatalf("expected lookup error, got %v", err)
	}
}

func TestHandleRecordsDescriptionAssigneeAndDueDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	description := "new description"
	assigneeID := int64(20)
	dueDate := time.Now().UTC().Truncate(time.Second)
	current := &models.Task{ID: 100, Title: "Implement API", Status: models.TaskStatusTodo, TeamID: 1, CreatedBy: 10}

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(current, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleOwner, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), assigneeID).Return(models.RoleMember, nil)
	tasks.EXPECT().
		UpdateWithHistory(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ *models.Task, history []models.TaskHistory) error {
			if len(history) != 3 {
				t.Fatalf("expected 3 history records, got %#v", history)
			}
			return nil
		})
	cache.EXPECT().DeleteTeamTasks(gomock.Any(), int64(1)).Return(nil)

	updated, err := uc.Update(context.Background(), 10, 100, Input{
		Description: &description,
		AssigneeID:  Optional[int64]{Set: true, Valid: true, Value: assigneeID},
		DueDate:     Optional[time.Time]{Set: true, Valid: true, Value: dueDate},
	})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if updated.AssigneeID == nil || *updated.AssigneeID != assigneeID || updated.DueDate == nil || !updated.DueDate.Equal(dueDate) {
		t.Fatalf("unexpected task: %#v", updated)
	}
}

func TestHandleClearsAssigneeAndDueDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	assigneeID := int64(20)
	dueDate := time.Date(2026, 6, 22, 10, 30, 0, 0, time.UTC)
	current := &models.Task{
		ID:         100,
		Title:      "Implement API",
		Status:     models.TaskStatusTodo,
		TeamID:     1,
		CreatedBy:  10,
		AssigneeID: &assigneeID,
		DueDate:    &dueDate,
	}

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(current, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleOwner, nil)
	tasks.EXPECT().
		UpdateWithHistory(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, task *models.Task, history []models.TaskHistory) error {
			if task.AssigneeID != nil || task.DueDate != nil {
				t.Fatalf("expected nullable fields to be cleared: %#v", task)
			}
			if len(history) != 2 {
				t.Fatalf("expected 2 history records, got %#v", history)
			}
			if history[0].Field != "assignee_id" || history[0].OldValue != "20" || history[0].NewValue != "" {
				t.Fatalf("unexpected assignee history: %#v", history)
			}
			if history[1].Field != "due_date" || history[1].OldValue != dueDate.Format(time.RFC3339) || history[1].NewValue != "" {
				t.Fatalf("unexpected due date history: %#v", history)
			}
			return nil
		})
	cache.EXPECT().DeleteTeamTasks(gomock.Any(), int64(1)).Return(nil)

	updated, err := uc.Update(context.Background(), 10, 100, Input{
		AssigneeID: Optional[int64]{Set: true},
		DueDate:    Optional[time.Time]{Set: true},
	})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if updated.AssigneeID != nil || updated.DueDate != nil {
		t.Fatalf("unexpected task: %#v", updated)
	}
}

func TestHandleSkipsHistoryWhenOptionalFieldsAreSame(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	assigneeID := int64(20)
	dueDate := time.Date(2026, 6, 22, 10, 30, 0, 0, time.UTC)
	current := &models.Task{
		ID:         100,
		Title:      "Implement API",
		Status:     models.TaskStatusTodo,
		TeamID:     1,
		CreatedBy:  10,
		AssigneeID: &assigneeID,
		DueDate:    &dueDate,
	}

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(current, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleOwner, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), assigneeID).Return(models.RoleMember, nil)

	updated, err := uc.Update(context.Background(), 10, 100, Input{
		AssigneeID: Optional[int64]{Set: true, Valid: true, Value: assigneeID},
		DueDate:    Optional[time.Time]{Set: true, Valid: true, Value: dueDate},
	})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if updated != current {
		t.Fatalf("expected current task pointer, got %#v", updated)
	}
}

func TestHandleUpdatesTitle(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	title := "Updated API"
	current := &models.Task{ID: 100, Title: "Implement API", Status: models.TaskStatusTodo, TeamID: 1, CreatedBy: 10}

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(current, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleOwner, nil)
	tasks.EXPECT().
		UpdateWithHistory(gomock.Any(), gomock.AssignableToTypeOf(&models.Task{}), gomock.Any()).
		DoAndReturn(func(_ context.Context, task *models.Task, history []models.TaskHistory) error {
			if task.Title != title {
				t.Fatalf("title was not updated: %#v", task)
			}
			if len(history) != 1 || history[0].Field != "title" || history[0].OldValue != "Implement API" || history[0].NewValue != title {
				t.Fatalf("unexpected history: %#v", history)
			}
			return nil
		})
	cache.EXPECT().DeleteTeamTasks(gomock.Any(), int64(1)).Return(nil)

	updated, err := uc.Update(context.Background(), 10, 100, Input{Title: &title})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if updated.Title != title {
		t.Fatalf("unexpected task: %#v", updated)
	}
}

func TestHandleReturnsTaskUpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	status := models.TaskStatusDone
	updateErr := errors.New("task update failed")
	current := &models.Task{ID: 100, Title: "Implement API", Status: models.TaskStatusTodo, TeamID: 1, CreatedBy: 10}

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(current, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleOwner, nil)
	tasks.EXPECT().UpdateWithHistory(gomock.Any(), gomock.AssignableToTypeOf(&models.Task{}), gomock.Any()).Return(updateErr)

	_, err := uc.Update(context.Background(), 10, 100, Input{Status: &status})
	if !errors.Is(err, updateErr) {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestHandleReturnsExternalErrorWhenCacheInvalidationFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(tasks, teams, cache)
	status := models.TaskStatusDone
	current := &models.Task{ID: 100, Title: "Implement API", Status: models.TaskStatusTodo, TeamID: 1, CreatedBy: 10}

	tasks.EXPECT().GetByID(gomock.Any(), int64(100)).Return(current, nil)
	teams.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleOwner, nil)
	tasks.EXPECT().UpdateWithHistory(gomock.Any(), gomock.AssignableToTypeOf(&models.Task{}), gomock.Any()).Return(nil)
	cache.EXPECT().DeleteTeamTasks(gomock.Any(), int64(1)).Return(errors.New("redis unavailable"))

	_, err := uc.Update(context.Background(), 10, 100, Input{Status: &status})
	if !errors.Is(err, domain.ErrExternal) {
		t.Fatalf("expected external error, got %v", err)
	}
}
