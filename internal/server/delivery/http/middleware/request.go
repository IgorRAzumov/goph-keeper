package middleware

import (
	"goph-keeper/internal/logging"
	"net/http"
	"time"
)

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

// RequestLogMiddleware логирует каждый HTTP-запрос (метод, путь, длительность).
func RequestLogMiddleware(log logging.Logger, next http.Handler) http.Handler {
	if log == nil {
		log = noopLogger{}
	}
	if next == nil {
		next = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"dur", time.Since(start),
		)
	})
}
