package models

import "time"

type Task struct {
	ID          int64
	Title       string
	Description string
	Status      TaskStatus
	AssigneeID  *int64
	TeamID      int64
	CreatedBy   int64
	DueDate     *time.Time
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
