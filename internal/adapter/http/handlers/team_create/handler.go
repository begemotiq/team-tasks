package team_create

import (
	"net/http"

	"task-service/internal/adapter/http/request"
	"task-service/internal/adapter/http/response"
)

type Handler struct {
	teams teamCreator
}

func New(teams teamCreator) *Handler {
	return &Handler{teams: teams}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := request.UserID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	var payload request.CreateTeamRequest
	if err := request.DecodeJSON(r, &payload); err != nil {
		response.Error(w, err)
		return
	}
	if err := payload.Validate(); err != nil {
		response.Error(w, err)
		return
	}
	team, err := h.teams.Create(r.Context(), userID, payload.ToInput())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, response.NewTeam(*team))
}
