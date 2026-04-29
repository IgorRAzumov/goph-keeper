package user

import (
	"context"
	"errors"
	"testing"

	"goph-keeper/internal/domain/common"
	usersvc "goph-keeper/internal/domain/user/service"
	"goph-keeper/internal/infrastructure/memory"
)

func TestRegisterUserUseCaseExecute(t *testing.T) {
	t.Parallel()

	repo := memory.NewUserRepository()
	service := usersvc.NewUserService(repo)
	usecase := NewUserUsecase(service)

	out, err := usecase.Execute(context.Background(), RegisterUserInput{
		Login:    "alice",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if out.UserID == "" {
		t.Fatal("expected user id")
	}
}

func TestRegisterUserUseCaseRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	usecase := NewUserUsecase(usersvc.NewUserService(memory.NewUserRepository()))

	tests := []struct {
		name string
		in   RegisterUserInput
	}{
		{name: "empty login", in: RegisterUserInput{Password: "secret"}},
		{name: "blank login", in: RegisterUserInput{Login: "  ", Password: "secret"}},
		{name: "empty password", in: RegisterUserInput{Login: "alice"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := usecase.Execute(context.Background(), tt.in)
			if !errors.Is(err, common.ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestRegisterUserUseCaseReturnsNotImplementedWithoutService(t *testing.T) {
	t.Parallel()

	usecase := NewUserUsecase(nil)

	_, err := usecase.Execute(context.Background(), RegisterUserInput{Login: "alice", Password: "secret"})
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}
