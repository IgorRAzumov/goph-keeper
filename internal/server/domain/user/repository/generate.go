package repository

//go:generate go run go.uber.org/mock/mockgen@latest -destination=./mocks/user_repository_mock.go -package=mocks goph-keeper/internal/server/domain/user/repository UserRepository
