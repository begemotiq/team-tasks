package mysql

import (
	"context"
	"database/sql"
	"strings"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

type TaskRepository struct {
	db *sql.DB
}

type taskSQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *models.Task) (err error) {
	defer func() { recordDBError("tasks", "create", err) }()

	return createTask(ctx, r.db, task)
}

func (r *TaskRepository) CreateWithHistory(ctx context.Context, task *models.Task, history *models.TaskHistory) (err error) {
	defer func() { recordDBError("tasks", "create_with_history", err) }()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := createTask(ctx, tx, task); err != nil {
		return err
	}
	history.TaskID = task.ID
	if err := addTaskHistory(ctx, tx, history); err != nil {
		return err
	}
	return tx.Commit()
}

func createTask(ctx context.Context, exec taskSQLExecutor, task *models.Task) error {
	result, err := exec.ExecContext(ctx, `
		INSERT INTO tasks (title, description, status, assignee_id, team_id, created_by, due_date)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		task.Title, task.Description, task.Status, optionalInt(task.AssigneeID), task.TeamID, task.CreatedBy, optionalTime(task.DueDate),
	)
	if err != nil {
		return mapMySQLError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	created, err := scanTask(exec.QueryRowContext(ctx, taskSelectSQL()+" WHERE t.id = ?", id))
	if err != nil {
		return mapSQLError(err)
	}
	*task = *created
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (task *models.Task, err error) {
	defer func() { recordDBError("tasks", "get_by_id", err) }()

	row := r.db.QueryRowContext(ctx, taskSelectSQL()+" WHERE t.id = ?", id)
	task, err = scanTask(row)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return task, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *models.Task) (err error) {
	defer func() { recordDBError("tasks", "update", err) }()

	return updateTask(ctx, r.db, task)
}

func (r *TaskRepository) UpdateWithHistory(ctx context.Context, task *models.Task, history []models.TaskHistory) (err error) {
	defer func() { recordDBError("tasks", "update_with_history", err) }()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := updateTask(ctx, tx, task); err != nil {
		return err
	}
	for i := range history {
		history[i].TaskID = task.ID
		if err := addTaskHistory(ctx, tx, &history[i]); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func updateTask(ctx context.Context, exec taskSQLExecutor, task *models.Task) error {
	result, err := exec.ExecContext(ctx, `
		UPDATE tasks
		SET title = ?, description = ?, status = ?, assignee_id = ?, due_date = ?, version = version + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND version = ?`,
		task.Title, task.Description, task.Status, optionalInt(task.AssigneeID), optionalTime(task.DueDate), task.ID, task.Version,
	)
	if err != nil {
		return mapMySQLError(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrConflict
	}
	updated, err := scanTask(exec.QueryRowContext(ctx, taskSelectSQL()+" WHERE t.id = ?", task.ID))
	if err != nil {
		return mapSQLError(err)
	}
	*task = *updated
	return nil
}

func (r *TaskRepository) List(ctx context.Context, filter models.TaskFilter, requesterID int64) (list models.TaskList, err error) {
	defer func() { recordDBError("tasks", "list", err) }()

	filter = filter.Normalize()
	where := []string{"tm.user_id = ?"}
	args := []any{requesterID}
	if filter.TeamID != nil {
		where = append(where, "t.team_id = ?")
		args = append(args, *filter.TeamID)
	}
	if filter.Status != nil {
		where = append(where, "t.status = ?")
		args = append(args, *filter.Status)
	}
	if filter.AssigneeID != nil {
		where = append(where, "t.assignee_id = ?")
		args = append(args, *filter.AssigneeID)
	}
	if filter.Cursor != nil {
		where = append(where, "(t.created_at < ? OR (t.created_at = ? AND t.id < ?))")
		args = append(args, filter.Cursor.CreatedAt, filter.Cursor.CreatedAt, filter.Cursor.ID)
	}

	join := " JOIN team_members tm ON tm.team_id = t.team_id "
	whereSQL := " WHERE " + strings.Join(where, " AND ")
	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, filter.PageSize+1)
	rows, err := r.db.QueryContext(ctx,
		taskSelectSQL()+join+whereSQL+" ORDER BY t.created_at DESC, t.id DESC LIMIT ?",
		queryArgs...,
	)
	if err != nil {
		return models.TaskList{}, err
	}
	defer rows.Close()

	items := make([]models.Task, 0, filter.PageSize+1)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return models.TaskList{}, err
		}
		items = append(items, *task)
	}
	if err := rows.Err(); err != nil {
		return models.TaskList{}, err
	}
	hasMore := len(items) > filter.PageSize
	if hasMore {
		items = items[:filter.PageSize]
	}
	var nextCursor *models.TaskCursor
	if hasMore && len(items) > 0 {
		cursor := models.NewTaskCursor(items[len(items)-1])
		nextCursor = &cursor
	}
	return models.TaskList{Items: items, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (r *TaskRepository) AddHistory(ctx context.Context, history *models.TaskHistory) (err error) {
	defer func() { recordDBError("tasks", "add_history", err) }()

	return addTaskHistory(ctx, r.db, history)
}

func addTaskHistory(ctx context.Context, exec taskSQLExecutor, history *models.TaskHistory) error {
	result, err := exec.ExecContext(ctx, `
		INSERT INTO task_history (task_id, changed_by, field, old_value, new_value)
		VALUES (?, ?, ?, ?, ?)`,
		history.TaskID, history.ChangedBy, history.Field, history.OldValue, history.NewValue,
	)
	if err != nil {
		return mapMySQLError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	history.ID = id
	return nil
}

func (r *TaskRepository) History(ctx context.Context, taskID int64) (history []models.TaskHistory, err error) {
	defer func() { recordDBError("tasks", "history", err) }()

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, task_id, changed_by, field, old_value, new_value, created_at
		FROM task_history
		WHERE task_id = ?
		ORDER BY created_at DESC, id DESC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history = make([]models.TaskHistory, 0)
	for rows.Next() {
		var row taskHistoryRow
		if err := rows.Scan(&row.ID, &row.TaskID, &row.ChangedBy, &row.Field, &row.OldValue, &row.NewValue, &row.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, row.toDomain())
	}
	return history, rows.Err()
}

func (r *TaskRepository) TeamSummary(ctx context.Context, managerID int64) (summaries []models.TeamSummary, err error) {
	defer func() { recordDBError("tasks", "team_summary", err) }()

	rows, err := r.db.QueryContext(ctx, `
			SELECT
				t.id,
			t.name,
			COUNT(DISTINCT tm.user_id) AS members_count,
			COUNT(DISTINCT CASE
				WHEN ta.status = 'done' AND ta.updated_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
				THEN ta.id
				END) AS done_tasks_last_7_days
			FROM teams t
			JOIN team_members manager_tm ON manager_tm.team_id = t.id
			LEFT JOIN team_members tm ON tm.team_id = t.id
			LEFT JOIN tasks ta ON ta.team_id = t.id
			WHERE manager_tm.user_id = ?
				AND manager_tm.role IN (?, ?)
			GROUP BY t.id, t.name
			ORDER BY t.name`,
		managerID, models.RoleOwner, models.RoleAdmin,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.TeamSummary, 0)
	for rows.Next() {
		var row teamSummaryRow
		if err := rows.Scan(&row.TeamID, &row.TeamName, &row.MembersCount, &row.DoneTasksLast7Days); err != nil {
			return nil, err
		}
		result = append(result, row.toDomain())
	}
	return result, rows.Err()
}

func (r *TaskRepository) TopCreatorsByTeam(ctx context.Context, managerID int64) (creators []models.TopCreator, err error) {
	defer func() { recordDBError("tasks", "top_creators_by_team", err) }()

	rows, err := r.db.QueryContext(ctx, `
			SELECT team_id, team_name, user_id, user_name, tasks_created, rank_position
			FROM (
			SELECT
				t.id AS team_id,
				t.name AS team_name,
				u.id AS user_id,
				u.name AS user_name,
					COUNT(ts.id) AS tasks_created,
					DENSE_RANK() OVER (PARTITION BY t.id ORDER BY COUNT(ts.id) DESC) AS rank_position
				FROM teams t
				JOIN team_members manager_tm ON manager_tm.team_id = t.id
				JOIN tasks ts ON ts.team_id = t.id
				JOIN users u ON u.id = ts.created_by
				WHERE manager_tm.user_id = ?
					AND manager_tm.role IN (?, ?)
					AND ts.created_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH)
				GROUP BY t.id, t.name, u.id, u.name
			) ranked
			WHERE rank_position <= 3
			ORDER BY team_name, rank_position, tasks_created DESC`,
		managerID, models.RoleOwner, models.RoleAdmin,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.TopCreator, 0)
	for rows.Next() {
		var row topCreatorRow
		if err := rows.Scan(&row.TeamID, &row.TeamName, &row.UserID, &row.UserName, &row.TasksCreated, &row.RankPosition); err != nil {
			return nil, err
		}
		result = append(result, row.toDomain())
	}
	return result, rows.Err()
}

func (r *TaskRepository) InvalidAssignees(ctx context.Context, managerID int64) (tasks []models.Task, err error) {
	defer func() { recordDBError("tasks", "invalid_assignees", err) }()

	rows, err := r.db.QueryContext(ctx, taskSelectSQL()+`
			JOIN team_members manager_tm ON manager_tm.team_id = t.team_id
			WHERE t.assignee_id IS NOT NULL
				AND manager_tm.user_id = ?
				AND manager_tm.role IN (?, ?)
				AND NOT EXISTS (
					SELECT 1
					FROM team_members tm
					WHERE tm.team_id = t.team_id
						AND tm.user_id = t.assignee_id
				)
			ORDER BY t.id`,
		managerID, models.RoleOwner, models.RoleAdmin,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks = make([]models.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}
	return tasks, rows.Err()
}
