package common

import "github.com/google/uuid"

// NewID генерирует непрозрачный идентификатор (UUID v4).
func NewID() (string, error) {
	return uuid.NewString(), nil
}
