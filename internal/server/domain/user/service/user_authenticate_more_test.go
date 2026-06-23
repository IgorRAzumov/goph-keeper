package service

import (
	"context"
	"errors"
	"testing"

	"goph-keeper/internal/server/domain/common"
	usermocks "goph-keeper/internal/server/domain/user/repository/mocks"

	"go.uber.org/mock/gomock"
)

func TestAuthenticateRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	svc := NewUserService(usermocks.NewMockUserStore(ctrl))
	_, err := svc.Authenticate(context.Background(), "", "p")
	if !errors.Is(err, common.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestAuthenticateNotImplemented(t *testing.T) {
	t.Parallel()

	_, err := NewUserService(nil).Authenticate(context.Background(), "a", "b")
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected not implemented, got %v", err)
	}
}
