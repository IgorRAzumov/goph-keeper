package sync

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	appsync "goph-keeper/internal/application/sync"
	"goph-keeper/internal/delivery/http/handler/stub"
	httpmiddleware "goph-keeper/internal/delivery/http/middleware"
	"goph-keeper/internal/delivery/http/util"
	"goph-keeper/internal/domain/record/model"
	"goph-keeper/internal/logging"
)

// Pull возвращает обработчик GET /api/v1/sync (выгрузка изменений с сервера).
func Pull(log logging.Logger, usecase *appsync.Usecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if usecase == nil {
			stub.NotImplemented(writer, request)
			return
		}

		since, err := parseSinceQuery(request)
		if err != nil {
			util.WriteJSON(writer, http.StatusBadRequest, map[string]any{"error": "invalid since"})
			return
		}

		out, err := usecase.Pull(request.Context(), appsync.PullInput{
			OwnerID:      httpmiddleware.UserID(request.Context()),
			SinceVersion: since,
		})
		if err != nil {
			writeSyncError(log, writer, "sync pull failed", err)
			return
		}

		util.WriteJSON(writer, http.StatusOK, map[string]any{
			"records": recordsToDTO(out.Records),
		})
	}
}

// Push возвращает обработчик POST /api/v1/sync (загрузка изменений на сервер).
func Push(log logging.Logger, usecase *appsync.Usecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if usecase == nil {
			stub.NotImplemented(writer, request)
			return
		}

		ownerID := httpmiddleware.UserID(request.Context())
		if request.Body == nil {
			util.WriteJSON(writer, http.StatusBadRequest, map[string]any{"error": "empty body"})
			return
		}
		defer func(body io.ReadCloser) {
			if err := body.Close(); err != nil {
				log.Error("sync push body close", "err", err)
			}
		}(request.Body)

		var req pushRequest
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			util.WriteJSON(writer, http.StatusBadRequest, map[string]any{"error": "invalid json"})
			return
		}

		records, err := recordsFromDTO(req.Records)
		if err != nil {
			util.WriteJSON(writer, http.StatusBadRequest, map[string]any{"error": "invalid record"})
			return
		}

		out, err := usecase.Push(request.Context(), appsync.PushInput{
			OwnerID: ownerID,
			Records: records,
		})
		if err != nil {
			writeSyncError(log, writer, "sync push failed", err)
			return
		}

		status := http.StatusOK
		if len(out.Conflicts) > 0 {
			status = http.StatusConflict
		}
		util.WriteJSON(writer, status, map[string]any{
			"conflicts": recordsToDTO(out.Conflicts),
		})
	}
}

type pushRequest struct {
	Records []recordDTO `json:"records"`
}

type recordDTO struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Meta       string `json:"meta"`
	Ciphertext string `json:"ciphertext"`
	Version    int64  `json:"version"`
	Deleted    bool   `json:"deleted"`
}

func parseSinceQuery(request *http.Request) (int64, error) {
	raw := strings.TrimSpace(request.URL.Query().Get("since"))
	if raw == "" {
		return 0, nil
	}
	since, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || since < 0 {
		return 0, strconv.ErrSyntax
	}
	return since, nil
}

func recordsToDTO(records []*model.Record) []recordDTO {
	if len(records) == 0 {
		return []recordDTO{}
	}
	out := make([]recordDTO, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		out = append(out, recordDTO{
			ID:         record.ID,
			Type:       string(record.Type),
			Meta:       record.Meta,
			Ciphertext: base64.StdEncoding.EncodeToString(record.Ciphertext),
			Version:    record.Version,
			Deleted:    record.Deleted,
		})
	}
	return out
}

func recordsFromDTO(items []recordDTO) ([]*model.Record, error) {
	if len(items) == 0 {
		return []*model.Record{}, nil
	}
	out := make([]*model.Record, 0, len(items))
	for _, item := range items {
		ciphertext, err := base64.StdEncoding.DecodeString(item.Ciphertext)
		if err != nil {
			return nil, err
		}
		out = append(out, &model.Record{
			ID:         item.ID,
			Type:       model.RecordType(item.Type),
			Meta:       item.Meta,
			Ciphertext: ciphertext,
			Version:    item.Version,
			Deleted:    item.Deleted,
		})
	}
	return out, nil
}

func writeSyncError(log logging.Logger, writer http.ResponseWriter, msg string, err error) {
	if code, ok := util.StatusFromDomain(err); ok {
		util.WriteJSON(writer, code, map[string]any{"error": err.Error()})
		return
	}
	log.Error(msg, "err", err)
	util.WriteJSON(writer, http.StatusInternalServerError, map[string]any{"error": "internal error"})
}
