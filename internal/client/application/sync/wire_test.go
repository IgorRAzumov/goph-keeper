package sync_test

import (
	"testing"

	"goph-keeper/internal/client/application/sync"
	"goph-keeper/internal/client/model"
	"goph-keeper/internal/client/store"
	"goph-keeper/internal/contract"
)

func TestWireConversionRoundTrip(t *testing.T) {
	t.Parallel()

	record := store.Record{
		ID:         "r1",
		Type:       model.RecordTypeText,
		Meta:       "meta",
		Ciphertext: []byte("cipher"),
		Version:    7,
		Deleted:    true,
	}

	wire := sync.RecordToWire(record)
	got, err := sync.WireToRecord(wire)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != record.ID || got.Type != record.Type || got.Version != record.Version || got.Deleted != record.Deleted {
		t.Fatalf("unexpected record after round-trip: %+v", got)
	}
	if string(got.Ciphertext) != string(record.Ciphertext) {
		t.Fatalf("ciphertext mismatch: %q vs %q", got.Ciphertext, record.Ciphertext)
	}
}

func TestWireToRecordRejectsInvalidBase64(t *testing.T) {
	t.Parallel()

	_, err := sync.WireToRecord(contract.Record{ID: "r1", Ciphertext: "%%%"})
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestWireToRecordAcceptsEmptyCiphertext(t *testing.T) {
	t.Parallel()

	got, err := sync.WireToRecord(contract.Record{ID: "r1", Ciphertext: ""})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Ciphertext) != 0 {
		t.Fatalf("expected empty bytes, got %v", got.Ciphertext)
	}
}
