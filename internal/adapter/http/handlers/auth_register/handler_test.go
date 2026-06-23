package auth_register

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task-service/internal/domain/models"
	authregisterusecase "task-service/internal/usecase/auth_register"

	"go.uber.org/mock/gomock"
)

func TestHandleRegistersUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	registrar := NewMockuserRegistrar(ctrl)
	handler := New(registrar)
	body := `{"email":" OWNER@EXAMPLE.COM ","password":"password123","name":" Owner "}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/register", strings.NewReader(body))
	recorder := httptest.NewRecorder()

	registrar.EXPECT().
		Register(gomock.Any(), authregisterusecase.Input{Email: "owner@example.com", Password: "password123", Name: "Owner"}).
		Return(&authregisterusecase.Result{User: models.User{ID: 1, Email: "owner@example.com", Name: "Owner"}, Token: "token-1"}, nil)

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"token":"token-1"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestHandleRejectsInvalidRegisterPayload(t *testing.T) {
	handler := New(NewMockuserRegistrar(gomock.NewController(t)))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/register", strings.NewReader(`{"email":"bad","password":"short","name":"Owner"}`))
	recorder := httptest.NewRecorder()

	handler.Handle(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}
