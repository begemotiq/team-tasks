//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package auth_register

import (
	"context"

	"task-service/internal/domain/models"
)

type userCreator interface {
	Create(ctx context.Context, user *models.User) error
}

type passwordHasher interface {
	Hash(password string) (string, error)
}

type tokenIssuer interface {
	NewToken(user *models.User) (string, error)
}
