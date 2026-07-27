package dtos

import (
	"time"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
)

type CreatePaymentRequest struct {
	OrderID int64 `json:"order_id"`
}

type PaymentResponse struct {
	ID                int64     `json:"id"`
	OrderID           int64     `json:"order_id"`
	Amount            string    `json:"amount"`
	Currency          string    `json:"currency"`
	Status            string    `json:"status"`
	Provider          string    `json:"provider"`
	ProviderPaymentID *string   `json:"provider_payment_id,omitempty"`
	ConfirmationURL   *string   `json:"confirmation_url,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type WebhookRequest struct {
	OrderID           int64  `json:"order_id"`
	ProviderPaymentID string `json:"provider_payment_id"`
	Status            string `json:"status"`
}

func ToPaymentResponse(p db.Payment) PaymentResponse {
	resp := PaymentResponse{
		ID:       p.ID,
		OrderID:  p.OrderID,
		Amount:   NumericToStr(p.Amount),
		Currency: p.Currency,
		Status:   string(p.Status),
		Provider: p.Provider,
		CreatedAt: p.CreatedAt.Time,
		UpdatedAt: p.UpdatedAt.Time,
	}

	if p.ProviderPaymentID.Valid {
		resp.ProviderPaymentID = &p.ProviderPaymentID.String
	}

	return resp
}
