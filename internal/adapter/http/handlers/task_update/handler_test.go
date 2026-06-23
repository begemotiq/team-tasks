package task_update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	"task-service/internal/adapter/http/requestctx"
	"task-service/internal/domain/models"
	taskupdateusecase "task-service/internal/usecase/task_update"
)

func TestHandleUpdatesTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	handler := New(tasks)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/100", strings.NewReader(`{"status":"done"}`))
	request = withPathParam(request, "id", "100")
	request = request.WithContext(requestctx.WithUserID(request.Context(), 10))
	recorder := httptest.NewRecorder()
	status := models.TaskStatusDone

	tasks.EXPECT().
		Update(gomock.Any(), int64(10), int64(100), taskupdateusecase.Input{Status: &status}).
		Return(&models.Task{ID: 100, Title: "Implement API", Status: models.TaskStatusDone, TeamID: 1}, nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"status":"done"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestHandleClearsNullableFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskUpdater(ctrl)
	handler := New(tasks)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/100", strings.NewReader(`{"assignee_id":null,"due_date":null}`))
	request = withPathParam(request, "id", "100")
	request = request.WithContext(requestctx.WithUserID(request.Context(), 10))
	recorder := httptest.NewRecorder()

	tasks.EXPECT().
		Update(gomock.Any(), int64(10), int64(100), taskupdateusecase.Input{
			AssigneeID: taskupdateusecase.Optional[int64]{Set: true},
			DueDate:    taskupdateusecase.Optional[time.Time]{Set: true},
		}).
		Return(&models.Task{ID: 100, Title: "Implement API", Status: models.TaskStatusTodo, TeamID: 1}, nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
}

func withPathParam(request *http.Request, key string, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)
	return request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))
}
