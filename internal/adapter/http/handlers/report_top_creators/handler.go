package report_top_creators

import (
	"net/http"

	"task-service/internal/adapter/http/request"
	"task-service/internal/adapter/http/response"
)

type Handler struct {
	reports topCreatorsProvider
}

func New(reports topCreatorsProvider) *Handler {
	return &Handler{reports: reports}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := request.UserID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	result, err := h.reports.GetTopCreators(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, response.NewTopCreatorList(result))
}
