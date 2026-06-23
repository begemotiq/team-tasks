//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package team_delete

import "context"

type teamDeleter interface {
	Delete(ctx context.Context, ownerID, teamID int64) error
}
