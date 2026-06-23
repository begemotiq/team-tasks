package response

import (
	"time"

	"task-service/internal/domain/models"
)

type TeamResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamListResponse struct {
	Items []TeamResponse `json:"items"`
}

func NewTeam(team models.Team) TeamResponse {
	return TeamResponse{
		ID:        team.ID,
		Name:      team.Name,
		CreatedBy: team.CreatedBy,
		CreatedAt: team.CreatedAt,
	}
}

func NewTeamList(teams []models.Team) TeamListResponse {
	items := make([]TeamResponse, 0, len(teams))
	for _, team := range teams {
		items = append(items, NewTeam(team))
	}
	return TeamListResponse{Items: items}
}
