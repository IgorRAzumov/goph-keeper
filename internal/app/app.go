package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"goph-keeper/internal/application/user"
	"goph-keeper/internal/config"
	httpapi "goph-keeper/internal/delivery/http"
	usersvc "goph-keeper/internal/domain/user/service"
	"goph-keeper/internal/infrastructure/memory"
	"goph-keeper/internal/logging"
)

// App содержит запускаемые сервисы
type App struct {
	httpServer *httpapi.Server
	logger     logging.Logger
}

// New создаёт приложение по умолчанию
func New() (*App, error) {
	logger := newLogger()
	cfg := config.Load()
	return NewWith(cfg, logger)
}

// NewWith собирает граф приложения и запускает его
func NewWith(cfg config.Config, logger logging.Logger) (*App, error) {
	userService := usersvc.NewUserService(memory.NewUserRepository())
	registerUser := user.NewUserUsecase(userService)

	server, err := startServer(cfg, logger, registerUser)
	if err != nil {
		logger.Error("app init failed", "err", err)
		return nil, err
	}
	return &App{httpServer: server, logger: logger}, nil
}

// Logger возвращает логгер приложения
func (app *App) Logger() logging.Logger {
	if app == nil || app.logger == nil {
		return logging.Default()
	}
	return app.logger
}

// Run запуск приложения
func (app *App) Run(ctx context.Context) error {
	if app == nil || app.httpServer == nil {
		return fmt.Errorf("app: not initialized")
	}
	return app.httpServer.Run(ctx)
}

func startServer(
	cfg config.Config,
	logger logging.Logger,
	registerUser *user.Usecase,
) (*httpapi.Server, error) {
	server, err := httpapi.NewServer(httpapi.ServerConfig{
		Address:      cfg.HTTPAddr,
		Dependencies: httpapi.Dependencies{RegisterUser: registerUser},
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("http server: %w", err)
	}

	logger.Info("server listening", "addr", cfg.HTTPAddr)
	return server, nil
}

func newLogger() logging.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
