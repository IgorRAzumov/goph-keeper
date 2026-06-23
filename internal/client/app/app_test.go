package app_test

import (
	"context"
	"testing"

	clientapp "goph-keeper/internal/client/app"
	"goph-keeper/internal/client/model"
	"goph-keeper/internal/testutil/testserver"
)

func TestAppAddListDelete(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")

	app, err := clientapp.New("master", fixture.Server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Register(context.Background(), "user1", "pass"); err != nil {
		t.Fatal(err)
	}

	id, err := app.Add(context.Background(), clientapp.AddInput{
		Type: model.RecordTypeLogin, Meta: "site",
		Payload: map[string]string{"login": "u", "password": "p"},
	})
	if err != nil {
		t.Fatal(err)
	}

	list, err := app.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 record, got %d", len(list))
	}

	if err := app.Delete(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	list, err = app.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 after delete, got %d", len(list))
	}
}

func TestAppSyncPull(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "m")

	app, err := clientapp.New("m", fixture.Server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Login(context.Background(), "sync-user", "pass"); err != nil {
		if err := app.Register(context.Background(), "sync-user", "pass"); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestAppRequiresMasterPassword(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "")

	app, err := clientapp.New("", "")
	if err != nil {
		t.Fatal(err)
	}
	app.Config.AccessToken = "tok"
	_, err = app.List(context.Background())
	if err == nil {
		t.Fatal("expected master password error")
	}
}
