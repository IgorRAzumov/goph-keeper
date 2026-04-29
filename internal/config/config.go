// Пакет config загружает конфигурацию приложения из переменных окружения и значений по умолчанию.
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
		HTTPAddr: getenv("GOPHKEEPER_ADDR", "127.0.0.1:8080"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
