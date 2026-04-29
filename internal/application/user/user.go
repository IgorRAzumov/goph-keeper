package user

import (
	"context"
	"strings"

	"goph-keeper/internal/domain/common"
	usersvc "goph-keeper/internal/domain/user/service"
)

// Usecase регистрирует нового пользователя. Персистентность внедряется через порты.
type Usecase struct {
	userService *usersvc.UserService
}

// NewUserUsecase создаёт сценарий регистрации пользователя.
func NewUserUsecase(userService *usersvc.UserService) *Usecase {
	return &Usecase{userService: userService}
}

// Execute выполняет сценарий. Полная реализация появится после добавления аутентификации и хэширования паролей.
func (usecase *Usecase) Execute(ctx context.Context, in RegisterUserInput) (RegisterUserOutput, error) {
	if strings.TrimSpace(in.Login) == "" || in.Password == "" {
		return RegisterUserOutput{}, common.ErrInvalidInput
	}
	if usecase == nil || usecase.userService == nil {
		return RegisterUserOutput{}, common.ErrNotImplemented
	}

	userID, err := usecase.userService.Register(ctx, in.Login, in.Password)
	if err != nil {
		return RegisterUserOutput{}, err
	}
	return RegisterUserOutput{UserID: userID}, nil
}
