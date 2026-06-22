package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	appverify "goph-keeper/internal/server/application/auth/verify"
	"goph-keeper/internal/server/delivery/http/util"
	"goph-keeper/internal/server/domain/common"
)

// BearerAuth извлекает Authorization: Bearer <access JWT>, делегирует проверку сценарию verify
// и кладёт user/session id в контекст запроса.
func BearerAuth(verifier *appverify.Usecase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if verifier == nil {
				util.WriteJSON(writer, http.StatusNotImplemented, map[string]any{"error": "auth not configured"})
				return
			}

			authz := strings.TrimSpace(request.Header.Get("Authorization"))
			scheme, token, ok := strings.Cut(authz, " ")
			if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
				util.WriteJSON(writer, http.StatusUnauthorized, map[string]any{"error": "missing bearer token"})
				return
			}

			out, err := verifier.Execute(request.Context(), token, time.Now().UTC())
			if err != nil {
				if errors.Is(err, common.ErrUnauthorized) {
					util.WriteJSON(writer, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
					return
				}
				util.WriteJSON(writer, http.StatusInternalServerError, map[string]any{"error": "internal error"})
				return
			}

			ctx := request.Context()
			ctx = WithUserID(ctx, out.UserID)
			ctx = WithSessionID(ctx, out.SessionID)
			next.ServeHTTP(writer, request.WithContext(ctx))
		})
	}
}
