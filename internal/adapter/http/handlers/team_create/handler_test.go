package team_create

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task-service/internal/adapter/http/requestctx"
	"task-service/internal/domain/models"
	teamcreateusecase "task-service/internal/usecase/team_create"

	"go.uber.org/mock/gomock"
)

func TestHandleCreatesTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	teams := NewMockteamCreator(ctrl)
	handler := New(teams)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/teams", strings.NewReader(`{"name":" Backend "}`))
	request = request.WithContext(requestctx.WithUserID(request.Context(), 10))
	recorder := httptest.NewRecorder()

	teams.EXPECT().
		Create(gomock.Any(), int64(10), teamcreateusecase.Input{Name: "Backend"}).
		Return(&models.Team{ID: 1, Name: "Backend", CreatedBy: 10}, nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"name":"Backend"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
