package domain

import "time"

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"`
	User         *Account `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *RegisterRequest) Validate() error {
	if r.Email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}
	if len(r.Password) < 8 {
		return &ValidationError{Field: "password", Message: "password must be at least 8 characters"}
	}
	if r.DisplayName == "" {
		return &ValidationError{Field: "display_name", Message: "display_name is required"}
	}
	return nil
}

func (r *LoginRequest) Validate() error {
	if r.Email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}
	if r.Password == "" {
		return &ValidationError{Field: "password", Message: "password is required"}
	}
	return nil
}
