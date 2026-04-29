package httpapi

import (
	"goph-keeper/internal/application/user"
)

// Deps содержит сценарии и сервисы application-слоя для HTTP-слоя доставки.
// Эту структуру создаёт корень композиции; обработчики не должны создавать сценарии сами.
type Deps struct {
	// RegisterUser связывает регистрацию пользователя: HTTP → application.
	RegisterUser *user.RegisterUserUseCase
}
