package httpapi

import (
	"context"
	"errors"
	httpapi "goph-keeper/internal/delivery/http/middlewear"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Server — HTTP-точка входа в API (адаптер слоя доставки).
type Server struct {
	httpServer *http.Server
	l          net.Listener
	log        *slog.Logger
}

// ServerConfig конфигурирует HTTP-сервер.
type ServerConfig struct {
	// Address — TCP-адрес для прослушивания (например "127.0.0.1:8080").
	Address string
	// Dependencies прокидывает сценарии в HTTP-обработчики.
	Dependencies Deps
}

// NewServer создаёт HTTP-сервер и регистрирует роуты.
func NewServer(cfg ServerConfig, log *slog.Logger) (*Server, error) {
	if cfg.Address == "" {
		cfg.Address = "127.0.0.1:8080"
	}
	if log == nil {
		log = slog.Default()
	}

	var handler http.Handler = Router(log, cfg.Dependencies)

	hs := &http.Server{
		Addr:              cfg.Address,
		Handler:           httpapi.RequestLogMiddleware(log, handler),
		ReadHeaderTimeout: 5 * time.Second,
	}

	l, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, err
	}

	return &Server{httpServer: hs, l: l, log: log}, nil
}

// Addr возвращает фактически занятый адрес (может отличаться от cfg, если ОС выбрала порт).
func (s *Server) Addr() string {
	if s.l == nil {
		return ""
	}
	return s.l.Addr().String()
}

// Run обслуживает запросы до отмены ctx, затем корректно завершает работу.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.httpServer.Serve(s.l)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
