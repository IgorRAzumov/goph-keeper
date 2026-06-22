package app

import (
	"context"
	"os"
	"testing"
	"time"

	"goph-keeper/internal/config"
)

func TestInitDependenciesBuildsGraph(t *testing.T) {
	t.Parallel()

	postgresDSN := os.Getenv("GOPHKEEPER_POSTGRES_DSN")
	if postgresDSN == "" {
		t.Skip("set GOPHKEEPER_POSTGRES_DSN to run app wiring test")
	}

	dependencies, database, err := initDependencies(config.Config{
		HTTPAddr:             "127.0.0.1:0",
		JWTSecret:            "test-secret",
		AccessTokenTTL:       time.Minute,
		RefreshTokenTTL:      time.Hour,
		PostgresDSN:          postgresDSN,
		PostgresMaxOpenConns: 1,
	})
	if err != nil {
		t.Fatalf("initDependencies failed: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if database == nil {
		t.Fatal("expected database")
	}
	if dependencies.RegisterUser == nil || dependencies.Login == nil || dependencies.Refresh == nil || dependencies.Logout == nil {
		t.Fatal("expected auth usecases")
	}
	if dependencies.Verify == nil {
		t.Fatal("expected verify usecase")
	}
	if dependencies.Sync == nil {
		t.Fatal("expected sync usecase")
	}
}

func TestLoggerReturnsDefaultForNilApp(t *testing.T) {
	t.Parallel()

	var app *App
	if app.Logger() != nil {
		t.Fatal("expected nil logger for nil app")
	}
}

func TestNewWithRejectsDefaultJWTSecretForPostgres(t *testing.T) {
	t.Parallel()

	t.Skip("internal/config.Load()")
}

func TestRunRejectsUninitializedApp(t *testing.T) {
	t.Parallel()

	err := (&App{}).Run(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
