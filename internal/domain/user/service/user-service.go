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
type UserService struct {
	userRepository userrepo.UserRepository
}

// NewUserService создаёт сервис пользователей.
func NewUserService(userRepository userrepo.UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

// Register регистрирует нового пользователя по логину и паролю.
// Возвращает ErrConflict, если пользователь с таким логином уже существует.
func (service *UserService) Register(ctx context.Context, login, password string) (string, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return "", common.ErrInvalidInput
	}
	if service == nil || service.userRepository == nil {
		return "", common.ErrNotImplemented
	}

	if _, err := service.userRepository.GetByLogin(ctx, login); err == nil {
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

	user := &model.User{ID: id, Login: login, PasswordHash: passwordHash}
	if err := service.userRepository.Save(ctx, user); err != nil {
		return "", err
	}
	return id, nil
}
