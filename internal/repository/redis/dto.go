package redis

import (
	"time"

	"task-service/internal/domain/models"
)

type taskListDTO struct {
	Items      []taskDTO      `json:"items"`
	NextCursor *taskCursorDTO `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

func newTaskListDTO(list models.TaskList) taskListDTO {
	items := make([]taskDTO, 0, len(list.Items))
	for _, task := range list.Items {
		items = append(items, newTaskDTO(task))
	}
	return taskListDTO{
		Items:      items,
		NextCursor: newTaskCursorDTO(list.NextCursor),
		HasMore:    list.HasMore,
	}
}

func (d taskListDTO) toDomain() models.TaskList {
	items := make([]models.Task, 0, len(d.Items))
	for _, task := range d.Items {
		items = append(items, task.toDomain())
	}
	return models.TaskList{
		Items:      items,
		NextCursor: d.NextCursor.toDomain(),
		HasMore:    d.HasMore,
	}
}

type taskCursorDTO struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int64     `json:"id"`
}

func newTaskCursorDTO(cursor *models.TaskCursor) *taskCursorDTO {
	if cursor == nil {
		return nil
	}
	return &taskCursorDTO{
		CreatedAt: cursor.CreatedAt,
		ID:        cursor.ID,
	}
}

func (d *taskCursorDTO) toDomain() *models.TaskCursor {
	if d == nil {
		return nil
	}
	return &models.TaskCursor{
		CreatedAt: d.CreatedAt,
		ID:        d.ID,
	}
}

type taskDTO struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	AssigneeID  *int64     `json:"assignee_id,omitempty"`
	TeamID      int64      `json:"team_id"`
	CreatedBy   int64      `json:"created_by"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func newTaskDTO(task models.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		AssigneeID:  task.AssigneeID,
		TeamID:      task.TeamID,
		CreatedBy:   task.CreatedBy,
		DueDate:     task.DueDate,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func (d taskDTO) toDomain() models.Task {
	return models.Task{
		ID:          d.ID,
		Title:       d.Title,
		Description: d.Description,
		Status:      models.TaskStatus(d.Status),
		AssigneeID:  d.AssigneeID,
		TeamID:      d.TeamID,
		CreatedBy:   d.CreatedBy,
		DueDate:     d.DueDate,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
