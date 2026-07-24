package dtos

import (
	"time"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
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
	AvatarURL *string   `json:"avatar_url"`
	Phone     *string   `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

type StoreResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	LogoURL     *string   `json:"logo_url"`
	CreatedAt   time.Time `json:"created_at"`
}

func ToProfileResponse(u db.User) ProfileResponse {
	return ProfileResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      string(u.Role),
		Username:  u.Username,
		AvatarURL: textToPtr(u.AvatarUrl),
		CreatedAt: u.CreatedAt.Time,
		Phone: textToPtr(u.Phone),
	}
}

func ToStoreResponse(s db.Store) StoreResponse {
	return StoreResponse{
		ID:          s.ID,
		UserID:      s.UserID,
		Name:        s.Name,
		Description: textToPtr(s.Description),
		LogoURL:     textToPtr(s.LogoUrl),
		CreatedAt:   s.CreatedAt.Time,
	}
}

func textToPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}

	return nil
}
