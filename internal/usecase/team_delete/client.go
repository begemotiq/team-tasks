//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package team_delete

import (
	"context"

	"task-service/internal/domain/models"
)

type teamOwnerReader interface {
	GetMemberRole(ctx context.Context, teamID, userID int64) (models.Role, error)
}

type teamDeleter interface {
	Delete(ctx context.Context, teamID int64) error
}

type taskCacheInvalidator interface {
	DeleteTeamTasks(ctx context.Context, teamID int64) error
}
