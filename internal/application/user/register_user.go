package user

import (
	"context"
	"strings"

	"goph-keeper/internal/domain/common"
	usersvc "goph-keeper/internal/domain/user/service"
)

// RegisterUserUseCase регистрирует нового пользователя. Персистентность внедряется через порты.
type RegisterUserUseCase struct {
	userService *usersvc.UserService
}

// NewRegisterUserUseCase создаёт сценарий регистрации пользователя.
func NewRegisterUserUseCase(userService *usersvc.UserService) *RegisterUserUseCase {
	return &RegisterUserUseCase{userService: userService}
}

// Execute выполняет сценарий. Полная реализация появится после добавления аутентификации и хэширования паролей.
func (uc *RegisterUserUseCase) Execute(ctx context.Context, in RegisterUserInput) (RegisterUserOutput, error) {
	if strings.TrimSpace(in.Login) == "" || in.Password == "" {
		return RegisterUserOutput{}, common.ErrInvalidInput
	}
	if uc == nil || uc.userService == nil {
		return RegisterUserOutput{}, common.ErrNotImplemented
	}

	userID, err := uc.userService.Register(ctx, in.Login, in.Password)
	if err != nil {
		return RegisterUserOutput{}, err
	}
	return RegisterUserOutput{UserID: userID}, nil
}
