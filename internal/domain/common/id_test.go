package common

import (
	"encoding/hex"
	"testing"
)

func TestNewIDReturnsHexEncoded16Bytes(t *testing.T) {
	id, err := NewID()
	if err != nil {
		t.Fatalf("NewID failed: %v", err)
	}

	if len(id) != 32 {
		t.Fatalf("expected 32 hex chars, got %d", len(id))
	}
	if _, err := hex.DecodeString(id); err != nil {
		t.Fatalf("expected hex encoded id, got %q: %v", id, err)
	}
}

func TestNewIDReturnsDifferentValues(t *testing.T) {
	first, err := NewID()
	if err != nil {
		t.Fatalf("first NewID failed: %v", err)
	}
	second, err := NewID()
	if err != nil {
		t.Fatalf("second NewID failed: %v", err)
	}

	if first == second {
		t.Fatalf("expected different ids, got %q", first)
	}
}
