package httpout

import (
	"encoding/json"
	"errors"
	"net/http"

	"goph-keeper/internal/domain/common"
)

// WriteJSON пишет JSON-ответ с указанным HTTP-статусом.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// StatusFromDomain маппит доменные sentinel-ошибки на HTTP-коды.
func StatusFromDomain(err error) (int, bool) {
	switch {
	case errors.Is(err, common.ErrInvalidInput):
		return http.StatusBadRequest, true
	case errors.Is(err, common.ErrNotFound):
		return http.StatusNotFound, true
	case errors.Is(err, common.ErrConflict):
		return http.StatusConflict, true
	case errors.Is(err, common.ErrNotImplemented):
		return http.StatusNotImplemented, true
	default:
		return 0, false
	}
}

