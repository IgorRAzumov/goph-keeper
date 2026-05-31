package login

// Input содержит данные для авторизации пользователя.
type Input struct {
	Login    string
	Password string
}

// Output возвращается после успешной авторизации.
type Output struct {
	AccessToken  string
	RefreshToken string
}
