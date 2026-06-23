package auth_login

import (
	"context"
	"errors"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

type UseCase struct {
	users  userFinder
	hasher passwordComparer
	tokens tokenIssuer
}

func New(users userFinder, hasher passwordComparer, tokens tokenIssuer) *UseCase {
	return &UseCase{users: users, hasher: hasher, tokens: tokens}
}

type Input struct {
	Email    string
	Password string
}

type Result struct {
	User  models.User
	Token string
}

func (uc *UseCase) Login(ctx context.Context, input Input) (*Result, error) {
	user, err := uc.users.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}
	if err := uc.hasher.Compare(user.PasswordHash, input.Password); err != nil {
		return nil, err
	}
	token, err := uc.tokens.NewToken(user)
	if err != nil {
		return nil, err
	}
	return &Result{User: *user, Token: token}, nil
}
