package team_delete

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	"task-service/internal/adapter/http/requestctx"
)

func TestHandleDeletesTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	teams := NewMockteamDeleter(ctrl)
	handler := New(teams)
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/1", nil)
	request = withPathParam(request, "id", "1")
	request = request.WithContext(requestctx.WithUserID(request.Context(), 10))
	recorder := httptest.NewRecorder()

	teams.EXPECT().Delete(gomock.Any(), int64(10), int64(1)).Return(nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusNoContent, recorder.Code, recorder.Body.String())
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", recorder.Body.String())
	}
}

func withPathParam(request *http.Request, key string, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)
	return request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))
}
