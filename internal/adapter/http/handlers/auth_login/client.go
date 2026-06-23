//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package auth_login

import (
	"context"

	authloginusecase "task-service/internal/usecase/auth_login"
)

type authenticator interface {
	Login(ctx context.Context, input authloginusecase.Input) (*authloginusecase.Result, error)
}
