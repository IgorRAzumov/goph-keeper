package handler

import (
	"encoding/json"
	"goph-keeper/internal/application/user"
	"goph-keeper/internal/delivery/http/httpout"
	"io"
	"log/slog"
	"net/http"
)

type registerUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// AuthRegister возвращает обработчик, который делегирует работу сценарию RegisterUserUseCase.
func AuthRegister(log *slog.Logger, uc *user.RegisterUserUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if uc == nil {
			NotImplemented(w, r)
			return
		}
		if r.Body == nil {
			httpout.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "empty body"})
			return
		}
		defer func(body io.ReadCloser) {
			err := body.Close()
			if err != nil {
				log.Error("auth io readerCloser", slog.Any("err", err))
			}
		}(r.Body)

		var req registerUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpout.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
			return
		}

		out, err := uc.Execute(r.Context(), user.RegisterUserInput{
			Login:    req.Login,
			Password: req.Password,
		})
		if err != nil {
			if code, ok := httpout.StatusFromDomain(err); ok {
				httpout.WriteJSON(w, code, map[string]any{"error": err.Error()})
				return
			}
			log.Error("register_user failed", slog.Any("err", err))
			httpout.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
			return
		}

		_ = out
		httpout.WriteJSON(w, http.StatusCreated, map[string]any{"userId": out.UserID})
	}
}
