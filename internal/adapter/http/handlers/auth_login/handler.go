package auth_login

import (
	"net/http"

	"task-service/internal/adapter/http/request"
	"task-service/internal/adapter/http/response"
)

type Handler struct {
	auth authenticator
}

func New(auth authenticator) *Handler {
	return &Handler{auth: auth}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	var payload request.LoginRequest
	if err := request.DecodeJSON(r, &payload); err != nil {
		response.Error(w, err)
		return
	}
	if err := payload.Validate(); err != nil {
		response.Error(w, err)
		return
	}
	result, err := h.auth.Login(r.Context(), payload.ToInput())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, response.NewAuth(result.User, result.Token))
}
