package team_delete

import (
	"context"
	"errors"
	"testing"

	"task-service/internal/domain"
	"task-service/internal/domain/models"

	"go.uber.org/mock/gomock"
)

type teamDeleteStore struct {
	*MockteamOwnerReader
	*MockteamDeleter
}

func TestHandleDeletesTeamForOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerReader := NewMockteamOwnerReader(ctrl)
	deleter := NewMockteamDeleter(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(teamDeleteStore{MockteamOwnerReader: ownerReader, MockteamDeleter: deleter}, cache)

	ownerReader.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleOwner, nil)
	deleter.EXPECT().Delete(gomock.Any(), int64(1)).Return(nil)
	cache.EXPECT().DeleteTeamTasks(gomock.Any(), int64(1)).Return(nil)

	if err := uc.Delete(context.Background(), 10, 1); err != nil {
		t.Fatalf("handle failed: %v", err)
	}
}

func TestHandleRejectsAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerReader := NewMockteamOwnerReader(ctrl)
	deleter := NewMockteamDeleter(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(teamDeleteStore{MockteamOwnerReader: ownerReader, MockteamDeleter: deleter}, cache)

	ownerReader.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleAdmin, nil)

	err := uc.Delete(context.Background(), 10, 1)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestHandleReturnsExternalErrorWhenCacheInvalidationFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerReader := NewMockteamOwnerReader(ctrl)
	deleter := NewMockteamDeleter(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(teamDeleteStore{MockteamOwnerReader: ownerReader, MockteamDeleter: deleter}, cache)

	ownerReader.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleOwner, nil)
	deleter.EXPECT().Delete(gomock.Any(), int64(1)).Return(nil)
	cache.EXPECT().DeleteTeamTasks(gomock.Any(), int64(1)).Return(errors.New("redis unavailable"))

	err := uc.Delete(context.Background(), 10, 1)
	if !errors.Is(err, domain.ErrExternal) {
		t.Fatalf("expected external error, got %v", err)
	}
}

func TestHandleReturnsRoleLookupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerReader := NewMockteamOwnerReader(ctrl)
	deleter := NewMockteamDeleter(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(teamDeleteStore{MockteamOwnerReader: ownerReader, MockteamDeleter: deleter}, cache)

	ownerReader.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.Role(""), domain.ErrNotFound)

	err := uc.Delete(context.Background(), 10, 1)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestHandleReturnsDeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerReader := NewMockteamOwnerReader(ctrl)
	deleter := NewMockteamDeleter(ctrl)
	cache := NewMocktaskCacheInvalidator(ctrl)
	uc := New(teamDeleteStore{MockteamOwnerReader: ownerReader, MockteamDeleter: deleter}, cache)

	ownerReader.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(10)).Return(models.RoleOwner, nil)
	deleter.EXPECT().Delete(gomock.Any(), int64(1)).Return(domain.ErrNotFound)

	err := uc.Delete(context.Background(), 10, 1)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
