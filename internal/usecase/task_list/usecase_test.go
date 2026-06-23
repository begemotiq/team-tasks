package task_list

import (
	"context"
	"testing"
	"time"

	"task-service/internal/domain"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleReturnsCachedTeamList(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskLister(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskListCache(ctrl)
	uc := New(tasks, teams, cache)
	teamID := int64(1)
	status := models.TaskStatusTodo
	cached := models.TaskList{Items: []models.Task{{ID: 100}}}

	teams.EXPECT().GetMemberRole(gomock.Any(), teamID, int64(10)).Return(models.RoleMember, nil)
	cache.EXPECT().GetTaskList(gomock.Any(), "team_tasks:1:status=todo:assignee=:cursor=:size=20").Return(cached, true, nil)

	result, err := uc.List(context.Background(), 10, models.TaskFilter{TeamID: &teamID, Status: &status})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].ID != 100 {
		t.Fatalf("unexpected list: %#v", result)
	}
}

func TestHandleStoresTeamListInCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskLister(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskListCache(ctrl)
	uc := New(tasks, teams, cache)
	teamID := int64(1)
	list := models.TaskList{Items: []models.Task{{ID: 100}}, HasMore: true}

	teams.EXPECT().GetMemberRole(gomock.Any(), teamID, int64(10)).Return(models.RoleMember, nil)
	cache.EXPECT().GetTaskList(gomock.Any(), "team_tasks:1:status=:assignee=:cursor=:size=20").Return(models.TaskList{}, false, nil)
	tasks.EXPECT().List(gomock.Any(), models.TaskFilter{TeamID: &teamID, PageSize: 20}, int64(10)).Return(list, nil)
	cache.EXPECT().SetTaskList(gomock.Any(), "team_tasks:1:status=:assignee=:cursor=:size=20", list, teamTasksCacheTTL).Return(nil)

	result, err := uc.List(context.Background(), 10, models.TaskFilter{TeamID: &teamID})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if !result.HasMore {
		t.Fatalf("unexpected list: %#v", result)
	}
}

func TestHandleRejectsNonMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskLister(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskListCache(ctrl)
	uc := New(tasks, teams, cache)
	teamID := int64(1)

	teams.EXPECT().GetMemberRole(gomock.Any(), teamID, int64(10)).Return(models.Role(""), domain.ErrNotFound)

	_, err := uc.List(context.Background(), 10, models.TaskFilter{TeamID: &teamID})
	if err != domain.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestHandleWithoutTeamBypassesCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskLister(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskListCache(ctrl)
	uc := New(tasks, teams, cache)
	list := models.TaskList{Items: []models.Task{{ID: 100}}}

	tasks.EXPECT().List(gomock.Any(), models.TaskFilter{PageSize: 20}, int64(10)).Return(list, nil)

	result, err := uc.List(context.Background(), 10, models.TaskFilter{})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("unexpected list: %#v", result)
	}
}

func TestHandleUsesCursorInCacheKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskLister(ctrl)
	teams := NewMockteamMembershipReader(ctrl)
	cache := NewMocktaskListCache(ctrl)
	uc := New(tasks, teams, cache)
	teamID := int64(1)
	cursor := &models.TaskCursor{CreatedAt: time.Unix(1700000000, 123), ID: 55}
	list := models.TaskList{Items: []models.Task{{ID: 54}}}

	teams.EXPECT().GetMemberRole(gomock.Any(), teamID, int64(10)).Return(models.RoleMember, nil)
	cache.EXPECT().GetTaskList(gomock.Any(), "team_tasks:1:status=:assignee=:cursor=1700000000000000123:55:size=20").Return(models.TaskList{}, false, nil)
	tasks.EXPECT().List(gomock.Any(), models.TaskFilter{TeamID: &teamID, Cursor: cursor, PageSize: 20}, int64(10)).Return(list, nil)
	cache.EXPECT().SetTaskList(gomock.Any(), "team_tasks:1:status=:assignee=:cursor=1700000000000000123:55:size=20", list, teamTasksCacheTTL).Return(nil)

	if _, err := uc.List(context.Background(), 10, models.TaskFilter{TeamID: &teamID, Cursor: cursor}); err != nil {
		t.Fatalf("handle failed: %v", err)
	}
}
