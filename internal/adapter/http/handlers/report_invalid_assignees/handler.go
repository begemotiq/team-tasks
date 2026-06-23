package report_invalid_assignees

import (
	"net/http"

	"task-service/internal/adapter/http/request"
	"task-service/internal/adapter/http/response"
)

type Handler struct {
	reports invalidAssigneesProvider
}

func New(reports invalidAssigneesProvider) *Handler {
	return &Handler{reports: reports}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := request.UserID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	result, err := h.reports.GetInvalidAssignees(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, response.NewTaskItems(result))
}
