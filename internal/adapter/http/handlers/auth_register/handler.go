package auth_register

import (
	"net/http"

	"task-service/internal/adapter/http/request"
	"task-service/internal/adapter/http/response"
)

type Handler struct {
	registrar userRegistrar
}

func New(registrar userRegistrar) *Handler {
	return &Handler{registrar: registrar}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	var payload request.RegisterRequest
	if err := request.DecodeJSON(r, &payload); err != nil {
		response.Error(w, err)
		return
	}
	if err := payload.Validate(); err != nil {
		response.Error(w, err)
		return
	}
	result, err := h.registrar.Register(r.Context(), payload.ToInput())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, response.NewAuth(result.User, result.Token))
}
