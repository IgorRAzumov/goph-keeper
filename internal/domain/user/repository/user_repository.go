package repository

import (
	"context"

	"goph-keeper/internal/domain/user/model"
)

// UserRepository сохраняет и загружает пользователя.
type UserRepository interface {
	// Save сохраняет нового пользователя или обновляет существующего.
	Save(ctx context.Context, user *model.User) error
	// GetByLogin возвращает пользователя по логину или ErrNotFound.
	GetByLogin(ctx context.Context, login string) (*model.User, error)
}
