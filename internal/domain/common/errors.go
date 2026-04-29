package common

import "errors"

// Ошибки
var (
	// ErrNotImplemented -  порт/адаптер ещё не подключён.
	ErrNotImplemented = errors.New("not implemented")
	// ErrNotFound - запрошенная сущность не найдена.
	ErrNotFound = errors.New("not found")
	// ErrConflict - конфликт версий (optimistic locking) или уникальности.
	ErrConflict = errors.New("conflict")
	// ErrInvalidInput  - обязательные поля не заполнены или имеют неверный формат.
	ErrInvalidInput = errors.New("invalid input")
)
