//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package task_update

import (
	"context"

	"task-service/internal/domain/models"
)

type taskUpdater interface {
	GetByID(ctx context.Context, id int64) (*models.Task, error)
	UpdateWithHistory(ctx context.Context, task *models.Task, history []models.TaskHistory) error
}

type teamMembershipReader interface {
	GetMemberRole(ctx context.Context, teamID, userID int64) (models.Role, error)
}

type taskCacheInvalidator interface {
	DeleteTeamTasks(ctx context.Context, teamID int64) error
}
