package health

import (
	"net/http"

	"task-service/internal/adapter/http/response"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, response.NewStatus("ok"))
}
