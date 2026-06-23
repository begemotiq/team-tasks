package task_list

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"task-service/internal/domain/models"
)

const teamTasksCacheTTL = 5 * time.Minute

type UseCase struct {
	tasks taskLister
	teams teamMembershipReader
	cache taskListCache
}

func New(tasks taskLister, teams teamMembershipReader, cache taskListCache) *UseCase {
	return &UseCase{tasks: tasks, teams: teams, cache: cache}
}

func (uc *UseCase) List(ctx context.Context, actorID int64, filter models.TaskFilter) (models.TaskList, error) {
	filter = filter.Normalize()
	if filter.TeamID != nil {
		if _, err := uc.teams.GetMemberRole(ctx, *filter.TeamID, actorID); err != nil {
			return models.TaskList{}, err
		}
		key := cacheKey(*filter.TeamID, filter)
		cached, ok, err := uc.cache.GetTaskList(ctx, key)
		if err == nil && ok {
			return cached, nil
		}
		list, err := uc.tasks.List(ctx, filter, actorID)
		if err != nil {
			return models.TaskList{}, err
		}
		_ = uc.cache.SetTaskList(ctx, key, list, teamTasksCacheTTL)
		return list, nil
	}
	return uc.tasks.List(ctx, filter, actorID)
}

func cacheKey(teamID int64, filter models.TaskFilter) string {
	status := ""
	if filter.Status != nil {
		status = string(*filter.Status)
	}
	assignee := ""
	if filter.AssigneeID != nil {
		assignee = strconv.FormatInt(*filter.AssigneeID, 10)
	}
	cursor := ""
	if filter.Cursor != nil {
		cursor = fmt.Sprintf("%d:%d", filter.Cursor.CreatedAt.UnixNano(), filter.Cursor.ID)
	}
	return fmt.Sprintf("team_tasks:%d:status=%s:assignee=%s:cursor=%s:size=%d", teamID, status, assignee, cursor, filter.PageSize)
}
