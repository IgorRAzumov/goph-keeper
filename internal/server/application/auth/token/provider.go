// Package token содержит абстракции для работы с access-токенами в application-слое.
package token

import (
	"time"

	"goph-keeper/internal/server/security/jwt"
)

// Provider описывает операции выпуска и валидации access-токена.
// Concrete-реализация внедряется из composition root.
type Provider interface {
	IssueAccessToken(userID, sessionID string, ttl time.Duration, now time.Time) (string, error)
	ParseAccessToken(accessToken string, now time.Time) (jwt.Claims, error)
}
