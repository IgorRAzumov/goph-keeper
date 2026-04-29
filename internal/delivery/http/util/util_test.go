package util

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"goph-keeper/internal/domain/common"
)

func TestWriteJSONWritesStatusContentTypeAndBody(t *testing.T) {
	response := httptest.NewRecorder()

	WriteJSON(response, http.StatusCreated, map[string]string{"status": "ok"})

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("unexpected content type: %q", got)
	}

	var body map[string]string
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("unexpected body: %#v", body)
	}
}

func TestStatusFromDomainMapsKnownErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "invalid input", err: common.ErrInvalidInput, want: http.StatusBadRequest},
		{name: "not found", err: common.ErrNotFound, want: http.StatusNotFound},
		{name: "conflict", err: common.ErrConflict, want: http.StatusConflict},
		{name: "not implemented", err: common.ErrNotImplemented, want: http.StatusNotImplemented},
		{name: "wrapped", err: errors.Join(errors.New("repo"), common.ErrConflict), want: http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := StatusFromDomain(tt.err)
			if !ok {
				t.Fatal("expected domain error to be mapped")
			}
			if got != tt.want {
				t.Fatalf("expected status %d, got %d", tt.want, got)
			}
		})
	}
}

func TestStatusFromDomainIgnoresUnknownError(t *testing.T) {
	if code, ok := StatusFromDomain(errors.New("unknown")); ok || code != 0 {
		t.Fatalf("expected unknown error to be ignored, got code=%d ok=%v", code, ok)
	}
}
