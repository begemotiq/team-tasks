package team_list

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task-service/internal/adapter/http/requestctx"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleListsTeams(t *testing.T) {
	ctrl := gomock.NewController(t)
	teams := NewMockteamLister(ctrl)
	handler := New(teams)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)
	request = request.WithContext(requestctx.WithUserID(request.Context(), 10))
	recorder := httptest.NewRecorder()

	teams.EXPECT().
		List(gomock.Any(), int64(10)).
		Return([]models.Team{{ID: 1, Name: "Backend"}}, nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"items"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
