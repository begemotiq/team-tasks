package request

import (
	"strings"
	"time"

	"task-service/internal/domain/models"
	taskcreateusecase "task-service/internal/usecase/task_create"
	taskupdateusecase "task-service/internal/usecase/task_update"
)

type CreateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	AssigneeID  *int64     `json:"assignee_id"`
	TeamID      int64      `json:"team_id"`
	DueDate     *time.Time `json:"due_date"`
}

func (r *CreateTaskRequest) Validate() error {
	title, err := requiredString("title", r.Title)
	if err != nil {
		return err
	}
	r.Status = strings.TrimSpace(r.Status)
	if r.Status == "" {
		r.Status = string(models.TaskStatusTodo)
	}
	if !models.TaskStatus(r.Status).Valid() {
		return invalidInput("invalid task status")
	}
	if r.TeamID <= 0 {
		return invalidInput("team_id must be a positive integer")
	}
	if r.AssigneeID != nil && *r.AssigneeID <= 0 {
		return invalidInput("assignee_id must be a positive integer")
	}
	r.Title = title
	r.Description = strings.TrimSpace(r.Description)
	return nil
}

func (r CreateTaskRequest) ToInput() taskcreateusecase.Input {
	return taskcreateusecase.Input{
		Title:       r.Title,
		Description: r.Description,
		Status:      models.TaskStatus(r.Status),
		AssigneeID:  r.AssigneeID,
		TeamID:      r.TeamID,
		DueDate:     r.DueDate,
	}
}

type UpdateTaskRequest struct {
	Title       *string                 `json:"title"`
	Description *string                 `json:"description"`
	Status      *string                 `json:"status"`
	AssigneeID  OptionalJSON[int64]     `json:"assignee_id"`
	DueDate     OptionalJSON[time.Time] `json:"due_date"`
}

func (r *UpdateTaskRequest) Validate() error {
	if r.Title != nil {
		title, err := requiredString("title", *r.Title)
		if err != nil {
			return err
		}
		r.Title = &title
	}
	if r.Description != nil {
		description := strings.TrimSpace(*r.Description)
		r.Description = &description
	}
	if r.Status != nil {
		status := strings.TrimSpace(*r.Status)
		if !models.TaskStatus(status).Valid() {
			return invalidInput("invalid task status")
		}
		r.Status = &status
	}
	if r.AssigneeID.Set && r.AssigneeID.Valid && r.AssigneeID.Value <= 0 {
		return invalidInput("assignee_id must be a positive integer")
	}
	return nil
}

func (r UpdateTaskRequest) ToInput() taskupdateusecase.Input {
	var status *models.TaskStatus
	if r.Status != nil {
		value := models.TaskStatus(*r.Status)
		status = &value
	}
	return taskupdateusecase.Input{
		Title:       r.Title,
		Description: r.Description,
		Status:      status,
		AssigneeID:  toUseCaseOptional(r.AssigneeID),
		DueDate:     toUseCaseOptional(r.DueDate),
	}
}

func toUseCaseOptional[T any](value OptionalJSON[T]) taskupdateusecase.Optional[T] {
	return taskupdateusecase.Optional[T]{
		Set:   value.Set,
		Valid: value.Valid,
		Value: value.Value,
	}
}
