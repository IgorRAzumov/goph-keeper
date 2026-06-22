package repository

//go:generate go run go.uber.org/mock/mockgen@latest -destination=./mocks/session_repository_mock.go -package=mocks goph-keeper/internal/server/domain/session/repository SessionRepository
