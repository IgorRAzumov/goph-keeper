package postgres

import (
	"context"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/user/model"
)

// UserRepository — PostgreSQL-адаптер для domain.UserRepository (пока без подключения к БД).
type UserRepository struct{}

// NewUserRepository создаёт репозиторий, который позже будет работать через *sql.DB (когда появится подключение).
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Save сохранение
func (repository *UserRepository) Save(ctx context.Context, user *model.User) error {
	_, _ = ctx, user
	return common.ErrNotImplemented
}

// GetByLogin получение по логину
func (repository *UserRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	_, _ = ctx, login
	return nil, common.ErrNotImplemented
}
