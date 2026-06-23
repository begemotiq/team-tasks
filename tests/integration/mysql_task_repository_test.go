//go:build integration

package integration

import (
	"errors"
	"testing"
	"time"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

func TestMySQLTaskRepositoryCRUDAndList(t *testing.T) {
	t.Parallel()

	fixture := newFixture(t)
	owner := fixture.user("owner")
	assignee := fixture.user("assignee")
	team := fixture.team("backend", owner)
	fixture.member(team, assignee, models.RoleMember)

	dueDate := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	task := &models.Task{
		Title:       "Implement API " + fixture.suffix,
		Description: "Task description " + fixture.suffix,
		Status:      models.TaskStatusTodo,
		AssigneeID:  &assignee.ID,
		TeamID:      team.ID,
		CreatedBy:   owner.ID,
		DueDate:     &dueDate,
	}
	if err := fixture.repos.tasks.Create(fixture.ctx, task); err != nil {
		t.Fatal(err)
	}
	if task.ID == 0 {
		t.Fatal("created task id is empty")
	}
	if task.CreatedAt.IsZero() || task.UpdatedAt.IsZero() {
		t.Fatalf("task timestamps are empty: %#v", task)
	}

	found, err := fixture.repos.tasks.GetByID(fixture.ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.ID != task.ID || found.Title != task.Title || found.AssigneeID == nil || *found.AssigneeID != assignee.ID {
		t.Fatalf("unexpected task by id: %#v", found)
	}

	found.Status = models.TaskStatusDone
	found.AssigneeID = nil
	found.DueDate = nil
	oldVersion := found.Version
	if err := fixture.repos.tasks.Update(fixture.ctx, found); err != nil {
		t.Fatal(err)
	}

	updated, err := fixture.repos.tasks.GetByID(fixture.ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != models.TaskStatusDone || updated.AssigneeID != nil || updated.DueDate != nil {
		t.Fatalf("unexpected updated task: %#v", updated)
	}
	if updated.Version != oldVersion+1 {
		t.Fatalf("expected task version to increment from %d to %d, got %d", oldVersion, oldVersion+1, updated.Version)
	}

	todo := fixture.task("todo", team, owner, models.TaskStatusTodo, assignee)
	fixture.task("in-progress", team, owner, models.TaskStatusInProgress, assignee)
	status := models.TaskStatusTodo
	filter := models.TaskFilter{
		TeamID:     &team.ID,
		Status:     &status,
		AssigneeID: &assignee.ID,
		PageSize:   20,
	}

	list, err := fixture.repos.tasks.List(fixture.ctx, filter, owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 || list.Items[0].ID != todo.ID {
		t.Fatalf("unexpected filtered tasks: %#v", list)
	}

	_, err = fixture.repos.tasks.GetByID(fixture.ctx, 9_999_999_999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected missing task to return not found, got %v", err)
	}
}

func TestMySQLTaskRepositoryRejectsStaleUpdate(t *testing.T) {
	t.Parallel()

	fixture := newFixture(t)
	owner := fixture.user("owner")
	team := fixture.team("backend", owner)
	task := fixture.task("optimistic-lock", team, owner, models.TaskStatusTodo, nil)

	fresh := *task
	fresh.Title = "Fresh update " + fixture.suffix
	if err := fixture.repos.tasks.Update(fixture.ctx, &fresh); err != nil {
		t.Fatal(err)
	}

	stale := *task
	stale.Description = "stale update must not win"
	err := fixture.repos.tasks.Update(fixture.ctx, &stale)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected stale update conflict, got %v", err)
	}

	current, err := fixture.repos.tasks.GetByID(fixture.ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Title != fresh.Title || current.Description == stale.Description {
		t.Fatalf("stale update overwrote current task: %#v", current)
	}
}

func TestMySQLTaskRepositoryCursorPagination(t *testing.T) {
	t.Parallel()

	fixture := newFixture(t)
	owner := fixture.user("owner")
	team := fixture.team("backend", owner)

	first := fixture.task("first", team, owner, models.TaskStatusTodo, nil)
	second := fixture.task("second", team, owner, models.TaskStatusTodo, nil)

	filter := models.TaskFilter{TeamID: &team.ID, PageSize: 1}
	page, err := fixture.repos.tasks.List(fixture.ctx, filter, owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || !page.HasMore || page.NextCursor == nil {
		t.Fatalf("unexpected first page: %#v", page)
	}
	if page.Items[0].ID != second.ID {
		t.Fatalf("expected newest task %d on first page, got %#v", second.ID, page.Items)
	}

	filter.Cursor = page.NextCursor
	page, err = fixture.repos.tasks.List(fixture.ctx, filter, owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != first.ID {
		t.Fatalf("unexpected second page: %#v", page)
	}
}

func TestMySQLTaskRepositoryHistoryTransactions(t *testing.T) {
	t.Parallel()

	fixture := newFixture(t)
	owner := fixture.user("owner")
	team := fixture.team("backend", owner)

	task := &models.Task{
		Title:       "Atomic create " + fixture.suffix,
		Description: "Task description " + fixture.suffix,
		Status:      models.TaskStatusTodo,
		TeamID:      team.ID,
		CreatedBy:   owner.ID,
	}
	if err := fixture.repos.tasks.CreateWithHistory(fixture.ctx, task, &models.TaskHistory{
		ChangedBy: owner.ID,
		Field:     "created",
		OldValue:  "",
		NewValue:  string(models.TaskStatusTodo),
	}); err != nil {
		t.Fatal(err)
	}

	history, err := fixture.repos.tasks.History(fixture.ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].TaskID != task.ID || history[0].Field != "created" {
		t.Fatalf("unexpected create history: %#v", history)
	}

	updated := *task
	updated.Status = models.TaskStatusDone
	if err := fixture.repos.tasks.UpdateWithHistory(fixture.ctx, &updated, []models.TaskHistory{{
		ChangedBy: owner.ID,
		Field:     "status",
		OldValue:  string(models.TaskStatusTodo),
		NewValue:  string(models.TaskStatusDone),
	}}); err != nil {
		t.Fatal(err)
	}

	history, err = fixture.repos.tasks.History(fixture.ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 {
		t.Fatalf("unexpected update history: %#v", history)
	}

	rolledBackTask := &models.Task{
		Title:       "Rollback create " + fixture.suffix,
		Description: "Task description " + fixture.suffix,
		Status:      models.TaskStatusTodo,
		TeamID:      team.ID,
		CreatedBy:   owner.ID,
	}
	err = fixture.repos.tasks.CreateWithHistory(fixture.ctx, rolledBackTask, &models.TaskHistory{
		ChangedBy: owner.ID + 9_999_999,
		Field:     "created",
		OldValue:  "",
		NewValue:  string(models.TaskStatusTodo),
	})
	if err == nil {
		t.Fatal("expected create with invalid history author to fail")
	}
	if rolledBackTask.ID != 0 {
		_, err = fixture.repos.tasks.GetByID(fixture.ctx, rolledBackTask.ID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected task to be rolled back, got %v", err)
		}
	}

	current, err := fixture.repos.tasks.GetByID(fixture.ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	failedUpdate := *current
	failedUpdate.Status = models.TaskStatusInProgress
	err = fixture.repos.tasks.UpdateWithHistory(fixture.ctx, &failedUpdate, []models.TaskHistory{{
		ChangedBy: owner.ID + 9_999_999,
		Field:     "status",
		OldValue:  string(models.TaskStatusDone),
		NewValue:  string(models.TaskStatusInProgress),
	}})
	if err == nil {
		t.Fatal("expected update with invalid history author to fail")
	}

	current, err = fixture.repos.tasks.GetByID(fixture.ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Status != models.TaskStatusDone {
		t.Fatalf("expected failed update to be rolled back, got status %q", current.Status)
	}
}
