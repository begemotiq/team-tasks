package models

type TeamSummary struct {
	TeamID             int64
	TeamName           string
	MembersCount       int64
	DoneTasksLast7Days int64
}
