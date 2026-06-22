package service

import (
	"context"
	"time"

	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/session/model"
	sessionrepo "goph-keeper/internal/server/domain/session/repository"
)

// SessionService управляет refresh-сессиями.
type SessionService struct {
	sessionRepository sessionrepo.SessionRepository
}

// NewSessionService создаёт сервис сессий.
func NewSessionService(sessionRepository sessionrepo.SessionRepository) *SessionService {
	return &SessionService{sessionRepository: sessionRepository}
}

// CreateSession создаёт новую сессию с хэшем refresh-токена.
func (service *SessionService) CreateSession(ctx context.Context, userID, refreshToken string, refreshTTL time.Duration, now time.Time) (sessionID string, err error) {
	if service == nil || service.sessionRepository == nil {
		return "", common.ErrNotImplemented
	}
	if userID == "" || refreshToken == "" || refreshTTL <= 0 {
		return "", common.ErrInvalidInput
	}

	sessionID, err = common.NewID()
	if err != nil {
		return "", err
	}

	session := &model.Session{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: model.HashRefreshToken(refreshToken),
		RefreshExpiresAt: now.Add(refreshTTL),
	}
	if err := service.sessionRepository.Save(ctx, session); err != nil {
		return "", err
	}
	return sessionID, nil
}

// RotateSession атомарно удаляет старую сессию и создаёт новую (refresh rotation).
func (service *SessionService) RotateSession(ctx context.Context, oldSessionID, userID, refreshToken string, refreshTTL time.Duration, now time.Time) (newSessionID string, err error) {
	if service == nil || service.sessionRepository == nil {
		return "", common.ErrNotImplemented
	}
	if oldSessionID == "" || userID == "" || refreshToken == "" || refreshTTL <= 0 {
		return "", common.ErrInvalidInput
	}

	newSessionID, err = common.NewID()
	if err != nil {
		return "", err
	}

	session := &model.Session{
		ID:               newSessionID,
		UserID:           userID,
		RefreshTokenHash: model.HashRefreshToken(refreshToken),
		RefreshExpiresAt: now.Add(refreshTTL),
	}
	if err := service.sessionRepository.Rotate(ctx, oldSessionID, session); err != nil {
		return "", err
	}
	return newSessionID, nil
}
