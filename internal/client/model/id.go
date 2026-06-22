package model

import "github.com/google/uuid"

// NewID генерирует непрозрачный идентификатор записи (UUID v4).
func NewID() (string, error) {
	return uuid.NewString(), nil
}
