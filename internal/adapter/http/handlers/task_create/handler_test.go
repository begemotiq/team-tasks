package task_create

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task-service/internal/adapter/http/requestctx"
	"task-service/internal/domain/models"
	taskcreateusecase "task-service/internal/usecase/task_create"

	"go.uber.org/mock/gomock"
)

func TestHandleCreatesTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskCreator(ctrl)
	handler := New(tasks)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{"title":" Implement API ","team_id":1}`))
	request = request.WithContext(requestctx.WithUserID(request.Context(), 10))
	recorder := httptest.NewRecorder()

	tasks.EXPECT().
		Create(gomock.Any(), int64(10), taskcreateusecase.Input{Title: "Implement API", Status: models.TaskStatusTodo, TeamID: 1}).
		Return(&models.Task{ID: 100, Title: "Implement API", Status: models.TaskStatusTodo, TeamID: 1, CreatedBy: 10}, nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"title":"Implement API"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
