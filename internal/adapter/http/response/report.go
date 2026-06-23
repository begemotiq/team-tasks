package response

import "task-service/internal/domain/models"

type TeamSummaryResponse struct {
	TeamID             int64  `json:"team_id"`
	TeamName           string `json:"team_name"`
	MembersCount       int64  `json:"members_count"`
	DoneTasksLast7Days int64  `json:"done_tasks_last_7_days"`
}

type TeamSummaryListResponse struct {
	Items []TeamSummaryResponse `json:"items"`
}

type TopCreatorResponse struct {
	TeamID       int64  `json:"team_id"`
	TeamName     string `json:"team_name"`
	UserID       int64  `json:"user_id"`
	UserName     string `json:"user_name"`
	TasksCreated int64  `json:"tasks_created"`
	RankPosition int64  `json:"rank_position"`
}

type TopCreatorListResponse struct {
	Items []TopCreatorResponse `json:"items"`
}

func NewTeamSummary(summary models.TeamSummary) TeamSummaryResponse {
	return TeamSummaryResponse{
		TeamID:             summary.TeamID,
		TeamName:           summary.TeamName,
		MembersCount:       summary.MembersCount,
		DoneTasksLast7Days: summary.DoneTasksLast7Days,
	}
}

func NewTeamSummaryList(summaries []models.TeamSummary) TeamSummaryListResponse {
	items := make([]TeamSummaryResponse, 0, len(summaries))
	for _, summary := range summaries {
		items = append(items, NewTeamSummary(summary))
	}
	return TeamSummaryListResponse{Items: items}
}

func NewTopCreator(creator models.TopCreator) TopCreatorResponse {
	return TopCreatorResponse{
		TeamID:       creator.TeamID,
		TeamName:     creator.TeamName,
		UserID:       creator.UserID,
		UserName:     creator.UserName,
		TasksCreated: creator.TasksCreated,
		RankPosition: creator.RankPosition,
	}
}

func NewTopCreatorList(creators []models.TopCreator) TopCreatorListResponse {
	items := make([]TopCreatorResponse, 0, len(creators))
	for _, creator := range creators {
		items = append(items, NewTopCreator(creator))
	}
	return TopCreatorListResponse{Items: items}
}
