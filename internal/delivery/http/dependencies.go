package httpapi

import (
	applogin "goph-keeper/internal/application/auth/login"
	applogout "goph-keeper/internal/application/auth/logout"
	apprefresh "goph-keeper/internal/application/auth/refresh"
	appregister "goph-keeper/internal/application/auth/register"
	sessionrepo "goph-keeper/internal/domain/session/repository"
	"goph-keeper/internal/security/jwt"
)

// Dependencies содержит сценарии и сервисы application-слоя для HTTP-слоя доставки.
type Dependencies struct {
	// RegisterUser связывает регистрацию пользователя: HTTP → application.
	RegisterUser *appregister.Usecase
	// Login связывает логин: HTTP → application.
	Login *applogin.LoginUsecase
	// Refresh связывает refresh: HTTP → application.
	Refresh *apprefresh.RefreshUsecase
	// Logout завершает сессию по access JWT.
	Logout *applogout.Usecase
	// JWT — провайдер для middleware (проверка access-токена на защищённых маршрутах).
	JWT *jwt.Provider
	// Sessions — хранилище серверных сессий для проверки отозванных access-токенов.
	Sessions sessionrepo.SessionRepository
}
