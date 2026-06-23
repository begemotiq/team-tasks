package task_history

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	"task-service/internal/adapter/http/requestctx"
	"task-service/internal/domain/models"
)

func TestHandleReturnsTaskHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	tasks := NewMocktaskHistoryReader(ctrl)
	handler := New(tasks)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/100/history", nil)
	request = withPathParam(request, "id", "100")
	request = request.WithContext(requestctx.WithUserID(request.Context(), 10))
	recorder := httptest.NewRecorder()

	tasks.EXPECT().
		GetHistory(gomock.Any(), int64(10), int64(100)).
		Return([]models.TaskHistory{{ID: 1, TaskID: 100, Field: "created"}}, nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"field":"created"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func withPathParam(request *http.Request, key string, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)
	return request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))
}
