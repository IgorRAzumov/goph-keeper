package health

import (
	"goph-keeper/internal/buildinfo"
	"goph-keeper/internal/delivery/http/util"
	"net/http"
)

// Health возвращает состояние сервера и метаданные сборки.
func Health(w http.ResponseWriter, request *http.Request) {
	util.WriteJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"version":   buildinfo.Version,
		"buildDate": buildinfo.BuildDate,
	})
}
