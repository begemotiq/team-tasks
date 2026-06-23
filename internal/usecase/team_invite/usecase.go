package team_invite

import (
	"context"
	"fmt"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

type UseCase struct {
	teams   teamReader
	members teamMemberInviter
	users   inviteUserFinder
}

func New(teams teamInviteStore, users inviteUserFinder) *UseCase {
	return &UseCase{teams: teams, members: teams, users: users}
}

type Input struct {
	Email string
	Role  models.Role
}

func (uc *UseCase) Invite(ctx context.Context, inviterID, teamID int64, input Input) error {
	role, err := uc.teams.GetMemberRole(ctx, teamID, inviterID)
	if err != nil {
		return err
	}
	if !role.CanInvite() {
		return domain.ErrForbidden
	}
	inviteRole := input.Role
	if inviteRole == "" {
		inviteRole = models.RoleMember
	}
	if inviteRole == models.RoleOwner {
		return fmt.Errorf("%w: invalid invite role", domain.ErrInvalidInput)
	}
	user, err := uc.users.FindByEmail(ctx, input.Email)
	if err != nil {
		return err
	}
	team, err := uc.teams.FindByID(ctx, teamID)
	if err != nil {
		return err
	}
	event, err := models.NewTeamInviteEmailEvent(user.Email, team.Name)
	if err != nil {
		return err
	}
	return uc.members.AddMemberWithOutboxEvent(ctx, teamID, user.ID, inviteRole, event)
}
