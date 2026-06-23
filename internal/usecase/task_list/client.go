//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package task_list

import (
	"context"
	"time"

	"task-service/internal/domain/models"
)

type taskLister interface {
	List(ctx context.Context, filter models.TaskFilter, requesterID int64) (models.TaskList, error)
}

type teamMembershipReader interface {
	GetMemberRole(ctx context.Context, teamID, userID int64) (models.Role, error)
}

type taskListCache interface {
	GetTaskList(ctx context.Context, key string) (models.TaskList, bool, error)
	SetTaskList(ctx context.Context, key string, value models.TaskList, ttl time.Duration) error
}
