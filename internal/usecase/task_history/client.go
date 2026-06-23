//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package task_history

import (
	"context"

	"task-service/internal/domain/models"
)

type taskHistoryReader interface {
	GetByID(ctx context.Context, id int64) (*models.Task, error)
	History(ctx context.Context, taskID int64) ([]models.TaskHistory, error)
}

type teamMembershipReader interface {
	GetMemberRole(ctx context.Context, teamID, userID int64) (models.Role, error)
}
