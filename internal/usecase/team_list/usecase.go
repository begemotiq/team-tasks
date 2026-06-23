package team_list

import (
	"context"

	"task-service/internal/domain/models"
)

type UseCase struct {
	teams teamLister
}

func New(teams teamLister) *UseCase {
	return &UseCase{teams: teams}
}

func (uc *UseCase) List(ctx context.Context, userID int64) ([]models.Team, error) {
	return uc.teams.ListByUser(ctx, userID)
}
