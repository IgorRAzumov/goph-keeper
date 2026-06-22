package register

import (
	"context"
	"strings"

	"goph-keeper/internal/server/domain/common"
	usersvc "goph-keeper/internal/server/domain/user/service"
)

// Usecase регистрирует нового пользователя. Персистентность внедряется через порты.
type Usecase struct {
	userService *usersvc.UserService
}

// NewUserUsecase создаёт сценарий регистрации пользователя.
func NewUserUsecase(userService *usersvc.UserService) *Usecase {
	return &Usecase{userService: userService}
}

func (usecase *Usecase) Execute(ctx context.Context, in Input) (Output, error) {
	if usecase == nil || usecase.userService == nil {
		return Output{}, common.ErrNotImplemented
	}
	if strings.TrimSpace(in.Login) == "" || in.Password == "" {
		return Output{}, common.ErrInvalidInput
	}

	userID, err := usecase.userService.Register(ctx, in.Login, in.Password, in.MasterSalt)
	if err != nil {
		return Output{}, err
	}
	return Output{UserID: userID}, nil
}
