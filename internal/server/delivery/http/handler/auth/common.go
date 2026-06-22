package auth

import (
	"encoding/json"
	"io"
	"net/http"

	"goph-keeper/internal/contract"
	"goph-keeper/internal/logging"
	"goph-keeper/internal/server/delivery/http/util"
)

func decodeRequestBody(log logging.Logger, writer http.ResponseWriter, request *http.Request, logMsg string, out any) bool {
	if request.Body == nil {
		util.WriteJSON(writer, http.StatusBadRequest, contract.ErrorResponse{Error: "empty body"})
		return false
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			log.Error(logMsg, "err", err)
		}
	}(request.Body)

	if err := json.NewDecoder(request.Body).Decode(out); err != nil {
		util.WriteJSON(writer, http.StatusBadRequest, contract.ErrorResponse{Error: "invalid json"})
		return false
	}
	return true
}

func writeUsecaseError(log logging.Logger, writer http.ResponseWriter, logMsg string, err error) {
	if code, ok := util.StatusFromDomain(err); ok {
		util.WriteJSON(writer, code, contract.ErrorResponse{Error: err.Error()})
		return
	}
	log.Error(logMsg, "err", err)
	util.WriteJSON(writer, http.StatusInternalServerError, contract.ErrorResponse{Error: "internal error"})
}
