package httpapi

import (
	"goph-keeper/internal/application/user"
)

// Dependensies содержит сценарии и сервисы application-слоя для HTTP-слоя доставки.
type Dependensies struct {
	// RegisterUser связывает регистрацию пользователя: HTTP → application.
	RegisterUser *user.Usecase
}
