package register

import (
	"context"
	"errors"
	"testing"

	"goph-keeper/internal/server/domain/common"
	usermocks "goph-keeper/internal/server/domain/user/repository/mocks"
	usersvc "goph-keeper/internal/server/domain/user/service"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRegisterUserUseCaseExecute(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	userRepository := usermocks.NewMockUserRepository(controller)
	userRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	userService := usersvc.NewUserService(userRepository)
	usecase := NewUserUsecase(userService)

	out, err := usecase.Execute(context.Background(), Input{
		Login:    "alice",
		Password: "secret",
	})
	require.NoError(t, err)
	require.NotEmpty(t, out.UserID)
}

func TestRegisterUserUseCaseRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	userRepository := usermocks.NewMockUserRepository(controller)
	usecase := NewUserUsecase(usersvc.NewUserService(userRepository))

	tests := []struct {
		name string
		in   Input
	}{
		{name: "empty login", in: Input{Password: "secret"}},
		{name: "blank login", in: Input{Login: "  ", Password: "secret"}},
		{name: "empty password", in: Input{Login: "alice"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := usecase.Execute(context.Background(), tt.in)
			require.ErrorIs(t, err, common.ErrInvalidInput)
		})
	}
}

func TestRegisterUserUseCaseReturnsNotImplementedWithoutService(t *testing.T) {
	t.Parallel()

	usecase := NewUserUsecase(nil)

	_, err := usecase.Execute(context.Background(), Input{Login: "alice", Password: "secret"})
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}
