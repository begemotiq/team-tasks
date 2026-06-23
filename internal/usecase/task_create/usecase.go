package task_create

import (
	"context"
	"errors"
	"fmt"
	"time"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

type UseCase struct {
	tasks taskCreator
	teams teamMembershipReader
	cache taskCacheInvalidator
}

func New(tasks taskCreator, teams teamMembershipReader, cache taskCacheInvalidator) *UseCase {
	return &UseCase{tasks: tasks, teams: teams, cache: cache}
}

type Input struct {
	Title       string
	Description string
	Status      models.TaskStatus
	AssigneeID  *int64
	TeamID      int64
	DueDate     *time.Time
}

func (uc *UseCase) Create(ctx context.Context, actorID int64, input Input) (*models.Task, error) {
	if _, err := uc.teams.GetMemberRole(ctx, input.TeamID, actorID); err != nil {
		return nil, err
	}
	if err := uc.ensureAssigneeIsTeamMember(ctx, input.TeamID, input.AssigneeID); err != nil {
		return nil, err
	}
	task := &models.Task{
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		AssigneeID:  input.AssigneeID,
		TeamID:      input.TeamID,
		CreatedBy:   actorID,
		DueDate:     input.DueDate,
	}
	history := &models.TaskHistory{
		ChangedBy: actorID,
		Field:     "created",
		OldValue:  "",
		NewValue:  string(task.Status),
	}
	if err := uc.tasks.CreateWithHistory(ctx, task, history); err != nil {
		return nil, err
	}
	if err := uc.cache.DeleteTeamTasks(ctx, task.TeamID); err != nil {
		return nil, fmt.Errorf("%w: invalidate task cache: %v", domain.ErrExternal, err)
	}
	return task, nil
}

func (uc *UseCase) ensureAssigneeIsTeamMember(ctx context.Context, teamID int64, assigneeID *int64) error {
	if assigneeID == nil {
		return nil
	}
	if _, err := uc.teams.GetMemberRole(ctx, teamID, *assigneeID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("%w: assignee must be a team member", domain.ErrInvalidInput)
		}
		return err
	}
	return nil
}
