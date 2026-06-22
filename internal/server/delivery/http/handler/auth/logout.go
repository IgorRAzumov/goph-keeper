package auth

import (
	"net/http"
	"strings"

	"goph-keeper/internal/logging"
	applogout "goph-keeper/internal/server/application/auth/logout"
	"goph-keeper/internal/server/delivery/http/handler/stub"
	"goph-keeper/internal/server/delivery/http/util"
)

// Logout возвращает обработчик выхода (отзыв refresh-сессии по access JWT).
func Logout(log logging.Logger, usecase *applogout.Usecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if usecase == nil {
			stub.NotImplemented(writer, request)
			return
		}

		authz := strings.TrimSpace(request.Header.Get("Authorization"))
		scheme, token, ok := strings.Cut(authz, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
			util.WriteJSON(writer, http.StatusUnauthorized, map[string]any{"error": "missing bearer token"})
			return
		}

		err := usecase.Execute(request.Context(), applogout.Input{AccessToken: strings.TrimSpace(token)})
		if err != nil {
			if code, ok := util.StatusFromDomain(err); ok {
				util.WriteJSON(writer, code, map[string]any{"error": err.Error()})
				return
			}
			log.Error("logout failed", "err", err)
			util.WriteJSON(writer, http.StatusInternalServerError, map[string]any{"error": "internal error"})
			return
		}

		writer.WriteHeader(http.StatusNoContent)
	}
}
