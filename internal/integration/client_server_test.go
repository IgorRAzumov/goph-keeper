package integration_test

import (
	"context"
	"path/filepath"
	"testing"

	"goph-keeper/internal/client/api"
	clientapp "goph-keeper/internal/client/app"
	"goph-keeper/internal/client/model"
	"goph-keeper/internal/testutil/testserver"
)

func TestClientServerRegisterLoginSyncFlow(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master-secret")

	app, err := clientapp.New("master-secret")
	if err != nil {
		t.Fatal(err)
	}
	app.API = api.NewClient(fixture.Server.URL)

	if err := app.Register(context.Background(), "alice", "account-pass", fixture.Server.URL); err != nil {
		t.Fatalf("register: %v", err)
	}
	if app.Config.AccessToken == "" {
		t.Fatal("expected access token after register")
	}

	id, err := app.Add(context.Background(), clientapp.AddInput{
		Type: model.RecordTypeText, Meta: "note", Payload: "hello integration",
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	got, err := app.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Payload != `{"text":"hello integration"}` {
		t.Fatalf("unexpected payload: %s", got.Payload)
	}

	// второй клиент с тем же аккаунтом подтягивает sync
	dir2 := filepath.Join(dir, "client2")
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir2)
	app2, err := clientapp.New("master-secret")
	if err != nil {
		t.Fatal(err)
	}
	app2.API = api.NewClient(fixture.Server.URL)
	if err := app2.Login(context.Background(), "alice", "account-pass", fixture.Server.URL); err != nil {
		t.Fatalf("login client2: %v", err)
	}
	// скопировать соль с первого клиента (в реальности та же учётка — та же соль при первом login)
	app2.Config.MasterSalt = app.Config.MasterSalt
	_ = app2.Save()

	records, err := app2.List(context.Background())
	if err != nil {
		t.Fatalf("list client2: %v", err)
	}
	if len(records) != 1 || records[0].ID != id {
		t.Fatalf("expected synced record, got %+v", records)
	}
}

func TestClientServerLogout(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")

	app, err := clientapp.New("master")
	if err != nil {
		t.Fatal(err)
	}
	app.API = api.NewClient(fixture.Server.URL)
	if err := app.Register(context.Background(), "bob", "pass", fixture.Server.URL); err != nil {
		t.Fatal(err)
	}
	if err := app.Logout(context.Background()); err != nil {
		t.Fatal(err)
	}
	if app.Config.AccessToken != "" {
		t.Fatal("expected cleared tokens")
	}
}
