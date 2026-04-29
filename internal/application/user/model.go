package user

// RegisterUserInput содержит данные для регистрации пользователя.
type RegisterUserInput struct {
	Login    string
	Password string
}

// RegisterUserOutput возвращается после успешной регистрации.
type RegisterUserOutput struct {
	UserID string
}
