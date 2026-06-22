package model

import "errors"

// Sentinel-ошибки клиента.
var (
	// ErrNotFound — запись отсутствует в локальном хранилище.
	ErrNotFound = errors.New("not found")
	// ErrInvalidInput — обязательные поля не заполнены или имеют неверный формат.
	ErrInvalidInput = errors.New("invalid input")
	// ErrUnauthorized — нет действующей сессии: требуется вход.
	ErrUnauthorized = errors.New("unauthorized")
)
