package sync

import (
	"encoding/base64"

	"goph-keeper/internal/client/model"
	"goph-keeper/internal/client/store"
	"goph-keeper/internal/contract"
)

// RecordToWire преобразует локальную запись в формат sync API.
func RecordToWire(record store.Record) contract.Record {
	return contract.Record{
		ID:         record.ID,
		Type:       string(record.Type),
		Meta:       record.Meta,
		Ciphertext: encodeB64(record.Ciphertext),
		Version:    record.Version,
		Deleted:    record.Deleted,
	}
}

// WireToRecord преобразует запись с провода в локальный формат.
func WireToRecord(wire contract.Record) (store.Record, error) {
	cipher, err := decodeB64(wire.Ciphertext)
	if err != nil {
		return store.Record{}, err
	}
	return store.Record{
		ID:         wire.ID,
		Type:       model.RecordType(wire.Type),
		Meta:       wire.Meta,
		Ciphertext: cipher,
		Version:    wire.Version,
		Deleted:    wire.Deleted,
	}, nil
}

func encodeB64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func decodeB64(s string) ([]byte, error) {
	if s == "" {
		return []byte{}, nil
	}
	return base64.StdEncoding.DecodeString(s)
}
