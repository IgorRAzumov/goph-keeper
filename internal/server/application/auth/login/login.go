package login

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	apptoken "goph-keeper/internal/server/application/auth/token"
	"goph-keeper/internal/server/domain/common"
	sessionsvc "goph-keeper/internal/server/domain/session/service"
	usersvc "goph-keeper/internal/server/domain/user/service"
)

// LoginUsecase выполняет аутентификацию и выдаёт пару токенов.
type LoginUsecase struct {
	userService    *usersvc.UserService
	sessionService *sessionsvc.SessionService
	jwt            apptoken.Provider
	accessTTL      time.Duration
	refreshTTL     time.Duration
}

// NewLoginUsecase создаёт сценарий логина.
// jwtProvider, accessTTL и refreshTTL внедряются извне, чтобы слой сценариев не зависел от формата конфигурации.
func NewLoginUsecase(
	userService *usersvc.UserService,
	sessionService *sessionsvc.SessionService,
	jwtProvider apptoken.Provider,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *LoginUsecase {
	return &LoginUsecase{
		userService:    userService,
		sessionService: sessionService,
		jwt:            jwtProvider,
		accessTTL:      accessTTL,
		refreshTTL:     refreshTTL,
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
	masterSalt := strings.TrimSpace(user.MasterSalt)
	if masterSalt == "" {
		masterSalt, err = generateMasterSalt()
		if err != nil {
			return Output{}, err
		}
		if err := usecase.userService.SetMasterSalt(ctx, user.ID, masterSalt); err != nil {
			return Output{}, err
		}
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

	return Output{AccessToken: access, RefreshToken: refreshToken, MasterSalt: masterSalt}, nil
}

func generateMasterSalt() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}
