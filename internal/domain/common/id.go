package common

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID генерирует непрозрачный идентификатор (32 hex-символа).
func NewID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
