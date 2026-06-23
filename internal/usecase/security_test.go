package usecase_test

import (
	"errors"
	"testing"
	"time"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
	"task-service/internal/usecase"
)

func TestJWTManagerIssuesAndParsesToken(t *testing.T) {
	manager := usecase.JWTManager{Secret: []byte("secret"), TTL: time.Hour}
	token, err := manager.NewToken(&models.User{ID: 42, Email: "owner@example.com"})
	if err != nil {
		t.Fatalf("new token failed: %v", err)
	}
	userID, err := manager.ParseToken(token)
	if err != nil {
		t.Fatalf("parse token failed: %v", err)
	}
	if userID != 42 {
		t.Fatalf("unexpected user id: %d", userID)
	}
	if _, err := manager.ParseToken("not-a-token"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized for bad token, got %v", err)
	}
}

func TestJWTManagerRequiresSecret(t *testing.T) {
	manager := usecase.JWTManager{}

	_, err := manager.NewToken(&models.User{ID: 1, Email: "owner@example.com"})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestBcryptHasherHashAndCompare(t *testing.T) {
	hasher := usecase.BcryptHasher{Cost: 4}

	hash, err := hasher.Hash("password123")
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if err := hasher.Compare(hash, "password123"); err != nil {
		t.Fatalf("compare failed: %v", err)
	}
	if err := hasher.Compare(hash, "wrong-password"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized for wrong password, got %v", err)
	}
}

func TestBcryptHasherUsesDefaultCost(t *testing.T) {
	hasher := usecase.BcryptHasher{}

	hash, err := hasher.Hash("password123")
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if err := hasher.Compare(hash, "password123"); err != nil {
		t.Fatalf("compare failed: %v", err)
	}
}

func TestBcryptHasherReturnsHashError(t *testing.T) {
	hasher := usecase.BcryptHasher{Cost: 100}

	if _, err := hasher.Hash("password123"); err == nil {
		t.Fatal("expected hash error")
	}
}

func TestJWTManagerRejectsTokenWithMissingUserID(t *testing.T) {
	manager := usecase.JWTManager{Secret: []byte("secret"), TTL: time.Hour}
	token, err := manager.NewToken(&models.User{Email: "owner@example.com"})
	if err != nil {
		t.Fatalf("new token failed: %v", err)
	}

	if _, err := manager.ParseToken(token); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}
