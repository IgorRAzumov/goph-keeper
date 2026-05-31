package logging

// Logger — минимальный интерфейс структурного логгера.
type Logger interface {
	// Info пишет информационное сообщение.
	Info(msg string, args ...any)
	// Error пишет сообщение об ошибке.
	Error(msg string, args ...any)
}
