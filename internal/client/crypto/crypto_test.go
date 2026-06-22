package crypto_test

import (
	"bytes"
	"testing"

	"goph-keeper/internal/client/crypto"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()

	salt, err := crypto.NewSalt()
	if err != nil {
		t.Fatal(err)
	}
	key, err := crypto.DeriveKey("master", salt)
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte(`{"text":"secret"}`)
	cipher, err := crypto.Encrypt(key, plain)
	if err != nil {
		t.Fatal(err)
	}
	got, err := crypto.Decrypt(key, cipher)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plain, got) {
		t.Fatalf("expected %q, got %q", plain, got)
	}
}

func TestDecryptRejectsShortCiphertext(t *testing.T) {
	t.Parallel()

	salt, _ := crypto.NewSalt()
	key, _ := crypto.DeriveKey("master", salt)
	_, err := crypto.Decrypt(key, []byte{1, 2, 3})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeriveKeyRejectsEmptyPassword(t *testing.T) {
	t.Parallel()

	_, err := crypto.DeriveKey("", "c2FsdA==")
	if err == nil {
		t.Fatal("expected error")
	}
}
