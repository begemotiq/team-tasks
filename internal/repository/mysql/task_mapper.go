package mysql

import (
	"time"

	"task-service/internal/domain/models"
)

type scanner interface {
	Scan(dest ...any) error
}

func taskSelectSQL() string {
	return `SELECT t.id, t.title, t.description, t.status, t.assignee_id, t.team_id, t.created_by, t.due_date, t.version, t.created_at, t.updated_at FROM tasks t`
}

func scanTask(row scanner) (*models.Task, error) {
	var task taskRow
	err := row.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.AssigneeID,
		&task.TeamID,
		&task.CreatedBy,
		&task.DueDate,
		&task.Version,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	domainTask := task.toDomain()
	return &domainTask, nil
}

func optionalInt(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func optionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}
