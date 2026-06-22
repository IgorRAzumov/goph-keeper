package api

import (
	"context"
	"fmt"
	"net/http"

	"goph-keeper/internal/contract"
)

// Register регистрирует пользователя.
func (client *Client) Register(ctx context.Context, login, password, masterSalt string) (string, error) {
	var out contract.RegisterResponse
	if err := client.postJSON(ctx, "/api/v1/auth/register", contract.RegisterRequest{
		Login: login, Password: password, MasterSalt: masterSalt,
	}, "", &out); err != nil {
		return "", err
	}
	if out.UserID == "" {
		return "", fmt.Errorf("api: register: empty user id")
	}
	return out.UserID, nil
}

// Login возвращает пару токенов и master_salt.
func (client *Client) Login(ctx context.Context, login, password string) (access, refresh, masterSalt string, err error) {
	var out contract.TokenResponse
	if err := client.postJSON(ctx, "/api/v1/auth/login", contract.LoginRequest{
		Login: login, Password: password,
	}, "", &out); err != nil {
		return "", "", "", err
	}
	if out.AccessToken == "" {
		return "", "", "", fmt.Errorf("api: login: empty access token")
	}
	return out.AccessToken, out.RefreshToken, out.MasterSalt, nil
}

// Refresh обновляет пару токенов.
func (client *Client) Refresh(ctx context.Context, refreshToken string) (access, refresh string, err error) {
	var out contract.TokenResponse
	if err := client.postJSON(ctx, "/api/v1/auth/refresh", contract.RefreshRequest{
		RefreshToken: refreshToken,
	}, "", &out); err != nil {
		return "", "", err
	}
	if out.AccessToken == "" {
		return "", "", fmt.Errorf("api: refresh: empty access token")
	}
	return out.AccessToken, out.RefreshToken, nil
}

// Logout отзывает сессию.
func (client *Client) Logout(ctx context.Context, accessToken string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.BaseURL+"/api/v1/auth/logout", http.NoBody)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response, err := client.do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusNoContent {
		return nil
	}
	return readAPIError(response)
}
