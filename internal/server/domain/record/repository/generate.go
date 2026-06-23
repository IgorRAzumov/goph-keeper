package repository

//go:generate go run go.uber.org/mock/mockgen@latest -destination=./mocks/record_store_mock.go -package=mocks goph-keeper/internal/server/domain/record/repository RecordStore
