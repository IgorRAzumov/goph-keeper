package config

import (
	"testing"
	"time"
)

func TestLoadUsesLocalDefaults(t *testing.T) {
	t.Setenv("GOPHKEEPER_ADDR", "")
	t.Setenv("GOPHKEEPER_JWT_SECRET", "")
	t.Setenv("GOPHKEEPER_ACCESS_TTL", "")
	t.Setenv("GOPHKEEPER_REFRESH_TTL", "")
	t.Setenv("GOPHKEEPER_POSTGRES_DSN", "")
	t.Setenv("GOPHKEEPER_POSTGRES_MAX_OPEN_CONNS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.HTTPAddr != "127.0.0.1:8080" {
		t.Fatalf("unexpected HTTP addr: %q", cfg.HTTPAddr)
	}
	if cfg.JWTSecret != DefaultJWTSecret {
		t.Fatalf("unexpected jwt secret: %q", cfg.JWTSecret)
	}
	if cfg.PostgresDSN != DefaultPostgresDSN {
		t.Fatalf("unexpected postgres dsn: %q", cfg.PostgresDSN)
	}
}

func TestLoadUsesEnvValues(t *testing.T) {
	t.Setenv("GOPHKEEPER_ADDR", "127.0.0.1:9090")
	t.Setenv("GOPHKEEPER_JWT_SECRET", "secret")
	t.Setenv("GOPHKEEPER_ACCESS_TTL", "1m")
	t.Setenv("GOPHKEEPER_REFRESH_TTL", "24h")
	t.Setenv("GOPHKEEPER_POSTGRES_DSN", "postgres://u:p@localhost:5432/db")
	t.Setenv("GOPHKEEPER_POSTGRES_MAX_OPEN_CONNS", "25")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.HTTPAddr != "127.0.0.1:9090" {
		t.Fatalf("expected env HTTP addr, got %q", cfg.HTTPAddr)
	}
	if cfg.JWTSecret != "secret" {
		t.Fatalf("expected env jwt secret, got %q", cfg.JWTSecret)
	}
	if cfg.AccessTokenTTL != time.Minute {
		t.Fatalf("expected env access ttl 1m, got %s", cfg.AccessTokenTTL)
	}
	if cfg.RefreshTokenTTL != 24*time.Hour {
		t.Fatalf("expected env refresh ttl 24h, got %s", cfg.RefreshTokenTTL)
	}
	if cfg.PostgresDSN != "postgres://u:p@localhost:5432/db" {
		t.Fatalf("unexpected postgres dsn: %q", cfg.PostgresDSN)
	}
	if cfg.PostgresMaxOpenConns != 25 {
		t.Fatalf("expected env max open conns 25, got %d", cfg.PostgresMaxOpenConns)
	}
}
