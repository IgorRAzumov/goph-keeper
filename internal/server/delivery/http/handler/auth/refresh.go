package auth

import (
	"net/http"

	"goph-keeper/internal/contract"
	"goph-keeper/internal/logging"
	apprefresh "goph-keeper/internal/server/application/auth/refresh"
	"goph-keeper/internal/server/delivery/http/handler/stub"
	"goph-keeper/internal/server/delivery/http/util"
)

// Refresh возвращает обработчик refresh.
func Refresh(log logging.Logger, usecase *apprefresh.RefreshUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if usecase == nil {
			stub.NotImplemented(writer, request)
			return
		}

		var req contract.RefreshRequest
		if ok := decodeRequestBody(log, writer, request, "auth refresh body close", &req); !ok {
			return
		}

		out, err := usecase.Execute(request.Context(), apprefresh.Input{RefreshToken: req.RefreshToken})
		if err != nil {
			writeUsecaseError(log, writer, "refresh failed", err)
			return
		}

		util.WriteJSON(writer, http.StatusOK, contract.TokenResponse{
			AccessToken:  out.AccessToken,
			RefreshToken: out.RefreshToken,
		})
	}
}
