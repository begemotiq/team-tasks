package models

type TaskList struct {
	Items      []Task
	NextCursor *TaskCursor
	HasMore    bool
}
