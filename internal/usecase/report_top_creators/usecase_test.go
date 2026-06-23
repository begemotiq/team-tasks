package report_top_creators

import (
	"context"
	"errors"
	"testing"

	"task-service/internal/domain"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

func TestHandleReturnsTopCreators(t *testing.T) {
	ctrl := gomock.NewController(t)
	reports := NewMocktopCreatorsReader(ctrl)
	access := NewMockreportAccessChecker(ctrl)
	uc := New(reports, access)
	expected := []models.TopCreator{{TeamID: 1, UserID: 10, TasksCreated: 3}}

	access.EXPECT().HasManagementRole(gomock.Any(), int64(10)).Return(true, nil)
	reports.EXPECT().TopCreatorsByTeam(gomock.Any(), int64(10)).Return(expected, nil)

	result, err := uc.GetTopCreators(context.Background(), 10)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if len(result) != 1 || result[0].UserID != 10 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestHandleRejectsUserWithoutManagementRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	reports := NewMocktopCreatorsReader(ctrl)
	access := NewMockreportAccessChecker(ctrl)
	uc := New(reports, access)

	access.EXPECT().HasManagementRole(gomock.Any(), int64(10)).Return(false, nil)

	_, err := uc.GetTopCreators(context.Background(), 10)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestHandleReturnsAccessCheckError(t *testing.T) {
	ctrl := gomock.NewController(t)
	reports := NewMocktopCreatorsReader(ctrl)
	access := NewMockreportAccessChecker(ctrl)
	uc := New(reports, access)
	accessErr := errors.New("access failed")

	access.EXPECT().HasManagementRole(gomock.Any(), int64(10)).Return(false, accessErr)

	_, err := uc.GetTopCreators(context.Background(), 10)
	if !errors.Is(err, accessErr) {
		t.Fatalf("expected access error, got %v", err)
	}
}

func TestHandleReturnsReportError(t *testing.T) {
	ctrl := gomock.NewController(t)
	reports := NewMocktopCreatorsReader(ctrl)
	access := NewMockreportAccessChecker(ctrl)
	uc := New(reports, access)
	reportErr := errors.New("report failed")

	access.EXPECT().HasManagementRole(gomock.Any(), int64(10)).Return(true, nil)
	reports.EXPECT().TopCreatorsByTeam(gomock.Any(), int64(10)).Return(nil, reportErr)

	_, err := uc.GetTopCreators(context.Background(), 10)
	if !errors.Is(err, reportErr) {
		t.Fatalf("expected report error, got %v", err)
	}
}
