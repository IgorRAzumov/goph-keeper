package httpapi

import (
	"goph-keeper/internal/application/user"
)

// Dependencies содержит сценарии и сервисы application-слоя для HTTP-слоя доставки.
type Dependencies struct {
	// RegisterUser связывает регистрацию пользователя: HTTP → application.
	RegisterUser *user.Usecase
}
