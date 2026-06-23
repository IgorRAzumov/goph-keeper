package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	clientapp "goph-keeper/internal/client/app"
	"goph-keeper/internal/client/model"

	"golang.org/x/term"
)

func openApp() (*clientapp.App, error) {
	master, err := readMasterPassword()
	if err != nil {
		return nil, err
	}
	app, err := clientapp.New(master, "")
	if err != nil {
		return nil, err
	}
	if app.Config.AccessToken == "" && app.Config.RefreshToken == "" {
		return nil, errors.New("not logged in: run login first")
	}
	return app, nil
}

func readMasterPassword() (string, error) {
	if masterPassword := strings.TrimSpace(os.Getenv("GOPHKEEPER_MASTER_PASSWORD")); masterPassword != "" {
		return masterPassword, nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", errors.New("set GOPHKEEPER_MASTER_PASSWORD for non-interactive mode")
	}
	return readPassword("master password: ")
}

func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	if len(password) == 0 {
		return "", model.ErrInvalidInput
	}
	return string(password), nil
}
