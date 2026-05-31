package authtest

import (
	"time"

	"goph-keeper/internal/config"
)

// UnitTestConfig возвращает конфиг для unit-тестов auth-сценариев (JWT и TTL).
func UnitTestConfig() config.Config {
	return config.Config{
		JWTSecret:       "unit-test-jwt-secret-not-for-production",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 720 * time.Hour,
	}
}
