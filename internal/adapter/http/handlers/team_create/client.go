//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package team_create

import (
	"context"

	"task-service/internal/domain/models"
	teamcreateusecase "task-service/internal/usecase/team_create"
)

type teamCreator interface {
	Create(ctx context.Context, ownerID int64, input teamcreateusecase.Input) (*models.Team, error)
}
