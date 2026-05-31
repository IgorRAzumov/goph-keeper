package login

import (
	"context"
	"strings"
	"time"

	"goph-keeper/internal/domain/common"
	sessionsvc "goph-keeper/internal/domain/session/service"
	usersvc "goph-keeper/internal/domain/user/service"
	"goph-keeper/internal/security/jwt"

	"goph-keeper/internal/config"
)

// LoginUsecase выполняет аутентификацию и выдаёт пару токенов.
type LoginUsecase struct {
	userService    *usersvc.UserService
	sessionService *sessionsvc.SessionService
	jwt            *jwt.Provider
	accessTTL      time.Duration
	refreshTTL     time.Duration
}

// NewLoginUsecase создаёт сценарий логина.
func NewLoginUsecase(cfg config.Config, userService *usersvc.UserService, sessionService *sessionsvc.SessionService) *LoginUsecase {
	return &LoginUsecase{
		userService:    userService,
		sessionService: sessionService,
		jwt:            jwt.NewProvider(cfg.JWTSecret),
		accessTTL:      cfg.AccessTokenTTL,
		refreshTTL:     cfg.RefreshTokenTTL,
	}
}

func (usecase *LoginUsecase) Execute(ctx context.Context, input Input) (Output, error) {
	if strings.TrimSpace(input.Login) == "" || input.Password == "" {
		return Output{}, common.ErrInvalidInput
	}

	user, err := usecase.userService.Authenticate(ctx, input.Login, input.Password)
	if err != nil {
		return Output{}, err
	}

	now := time.Now().UTC()
	refreshToken, err := common.NewID()
	if err != nil {
		return Output{}, err
	}

	sessionID, err := usecase.sessionService.CreateSession(ctx, user.ID, refreshToken, usecase.refreshTTL, now)
	if err != nil {
		return Output{}, err
	}

	access, err := usecase.jwt.IssueAccessToken(user.ID, sessionID, usecase.accessTTL, now)
	if err != nil {
		return Output{}, err
	}

	return Output{AccessToken: access, RefreshToken: refreshToken}, nil
}
