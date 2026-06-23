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

func TestAuthenticateSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := usermocks.NewMockUserStore(ctrl)
	repo.EXPECT().GetByLogin(gomock.Any(), "alice").Return(&model.User{
		ID: "u1", Login: "alice", PasswordHash: hash,
	}, nil)

	user, err := NewUserService(repo).Authenticate(context.Background(), "alice", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != "u1" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestAuthenticateWrongPassword(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	repo := usermocks.NewMockUserStore(ctrl)
	repo.EXPECT().GetByLogin(gomock.Any(), "alice").Return(&model.User{
		ID: "u1", Login: "alice", PasswordHash: hash,
	}, nil)

	_, err := NewUserService(repo).Authenticate(context.Background(), "alice", "wrong")
	if !errors.Is(err, common.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestAuthenticateUserNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := usermocks.NewMockUserStore(ctrl)
	repo.EXPECT().GetByLogin(gomock.Any(), "nobody").Return(nil, common.ErrNotFound)

	_, err := NewUserService(repo).Authenticate(context.Background(), "nobody", "x")
	if !errors.Is(err, common.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}
