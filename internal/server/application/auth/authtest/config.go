// Package authtest предоставляет общие параметры JWT и TTL для unit-тестов auth-сценариев,
// не завязываясь на формат серверной конфигурации.
package authtest

import (
	"time"

	"goph-keeper/internal/server/security/jwt"
)

const (
	// JWTSecret — секрет подписи токенов в тестах (не для продакшена).
	JWTSecret = "unit-test-jwt-secret-not-for-production"
	// AccessTTL — срок жизни access-токена в тестах.
	AccessTTL = 15 * time.Minute
	// RefreshTTL — срок жизни refresh-токена в тестах.
	RefreshTTL = 720 * time.Hour
)

// Provider возвращает JWT-провайдер с тестовым секретом.
func Provider() *jwt.Provider {
	return jwt.NewProvider(JWTSecret)
}
