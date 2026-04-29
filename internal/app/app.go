// Пакет app — корень композиции: связывает адаптеры инфраструктуры, сценарии и HTTP-слой доставки.
// Точки входа cmd должны оставаться тонкими и делегировать жизненный цикл в App.
package app

import (
	"context"
	"fmt"
	"goph-keeper/internal/application/user"
	"log/slog"
	"os"

	"goph-keeper/internal/config"
	httpapi "goph-keeper/internal/delivery/http"
	usersvc "goph-keeper/internal/domain/user/service"
	"goph-keeper/internal/infrastructure/memory"
)

// App содержит запускаемые сервисы, собранные из конфигурации и адаптеров.
type App struct {
	http *httpapi.Server
	log  *slog.Logger
}

// New создаёт приложение по умолчанию: загружает конфигурацию из окружения,
// создаёт логгер в stdout (текстовый обработчик) на уровне INFO и связывает адаптеры с HTTP-слоем.
func New() (*App, error) {
	log := newLogger()
	cfg := config.Load()
	return NewWith(cfg, log)
}

// NewWith собирает граф приложения для тестов или нестандартной проводки (явные config + logger).
func NewWith(cfg config.Config, log *slog.Logger) (*App, error) {
	userRepo := memory.NewUserRepository()
	userService := usersvc.NewUserService(userRepo)
	registerUser := user.NewRegisterUserUseCase(userService)

	server, app, err := startServer(cfg, log, registerUser)
	if err != nil {
		return app, err
	}
	return &App{http: server, log: log}, nil
}

// Logger возвращает логгер приложения
func (app *App) Logger() *slog.Logger {
	if app == nil || app.log == nil {
		return slog.Default()
	}
	return app.log
}

// Run запуск приложения
func (app *App) Run(ctx context.Context) error {
	if app == nil || app.http == nil {
		return fmt.Errorf("app: not initialized")
	}
	return app.http.Run(ctx)
}

func startServer(
	cfg config.Config,
	log *slog.Logger,
	registerUser *user.RegisterUserUseCase,
) (*httpapi.Server, *App, error) {
	server, err := httpapi.NewServer(httpapi.ServerConfig{
		Address:      cfg.HTTPAddr,
		Dependencies: httpapi.Deps{RegisterUser: registerUser},
	}, log)
	if err != nil {
		return nil, nil, fmt.Errorf("http server: %w", err)
	}

	log.Info("server listening", slog.String("addr", cfg.HTTPAddr))
	return server, nil, nil
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
