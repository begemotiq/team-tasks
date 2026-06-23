package mysql

import (
	"database/sql"
	"time"

	"task-service/internal/domain/models"
)

type userRow struct {
	ID           int64
	Email        string
	PasswordHash string
	Name         string
	CreatedAt    time.Time
}

func (r userRow) toDomain() models.User {
	return models.User{
		ID:           r.ID,
		Email:        r.Email,
		PasswordHash: r.PasswordHash,
		Name:         r.Name,
		CreatedAt:    r.CreatedAt,
	}
}

type teamRow struct {
	ID        int64
	Name      string
	CreatedBy int64
	CreatedAt time.Time
}

func (r teamRow) toDomain() models.Team {
	return models.Team{
		ID:        r.ID,
		Name:      r.Name,
		CreatedBy: r.CreatedBy,
		CreatedAt: r.CreatedAt,
	}
}

type taskRow struct {
	ID          int64
	Title       string
	Description string
	Status      string
	AssigneeID  sql.NullInt64
	TeamID      int64
	CreatedBy   int64
	DueDate     sql.NullTime
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r taskRow) toDomain() models.Task {
	task := models.Task{
		ID:          r.ID,
		Title:       r.Title,
		Description: r.Description,
		Status:      models.TaskStatus(r.Status),
		TeamID:      r.TeamID,
		CreatedBy:   r.CreatedBy,
		Version:     r.Version,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	if r.AssigneeID.Valid {
		assigneeID := r.AssigneeID.Int64
		task.AssigneeID = &assigneeID
	}
	if r.DueDate.Valid {
		dueDate := r.DueDate.Time
		task.DueDate = &dueDate
	}
	return task
}

type taskHistoryRow struct {
	ID        int64
	TaskID    int64
	ChangedBy int64
	Field     string
	OldValue  string
	NewValue  string
	CreatedAt time.Time
}

type outboxEventRow struct {
	ID          int64
	Type        string
	Payload     []byte
	Status      string
	Attempts    int
	ClaimToken  sql.NullString
	AvailableAt time.Time
	ProcessedAt sql.NullTime
	LastError   sql.NullString
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r outboxEventRow) toDomain() models.OutboxEvent {
	event := models.OutboxEvent{
		ID:          r.ID,
		Type:        models.OutboxEventType(r.Type),
		Payload:     r.Payload,
		Status:      models.OutboxEventStatus(r.Status),
		Attempts:    r.Attempts,
		AvailableAt: r.AvailableAt,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	if r.ProcessedAt.Valid {
		processedAt := r.ProcessedAt.Time
		event.ProcessedAt = &processedAt
	}
	if r.ClaimToken.Valid {
		event.ClaimToken = r.ClaimToken.String
	}
	if r.LastError.Valid {
		event.LastError = r.LastError.String
	}
	return event
}

func (r taskHistoryRow) toDomain() models.TaskHistory {
	return models.TaskHistory{
		ID:        r.ID,
		TaskID:    r.TaskID,
		ChangedBy: r.ChangedBy,
		Field:     r.Field,
		OldValue:  r.OldValue,
		NewValue:  r.NewValue,
		CreatedAt: r.CreatedAt,
	}
}

type teamSummaryRow struct {
	TeamID             int64
	TeamName           string
	MembersCount       int64
	DoneTasksLast7Days int64
}

func (r teamSummaryRow) toDomain() models.TeamSummary {
	return models.TeamSummary{
		TeamID:             r.TeamID,
		TeamName:           r.TeamName,
		MembersCount:       r.MembersCount,
		DoneTasksLast7Days: r.DoneTasksLast7Days,
	}
}

type topCreatorRow struct {
	TeamID       int64
	TeamName     string
	UserID       int64
	UserName     string
	TasksCreated int64
	RankPosition int64
}

func (r topCreatorRow) toDomain() models.TopCreator {
	return models.TopCreator{
		TeamID:       r.TeamID,
		TeamName:     r.TeamName,
		UserID:       r.UserID,
		UserName:     r.UserName,
		TasksCreated: r.TasksCreated,
		RankPosition: r.RankPosition,
	}
}
