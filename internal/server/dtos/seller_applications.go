package dtos

import (
	"time"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
)

type CreateSellerApplicationRequest struct {
	StoreName   string  `json:"store_name"`
	Description *string `json:"description,omitempty"`
}

type UpdateApplicationStatusRequest struct {
	Status string `json:"status"` // approved | rejected
	Reason string `json:"reason,omitempty"`
}

type SellerApplicationResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	StoreName   string    `json:"store_name"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToSellerApplicationResponse(app db.SellerApplication) SellerApplicationResponse {
	resp := SellerApplicationResponse{
		ID:        app.ID,
		UserID:    app.UserID,
		StoreName: app.StoreName,
		Status:    app.Status,
		CreatedAt: app.CreatedAt.Time,
		UpdatedAt: app.UpdatedAt.Time,
	}
	if app.Description.Valid {
		resp.Description = app.Description.String
	}
	return resp
}
