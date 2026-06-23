//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package team_create

import (
	"context"

	"task-service/internal/domain/models"
)

type teamCreator interface {
	CreateWithOwner(ctx context.Context, team *models.Team, ownerID int64) error
}
