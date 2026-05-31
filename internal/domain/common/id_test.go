package common

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewIDReturnsUUID(t *testing.T) {
	id, err := NewID()
	if err != nil {
		t.Fatalf("NewID failed: %v", err)
	}

	if _, err := uuid.Parse(id); err != nil {
		t.Fatalf("expected uuid, got %q: %v", id, err)
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
