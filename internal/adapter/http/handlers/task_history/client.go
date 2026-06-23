//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package task_history

import (
	"context"

	"task-service/internal/domain/models"
)

type taskHistoryReader interface {
	GetHistory(ctx context.Context, actorID int64, taskID int64) ([]models.TaskHistory, error)
}
