package dtos

import (
	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
)

type RegisterFormRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type LoginFormRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshFormRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Username string `json:"username"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func ToUserResponse(u db.User) UserResponse {
	return UserResponse{
		ID:       u.ID,
		Email:    u.Email,
		Role:     string(u.Role),
		Username: u.Username,
	}
}
