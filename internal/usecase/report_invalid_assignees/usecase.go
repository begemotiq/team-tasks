package report_invalid_assignees

import (
	"context"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

type UseCase struct {
	reports invalidAssigneesReader
	access  reportAccessChecker
}

func New(reports invalidAssigneesReader, access reportAccessChecker) *UseCase {
	return &UseCase{reports: reports, access: access}
}

func (uc *UseCase) GetInvalidAssignees(ctx context.Context, managerID int64) ([]models.Task, error) {
	allowed, err := uc.access.HasManagementRole(ctx, managerID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	return uc.reports.InvalidAssignees(ctx, managerID)
}
