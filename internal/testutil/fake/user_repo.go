package fake

import (
	"context"
	"sync"

	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/user/model"
)

// UserRepo — in-memory UserStore для тестов.
type UserRepo struct {
	mu      sync.Mutex
	byID    map[string]*model.User
	byLogin map[string]string
}

// NewUserRepo создаёт пустой in-memory репозиторий пользователей.
func NewUserRepo() *UserRepo {
	return &UserRepo{
		byID:    map[string]*model.User{},
		byLogin: map[string]string{},
	}
}

func (r *UserRepo) Save(_ context.Context, user *model.User) error {
	if user == nil || user.ID == "" || user.Login == "" {
		return common.ErrInvalidInput
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.byLogin[user.Login]; ok && existing != user.ID {
		return common.ErrConflict
	}
	copyUser := *user
	r.byID[user.ID] = &copyUser
	r.byLogin[user.Login] = user.ID
	return nil
}

func (r *UserRepo) GetByLogin(_ context.Context, login string) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byLogin[login]
	if !ok {
		return nil, common.ErrNotFound
	}
	u := *r.byID[id]
	return &u, nil
}

func (r *UserRepo) SetMasterSalt(_ context.Context, userID, masterSalt string) error {
	if userID == "" || masterSalt == "" {
		return common.ErrInvalidInput
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.byID[userID]
	if !ok {
		return common.ErrNotFound
	}
	user.MasterSalt = masterSalt
	return nil
}
