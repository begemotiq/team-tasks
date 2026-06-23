package task_list

import (
	"net/http"

	"task-service/internal/adapter/http/request"
	"task-service/internal/adapter/http/response"
)

type Handler struct {
	tasks taskLister
}

func New(tasks taskLister) *Handler {
	return &Handler{tasks: tasks}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := request.UserID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	filter, err := request.TaskFilterFromQuery(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	list, err := h.tasks.List(r.Context(), userID, filter)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, response.NewTaskList(list))
}
