package models

type TaskFilter struct {
	TeamID     *int64
	Status     *TaskStatus
	AssigneeID *int64
	Cursor     *TaskCursor
	PageSize   int
}

func (f TaskFilter) Normalize() TaskFilter {
	if f.PageSize < 1 {
		f.PageSize = 20
	}
	if f.PageSize > 100 {
		f.PageSize = 100
	}
	return f
}
