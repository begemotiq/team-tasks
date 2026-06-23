package auth_register

import (
	"context"
	"errors"
	"testing"

	"task-service/internal/domain"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleCreatesUserAndIssuesToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := NewMockuserCreator(ctrl)
	hasher := NewMockpasswordHasher(ctrl)
	tokens := NewMocktokenIssuer(ctrl)
	uc := New(users, hasher, tokens)

	hasher.EXPECT().Hash("password123").Return("hash", nil)
	users.EXPECT().
		Create(gomock.Any(), gomock.AssignableToTypeOf(&models.User{})).
		DoAndReturn(func(_ context.Context, user *models.User) error {
			if user.Email != "owner@example.com" || user.Name != "Owner" || user.PasswordHash != "hash" {
				t.Fatalf("unexpected user before create: %#v", user)
			}
			user.ID = 1
			return nil
		})
	tokens.EXPECT().
		NewToken(gomock.AssignableToTypeOf(&models.User{})).
		Return("token-1", nil)

	result, err := uc.Register(context.Background(), Input{
		Email:    "owner@example.com",
		Password: "password123",
		Name:     "Owner",
	})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if result.User.ID != 1 || result.Token != "token-1" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestHandlePropagatesCreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := NewMockuserCreator(ctrl)
	hasher := NewMockpasswordHasher(ctrl)
	tokens := NewMocktokenIssuer(ctrl)
	uc := New(users, hasher, tokens)

	hasher.EXPECT().Hash("password123").Return("hash", nil)
	users.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(domain.ErrConflict)

	_, err := uc.Register(context.Background(), Input{
		Email:    "owner@example.com",
		Password: "password123",
		Name:     "Owner",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestHandlePropagatesHashError(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := NewMockuserCreator(ctrl)
	hasher := NewMockpasswordHasher(ctrl)
	tokens := NewMocktokenIssuer(ctrl)
	uc := New(users, hasher, tokens)
	hashErr := errors.New("hash failed")

	hasher.EXPECT().Hash("password123").Return("", hashErr)

	_, err := uc.Register(context.Background(), Input{
		Email:    "owner@example.com",
		Password: "password123",
		Name:     "Owner",
	})
	if !errors.Is(err, hashErr) {
		t.Fatalf("expected hash error, got %v", err)
	}
}

func TestHandlePropagatesTokenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := NewMockuserCreator(ctrl)
	hasher := NewMockpasswordHasher(ctrl)
	tokens := NewMocktokenIssuer(ctrl)
	uc := New(users, hasher, tokens)
	tokenErr := errors.New("token failed")

	hasher.EXPECT().Hash("password123").Return("hash", nil)
	users.EXPECT().
		Create(gomock.Any(), gomock.AssignableToTypeOf(&models.User{})).
		DoAndReturn(func(_ context.Context, user *models.User) error {
			user.ID = 1
			return nil
		})
	tokens.EXPECT().NewToken(gomock.AssignableToTypeOf(&models.User{})).Return("", tokenErr)

	_, err := uc.Register(context.Background(), Input{
		Email:    "owner@example.com",
		Password: "password123",
		Name:     "Owner",
	})
	if !errors.Is(err, tokenErr) {
		t.Fatalf("expected token error, got %v", err)
	}
}
