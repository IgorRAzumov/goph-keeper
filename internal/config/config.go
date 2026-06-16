package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// DefaultPostgresDSN — локальный Postgres на стандартном порту (127.0.0.1:5432).
// Переопределите GOPHKEEPER_POSTGRES_DSN, если другой пользователь/пароль/БД.
const DefaultPostgresDSN = "postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable"

// DefaultJWTSecret — секрет по умолчанию для локальной разработки.
const DefaultJWTSecret = "local-dev-jwt-secret"

// Config содержит настройки процесса для API-сервера и будущих адаптеров.
type Config struct {
	// HTTPAddr — TCP-адрес для HTTP-листенера (например "127.0.0.1:8080").
	HTTPAddr string
	// JWTSecret — секрет для подписи access JWT (HS256). Для dev можно оставить дефолт, для prod обязателен свой.
	JWTSecret string
	// AccessTokenTTL — TTL access-токена (например "15m").
	AccessTokenTTL time.Duration
	// RefreshTokenTTL — TTL refresh-токена (например "720h").
	RefreshTokenTTL time.Duration

	// PostgresDSN — строка подключения Postgres (pgx), например postgres://user:pass@localhost:5432/gophkeeper?sslmode=disable
	PostgresDSN string
	// PostgresMaxOpenConns — лимит открытых соединений пула (только для postgres).
	PostgresMaxOpenConns int
}

// Load читает конфигурацию из переменных окружения, подставляя значения по умолчанию при отсутствии.
func Load() (Config, error) {
	accessTTL := mustParseDuration(getEnv("GOPHKEEPER_ACCESS_TTL", "15m"))
	refreshTTL := mustParseDuration(getEnv("GOPHKEEPER_REFRESH_TTL", "720h"))
	maxOpen := mustParseInt(getEnv("GOPHKEEPER_POSTGRES_MAX_OPEN_CONNS", "10"))

	configuration := Config{
		HTTPAddr:             getEnv("GOPHKEEPER_ADDR", "127.0.0.1:8080"),
		JWTSecret:            getEnv("GOPHKEEPER_JWT_SECRET", DefaultJWTSecret),
		AccessTokenTTL:       accessTTL,
		RefreshTokenTTL:      refreshTTL,
		PostgresDSN:          getEnv("GOPHKEEPER_POSTGRES_DSN", DefaultPostgresDSN),
		PostgresMaxOpenConns: maxOpen,
	}

	if err := validateConfig(configuration); err != nil {
		return Config{}, err
	}

	return configuration, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func mustParseDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		// Дефолты в Load() уже валидные; если пользователь ошибся в env — лучше упасть явно.
		panic("invalid duration: " + value)
	}
	return duration
}

func mustParseInt(value string) int {
	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		panic("invalid parse Int: " + value)
	}
	return number
}
func validateConfig(configuration Config) error {
	if strings.TrimSpace(configuration.PostgresDSN) == "" {
		return fmt.Errorf("config: postgres requires GOPHKEEPER_POSTGRES_DSN")
	}
	return nil
}
