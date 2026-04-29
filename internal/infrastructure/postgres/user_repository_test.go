package postgres

import (
	"context"
	"errors"
	"testing"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/user/model"
)

func TestUserRepositoryIsExplicitlyNotImplemented(t *testing.T) {
	t.Parallel()

	repo := NewUserRepository()

	if err := repo.Save(context.Background(), &model.User{}); !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented from Save, got %v", err)
	}
	if _, err := repo.GetByLogin(context.Background(), "alice"); !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented from GetByLogin, got %v", err)
	}
}
