//go:build integration

package integration

import (
	"errors"
	"testing"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

func TestMySQLTeamRepository(t *testing.T) {
	t.Parallel()

	fixture := newFixture(t)
	owner := fixture.user("owner")
	admin := fixture.user("admin")
	member := fixture.user("member")
	outsider := fixture.user("outsider")

	team := fixture.team("backend", owner)
	if team.ID == 0 {
		t.Fatal("created team id is empty")
	}
	if team.CreatedBy != owner.ID {
		t.Fatalf("expected team owner %d, got %d", owner.ID, team.CreatedBy)
	}
	if team.CreatedAt.IsZero() {
		t.Fatal("created team timestamp is empty")
	}

	found, err := fixture.repos.teams.FindByID(fixture.ctx, team.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.ID != team.ID || found.Name != team.Name {
		t.Fatalf("unexpected team by id: %#v", found)
	}

	role, err := fixture.repos.teams.GetMemberRole(fixture.ctx, team.ID, owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if role != models.RoleOwner {
		t.Fatalf("expected owner role, got %q", role)
	}

	fixture.member(team, admin, models.RoleAdmin)
	fixture.member(team, member, models.RoleMember)

	role, err = fixture.repos.teams.GetMemberRole(fixture.ctx, team.ID, admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if role != models.RoleAdmin {
		t.Fatalf("expected admin role, got %q", role)
	}

	teams, err := fixture.repos.teams.ListByUser(fixture.ctx, member.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(teams) != 1 || teams[0].ID != team.ID {
		t.Fatalf("unexpected teams by member: %#v", teams)
	}

	hasManagementRole, err := fixture.repos.teams.HasManagementRole(fixture.ctx, owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !hasManagementRole {
		t.Fatal("owner must have management role")
	}

	hasManagementRole, err = fixture.repos.teams.HasManagementRole(fixture.ctx, admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !hasManagementRole {
		t.Fatal("admin must have management role")
	}

	hasManagementRole, err = fixture.repos.teams.HasManagementRole(fixture.ctx, member.ID)
	if err != nil {
		t.Fatal(err)
	}
	if hasManagementRole {
		t.Fatal("member must not have management role")
	}

	_, err = fixture.repos.teams.GetMemberRole(fixture.ctx, team.ID, outsider.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found for outsider role, got %v", err)
	}
}

func TestMySQLTeamRepositoryDelete(t *testing.T) {
	t.Parallel()

	fixture := newFixture(t)
	owner := fixture.user("owner")
	team := fixture.team("backend", owner)

	if err := fixture.repos.teams.Delete(fixture.ctx, team.ID); err != nil {
		t.Fatal(err)
	}

	_, err := fixture.repos.teams.FindByID(fixture.ctx, team.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected deleted team to be missing, got %v", err)
	}

	if err := fixture.repos.teams.Delete(fixture.ctx, team.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found on repeated delete, got %v", err)
	}
}
