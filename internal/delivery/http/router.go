package httpapi

import (
	"goph-keeper/internal/delivery/http/handler/auth"
	"goph-keeper/internal/delivery/http/handler/health"
	"goph-keeper/internal/delivery/http/handler/stub"
	httpmiddleware "goph-keeper/internal/delivery/http/middleware"

	"github.com/go-chi/chi/v5"

	"goph-keeper/internal/logging"
)

// Router собирает и возвращает HTTP-роутер со всеми зарегистрированными маршрутами.
func Router(log logging.Logger, deps Dependencies) chi.Router {
	router := chi.NewRouter()

	router.Get("/healthz", health.Health)

	router.Route("/api/v1", func(internalRouter chi.Router) {
		internalRouter.Route("/auth", func(r chi.Router) {
			r.Post("/register", auth.Register(log, deps.RegisterUser))
			r.Post("/login", auth.Login(log, deps.Login))
			r.Post("/refresh", auth.Refresh(log, deps.Refresh))
			r.Post("/logout", auth.Logout(log, deps.Logout))
		})

		internalRouter.Route("/sync", func(r chi.Router) {
			r.Use(httpmiddleware.BearerAuth(deps.JWT, deps.Sessions))
			r.Get("/", stub.NotImplemented)
			r.Post("/", stub.NotImplemented)
		})
	})

	return router
}
