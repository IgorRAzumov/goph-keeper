package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// File хранит настройки клиента и токены сессии (без мастер-пароля).
type File struct {
	ServerURL    string `json:"server_url"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Login        string `json:"login,omitempty"`
	MasterSalt   string `json:"master_salt,omitempty"`
}

// DefaultServerURL — адрес API по умолчанию.
const DefaultServerURL = "http://127.0.0.1:8080"

// Dir возвращает каталог конфигурации (~/.gophkeeper или GOPHKEEPER_CONFIG_DIR).
func Dir() (string, error) {
	if custom := strings.TrimSpace(os.Getenv("GOPHKEEPER_CONFIG_DIR")); custom != "" {
		return custom, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gophkeeper"), nil
}

// Path возвращает путь к config.json.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// DataPath возвращает путь к локальному хранилищу записей.
func DataPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "data.json"), nil
}

// Load читает config.json; отсутствующий файл — дефолты без ошибки.
func Load() (File, error) {
	path, err := Path()
	if err != nil {
		return File{}, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return File{ServerURL: serverURLFromEnv()}, nil
	}
	if err != nil {
		return File{}, err
	}
	var cfg File
	if err := json.Unmarshal(data, &cfg); err != nil {
		return File{}, fmt.Errorf("config: parse: %w", err)
	}
	if strings.TrimSpace(cfg.ServerURL) == "" {
		cfg.ServerURL = serverURLFromEnv()
	}
	return cfg, nil
}

// Save записывает config.json (каталог создаётся при необходимости).
func Save(cfg File) error {
	if strings.TrimSpace(cfg.ServerURL) == "" {
		cfg.ServerURL = DefaultServerURL
	}
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func serverURLFromEnv() string {
	if v := strings.TrimSpace(os.Getenv("GOPHKEEPER_SERVER_URL")); v != "" {
		return v
	}
	return DefaultServerURL
}
