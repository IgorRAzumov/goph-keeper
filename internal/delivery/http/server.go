package httpapi

import (
	"context"
	"errors"
	httpapi "goph-keeper/internal/delivery/http/middlewear"
	"goph-keeper/internal/logging"
	"net"
	"net/http"
	"time"
)

// Server — HTTP-точка входа в API
type Server struct {
	httpServer *http.Server
	listener   net.Listener
	logger     logging.Logger
}

// ServerConfig конфигурирует HTTP-сервер
type ServerConfig struct {
	// Address — TCP-адрес для прослушивания (например "127.0.0.1:8080").
	Address string
	// Dependencies прокидывает сценарии в HTTP-обработчики.
	Dependencies Dependensies
}

// NewServer создаёт HTTP-сервер и регистрирует роуты.
func NewServer(cfg ServerConfig, logger logging.Logger) (*Server, error) {
	if cfg.Address == "" {
		cfg.Address = "127.0.0.1:8080"
	}

	var handler http.Handler = Router(logger, cfg.Dependencies)

	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           httpapi.RequestLogMiddleware(logger, handler),
		ReadHeaderTimeout: 5 * time.Second,
	}

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, err
	}

	return &Server{httpServer: server, listener: listener, logger: logger}, nil
}

// Addr возвращает фактически занятый адрес
func (server *Server) Addr() string {
	if server.listener == nil {
		return ""
	}
	return server.listener.Addr().String()
}

// Run обслуживает запросы до отмены ctx, затем корректно завершает работу.
func (server *Server) Run(ctx context.Context) error {
	errChannel := make(chan error, 1)
	go func() {
		errChannel <- server.httpServer.Serve(server.listener)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.httpServer.Shutdown(shutdownCtx)
	case err := <-errChannel:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
