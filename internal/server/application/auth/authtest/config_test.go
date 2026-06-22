package authtest_test

import (
	"testing"

	"goph-keeper/internal/server/application/auth/authtest"
)

func TestProviderAndDefaults(t *testing.T) {
	if authtest.JWTSecret == "" || authtest.AccessTTL == 0 || authtest.RefreshTTL == 0 {
		t.Fatalf("unexpected defaults: secret=%q access=%v refresh=%v",
			authtest.JWTSecret, authtest.AccessTTL, authtest.RefreshTTL)
	}
	if authtest.Provider() == nil {
		t.Fatal("expected non-nil provider")
	}
}
