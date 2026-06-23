package report_team_summary

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task-service/internal/adapter/http/requestctx"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleReturnsTeamSummary(t *testing.T) {
	ctrl := gomock.NewController(t)
	reports := NewMockteamSummaryProvider(ctrl)
	handler := New(reports)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/reports/team-summary", nil)
	request = request.WithContext(requestctx.WithUserID(request.Context(), 10))
	recorder := httptest.NewRecorder()

	reports.EXPECT().
		GetTeamSummary(gomock.Any(), int64(10)).
		Return([]models.TeamSummary{{TeamID: 1, TeamName: "Backend", MembersCount: 2}}, nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"team_name":"Backend"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
