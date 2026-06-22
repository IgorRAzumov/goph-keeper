package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"goph-keeper/internal/client/api"
	clientapp "goph-keeper/internal/client/app"
	"goph-keeper/internal/client/model"
	"goph-keeper/internal/config"
	applogin "goph-keeper/internal/server/application/auth/login"
	applogout "goph-keeper/internal/server/application/auth/logout"
	apprefresh "goph-keeper/internal/server/application/auth/refresh"
	appregister "goph-keeper/internal/server/application/auth/register"
	appverify "goph-keeper/internal/server/application/auth/verify"
	appsync "goph-keeper/internal/server/application/sync"
	httpapi "goph-keeper/internal/server/delivery/http"
	recordsvc "goph-keeper/internal/server/domain/record/service"
	sessionsvc "goph-keeper/internal/server/domain/session/service"
	usersvc "goph-keeper/internal/server/domain/user/service"
	"goph-keeper/internal/server/infrastructure/postgres"
	"goph-keeper/internal/server/security/jwt"
	"io"
	"log/slog"
	"net/http/httptest"
)

func TestPostgresClientServerFlow(t *testing.T) {
	dsn := os.Getenv("GOPHKEEPER_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set GOPHKEEPER_POSTGRES_DSN for postgres integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := postgres.OpenPool(dsn, 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := postgres.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		JWTSecret:       "integration-jwt-secret",
		AccessTokenTTL:  time.Minute,
		RefreshTokenTTL: time.Hour,
	}
	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	recordRepo := postgres.NewRecordRepository(db)

	jwtProvider := jwt.NewProvider(cfg.JWTSecret)
	deps := httpapi.Dependencies{
		RegisterUser: appregister.NewUserUsecase(usersvc.NewUserService(userRepo)),
		Login:        applogin.NewLoginUsecase(usersvc.NewUserService(userRepo), sessionsvc.NewSessionService(sessionRepo), jwtProvider, cfg.AccessTokenTTL, cfg.RefreshTokenTTL),
		Refresh:      apprefresh.NewRefreshUsecase(sessionRepo, sessionsvc.NewSessionService(sessionRepo), jwtProvider, cfg.AccessTokenTTL, cfg.RefreshTokenTTL),
		Logout:       applogout.NewLogoutUsecase(sessionRepo, jwtProvider),
		Verify:       appverify.NewUsecase(jwtProvider, sessionRepo),
		Sync:         appsync.NewUsecase(recordsvc.NewRecordService(recordRepo)),
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httptest.NewServer(httpapi.Router(logger, deps))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "pg-master")

	app, err := clientapp.New("pg-master")
	if err != nil {
		t.Fatal(err)
	}
	app.API = api.NewClient(srv.URL)
	login := "intuser_" + time.Now().Format("150405")
	if err := app.Register(ctx, login, "pass", srv.URL); err != nil {
		t.Fatal(err)
	}
	id, err := app.Add(ctx, clientapp.AddInput{Type: model.RecordTypeText, Meta: "m", Payload: "data"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := app.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Meta != "m" {
		t.Fatalf("unexpected: %+v", got)
	}
}
