package report_top_creators

import (
	"context"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

type UseCase struct {
	reports topCreatorsReader
	access  reportAccessChecker
}

func New(reports topCreatorsReader, access reportAccessChecker) *UseCase {
	return &UseCase{reports: reports, access: access}
}

func (uc *UseCase) GetTopCreators(ctx context.Context, managerID int64) ([]models.TopCreator, error) {
	allowed, err := uc.access.HasManagementRole(ctx, managerID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	return uc.reports.TopCreatorsByTeam(ctx, managerID)
}
