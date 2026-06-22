package repository

import (
	"context"

	"goph-keeper/internal/server/domain/user/model"
)

// UserRepository сохраняет и загружает пользователя.
type UserRepository interface {
	// Save сохраняет нового пользователя или обновляет существующего.
	Save(ctx context.Context, user *model.User) error
	// GetByLogin возвращает пользователя по логину или ErrNotFound.
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	// SetMasterSalt сохраняет master_salt для пользователя по id.
	SetMasterSalt(ctx context.Context, userID, masterSalt string) error
}
