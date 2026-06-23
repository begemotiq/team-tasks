package task_history

import (
	"context"

	"task-service/internal/domain/models"
)

type UseCase struct {
	tasks taskHistoryReader
	teams teamMembershipReader
}

func New(tasks taskHistoryReader, teams teamMembershipReader) *UseCase {
	return &UseCase{tasks: tasks, teams: teams}
}

func (uc *UseCase) GetHistory(ctx context.Context, actorID, taskID int64) ([]models.TaskHistory, error) {
	task, err := uc.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if _, err := uc.teams.GetMemberRole(ctx, task.TeamID, actorID); err != nil {
		return nil, err
	}
	return uc.tasks.History(ctx, taskID)
}
