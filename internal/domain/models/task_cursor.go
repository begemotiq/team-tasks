package models

import "time"

type TaskCursor struct {
	CreatedAt time.Time
	ID        int64
}

func NewTaskCursor(task Task) TaskCursor {
	return TaskCursor{
		CreatedAt: task.CreatedAt,
		ID:        task.ID,
	}
}

func (c TaskCursor) Valid() bool {
	return !c.CreatedAt.IsZero() && c.ID > 0
}
