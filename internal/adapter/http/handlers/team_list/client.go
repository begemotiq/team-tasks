//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package team_list

import (
	"context"

	"task-service/internal/domain/models"
)

type teamLister interface {
	List(ctx context.Context, userID int64) ([]models.Team, error)
}
