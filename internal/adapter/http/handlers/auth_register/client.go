//go:generate go run go.uber.org/mock/mockgen -source=client.go -destination=mock_client_test.go -package=$GOPACKAGE

package auth_register

import (
	"context"

	authregisterusecase "task-service/internal/usecase/auth_register"
)

type userRegistrar interface {
	Register(ctx context.Context, input authregisterusecase.Input) (*authregisterusecase.Result, error)
}
