//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package report_invalid_assignees

import (
	"context"

	"task-service/internal/domain/models"
)

type invalidAssigneesProvider interface {
	GetInvalidAssignees(ctx context.Context, managerID int64) ([]models.Task, error)
}
