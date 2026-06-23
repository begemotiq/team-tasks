//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package report_team_summary

import (
	"context"

	"task-service/internal/domain/models"
)

type teamSummaryReader interface {
	TeamSummary(ctx context.Context, managerID int64) ([]models.TeamSummary, error)
}

type reportAccessChecker interface {
	HasManagementRole(ctx context.Context, userID int64) (bool, error)
}
