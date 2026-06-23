package response

import (
	"time"

	"task-service/internal/domain/models"
)

type TaskHistoryResponse struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	ChangedBy int64     `json:"changed_by"`
	Field     string    `json:"field"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskHistoryListResponse struct {
	Items []TaskHistoryResponse `json:"items"`
}

func NewTaskHistory(history models.TaskHistory) TaskHistoryResponse {
	return TaskHistoryResponse{
		ID:        history.ID,
		TaskID:    history.TaskID,
		ChangedBy: history.ChangedBy,
		Field:     history.Field,
		OldValue:  history.OldValue,
		NewValue:  history.NewValue,
		CreatedAt: history.CreatedAt,
	}
}

func NewTaskHistoryList(history []models.TaskHistory) TaskHistoryListResponse {
	items := make([]TaskHistoryResponse, 0, len(history))
	for _, item := range history {
		items = append(items, NewTaskHistory(item))
	}
	return TaskHistoryListResponse{Items: items}
}
