package task_history

import (
	"net/http"

	"task-service/internal/adapter/http/request"
	"task-service/internal/adapter/http/response"
)

type Handler struct {
	tasks taskHistoryReader
}

func New(tasks taskHistoryReader) *Handler {
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
	history, err := h.tasks.GetHistory(r.Context(), userID, taskID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, response.NewTaskHistoryList(history))
}
