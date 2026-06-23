package team_delete

import (
	"context"
	"fmt"

	"task-service/internal/domain"
)

type UseCase struct {
	teams teamStore
	cache taskCacheInvalidator
}

type teamStore interface {
	teamOwnerReader
	teamDeleter
}

func New(teams teamStore, cache taskCacheInvalidator) *UseCase {
	return &UseCase{teams: teams, cache: cache}
}

func (uc *UseCase) Delete(ctx context.Context, ownerID, teamID int64) error {
	role, err := uc.teams.GetMemberRole(ctx, teamID, ownerID)
	if err != nil {
		return err
	}
	if !role.CanDeleteTeam() {
		return domain.ErrForbidden
	}
	if err := uc.teams.Delete(ctx, teamID); err != nil {
		return err
	}
	if err := uc.cache.DeleteTeamTasks(ctx, teamID); err != nil {
		return fmt.Errorf("%w: invalidate task cache: %v", domain.ErrExternal, err)
	}
	return nil
}
