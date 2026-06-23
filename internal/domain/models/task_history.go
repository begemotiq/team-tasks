package models

import "time"

type TaskHistory struct {
	ID        int64
	TaskID    int64
	ChangedBy int64
	Field     string
	OldValue  string
	NewValue  string
	CreatedAt time.Time
}
