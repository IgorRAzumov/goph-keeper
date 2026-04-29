package service

import (
	"context"
	"errors"
	"testing"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/user/model"

	"golang.org/x/crypto/bcrypt"
)

type userRepoStub struct {
	getByLogin func(context.Context, string) (*model.User, error)
	save       func(context.Context, *model.User) error
}

func (r userRepoStub) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	return r.getByLogin(ctx, login)
}

func (r userRepoStub) Save(ctx context.Context, user *model.User) error {
	return r.save(ctx, user)
}

func TestRegisterHashesPasswordAndSavesUser(t *testing.T) {
	t.Parallel()

	var saved *model.User
	service := NewUserService(userRepoStub{
		getByLogin: func(context.Context, string) (*model.User, error) {
			return nil, common.ErrNotFound
		},
		save: func(_ context.Context, user *model.User) error {
			saved = user
			return nil
		},
	})

	id, err := service.Register(context.Background(), " alice ", "secret")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if id == "" {
		t.Fatal("expected user id")
	}
	if saved == nil {
		t.Fatal("expected user to be saved")
	}
	if saved.Login != "alice" {
		t.Fatalf("expected trimmed login, got %q", saved.Login)
	}
	if string(saved.PasswordHash) == "secret" {
		t.Fatal("password must not be stored in plain text")
	}
	if err := bcrypt.CompareHashAndPassword(saved.PasswordHash, []byte("secret")); err != nil {
		t.Fatalf("password hash mismatch: %v", err)
	}
}

func TestRegisterReturnsInfrastructureError(t *testing.T) {
	t.Parallel()

	service := NewUserService(userRepoStub{
		getByLogin: func(context.Context, string) (*model.User, error) {
			return nil, common.ErrNotImplemented
		},
		save: func(context.Context, *model.User) error {
			t.Fatal("save must not be called when lookup failed")
			return nil
		},
	})

	_, err := service.Register(context.Background(), "alice", "secret")
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}
