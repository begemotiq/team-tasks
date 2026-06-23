//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package task_create

import (
	"context"

	"task-service/internal/domain/models"
	taskcreateusecase "task-service/internal/usecase/task_create"
)

type taskCreator interface {
	Create(ctx context.Context, actorID int64, input taskcreateusecase.Input) (*models.Task, error)
}
