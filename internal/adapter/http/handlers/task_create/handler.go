package task_create

import (
	"net/http"

	"task-service/internal/adapter/http/request"
	"task-service/internal/adapter/http/response"
)

type Handler struct {
	tasks taskCreator
}

func New(tasks taskCreator) *Handler {
	return &Handler{tasks: tasks}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := request.UserID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	var payload request.CreateTaskRequest
	if err := request.DecodeJSON(r, &payload); err != nil {
		response.Error(w, err)
		return
	}
	if err := payload.Validate(); err != nil {
		response.Error(w, err)
		return
	}
	task, err := h.tasks.Create(r.Context(), userID, payload.ToInput())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, response.NewTask(*task))
}
