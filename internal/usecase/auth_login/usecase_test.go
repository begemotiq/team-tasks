package auth_login

import (
	"context"
	"errors"
	"testing"

	"task-service/internal/domain"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleIssuesToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := NewMockuserFinder(ctrl)
	hasher := NewMockpasswordComparer(ctrl)
	tokens := NewMocktokenIssuer(ctrl)
	uc := New(users, hasher, tokens)
	user := &models.User{ID: 1, Email: "owner@example.com", PasswordHash: "hash"}

	users.EXPECT().FindByEmail(gomock.Any(), "owner@example.com").Return(user, nil)
	hasher.EXPECT().Compare("hash", "password123").Return(nil)
	tokens.EXPECT().NewToken(user).Return("token-1", nil)

	result, err := uc.Login(context.Background(), Input{Email: "owner@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if result.User.ID != 1 || result.Token != "token-1" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestHandleHidesMissingUserAsUnauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := NewMockuserFinder(ctrl)
	hasher := NewMockpasswordComparer(ctrl)
	tokens := NewMocktokenIssuer(ctrl)
	uc := New(users, hasher, tokens)

	users.EXPECT().FindByEmail(gomock.Any(), "owner@example.com").Return(nil, domain.ErrNotFound)

	_, err := uc.Login(context.Background(), Input{Email: "owner@example.com", Password: "password123"})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestHandleReturnsUserLookupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := NewMockuserFinder(ctrl)
	hasher := NewMockpasswordComparer(ctrl)
	tokens := NewMocktokenIssuer(ctrl)
	uc := New(users, hasher, tokens)
	dbErr := errors.New("database is down")

	users.EXPECT().FindByEmail(gomock.Any(), "owner@example.com").Return(nil, dbErr)

	_, err := uc.Login(context.Background(), Input{Email: "owner@example.com", Password: "password123"})
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected lookup error, got %v", err)
	}
	if errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("lookup error must not be masked as unauthorized: %v", err)
	}
}

func TestHandleRejectsWrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := NewMockuserFinder(ctrl)
	hasher := NewMockpasswordComparer(ctrl)
	tokens := NewMocktokenIssuer(ctrl)
	uc := New(users, hasher, tokens)
	user := &models.User{ID: 1, Email: "owner@example.com", PasswordHash: "hash"}

	users.EXPECT().FindByEmail(gomock.Any(), "owner@example.com").Return(user, nil)
	hasher.EXPECT().Compare("hash", "wrong-password").Return(domain.ErrUnauthorized)

	_, err := uc.Login(context.Background(), Input{Email: "owner@example.com", Password: "wrong-password"})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}
