package team_invite

import (
	"net/http"

	"task-service/internal/adapter/http/request"
	"task-service/internal/adapter/http/response"
)

type Handler struct {
	teams teamInviter
}

func New(teams teamInviter) *Handler {
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
	var payload request.InviteRequest
	if err := request.DecodeJSON(r, &payload); err != nil {
		response.Error(w, err)
		return
	}
	if err := payload.Validate(); err != nil {
		response.Error(w, err)
		return
	}
	if err := h.teams.Invite(r.Context(), userID, teamID, payload.ToInput()); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusAccepted, response.NewStatus("invited"))
}
