//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package task_update

import (
	"context"

	"task-service/internal/domain/models"
	taskupdateusecase "task-service/internal/usecase/task_update"
)

type taskUpdater interface {
	Update(ctx context.Context, actorID int64, taskID int64, input taskupdateusecase.Input) (*models.Task, error)
}
