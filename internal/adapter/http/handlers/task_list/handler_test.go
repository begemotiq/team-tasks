package task_list

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"task-service/internal/adapter/http/pagination"
	"task-service/internal/adapter/http/requestctx"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleListsTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskLister(ctrl)
	handler := New(tasks)
	createdAt := time.Date(2026, 6, 22, 10, 30, 0, 0, time.UTC)
	cursor := &models.TaskCursor{CreatedAt: createdAt, ID: 90}
	nextCursor := &models.TaskCursor{CreatedAt: createdAt.Add(-time.Minute), ID: 80}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?team_id=1&status=todo&cursor="+pagination.EncodeTaskCursor(cursor)+"&page_size=10", nil)
	request = request.WithContext(requestctx.WithUserID(request.Context(), 10))
	recorder := httptest.NewRecorder()
	teamID := int64(1)
	status := models.TaskStatusTodo

	tasks.EXPECT().
		List(gomock.Any(), int64(10), models.TaskFilter{TeamID: &teamID, Status: &status, Cursor: cursor, PageSize: 10}).
		Return(models.TaskList{Items: []models.Task{{ID: 100, Title: "Implement API"}}, NextCursor: nextCursor, HasMore: true}, nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"has_more":true`) || !strings.Contains(recorder.Body.String(), pagination.EncodeTaskCursor(nextCursor)) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
