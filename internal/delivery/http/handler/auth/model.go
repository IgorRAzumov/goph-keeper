package auth

type registerUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
