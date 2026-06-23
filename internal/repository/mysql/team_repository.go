package mysql

import (
	"context"
	"database/sql"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

type TeamRepository struct {
	db *sql.DB
}

type teamSQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateWithOwner(ctx context.Context, team *models.Team, ownerID int64) (err error) {
	defer func() { recordDBError("teams", "create_with_owner", err) }()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, "INSERT INTO teams (name, created_by) VALUES (?, ?)", team.Name, ownerID)
	if err != nil {
		return mapMySQLError(err)
	}
	teamID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)",
		teamID, ownerID, models.RoleOwner,
	); err != nil {
		return mapMySQLError(err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	created, err := r.FindByID(ctx, teamID)
	if err != nil {
		return err
	}
	*team = *created
	return nil
}

func (r *TeamRepository) FindByID(ctx context.Context, id int64) (team *models.Team, err error) {
	defer func() { recordDBError("teams", "find_by_id", err) }()

	var row teamRow
	err = r.db.QueryRowContext(ctx,
		"SELECT id, name, created_by, created_at FROM teams WHERE id = ?",
		id,
	).Scan(&row.ID, &row.Name, &row.CreatedBy, &row.CreatedAt)
	if err != nil {
		return nil, mapSQLError(err)
	}
	result := row.toDomain()
	return &result, nil
}

func (r *TeamRepository) ListByUser(ctx context.Context, userID int64) (teams []models.Team, err error) {
	defer func() { recordDBError("teams", "list_by_user", err) }()

	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.name, t.created_by, t.created_at
		FROM teams t
		JOIN team_members tm ON tm.team_id = t.id
		WHERE tm.user_id = ?
		ORDER BY t.created_at DESC, t.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows, &err)

	teams = make([]models.Team, 0)
	for rows.Next() {
		var row teamRow
		if err := rows.Scan(&row.ID, &row.Name, &row.CreatedBy, &row.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, row.toDomain())
	}
	return teams, rows.Err()
}

func (r *TeamRepository) AddMember(ctx context.Context, teamID, userID int64, role models.Role) (err error) {
	defer func() { recordDBError("teams", "add_member", err) }()

	return addTeamMember(ctx, r.db, teamID, userID, role)
}

func (r *TeamRepository) AddMemberWithOutboxEvent(ctx context.Context, teamID, userID int64, role models.Role, event *models.OutboxEvent) (err error) {
	defer func() { recordDBError("teams", "add_member_with_outbox_event", err) }()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := addTeamMember(ctx, tx, teamID, userID, role); err != nil {
		return err
	}
	if err := addOutboxEvent(ctx, tx, event); err != nil {
		return err
	}
	return tx.Commit()
}

func addTeamMember(ctx context.Context, exec teamSQLExecutor, teamID, userID int64, role models.Role) error {
	_, err := exec.ExecContext(ctx,
		"INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)",
		teamID, userID, role,
	)
	if err != nil {
		return mapMySQLError(err)
	}
	return nil
}

func (r *TeamRepository) Delete(ctx context.Context, teamID int64) (err error) {
	defer func() { recordDBError("teams", "delete", err) }()

	result, err := r.db.ExecContext(ctx, "DELETE FROM teams WHERE id = ?", teamID)
	if err != nil {
		return mapMySQLError(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TeamRepository) GetMemberRole(ctx context.Context, teamID, userID int64) (role models.Role, err error) {
	defer func() { recordDBError("teams", "get_member_role", err) }()

	var rawRole string
	err = r.db.QueryRowContext(ctx,
		"SELECT role FROM team_members WHERE team_id = ? AND user_id = ?",
		teamID, userID,
	).Scan(&rawRole)
	if err != nil {
		return "", mapSQLError(err)
	}
	return models.Role(rawRole), nil
}

func (r *TeamRepository) HasManagementRole(ctx context.Context, userID int64) (exists bool, err error) {
	defer func() { recordDBError("teams", "has_management_role", err) }()

	err = r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM team_members
			WHERE user_id = ?
				AND role IN (?, ?)
		)`,
		userID, models.RoleOwner, models.RoleAdmin,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
