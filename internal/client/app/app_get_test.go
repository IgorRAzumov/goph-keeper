package app_test

import (
	"context"
	"errors"
	"testing"

	clientapp "goph-keeper/internal/client/app"
	"goph-keeper/internal/client/model"
	"goph-keeper/internal/testutil/testserver"
)

func TestAppGetNotFound(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")

	app, err := clientapp.New("master", fixture.Server.URL)
	if err != nil {
		t.Fatal(err)
	}
	_ = app.Register(context.Background(), "nf", "pass")
	_, err = app.Get(context.Background(), "00000000-0000-4000-8000-000000000001")
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
