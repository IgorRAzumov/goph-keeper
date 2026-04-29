package httpapi

import (
	"goph-keeper/internal/logging"
	"net/http"
	"time"
)

func RequestLogMiddleware(log logging.Logger, next http.Handler) http.Handler {
	if log == nil {
		log = logging.Default()
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
