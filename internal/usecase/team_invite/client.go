//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package team_invite

import (
	"context"

	"task-service/internal/domain/models"
)

type teamReader interface {
	FindByID(ctx context.Context, id int64) (*models.Team, error)
	GetMemberRole(ctx context.Context, teamID, userID int64) (models.Role, error)
}

type teamInviteStore interface {
	teamReader
	teamMemberInviter
}

type teamMemberInviter interface {
	AddMemberWithOutboxEvent(ctx context.Context, teamID, userID int64, role models.Role, event *models.OutboxEvent) error
}

type inviteUserFinder interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}
