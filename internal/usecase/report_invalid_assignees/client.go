//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package report_invalid_assignees

import (
	"context"

	"task-service/internal/domain/models"
)

type invalidAssigneesReader interface {
	InvalidAssignees(ctx context.Context, managerID int64) ([]models.Task, error)
}

type reportAccessChecker interface {
	HasManagementRole(ctx context.Context, userID int64) (bool, error)
}
