package handler

import (
	"net/http"

	"goph-keeper/internal/buildinfo"
	"goph-keeper/internal/delivery/http/httpout"
)

// Healthz возвращает состояние сервера и метаданные сборки.
func Healthz(w http.ResponseWriter, r *http.Request) {
	httpout.WriteJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"version":   buildinfo.Version,
		"buildDate": buildinfo.BuildDate,
	})
}
