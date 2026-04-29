package memory

import (
	"context"
	"errors"
	"testing"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/user/model"
)

func TestUserRepositorySaveAndGetByLogin(t *testing.T) {
	t.Parallel()

	repo := NewUserRepository()
	user := &model.User{ID: "user-1", Login: "alice", PasswordHash: []byte("hash")}

	if err := repo.Save(context.Background(), user); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := repo.GetByLogin(context.Background(), "alice")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.ID != user.ID || got.Login != user.Login || string(got.PasswordHash) != string(user.PasswordHash) {
		t.Fatalf("unexpected user: %#v", got)
	}
}

func TestUserRepositoryReturnsCopies(t *testing.T) {
	t.Parallel()

	repo := NewUserRepository()
	if err := repo.Save(context.Background(), &model.User{
		ID:           "user-1",
		Login:        "alice",
		PasswordHash: []byte("hash"),
	}); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	first, err := repo.GetByLogin(context.Background(), "alice")
	if err != nil {
		t.Fatalf("first get failed: %v", err)
	}
	first.PasswordHash[0] = 'X'

	second, err := repo.GetByLogin(context.Background(), "alice")
	if err != nil {
		t.Fatalf("second get failed: %v", err)
	}
	if string(second.PasswordHash) != "hash" {
		t.Fatalf("repository leaked mutable state: %q", second.PasswordHash)
	}
}

func TestUserRepositoryRejectsInvalidUser(t *testing.T) {
	t.Parallel()

	repo := NewUserRepository()

	err := repo.Save(context.Background(), &model.User{ID: "user-1", Login: "alice"})
	if !errors.Is(err, common.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUserRepositoryRejectsDuplicateLogin(t *testing.T) {
	t.Parallel()

	repo := NewUserRepository()
	user := &model.User{ID: "user-1", Login: "alice", PasswordHash: []byte("hash")}

	if err := repo.Save(context.Background(), user); err != nil {
		t.Fatalf("first save failed: %v", err)
	}
	err := repo.Save(context.Background(), &model.User{ID: "user-2", Login: "alice", PasswordHash: []byte("hash")})
	if !errors.Is(err, common.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestUserRepositoryGetByLoginErrors(t *testing.T) {
	t.Parallel()

	repo := NewUserRepository()

	if _, err := repo.GetByLogin(context.Background(), ""); !errors.Is(err, common.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if _, err := repo.GetByLogin(context.Background(), "missing"); !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
