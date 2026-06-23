package response

import (
	"time"

	"task-service/internal/adapter/http/pagination"
	"task-service/internal/domain/models"
)

type TaskResponse struct {
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

type TaskListResponse struct {
	Items      []TaskResponse `json:"items"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

type TaskItemsResponse struct {
	Items []TaskResponse `json:"items"`
}

func NewTask(task models.Task) TaskResponse {
	return TaskResponse{
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

func NewTaskList(list models.TaskList) TaskListResponse {
	return TaskListResponse{
		Items:      NewTasks(list.Items),
		NextCursor: pagination.EncodeTaskCursor(list.NextCursor),
		HasMore:    list.HasMore,
	}
}

func NewTaskItems(tasks []models.Task) TaskItemsResponse {
	return TaskItemsResponse{Items: NewTasks(tasks)}
}

func NewTasks(tasks []models.Task) []TaskResponse {
	items := make([]TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, NewTask(task))
	}
	return items
}
