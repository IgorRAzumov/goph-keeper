package service

import (
	"context"
	"errors"
	"testing"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/user/model"
	usermocks "goph-keeper/internal/domain/user/repository/mocks"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterHashesPasswordAndSavesUser(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	var saved *model.User
	repo := usermocks.NewMockUserRepository(ctrl)
	repo.EXPECT().GetByLogin(gomock.Any(), "alice").Return(nil, common.ErrNotFound)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, user *model.User) error {
			saved = user
			return nil
		})

	service := NewUserService(repo)

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

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := usermocks.NewMockUserRepository(ctrl)
	repo.EXPECT().GetByLogin(gomock.Any(), "alice").Return(nil, common.ErrNotImplemented)

	service := NewUserService(repo)

	_, err := service.Register(context.Background(), "alice", "secret")
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	service := NewUserService(usermocks.NewMockUserRepository(ctrl))

	tests := []struct {
		name     string
		login    string
		password string
	}{
		{name: "empty login", password: "secret"},
		{name: "blank login", login: "  ", password: "secret"},
		{name: "empty password", login: "alice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Register(context.Background(), tt.login, tt.password)
			if !errors.Is(err, common.ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestRegisterReturnsNotImplementedWithoutRepository(t *testing.T) {
	t.Parallel()

	service := NewUserService(nil)

	_, err := service.Register(context.Background(), "alice", "secret")
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

func TestRegisterReturnsConflictWhenLoginExists(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := usermocks.NewMockUserRepository(ctrl)
	repo.EXPECT().
		GetByLogin(gomock.Any(), "alice").
		Return(&model.User{ID: "user-1", Login: "alice"}, nil)

	service := NewUserService(repo)

	_, err := service.Register(context.Background(), "alice", "secret")
	if !errors.Is(err, common.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestRegisterReturnsSaveError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	saveErr := errors.New("save failed")
	repo := usermocks.NewMockUserRepository(ctrl)
	repo.EXPECT().GetByLogin(gomock.Any(), "alice").Return(nil, common.ErrNotFound)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(saveErr)

	service := NewUserService(repo)

	_, err := service.Register(context.Background(), "alice", "secret")
	if !errors.Is(err, saveErr) {
		t.Fatalf("expected save error, got %v", err)
	}
}
