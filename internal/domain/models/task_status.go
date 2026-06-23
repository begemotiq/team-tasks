package models

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

func (s TaskStatus) Valid() bool {
	return s == TaskStatusTodo || s == TaskStatusInProgress || s == TaskStatusDone
}
