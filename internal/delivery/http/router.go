package httpapi

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"goph-keeper/internal/delivery/http/handler"
)

// Router собирает и возвращает HTTP-роутер со всеми зарегистрированными маршрутами.
func Router(log *slog.Logger, deps Deps) chi.Router {
	r := chi.NewRouter()

	r.Get("/healthz", handler.Healthz)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handler.AuthRegister(log, deps.RegisterUser))
			r.Post("/login", handler.NotImplemented)
			r.Post("/refresh", handler.NotImplemented)
			r.Post("/logout", handler.NotImplemented)
		})

		r.Route("/sync", func(r chi.Router) {
			r.Get("/", handler.NotImplemented)
			r.Post("/", handler.NotImplemented)
		})
	})

	return r
}
