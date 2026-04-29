package auth

import (
	"encoding/json"
	"goph-keeper/internal/application/user"
	"goph-keeper/internal/delivery/http/handler/stub"
	"goph-keeper/internal/delivery/http/util"
	"goph-keeper/internal/logging"
	"io"
	"net/http"
)

// Register возвращает обработчик отвечающий за регистрацию нового пользователя
func Register(log logging.Logger, usecase *user.Usecase) http.HandlerFunc {
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
			err := body.Close()
			if err != nil {
				log.Error("auth io readerCloser", "err", err)
			}
		}(request.Body)

		var req registerUserRequest
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			util.WriteJSON(writer, http.StatusBadRequest, map[string]any{"error": "invalid json"})
			return
		}

		output, err := usecase.Execute(request.Context(), user.RegisterUserInput{
			Login:    req.Login,
			Password: req.Password,
		})
		if err != nil {
			if code, ok := util.StatusFromDomain(err); ok {
				util.WriteJSON(writer, code, map[string]any{"error": err.Error()})
				return
			}
			log.Error("register_user failed", "err", err)
			util.WriteJSON(writer, http.StatusInternalServerError, map[string]any{"error": "internal error"})
			return
		}

		util.WriteJSON(writer, http.StatusCreated, map[string]any{"userId": output.UserID})
	}
}
