package mysql

import (
	"context"
	"database/sql"

	"task-service/internal/domain/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) (err error) {
	defer func() { recordDBError("users", "create", err) }()

	result, err := r.db.ExecContext(ctx,
		"INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)",
		user.Email, user.PasswordHash, user.Name,
	)
	if err != nil {
		return mapMySQLError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	created, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	*user = *created
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (user *models.User, err error) {
	defer func() { recordDBError("users", "find_by_email", err) }()

	return r.scanUser(ctx, "SELECT id, email, password_hash, name, created_at FROM users WHERE email = ?", email)
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (user *models.User, err error) {
	defer func() { recordDBError("users", "find_by_id", err) }()

	return r.scanUser(ctx, "SELECT id, email, password_hash, name, created_at FROM users WHERE id = ?", id)
}

func (r *UserRepository) scanUser(ctx context.Context, query string, args ...any) (*models.User, error) {
	var row userRow
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&row.ID, &row.Email, &row.PasswordHash, &row.Name, &row.CreatedAt)
	if err != nil {
		return nil, mapSQLError(err)
	}
	user := row.toDomain()
	return &user, nil
}
