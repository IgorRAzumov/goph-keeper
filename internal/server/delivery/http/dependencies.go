package httpapi

import (
	applogin "goph-keeper/internal/server/application/auth/login"
	applogout "goph-keeper/internal/server/application/auth/logout"
	apprefresh "goph-keeper/internal/server/application/auth/refresh"
	appregister "goph-keeper/internal/server/application/auth/register"
	appverify "goph-keeper/internal/server/application/auth/verify"
	appsync "goph-keeper/internal/server/application/sync"
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
	// Verify — проверка access-токена и сессии для защищённых маршрутов (middleware).
	Verify *appverify.Usecase
	// Sync — pull/push зашифрованных записей владельца.
	Sync *appsync.Usecase
}
