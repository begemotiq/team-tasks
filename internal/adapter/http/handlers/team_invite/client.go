//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package team_invite

import (
	"context"

	teaminviteusecase "task-service/internal/usecase/team_invite"
)

type teamInviter interface {
	Invite(ctx context.Context, inviterID int64, teamID int64, input teaminviteusecase.Input) error
}
