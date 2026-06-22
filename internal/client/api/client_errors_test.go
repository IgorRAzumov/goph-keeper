package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"goph-keeper/internal/client/api"
)

func TestClientRegisterAPIError(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "conflict"})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	_, err := api.NewClient(srv.URL).Register(context.Background(), "a", "b", "salt")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClientPullError(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sync/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	_, err := api.NewClient(srv.URL).Pull(context.Background(), "bad", 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClientLoginMissingToken(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	_, _, _, err := api.NewClient(srv.URL).Login(context.Background(), "a", "b")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAPIErrorIsConflictThroughWrap(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("decode failed: %w", &api.Error{Status: http.StatusConflict, Message: "version mismatch"})
	if !errors.Is(wrapped, api.ErrConflict) {
		t.Fatal("expected ErrConflict through error wrap chain")
	}
}
