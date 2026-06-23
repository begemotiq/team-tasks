package mysql

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"strings"
	"time"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

const maxOutboxErrorMessageLength = 2048

type OutboxRepository struct {
	db *sql.DB
}

type outboxSQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) ClaimPending(ctx context.Context, limit int) (events []models.OutboxEvent, err error) {
	defer func() { recordDBError("outbox", "claim_pending", err) }()

	if limit <= 0 {
		limit = 10
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.QueryContext(ctx, `
		SELECT id, event_type, payload, status, attempts, claim_token, available_at, processed_at, last_error, created_at, updated_at
		FROM outbox_events
		WHERE (status = ? AND available_at <= CURRENT_TIMESTAMP)
			OR (status = ? AND updated_at < DATE_SUB(CURRENT_TIMESTAMP, INTERVAL 5 MINUTE))
		ORDER BY id
		LIMIT ?
		FOR UPDATE SKIP LOCKED`,
		models.OutboxEventStatusPending,
		models.OutboxEventStatusProcessing,
		limit,
	)
	if err != nil {
		return nil, err
	}
	rowsClosed := false
	defer func() {
		if !rowsClosed {
			closeRows(rows, &err)
		}
	}()

	events = make([]models.OutboxEvent, 0, limit)
	ids := make([]any, 0, limit)
	for rows.Next() {
		event, err := scanOutboxEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
		ids = append(ids, event.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if closeErr := rows.Close(); closeErr != nil {
		rowsClosed = true
		return nil, closeErr
	}
	rowsClosed = true
	if len(events) == 0 {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return events, nil
	}

	claimToken, err := newOutboxClaimToken()
	if err != nil {
		return nil, err
	}
	args := append([]any{models.OutboxEventStatusProcessing, claimToken}, ids...)
	if _, err := tx.ExecContext(ctx, `
		UPDATE outbox_events
		SET status = ?, claim_token = ?, attempts = attempts + 1
		WHERE id IN (`+placeholders(len(ids))+`)`,
		args...,
	); err != nil {
		return nil, err
	}
	for i := range events {
		events[i].Status = models.OutboxEventStatusProcessing
		events[i].ClaimToken = claimToken
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *OutboxRepository) MarkProcessed(ctx context.Context, id int64, claimToken string) (err error) {
	defer func() { recordDBError("outbox", "mark_processed", err) }()

	result, err := r.db.ExecContext(ctx, `
		UPDATE outbox_events
		SET status = ?, claim_token = NULL, processed_at = CURRENT_TIMESTAMP, last_error = NULL
		WHERE id = ? AND status = ? AND claim_token = ?`,
		models.OutboxEventStatusProcessed, id, models.OutboxEventStatusProcessing, claimToken,
	)
	if err != nil {
		return err
	}
	return ensureAffected(result)
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id int64, claimToken string, retryAt time.Time, message string) (err error) {
	defer func() { recordDBError("outbox", "mark_failed", err) }()

	message = truncateString(message, maxOutboxErrorMessageLength)
	result, err := r.db.ExecContext(ctx, `
		UPDATE outbox_events
		SET status = ?, claim_token = NULL, available_at = ?, last_error = ?
		WHERE id = ? AND status = ? AND claim_token = ?`,
		models.OutboxEventStatusPending, retryAt, message, id, models.OutboxEventStatusProcessing, claimToken,
	)
	if err != nil {
		return err
	}
	return ensureAffected(result)
}

func (r *OutboxRepository) MarkDeadLetter(ctx context.Context, id int64, claimToken string, message string) (err error) {
	defer func() { recordDBError("outbox", "mark_dead_letter", err) }()

	message = truncateString(message, maxOutboxErrorMessageLength)
	result, err := r.db.ExecContext(ctx, `
		UPDATE outbox_events
		SET status = ?, claim_token = NULL, last_error = ?
		WHERE id = ? AND status = ? AND claim_token = ?`,
		models.OutboxEventStatusDeadLetter, message, id, models.OutboxEventStatusProcessing, claimToken,
	)
	if err != nil {
		return err
	}
	return ensureAffected(result)
}

func (r *OutboxRepository) DeleteProcessedBefore(ctx context.Context, before time.Time) (deleted int64, err error) {
	defer func() { recordDBError("outbox", "delete_processed_before", err) }()

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM outbox_events
		WHERE status = ?
			AND processed_at IS NOT NULL
			AND processed_at < ?`,
		models.OutboxEventStatusProcessed, before,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func addOutboxEvent(ctx context.Context, exec outboxSQLExecutor, event *models.OutboxEvent) error {
	result, err := exec.ExecContext(ctx, `
		INSERT INTO outbox_events (event_type, payload, status)
		VALUES (?, ?, ?)`,
		event.Type, event.Payload, models.OutboxEventStatusPending,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	event.ID = id
	return nil
}

func scanOutboxEvent(row scanner) (models.OutboxEvent, error) {
	var event outboxEventRow
	err := row.Scan(
		&event.ID,
		&event.Type,
		&event.Payload,
		&event.Status,
		&event.Attempts,
		&event.ClaimToken,
		&event.AvailableAt,
		&event.ProcessedAt,
		&event.LastError,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return models.OutboxEvent{}, err
	}
	return event.toDomain(), nil
}

func newOutboxClaimToken() (string, error) {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(data[:]), nil
}

func ensureAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}

func truncateString(value string, maxLength int) string {
	if maxLength <= 0 || len(value) <= maxLength {
		return value
	}
	lastBoundary := 0
	for boundary := range value {
		if boundary > maxLength {
			return value[:lastBoundary]
		}
		lastBoundary = boundary
	}
	if lastBoundary < maxLength {
		return value[:lastBoundary]
	}
	return value[:maxLength]
}
