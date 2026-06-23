//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package auth_login

import (
	"context"

	"task-service/internal/domain/models"
)

type userFinder interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type passwordComparer interface {
	Compare(hash string, password string) error
}

type tokenIssuer interface {
	NewToken(user *models.User) (string, error)
}
