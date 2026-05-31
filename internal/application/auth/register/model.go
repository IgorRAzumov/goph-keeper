package register

// Input содержит данные для регистрации пользователя.
type Input struct {
	Login    string
	Password string
}

// Output возвращается после успешной регистрации.
type Output struct {
	UserID string
}
