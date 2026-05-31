package repository

//go:generate go run go.uber.org/mock/mockgen@latest -destination=./mocks/record_repository_mock.go -package=mocks goph-keeper/internal/domain/record/repository RecordRepository
