package team_delete

import (
	"net/http"

	"task-service/internal/adapter/http/request"
	"task-service/internal/adapter/http/response"
)

type Handler struct {
	teams teamDeleter
}

func New(teams teamDeleter) *Handler {
	return &Handler{teams: teams}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := request.UserID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	teamID, err := request.PathID(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}
	if err := h.teams.Delete(r.Context(), userID, teamID); err != nil {
		response.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
