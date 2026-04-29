package config

import "testing"

func TestLoadUsesDefaultHTTPAddr(t *testing.T) {
	t.Setenv("GOPHKEEPER_ADDR", "")

	cfg := Load()

	if cfg.HTTPAddr != "127.0.0.1:8080" {
		t.Fatalf("expected default HTTP addr, got %q", cfg.HTTPAddr)
	}
}

func TestLoadUsesEnvHTTPAddr(t *testing.T) {
	t.Setenv("GOPHKEEPER_ADDR", "127.0.0.1:9090")

	cfg := Load()

	if cfg.HTTPAddr != "127.0.0.1:9090" {
		t.Fatalf("expected env HTTP addr, got %q", cfg.HTTPAddr)
	}
}
