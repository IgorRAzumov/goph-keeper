package session

import (
	"context"

	"goph-keeper/internal/client/application/state"
	"goph-keeper/internal/client/model"
)

// Service проверяет и обновляет access/refresh токены.
type Service struct {
	State *state.State
}

// New создаёт сервис сессии.
func New(state *state.State) *Service {
	return &Service{State: state}
}

// Ensure возвращает nil, если есть валидная сессия (access или refresh).
func (service *Service) Ensure(ctx context.Context) error {
	if service.State.Config.AccessToken != "" {
		return nil
	}
	if service.State.Config.RefreshToken != "" {
		return service.Refresh(ctx)
	}
	return model.ErrUnauthorized
}

// Refresh обновляет пару токенов по refresh_token.
func (service *Service) Refresh(ctx context.Context) error {
	if service.State.Config.RefreshToken == "" {
		return model.ErrUnauthorized
	}
	access, refresh, err := service.State.API.Refresh(ctx, service.State.Config.RefreshToken)
	if err != nil {
		return err
	}
	service.State.Config.AccessToken = access
	service.State.Config.RefreshToken = refresh
	return service.State.Save()
}
