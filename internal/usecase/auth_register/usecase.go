package auth_register

import (
	"context"

	"task-service/internal/domain/models"
)

type UseCase struct {
	users  userCreator
	hasher passwordHasher
	tokens tokenIssuer
}

func New(users userCreator, hasher passwordHasher, tokens tokenIssuer) *UseCase {
	return &UseCase{users: users, hasher: hasher, tokens: tokens}
}

type Input struct {
	Email    string
	Password string
	Name     string
}

type Result struct {
	User  models.User
	Token string
}

func (uc *UseCase) Register(ctx context.Context, input Input) (*Result, error) {
	passwordHash, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Email:        input.Email,
		Name:         input.Name,
		PasswordHash: passwordHash,
	}
	if err := uc.users.Create(ctx, user); err != nil {
		return nil, err
	}
	token, err := uc.tokens.NewToken(user)
	if err != nil {
		return nil, err
	}
	return &Result{User: *user, Token: token}, nil
}
