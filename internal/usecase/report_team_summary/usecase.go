package report_team_summary

import (
	"context"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

type UseCase struct {
	reports teamSummaryReader
	access  reportAccessChecker
}

func New(reports teamSummaryReader, access reportAccessChecker) *UseCase {
	return &UseCase{reports: reports, access: access}
}

func (uc *UseCase) GetTeamSummary(ctx context.Context, managerID int64) ([]models.TeamSummary, error) {
	allowed, err := uc.access.HasManagementRole(ctx, managerID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	return uc.reports.TeamSummary(ctx, managerID)
}
