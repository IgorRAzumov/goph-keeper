package contract

// RegisterRequest — тело POST /api/v1/auth/register.
type RegisterRequest struct {
	Login      string `json:"login"`
	Password   string `json:"password"`
	MasterSalt string `json:"master_salt"`
}

// RegisterResponse — ответ POST /api/v1/auth/register.
type RegisterResponse struct {
	UserID string `json:"userId"`
}

// LoginRequest — тело POST /api/v1/auth/login.
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// RefreshRequest — тело POST /api/v1/auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenResponse — ответ login и refresh (master_salt только при login).
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	MasterSalt   string `json:"master_salt,omitempty"`
}
