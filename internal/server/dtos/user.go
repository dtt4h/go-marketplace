package dtos

import (
	"time"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
)

type UpdateProfileRequest struct {
	Username  *string `json:"username,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type CreateStoreRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
}

type ProfileResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type StoreResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	LogoURL     *string   `json:"logo_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func ToProfileResponse(u db.User) ProfileResponse {
	return ProfileResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      string(u.Role),
		Username:  u.Username,
		AvatarURL: pgutil.TextToPtr(u.AvatarUrl),
		Phone:     pgutil.TextToPtr(u.Phone),
		CreatedAt: u.CreatedAt.Time,
	}
}

func ToStoreResponse(s db.Store) StoreResponse {
	return StoreResponse{
		ID:          s.ID,
		UserID:      s.UserID,
		Name:        s.Name,
		Description: pgutil.TextToPtr(s.Description),
		LogoURL:     pgutil.TextToPtr(s.LogoUrl),
		CreatedAt:   s.CreatedAt.Time,
	}
}
