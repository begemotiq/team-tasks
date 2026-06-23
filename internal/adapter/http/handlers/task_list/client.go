//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package task_list

import (
	"context"

	"task-service/internal/domain/models"
)

type taskLister interface {
	List(ctx context.Context, actorID int64, filter models.TaskFilter) (models.TaskList, error)
}
