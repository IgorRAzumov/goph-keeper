// Пакет memory содержит in-memory адаптеры для локальной разработки и тестов.
package memory

import (
	"context"
	"sync"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/user/model"
)

// UserRepository хранит пользователей в памяти процесса.
type UserRepository struct {
	mu      sync.RWMutex
	byLogin map[string]*model.User
}

// NewUserRepository создаёт пустой in-memory репозиторий пользователей.
func NewUserRepository() *UserRepository {
	return &UserRepository{byLogin: make(map[string]*model.User)}
}

// Save сохраняет нового пользователя.
func (repository *UserRepository) Save(ctx context.Context, user *model.User) error {
	if user == nil || user.ID == "" || user.Login == "" || len(user.PasswordHash) == 0 {
		return common.ErrInvalidInput
	}

	repository.mu.Lock()
	defer repository.mu.Unlock()

	if _, ok := repository.byLogin[user.Login]; ok {
		return common.ErrConflict
	}
	repository.byLogin[user.Login] = cloneUser(user)
	return nil
}

// GetByLogin возвращает пользователя по логину.
func (repository *UserRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	if login == "" {
		return nil, common.ErrInvalidInput
	}

	repository.mu.RLock()
	defer repository.mu.RUnlock()

	user, ok := repository.byLogin[login]
	if !ok {
		return nil, common.ErrNotFound
	}
	return cloneUser(user), nil
}

func cloneUser(user *model.User) *model.User {
	if user == nil {
		return nil
	}

	clone := *user
	clone.PasswordHash = append([]byte(nil), user.PasswordHash...)
	return &clone
}
