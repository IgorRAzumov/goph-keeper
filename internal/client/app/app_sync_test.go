package app_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	clientapp "goph-keeper/internal/client/app"
	clientcfg "goph-keeper/internal/client/config"
	clientcrypto "goph-keeper/internal/client/crypto"
	"goph-keeper/internal/client/model"
	"goph-keeper/internal/testutil/testserver"
)

func TestAppSyncWithRefreshOnPull(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")

	app, err := clientapp.New("master", fixture.Server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Register(context.Background(), "refresh-user", "pass"); err != nil {
		t.Fatal(err)
	}
	// симулируем протухший access, оставляем refresh
	app.Config.AccessToken = "expired"
	_ = app.Save()
	if err := app.Sync(context.Background()); err != nil {
		t.Fatalf("sync after refresh: %v", err)
	}
}

func TestAppLogoutWithoutToken(t *testing.T) {
	app, err := clientapp.New("", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Logout(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestAppMasterKeyFromEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "from-env")

	cfg := clientcfg.File{MasterSalt: mustSalt(t)}
	if err := clientcfg.Save(cfg); err != nil {
		t.Fatal(err)
	}
	app, err := clientapp.New("", "")
	if err != nil {
		t.Fatal(err)
	}
	app.Config.AccessToken = "x"
	records, err := app.List(context.Background())
	if err != nil {
		t.Fatalf("list with env master key: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected empty list, got %d", len(records))
	}
}

func TestCLIPrerequisiteDataPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	path, err := clientcfg.DataPath()
	if err != nil || path == "" {
		t.Fatalf("data path: %q err=%v", path, err)
	}
}

func TestReadRecordIDFromStore(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")

	app, err := clientapp.New("master", fixture.Server.URL)
	if err != nil {
		t.Fatal(err)
	}
	_ = app.Register(context.Background(), "id-user", "pass")
	id, _ := app.Add(context.Background(), clientapp.AddInput{
		Type: model.RecordTypeText, Meta: "m", Payload: "x",
	})
	dataPath, _ := clientcfg.DataPath()
	raw, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(raw), id) {
		t.Fatalf("expected id in store file")
	}
	_ = filepath.Dir(dataPath)
}

func TestSyncRebasesDirtyOnGlobalVersionConflict(t *testing.T) {
	fixture := testserver.New(t)

	dirA := filepath.Join(t.TempDir(), "a")
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dirA)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")

	appA, err := clientapp.New("master", fixture.Server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := appA.Register(context.Background(), "rebase-user", "pass"); err != nil {
		t.Fatal(err)
	}
	if _, err := appA.Add(context.Background(), clientapp.AddInput{
		Type: model.RecordTypeText, Meta: "a1", Payload: "one",
	}); err != nil {
		t.Fatal(err)
	}

	// Второй клиент подтягивает состояние на версии 1.
	dirB := filepath.Join(t.TempDir(), "b")
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dirB)
	appB, err := clientapp.New("master", fixture.Server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := appB.Login(context.Background(), "rebase-user", "pass"); err != nil {
		t.Fatal(err)
	}
	appB.Config.MasterSalt = appA.Config.MasterSalt
	if err := appB.Save(); err != nil {
		t.Fatal(err)
	}
	if err := appB.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}

	// Первый клиент пишет новую запись с версией 2.
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dirA)
	if _, err := appA.Add(context.Background(), clientapp.AddInput{
		Type: model.RecordTypeText, Meta: "a2", Payload: "two",
	}); err != nil {
		t.Fatal(err)
	}

	// Второй клиент локально создаёт запись с той же глобальной версией 2.
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dirB)
	if _, err := appB.Add(context.Background(), clientapp.AddInput{
		Type: model.RecordTypeText, Meta: "b-local", Payload: "three",
	}); err != nil {
		t.Fatal(err)
	}

	// Sync должен сам rebasing-ом поднять версию и завершиться без конфликта.
	if err := appB.Sync(context.Background()); err != nil {
		t.Fatalf("sync with rebase failed: %v", err)
	}
	if len(slices.Collect(appB.Data.DirtyRecords())) != 0 {
		t.Fatalf("expected no dirty records after rebase sync, got %d", len(slices.Collect(appB.Data.DirtyRecords())))
	}
}

func mustSalt(t *testing.T) string {
	t.Helper()
	salt, err := clientcrypto.NewSalt()
	if err != nil {
		t.Fatal(err)
	}
	return salt
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && json.Valid([]byte(s)) && stringIndex(s, sub))
}

func stringIndex(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
