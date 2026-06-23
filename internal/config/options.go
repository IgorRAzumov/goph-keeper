package config

import "time"

// Option переопределяет поле конфигурации после чтения env (functional options).
type Option func(*Config)

// WithHTTPAddr задаёт адрес HTTP-листенера.
func WithHTTPAddr(addr string) Option {
	return func(config *Config) { config.HTTPAddr = addr }
}

// WithJWTSecret задаёт секрет подписи JWT.
func WithJWTSecret(secret string) Option {
	return func(config *Config) { config.JWTSecret = secret }
}

// WithAccessTokenTTL задаёт TTL access-токена.
func WithAccessTokenTTL(ttl time.Duration) Option {
	return func(config *Config) { config.AccessTokenTTL = ttl }
}

// WithRefreshTokenTTL задаёт TTL refresh-токена.
func WithRefreshTokenTTL(ttl time.Duration) Option {
	return func(config *Config) { config.RefreshTokenTTL = ttl }
}

// WithPostgresDSN задаёт строку подключения Postgres.
func WithPostgresDSN(dsn string) Option {
	return func(config *Config) { config.PostgresDSN = dsn }
}

// WithPostgresMaxOpenConns задаёт лимит открытых соединений пула.
func WithPostgresMaxOpenConns(n int) Option {
	return func(config *Config) { config.PostgresMaxOpenConns = n }
}
