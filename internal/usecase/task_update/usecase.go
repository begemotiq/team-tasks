package task_update

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

type UseCase struct {
	tasks taskUpdater
	teams teamMembershipReader
	cache taskCacheInvalidator
}

func New(tasks taskUpdater, teams teamMembershipReader, cache taskCacheInvalidator) *UseCase {
	return &UseCase{tasks: tasks, teams: teams, cache: cache}
}

type Input struct {
	Title       *string
	Description *string
	Status      *models.TaskStatus
	AssigneeID  Optional[int64]
	DueDate     Optional[time.Time]
}

func (uc *UseCase) Update(ctx context.Context, actorID, taskID int64, input Input) (*models.Task, error) {
	current, err := uc.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	role, err := uc.teams.GetMemberRole(ctx, current.TeamID, actorID)
	if err != nil {
		return nil, err
	}
	if !role.CanManageTask() && current.CreatedBy != actorID && !sameOptionalID(current.AssigneeID, actorID) {
		return nil, domain.ErrForbidden
	}
	updated := *current
	changes := make([]models.TaskHistory, 0, 4)
	if input.Title != nil {
		if *input.Title != current.Title {
			changes = append(changes, history(taskID, actorID, "title", current.Title, *input.Title))
			updated.Title = *input.Title
		}
	}
	if input.Description != nil {
		if *input.Description != current.Description {
			changes = append(changes, history(taskID, actorID, "description", current.Description, *input.Description))
			updated.Description = *input.Description
		}
	}
	if input.Status != nil {
		if *input.Status != current.Status {
			changes = append(changes, history(taskID, actorID, "status", string(current.Status), string(*input.Status)))
			updated.Status = *input.Status
		}
	}
	if input.AssigneeID.Set {
		assigneeID := optionalInt64Ptr(input.AssigneeID)
		if assigneeID != nil {
			if err := uc.ensureAssigneeIsTeamMember(ctx, current.TeamID, *assigneeID); err != nil {
				return nil, err
			}
		}
		if !sameOptionalIDs(current.AssigneeID, assigneeID) {
			changes = append(changes, history(taskID, actorID, "assignee_id", formatOptionalID(current.AssigneeID), formatOptionalID(assigneeID)))
			updated.AssigneeID = assigneeID
		}
	}
	if input.DueDate.Set {
		dueDate := optionalTimePtr(input.DueDate)
		if !sameOptionalTimes(current.DueDate, dueDate) {
			changes = append(changes, history(taskID, actorID, "due_date", formatOptionalTime(current.DueDate), formatOptionalTime(dueDate)))
			updated.DueDate = dueDate
		}
	}
	if len(changes) == 0 {
		return current, nil
	}
	if err := uc.tasks.UpdateWithHistory(ctx, &updated, changes); err != nil {
		return nil, err
	}
	if err := uc.cache.DeleteTeamTasks(ctx, current.TeamID); err != nil {
		return nil, fmt.Errorf("%w: invalidate task cache: %v", domain.ErrExternal, err)
	}
	return &updated, nil
}

func (uc *UseCase) ensureAssigneeIsTeamMember(ctx context.Context, teamID, assigneeID int64) error {
	if _, err := uc.teams.GetMemberRole(ctx, teamID, assigneeID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("%w: assignee must be a team member", domain.ErrInvalidInput)
		}
		return err
	}
	return nil
}

func history(taskID, actorID int64, field, oldValue, newValue string) models.TaskHistory {
	return models.TaskHistory{TaskID: taskID, ChangedBy: actorID, Field: field, OldValue: oldValue, NewValue: newValue}
}

func sameOptionalID(id *int64, value int64) bool {
	return id != nil && *id == value
}

func sameOptionalIDs(left, right *int64) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func sameOptionalTimes(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Equal(*right)
}

func optionalInt64Ptr(value Optional[int64]) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Value
	return &result
}

func optionalTimePtr(value Optional[time.Time]) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Value
	return &result
}

func formatOptionalID(id *int64) string {
	if id == nil {
		return ""
	}
	return strconv.FormatInt(*id, 10)
}

func formatOptionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
