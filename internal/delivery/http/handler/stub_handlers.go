package handler

import (
	"net/http"

	"goph-keeper/internal/delivery/http/httpout"
)

// NotImplemented — временная заглушка для ещё не реализованных эндпойнтов.
func NotImplemented(w http.ResponseWriter, r *http.Request) {
	httpout.WriteJSON(w, http.StatusNotImplemented, map[string]any{
		"error": "not implemented",
	})
}
