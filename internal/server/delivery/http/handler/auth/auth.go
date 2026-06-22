package auth

import (
	"net/http"

	"goph-keeper/internal/contract"
	"goph-keeper/internal/logging"
	appregister "goph-keeper/internal/server/application/auth/register"
	"goph-keeper/internal/server/delivery/http/handler/stub"
	"goph-keeper/internal/server/delivery/http/util"
)

// Register возвращает обработчик отвечающий за регистрацию нового пользователя
func Register(log logging.Logger, usecase *appregister.Usecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if usecase == nil {
			stub.NotImplemented(writer, request)
			return
		}
		var req contract.RegisterRequest
		if ok := decodeRequestBody(log, writer, request, "auth register body close", &req); !ok {
			return
		}

		output, err := usecase.Execute(request.Context(), appregister.Input{
			Login:      req.Login,
			Password:   req.Password,
			MasterSalt: req.MasterSalt,
		})
		if err != nil {
			writeUsecaseError(log, writer, "register_user failed", err)
			return
		}

		util.WriteJSON(writer, http.StatusCreated, contract.RegisterResponse{UserID: output.UserID})
	}
}
