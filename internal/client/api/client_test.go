package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"goph-keeper/internal/client/api"
	"goph-keeper/internal/contract"
)

func TestClientLoginAndPull(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"access_token": "acc", "refresh_token": "ref", "master_salt": "salt",
		})
	})
	mux.HandleFunc("/api/v1/sync/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"records": []contract.Record{{ID: "r1", Type: "text", Version: 1}},
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := api.NewClient(srv.URL)
	access, refresh, masterSalt, err := client.Login(context.Background(), "alice", "secret")
	if err != nil || access != "acc" || refresh != "ref" || masterSalt != "salt" {
		t.Fatalf("login: access=%q refresh=%q err=%v", access, refresh, err)
	}
	records, err := client.Pull(context.Background(), access, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].ID != "r1" {
		t.Fatalf("unexpected records: %+v", records)
	}
}

func TestClientPushConflict(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sync/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"conflicts": []contract.Record{{ID: "r1", Version: 99}},
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := api.NewClient(srv.URL)
	conflicts, err := client.Push(context.Background(), "token", []contract.Record{{ID: "r1", Version: 1}})
	if err != api.ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if len(conflicts) != 1 || conflicts[0].Version != 99 {
		t.Fatalf("unexpected conflicts: %+v", conflicts)
	}
}
