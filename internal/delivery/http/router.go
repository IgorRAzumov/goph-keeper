package httpapi

import (
	"goph-keeper/internal/delivery/http/handler/auth"
	"goph-keeper/internal/delivery/http/handler/health"
	"goph-keeper/internal/delivery/http/handler/stub"

	"github.com/go-chi/chi/v5"

	"goph-keeper/internal/logging"
)

// Router собирает и возвращает HTTP-роутер со всеми зарегистрированными маршрутами.
func Router(log logging.Logger, deps Dependensies) chi.Router {
	router := chi.NewRouter()

	router.Get("/health", health.Health)

	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", auth.Register(log, deps.RegisterUser))
			r.Post("/login", stub.NotImplemented)
			r.Post("/refresh", stub.NotImplemented)
			r.Post("/logout", stub.NotImplemented)
		})

		r.Route("/sync", func(r chi.Router) {
			r.Get("/", stub.NotImplemented)
			r.Post("/", stub.NotImplemented)
		})
	})

	return router
}
