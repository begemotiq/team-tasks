package task_update

import (
	"net/http"

	"task-service/internal/adapter/http/request"
	"task-service/internal/adapter/http/response"
)

type Handler struct {
	tasks taskUpdater
}

func New(tasks taskUpdater) *Handler {
	return &Handler{tasks: tasks}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := request.UserID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	taskID, err := request.PathID(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}
	var payload request.UpdateTaskRequest
	if err := request.DecodeJSON(r, &payload); err != nil {
		response.Error(w, err)
		return
	}
	if err := payload.Validate(); err != nil {
		response.Error(w, err)
		return
	}
	task, err := h.tasks.Update(r.Context(), userID, taskID, payload.ToInput())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, response.NewTask(*task))
}
