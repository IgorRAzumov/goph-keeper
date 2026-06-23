package service

import (
	"context"
	"errors"
	"testing"

	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/user/model"
	usermocks "goph-keeper/internal/server/domain/user/repository/mocks"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterHashesPasswordAndSavesUser(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	var saved *model.User
	repo := usermocks.NewMockUserStore(ctrl)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, user *model.User) error {
			saved = user
			return nil
		})

	service := NewUserService(repo)

	id, err := service.Register(context.Background(), " alice ", "secret", "salt")
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

func TestRegisterReturnsSaveError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	saveErr := errors.New("save failed")
	repo := usermocks.NewMockUserStore(ctrl)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(saveErr)

	service := NewUserService(repo)

	_, err := service.Register(context.Background(), "alice", "secret", "salt")
	if !errors.Is(err, saveErr) {
		t.Fatalf("expected save error, got %v", err)
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	service := NewUserService(usermocks.NewMockUserStore(ctrl))

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
			_, err := service.Register(context.Background(), tt.login, tt.password, "salt")
			if !errors.Is(err, common.ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestRegisterReturnsNotImplementedWithoutRepository(t *testing.T) {
	t.Parallel()

	service := NewUserService(nil)

	_, err := service.Register(context.Background(), "alice", "secret", "salt")
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

func TestRegisterReturnsConflict(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := usermocks.NewMockUserStore(ctrl)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(common.ErrConflict)

	service := NewUserService(repo)

	_, err := service.Register(context.Background(), "alice", "secret", "salt")
	if !errors.Is(err, common.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestSetMasterSalt(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := usermocks.NewMockUserStore(ctrl)
	repo.EXPECT().SetMasterSalt(gomock.Any(), "u1", "salt").Return(nil)
	service := NewUserService(repo)

	if err := service.SetMasterSalt(context.Background(), "u1", "salt"); err != nil {
		t.Fatal(err)
	}
}

func TestSetMasterSaltRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	service := NewUserService(nil)
	if err := service.SetMasterSalt(context.Background(), "", "salt"); !errors.Is(err, common.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
