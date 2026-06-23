package team_invite

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	"task-service/internal/adapter/http/requestctx"
	teaminviteusecase "task-service/internal/usecase/team_invite"
)

func TestHandleInvitesUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	teams := NewMockteamInviter(ctrl)
	handler := New(teams)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/teams/1/invite", strings.NewReader(`{"email":" MEMBER@EXAMPLE.COM "}`))
	request = withPathParam(request, "id", "1")
	request = request.WithContext(requestctx.WithUserID(request.Context(), 10))
	recorder := httptest.NewRecorder()

	teams.EXPECT().
		Invite(gomock.Any(), int64(10), int64(1), teaminviteusecase.Input{Email: "member@example.com", Role: "member"}).
		Return(nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusAccepted, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"status":"invited"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func withPathParam(request *http.Request, key string, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)
	return request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))
}
