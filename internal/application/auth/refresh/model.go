package refresh

// Input содержит данные для refresh-сценария.
type Input struct {
	RefreshToken string
}

// Output возвращается после успешного refresh.
type Output struct {
	AccessToken  string
	RefreshToken string
}
