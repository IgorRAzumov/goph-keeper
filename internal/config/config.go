package config

import "os"

// Config содержит настройки процесса для API-сервера и будущих адаптеров.
type Config struct {
	// HTTPAddr — TCP-адрес для HTTP-листенера (например "127.0.0.1:8080").
	HTTPAddr string
}

// Load читает конфигурацию из переменных окружения, подставляя значения по умолчанию при отсутствии.
func Load() Config {
	return Config{
		HTTPAddr: getEnv("GOPHKEEPER_ADDR", "127.0.0.1:8080"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
