package team_create

import (
	"context"

	"task-service/internal/domain/models"
)

type UseCase struct {
	teams teamCreator
}

func New(teams teamCreator) *UseCase {
	return &UseCase{teams: teams}
}

type Input struct {
	Name string
}

func (uc *UseCase) Create(ctx context.Context, ownerID int64, input Input) (*models.Team, error) {
	team := &models.Team{Name: input.Name, CreatedBy: ownerID}
	if err := uc.teams.CreateWithOwner(ctx, team, ownerID); err != nil {
		return nil, err
	}
	return team, nil
}
