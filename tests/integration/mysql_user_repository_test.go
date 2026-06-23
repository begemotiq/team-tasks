//go:build integration

package integration

import (
	"errors"
	"testing"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

func TestMySQLUserRepository(t *testing.T) {
	t.Parallel()

	fixture := newFixture(t)

	created := fixture.user("owner")
	if created.ID == 0 {
		t.Fatal("created user id is empty")
	}
	if created.CreatedAt.IsZero() {
		t.Fatal("created user timestamp is empty")
	}

	byEmail, err := fixture.repos.users.FindByEmail(fixture.ctx, created.Email)
	if err != nil {
		t.Fatal(err)
	}
	if byEmail.ID != created.ID || byEmail.Email != created.Email {
		t.Fatalf("unexpected user by email: %#v", byEmail)
	}

	byID, err := fixture.repos.users.FindByID(fixture.ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if byID.ID != created.ID || byID.Name != created.Name {
		t.Fatalf("unexpected user by id: %#v", byID)
	}

	duplicate := &models.User{
		Email:        created.Email,
		PasswordHash: "hash",
		Name:         "Duplicate",
	}
	if err := fixture.repos.users.Create(fixture.ctx, duplicate); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict for duplicate email, got %v", err)
	}

	_, err = fixture.repos.users.FindByEmail(fixture.ctx, "missing-"+fixture.suffix+"@example.com")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found by email, got %v", err)
	}

	_, err = fixture.repos.users.FindByID(fixture.ctx, 9_999_999_999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found by id, got %v", err)
	}
}
