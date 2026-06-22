package app_test

import (
	"context"
	"errors"
	"testing"

	"goph-keeper/internal/client/api"
	clientapp "goph-keeper/internal/client/app"
	"goph-keeper/internal/client/model"
	"goph-keeper/internal/testutil/testserver"
)

func TestAppGetNotFound(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")

	app, err := clientapp.New("master")
	if err != nil {
		t.Fatal(err)
	}
	app.API = api.NewClient(fixture.Server.URL)
	_ = app.Register(context.Background(), "nf", "pass", fixture.Server.URL)
	_, err = app.Get(context.Background(), "00000000-0000-4000-8000-000000000001")
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
