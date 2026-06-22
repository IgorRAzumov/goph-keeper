package model

import (
	"crypto/sha256"
	"time"
)

// Session — серверная сессия пользователя (refresh-ротация).
type Session struct {
	// ID — идентификатор сессии (непрозрачная строка).
	ID string
	// UserID — владелец сессии.
	UserID string
	// RefreshTokenHash — SHA-256 хэш случайного refresh-токена (сырой токен нигде не храним).
	RefreshTokenHash []byte
	// RefreshExpiresAt — время истечения refresh-сессии.
	RefreshExpiresAt time.Time
}

// HashRefreshToken возвращает индексируемый SHA-256 hash случайного refresh-токена.
func HashRefreshToken(refreshToken string) []byte {
	sum := sha256.Sum256([]byte(refreshToken))
	return sum[:]
}
