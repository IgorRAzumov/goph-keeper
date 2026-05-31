package auth

import (
	"encoding/json"
	"io"
	"net/http"

	apprefresh "goph-keeper/internal/application/auth/refresh"
	"goph-keeper/internal/delivery/http/handler/stub"
	"goph-keeper/internal/delivery/http/util"
	"goph-keeper/internal/logging"
)

// Refresh возвращает обработчик refresh.
func Refresh(log logging.Logger, usecase *apprefresh.RefreshUsecase) http.HandlerFunc {
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
				log.Error("auth refresh body close", "err", err)
			}
		}(request.Body)

		var req refreshRequest
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			util.WriteJSON(writer, http.StatusBadRequest, map[string]any{"error": "invalid json"})
			return
		}

		out, err := usecase.Execute(request.Context(), apprefresh.Input{RefreshToken: req.RefreshToken})
		if err != nil {
			if code, ok := util.StatusFromDomain(err); ok {
				util.WriteJSON(writer, code, map[string]any{"error": err.Error()})
				return
			}
			log.Error("refresh failed", "err", err)
			util.WriteJSON(writer, http.StatusInternalServerError, map[string]any{"error": "internal error"})
			return
		}

		util.WriteJSON(writer, http.StatusOK, map[string]any{
			"access_token":  out.AccessToken,
			"refresh_token": out.RefreshToken,
		})
	}
}
