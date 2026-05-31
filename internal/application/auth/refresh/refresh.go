package refresh

import (
	"context"
	"errors"
	"strings"
	"time"

	"goph-keeper/internal/config"
	"goph-keeper/internal/domain/common"
	sessionrepo "goph-keeper/internal/domain/session/repository"
	sessionsvc "goph-keeper/internal/domain/session/service"
	"goph-keeper/internal/security/jwt"
)

// RefreshUsecase обновляет access/refresh токены по refresh-токену (rotation).
type RefreshUsecase struct {
	sessionRepository sessionrepo.SessionRepository
	sessionService    *sessionsvc.SessionService
	jwt               *jwt.Provider
	accessTTL         time.Duration
	refreshTTL        time.Duration
}

// NewRefreshUsecase создаёт сценарий refresh.
func NewRefreshUsecase(cfg config.Config, sessionRepository sessionrepo.SessionRepository, sessionService *sessionsvc.SessionService) *RefreshUsecase {
	return &RefreshUsecase{
		sessionRepository: sessionRepository,
		sessionService:    sessionService,
		jwt:               jwt.NewProvider(cfg.JWTSecret),
		accessTTL:         cfg.AccessTokenTTL,
		refreshTTL:        cfg.RefreshTokenTTL,
	}
}

func (usecase *RefreshUsecase) Execute(ctx context.Context, in Input) (Output, error) {
	if strings.TrimSpace(in.RefreshToken) == "" {
		return Output{}, common.ErrInvalidInput
	}
	if usecase == nil || usecase.sessionRepository == nil || usecase.sessionService == nil || usecase.jwt == nil {
		return Output{}, common.ErrNotImplemented
	}

	now := time.Now().UTC()
	session, err := usecase.sessionRepository.FindActiveByRefreshToken(ctx, in.RefreshToken, now)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return Output{}, common.ErrUnauthorized
		}
		return Output{}, err
	}

	newRefresh, err := common.NewID()
	if err != nil {
		return Output{}, err
	}

	newSessionID, err := usecase.sessionService.RotateSession(ctx, session.ID, session.UserID, newRefresh, usecase.refreshTTL, now)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return Output{}, common.ErrUnauthorized
		}
		return Output{}, err
	}

	access, err := usecase.jwt.IssueAccessToken(session.UserID, newSessionID, usecase.accessTTL, now)
	if err != nil {
		return Output{}, err
	}

	return Output{AccessToken: access, RefreshToken: newRefresh}, nil
}
