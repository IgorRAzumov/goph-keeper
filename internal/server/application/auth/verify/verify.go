// Package verify содержит сценарий проверки access-токена и активной сессии.
//
// Это политика аутентификации application-слоя: HTTP-доставка лишь извлекает
// токен из заголовка и вызывает Execute, не зная деталей JWT и хранилища сессий.
package verify

import (
	"context"
	"errors"
	"strings"
	"time"

	apptoken "goph-keeper/internal/server/application/auth/token"
	"goph-keeper/internal/server/domain/common"
	sessionrepo "goph-keeper/internal/server/domain/session/repository"
)

// Usecase проверяет access-токен и связанную с ним серверную сессию.
type Usecase struct {
	jwt      apptoken.Provider
	sessions sessionrepo.SessionRepository
}

// NewUsecase создаёт сценарий проверки доступа.
func NewUsecase(jwtProvider apptoken.Provider, sessions sessionrepo.SessionRepository) *Usecase {
	return &Usecase{jwt: jwtProvider, sessions: sessions}
}

// Output — идентификаторы владельца и сессии для аутентифицированного запроса.
type Output struct {
	UserID    string
	SessionID string
}

// Execute проверяет токен и сессию на момент now.
// Возвращает common.ErrUnauthorized при невалидном токене/сессии и common.ErrNotImplemented, если сценарий не сконфигурирован.
func (usecase *Usecase) Execute(ctx context.Context, accessToken string, now time.Time) (Output, error) {
	if usecase == nil || usecase.jwt == nil || usecase.sessions == nil {
		return Output{}, common.ErrNotImplemented
	}

	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return Output{}, common.ErrUnauthorized
	}

	claims, err := usecase.jwt.ParseAccessToken(accessToken, now)
	if err != nil {
		return Output{}, common.ErrUnauthorized
	}

	session, err := usecase.sessions.Get(ctx, claims.Sid)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return Output{}, common.ErrUnauthorized
		}
		return Output{}, err
	}

	if session.UserID != claims.Sub || !now.Before(session.RefreshExpiresAt) {
		return Output{}, common.ErrUnauthorized
	}

	return Output{UserID: claims.Sub, SessionID: claims.Sid}, nil
}
