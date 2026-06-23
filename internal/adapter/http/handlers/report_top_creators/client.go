//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package report_top_creators

import (
	"context"

	"task-service/internal/domain/models"
)

type topCreatorsProvider interface {
	GetTopCreators(ctx context.Context, managerID int64) ([]models.TopCreator, error)
}
