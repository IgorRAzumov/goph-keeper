package config_test

import (
	"testing"

	"goph-keeper/internal/client/config"
)

func TestDataPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	path, err := config.DataPath()
	if err != nil {
		t.Fatal(err)
	}
	if path == "" {
		t.Fatal("expected path")
	}
}
