package auth

import (
	"encoding/json"
	"io"
	"net/http"

	applogin "goph-keeper/internal/application/auth/login"
	"goph-keeper/internal/delivery/http/handler/stub"
	"goph-keeper/internal/delivery/http/util"
	"goph-keeper/internal/logging"
)

// Login возвращает обработчик логина.
func Login(log logging.Logger, usecase *applogin.LoginUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if usecase == nil {
			stub.NotImplemented(writer, request)
			return
		}
		if request.Body == nil {
			util.WriteJSON(writer, http.StatusBadRequest, map[string]any{"error": "empty body"})
			return
		}
		defer func(body io.ReadCloser) {
			if err := body.Close(); err != nil {
				log.Error("auth login body close", "err", err)
			}
		}(request.Body)

		var req loginRequest
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			util.WriteJSON(writer, http.StatusBadRequest, map[string]any{"error": "invalid json"})
			return
		}

		out, err := usecase.Execute(request.Context(), applogin.Input{
			Login:    req.Login,
			Password: req.Password,
		})
		if err != nil {
			if code, ok := util.StatusFromDomain(err); ok {
				util.WriteJSON(writer, code, map[string]any{"error": err.Error()})
				return
			}
			log.Error("login failed", "err", err)
			util.WriteJSON(writer, http.StatusInternalServerError, map[string]any{"error": "internal error"})
			return
		}

		util.WriteJSON(writer, http.StatusOK, map[string]any{
			"access_token":  out.AccessToken,
			"refresh_token": out.RefreshToken,
		})
	}
}
