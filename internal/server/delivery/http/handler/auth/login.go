package auth

import (
	"net/http"

	"goph-keeper/internal/contract"
	"goph-keeper/internal/logging"
	applogin "goph-keeper/internal/server/application/auth/login"
	"goph-keeper/internal/server/delivery/http/handler/stub"
	"goph-keeper/internal/server/delivery/http/util"
)

// Login возвращает обработчик логина.
func Login(log logging.Logger, usecase *applogin.LoginUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if usecase == nil {
			stub.NotImplemented(writer, request)
			return
		}

		var req contract.LoginRequest
		if ok := decodeRequestBody(log, writer, request, "auth login body close", &req); !ok {
			return
		}

		out, err := usecase.Execute(request.Context(), applogin.Input{
			Login:    req.Login,
			Password: req.Password,
		})
		if err != nil {
			writeUsecaseError(log, writer, "login failed", err)
			return
		}

		util.WriteJSON(writer, http.StatusOK, contract.TokenResponse{
			AccessToken:  out.AccessToken,
			RefreshToken: out.RefreshToken,
			MasterSalt:   out.MasterSalt,
		})
	}
}
