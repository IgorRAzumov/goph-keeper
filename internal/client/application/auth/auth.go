package auth

import (
	"context"
	"strings"

	"goph-keeper/internal/client/application/session"
	"goph-keeper/internal/client/application/state"
	clientcrypto "goph-keeper/internal/client/crypto"
)

// Usecase выполняет сценарии регистрации и аутентификации.
type Usecase struct {
	State   *state.State
	Session *session.Service
}

// New создаёт auth usecase.
func New(state *state.State, session *session.Service) *Usecase {
	return &Usecase{State: state, Session: session}
}

// Register регистрирует пользователя на сервере и выполняет login.
func (usecase *Usecase) Register(ctx context.Context, login, password string) error {
	masterSalt := strings.TrimSpace(usecase.State.Config.MasterSalt)
	if masterSalt == "" {
		salt, err := clientcrypto.NewSalt()
		if err != nil {
			return err
		}
		masterSalt = salt
	}
	if _, err := usecase.State.API.Register(ctx, login, password, masterSalt); err != nil {
		return err
	}
	usecase.State.Config.MasterSalt = masterSalt
	return usecase.Login(ctx, login, password)
}

// Login аутентифицирует пользователя и сохраняет токены.
func (usecase *Usecase) Login(ctx context.Context, login, password string) error {
	access, refresh, masterSalt, err := usecase.State.API.Login(ctx, login, password)
	if err != nil {
		return err
	}
	usecase.State.Config.Login = login
	usecase.State.Config.AccessToken = access
	usecase.State.Config.RefreshToken = refresh
	if strings.TrimSpace(masterSalt) != "" {
		usecase.State.Config.MasterSalt = strings.TrimSpace(masterSalt)
	}
	return usecase.State.Save()
}

// Logout завершает сессию на сервере и очищает токены локально.
func (usecase *Usecase) Logout(ctx context.Context) error {
	if usecase.State.Config.AccessToken == "" {
		return nil
	}
	err := usecase.State.API.Logout(ctx, usecase.State.Config.AccessToken)
	usecase.State.Config.AccessToken = ""
	usecase.State.Config.RefreshToken = ""
	_ = usecase.State.Save()
	return err
}
