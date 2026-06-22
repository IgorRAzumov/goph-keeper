package service

import (
	"context"
	"errors"
	"strings"

	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/user/model"
	userrepo "goph-keeper/internal/server/domain/user/repository"

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
func (service *UserService) Register(ctx context.Context, login, password, masterSalt string) (string, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return "", common.ErrInvalidInput
	}
	if service == nil || service.userRepository == nil {
		return "", common.ErrNotImplemented
	}

	id, err := common.NewID()
	if err != nil {
		return "", err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &model.User{ID: id, Login: login, PasswordHash: passwordHash, MasterSalt: strings.TrimSpace(masterSalt)}
	if err := service.userRepository.Save(ctx, user); err != nil {
		return "", err
	}
	return id, nil
}

// Authenticate проверяет логин/пароль и возвращает пользователя.
func (service *UserService) Authenticate(ctx context.Context, login, password string) (*model.User, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return nil, common.ErrInvalidInput
	}
	if service == nil || service.userRepository == nil {
		return nil, common.ErrNotImplemented
	}

	user, err := service.userRepository.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil, common.ErrUnauthorized
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return nil, common.ErrUnauthorized
	}
	return user, nil
}

// SetMasterSalt сохраняет master_salt для пользователя.
func (service *UserService) SetMasterSalt(ctx context.Context, userID, masterSalt string) error {
	userID = strings.TrimSpace(userID)
	masterSalt = strings.TrimSpace(masterSalt)
	if userID == "" || masterSalt == "" {
		return common.ErrInvalidInput
	}
	if service == nil || service.userRepository == nil {
		return common.ErrNotImplemented
	}
	return service.userRepository.SetMasterSalt(ctx, userID, masterSalt)
}
