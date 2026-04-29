package stub

import (
	"net/http"

	"goph-keeper/internal/delivery/http/util"
)

// NotImplemented — временная заглушка для ещё не реализованных эндпойнтов.
func NotImplemented(w http.ResponseWriter, request *http.Request) {
	util.WriteJSON(w, http.StatusNotImplemented, map[string]any{
		"error": "not implemented",
	})
}
