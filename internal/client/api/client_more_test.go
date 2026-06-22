package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"goph-keeper/internal/client/api"
)

func TestClientRegister(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"userId": "u1"})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	id, err := api.NewClient(srv.URL).Register(context.Background(), "alice", "secret", "salt")
	if err != nil || id != "u1" {
		t.Fatalf("register: id=%q err=%v", id, err)
	}
}

func TestClientRefreshAndLogout(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"access_token": "new-acc", "refresh_token": "new-ref",
		})
	})
	mux.HandleFunc("/api/v1/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := api.NewClient(srv.URL)
	acc, ref, err := client.Refresh(context.Background(), "old-ref")
	if err != nil || acc != "new-acc" {
		t.Fatalf("refresh: %v", err)
	}
	if err := client.Logout(context.Background(), acc); err != nil {
		t.Fatalf("logout: %v", err)
	}
	_ = ref
}

func TestAPIErrorString(t *testing.T) {
	t.Parallel()

	err := &api.Error{Status: 400, Message: "bad request"}
	if err.Error() == "" {
		t.Fatal("expected error string")
	}
}
