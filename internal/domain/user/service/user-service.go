package service

import (
	"context"
	"errors"
	"strings"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/user/model"
	userrepo "goph-keeper/internal/domain/user/repository"

	"golang.org/x/crypto/bcrypt"
)

// UserService инкапсулирует доменные операции над пользователями.
// Сценарии application-слоя вызывают сервис, а сервис работает через порт репозитория.
type UserService struct {
	repo userrepo.UserRepository
}

// NewUserService создаёт сервис пользователей.
func NewUserService(repo userrepo.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register регистрирует нового пользователя по логину и паролю.
// Возвращает ErrConflict, если пользователь с таким логином уже существует.
func (s *UserService) Register(ctx context.Context, login, password string) (string, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return "", common.ErrInvalidInput
	}
	if s == nil || s.repo == nil {
		return "", common.ErrNotImplemented
	}

	if _, err := s.repo.GetByLogin(ctx, login); err == nil {
		return "", common.ErrConflict
	} else if !errors.Is(err, common.ErrNotFound) {
		return "", err
	}

	id, err := common.NewID()
	if err != nil {
		return "", err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	u := &model.User{ID: id, Login: login, PasswordHash: passwordHash}
	if err := s.repo.Save(ctx, u); err != nil {
		return "", err
	}
	return id, nil
}
