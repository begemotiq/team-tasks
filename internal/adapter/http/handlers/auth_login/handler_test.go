package auth_login

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task-service/internal/domain/models"
	authloginusecase "task-service/internal/usecase/auth_login"

	"go.uber.org/mock/gomock"
)

func TestHandleLogsUserIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	auth := NewMockauthenticator(ctrl)
	handler := New(auth)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(`{"email":" OWNER@EXAMPLE.COM ","password":"password123"}`))
	recorder := httptest.NewRecorder()

	auth.EXPECT().
		Login(gomock.Any(), authloginusecase.Input{Email: "owner@example.com", Password: "password123"}).
		Return(&authloginusecase.Result{User: models.User{ID: 1, Email: "owner@example.com", Name: "Owner"}, Token: "token-1"}, nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"token":"token-1"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
