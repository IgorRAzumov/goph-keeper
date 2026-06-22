package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"goph-keeper/internal/config"
	"goph-keeper/internal/logging"
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
)

// App содержит запускаемые сервисы
type App struct {
	httpServer *httpapi.Server
	logger     logging.Logger
	database   *sql.DB
}

// New создаёт приложение по умолчанию
func New() (*App, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	configuration, err := config.Load()
	if err != nil {
		return nil, err
	}

	dependencies, database, err := initDependencies(configuration)
	if err != nil {
		logger.Error("database init failed", "err", err)
		return nil, err
	}

	server, err := startServer(configuration, logger, dependencies)
	if err != nil {
		_ = database.Close()
		logger.Error("server init failed", "err", err)
		return nil, err
	}
	return &App{httpServer: server, logger: logger, database: database}, nil
}

// Logger возвращает логгер приложения
func (app *App) Logger() logging.Logger {
	if app == nil {
		return nil
	}
	return app.logger
}

// Run запуск приложения
func (app *App) Run(ctx context.Context) error {
	if app == nil || app.httpServer == nil {
		return fmt.Errorf("app: not initialized")
	}
	err := app.httpServer.Run(ctx)
	if app.database != nil {
		_ = app.database.Close()
		app.database = nil
	}
	return err
}

func startServer(
	configuration config.Config,
	logger logging.Logger,
	dependencies httpapi.Dependencies,
) (*httpapi.Server, error) {
	server, err := httpapi.NewServer(httpapi.ServerConfig{
		Address:      configuration.HTTPAddr,
		Dependencies: dependencies,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("http server: %w", err)
	}

	logger.Info("server listening", "addr", configuration.HTTPAddr)
	return server, nil
}

func initDependencies(configuration config.Config) (httpapi.Dependencies, *sql.DB, error) {
	database, err := initDatabase(configuration)
	if err != nil {
		return httpapi.Dependencies{}, nil, err
	}

	sessionRepository := postgres.NewSessionRepository(database)
	sessionService := sessionsvc.NewSessionService(sessionRepository)
	userService := usersvc.NewUserService(postgres.NewUserRepository(database))

	jwtProvider := jwt.NewProvider(configuration.JWTSecret)

	registerUser := appregister.NewUserUsecase(userService)
	login := applogin.NewLoginUsecase(userService, sessionService, jwtProvider, configuration.AccessTokenTTL, configuration.RefreshTokenTTL)
	refresh := apprefresh.NewRefreshUsecase(sessionRepository, sessionService, jwtProvider, configuration.AccessTokenTTL, configuration.RefreshTokenTTL)
	logout := applogout.NewLogoutUsecase(sessionRepository, jwtProvider)
	verify := appverify.NewUsecase(jwtProvider, sessionRepository)

	recordRepository := postgres.NewRecordRepository(database)
	recordService := recordsvc.NewRecordService(recordRepository)
	syncUsecase := appsync.NewUsecase(recordService)

	return httpapi.Dependencies{
		RegisterUser: registerUser,
		Login:        login,
		Refresh:      refresh,
		Logout:       logout,
		Verify:       verify,
		Sync:         syncUsecase,
	}, database, nil

}

func initDatabase(configuration config.Config) (*sql.DB, error) {
	database, err := postgres.OpenPool(configuration.PostgresDSN, configuration.PostgresMaxOpenConns)
	if err != nil {
		return nil, fmt.Errorf("app: postgres open: %w", err)
	}

	migrationContext, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	err = postgres.Migrate(migrationContext, database)
	cancel()

	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("app: postgres migrate: %w", err)
	}
	return database, nil
}
