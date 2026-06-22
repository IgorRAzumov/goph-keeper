package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"goph-keeper/internal/client/config"
)

func TestLoadSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)

	cfg := config.File{
		ServerURL:    "http://example:9090",
		AccessToken:  "access",
		RefreshToken: "refresh",
		Login:        "alice",
		MasterSalt:   "salt",
	}
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	got, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.ServerURL != cfg.ServerURL || got.Login != cfg.Login {
		t.Fatalf("unexpected cfg: %+v", got)
	}
	path, err := config.Path()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("unexpected path dir: %s", path)
	}
}

func TestLoadMissingUsesDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_SERVER_URL", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerURL != config.DefaultServerURL {
		t.Fatalf("expected default server url, got %q", cfg.ServerURL)
	}
}

func TestLoadUsesServerURLFromEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_SERVER_URL", "http://custom:9999")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerURL != "http://custom:9999" {
		t.Fatalf("unexpected server url: %q", cfg.ServerURL)
	}
}

func TestDirUsesHomeWhenUnset(t *testing.T) {
	t.Setenv("GOPHKEEPER_CONFIG_DIR", "")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	dir, err := config.Dir()
	if err != nil {
		t.Fatal(err)
	}
	if dir != filepath.Join(home, ".gophkeeper") {
		t.Fatalf("unexpected dir: %s", dir)
	}
}
