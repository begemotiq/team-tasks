package report_top_creators

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task-service/internal/adapter/http/requestctx"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleReturnsTopCreators(t *testing.T) {
	ctrl := gomock.NewController(t)
	reports := NewMocktopCreatorsProvider(ctrl)
	handler := New(reports)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/reports/top-creators", nil)
	request = request.WithContext(requestctx.WithUserID(request.Context(), 10))
	recorder := httptest.NewRecorder()

	reports.EXPECT().
		GetTopCreators(gomock.Any(), int64(10)).
		Return([]models.TopCreator{{TeamID: 1, UserID: 10, UserName: "Owner", TasksCreated: 3}}, nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"user_name":"Owner"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
