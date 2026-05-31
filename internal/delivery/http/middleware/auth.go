package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"goph-keeper/internal/delivery/http/util"
	"goph-keeper/internal/domain/common"
	sessionrepo "goph-keeper/internal/domain/session/repository"
	"goph-keeper/internal/security/jwt"
)

// BearerAuth проверяет Authorization: Bearer <access JWT> и кладёт user/session id в контекст запроса.
func BearerAuth(provider *jwt.Provider, sessions sessionrepo.SessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if provider == nil || sessions == nil {
				util.WriteJSON(writer, http.StatusNotImplemented, map[string]any{"error": "auth not configured"})
				return
			}

			now := time.Now().UTC()
			authz := strings.TrimSpace(request.Header.Get("Authorization"))
			scheme, token, ok := strings.Cut(authz, " ")
			if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
				util.WriteJSON(writer, http.StatusUnauthorized, map[string]any{"error": "missing bearer token"})
				return
			}

			claims, err := provider.ParseAccessToken(strings.TrimSpace(token), now)
			if err != nil {
				util.WriteJSON(writer, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
				return
			}
			session, err := sessions.Get(request.Context(), claims.Sid)
			if err != nil {
				if errors.Is(err, common.ErrNotFound) {
					util.WriteJSON(writer, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
					return
				}
				util.WriteJSON(writer, http.StatusInternalServerError, map[string]any{"error": "internal error"})
				return
			}
			if session.UserID != claims.Sub || !now.Before(session.RefreshExpiresAt) {
				util.WriteJSON(writer, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
				return
			}

			ctx := request.Context()
			ctx = WithUserID(ctx, claims.Sub)
			ctx = WithSessionID(ctx, claims.Sid)
			next.ServeHTTP(writer, request.WithContext(ctx))
		})
	}
}
