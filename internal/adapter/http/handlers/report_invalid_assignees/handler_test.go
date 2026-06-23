package report_invalid_assignees

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task-service/internal/adapter/http/requestctx"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleReturnsInvalidAssignees(t *testing.T) {
	ctrl := gomock.NewController(t)
	reports := NewMockinvalidAssigneesProvider(ctrl)
	handler := New(reports)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/reports/invalid-assignees", nil)
	request = request.WithContext(requestctx.WithUserID(request.Context(), 10))
	recorder := httptest.NewRecorder()

	reports.EXPECT().
		GetInvalidAssignees(gomock.Any(), int64(10)).
		Return([]models.Task{{ID: 100, Title: "Invalid", Status: models.TaskStatusTodo}}, nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"title":"Invalid"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
