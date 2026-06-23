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

// Load собирает конфигурацию: дефолты → переменные окружения → опции → валидация.
func Load(opts ...Option) (Config, error) {
	cfg := defaults()

	if err := applyEnv(&cfg); err != nil {
		return Config{}, err
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func defaults() Config {
	return Config{
		HTTPAddr:             "127.0.0.1:8080",
		JWTSecret:            DefaultJWTSecret,
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      720 * time.Hour,
		PostgresDSN:          DefaultPostgresDSN,
		PostgresMaxOpenConns: 10,
	}
}

func applyEnv(cfg *Config) error {
	cfg.HTTPAddr = stringEnv("GOPHKEEPER_ADDR", cfg.HTTPAddr)
	cfg.JWTSecret = stringEnv("GOPHKEEPER_JWT_SECRET", cfg.JWTSecret)
	cfg.PostgresDSN = stringEnv("GOPHKEEPER_POSTGRES_DSN", cfg.PostgresDSN)

	if err := parseEnv(&cfg.AccessTokenTTL, "GOPHKEEPER_ACCESS_TTL", time.ParseDuration); err != nil {
		return err
	}
	if err := parseEnv(&cfg.RefreshTokenTTL, "GOPHKEEPER_REFRESH_TTL", time.ParseDuration); err != nil {
		return err
	}
	if err := parseEnv(&cfg.PostgresMaxOpenConns, "GOPHKEEPER_POSTGRES_MAX_OPEN_CONNS", parsePositiveInt); err != nil {
		return err
	}
	return nil
}

func (config Config) validate() error {
	if strings.TrimSpace(config.HTTPAddr) == "" {
		return fmt.Errorf("config: GOPHKEEPER_ADDR must not be empty")
	}
	if strings.TrimSpace(config.PostgresDSN) == "" {
		return fmt.Errorf("config: postgres requires GOPHKEEPER_POSTGRES_DSN")
	}
	if config.AccessTokenTTL <= 0 || config.RefreshTokenTTL <= 0 {
		return fmt.Errorf("config: token TTL must be > 0")
	}
	if config.PostgresMaxOpenConns <= 0 {
		return fmt.Errorf("config: GOPHKEEPER_POSTGRES_MAX_OPEN_CONNS must be > 0")
	}
	return nil
}

func stringEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func parseEnv[T any](dst *T, key string, parse func(string) (T, error)) error {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	value, err := parse(raw)
	if err != nil {
		return fmt.Errorf("config: %s=%q: %w", key, raw, err)
	}
	*dst = value
	return nil
}

func parsePositiveInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, fmt.Errorf("must be a positive integer")
	}
	return n, nil
}
