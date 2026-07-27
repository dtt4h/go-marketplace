package payments

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrPaymentExists      = errors.New("payment already exists for this order")
	ErrPaymentNotFound    = errors.New("payment not found")
	ErrAlreadyPaid        = errors.New("order is already paid")
	ErrInvalidWebhook     = errors.New("invalid webhook payload")
	ErrRefundFailed       = errors.New("refund failed")
)

type OrderFetcher interface {
	GetOrder(ctx context.Context, orderID int64) (db.Order, error)
	UpdateOrderStatus(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error)
}

type PaymentService interface {
	CreatePayment(ctx context.Context, userID int64, req dtos.CreatePaymentRequest) (dtos.PaymentResponse, error)
	GetPayment(ctx context.Context, userID, paymentID int64) (dtos.PaymentResponse, error)
	HandleWebhook(ctx context.Context, req dtos.WebhookRequest) error
	RefundPayment(ctx context.Context, userID, paymentID int64) (dtos.PaymentResponse, error)
}

type paymentService struct {
	repo   PaymentRepository
	orders OrderFetcher
	pool   *pgxpool.Pool
}

func NewPaymentService(repo PaymentRepository, orders OrderFetcher, pool *pgxpool.Pool) PaymentService {
	return &paymentService{repo: repo, orders: orders, pool: pool}
}

func (s *paymentService) CreatePayment(ctx context.Context, userID int64, req dtos.CreatePaymentRequest) (dtos.PaymentResponse, error) {
	order, err := s.orders.GetOrder(ctx, req.OrderID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.PaymentResponse{}, ErrOrderNotFound
		}
		return dtos.PaymentResponse{}, fmt.Errorf("get order: %w", err)
	}

	if order.UserID != userID {
		return dtos.PaymentResponse{}, ErrOrderNotFound
	}

	if order.Status != db.OrderStatusPending {
		return dtos.PaymentResponse{}, ErrAlreadyPaid
	}

	existing, err := s.repo.GetPaymentByOrderID(ctx, req.OrderID)
	if err == nil && existing.Status != db.PaymentStatusFailed {
		return dtos.PaymentResponse{}, ErrPaymentExists
	}

	amountStr := dtos.NumericToStr(order.Total)
	amount, err := toNumeric(amountStr)
	if err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("parse amount: %w", err)
	}

	payment, err := s.repo.CreatePayment(ctx, db.CreatePaymentParams{
		OrderID:           req.OrderID,
		Amount:            amount,
		Currency:          "RUB",
		Provider:          "mock",
		ProviderPaymentID: pgtype.Text{Valid: false},
	})
	if err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("create payment: %w", err)
	}

	resp := dtos.ToPaymentResponse(payment)
	confirmationURL := "https://pay.mock-gateway.ru/checkout/" + fmt.Sprintf("%d", payment.ID)
	resp.ConfirmationURL = &confirmationURL

	return resp, nil
}

func (s *paymentService) GetPayment(ctx context.Context, userID, paymentID int64) (dtos.PaymentResponse, error) {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.PaymentResponse{}, ErrPaymentNotFound
		}
		return dtos.PaymentResponse{}, fmt.Errorf("get payment: %w", err)
	}

	order, err := s.orders.GetOrder(ctx, payment.OrderID)
	if err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("get order: %w", err)
	}

	if order.UserID != userID {
		return dtos.PaymentResponse{}, ErrPaymentNotFound
	}

	return dtos.ToPaymentResponse(payment), nil
}

func (s *paymentService) HandleWebhook(ctx context.Context, req dtos.WebhookRequest) error {
	var status db.PaymentStatus
	switch req.Status {
	case "succeeded":
		status = db.PaymentStatusSucceeded
	case "failed":
		status = db.PaymentStatusFailed
	case "refunded":
		status = db.PaymentStatusRefunded
	default:
		return ErrInvalidWebhook
	}

	payment, err := s.repo.GetPaymentByOrderID(ctx, req.OrderID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return ErrPaymentNotFound
		}
		return fmt.Errorf("get payment by order: %w", err)
	}

	_, err = s.repo.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
		ID:     payment.ID,
		Status: status,
		ProviderPaymentID: pgtype.Text{String: req.ProviderPaymentID, Valid: req.ProviderPaymentID != ""},
	})
	if err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	if status == db.PaymentStatusSucceeded {
		if _, err := s.orders.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
			ID:     req.OrderID,
			Status: db.OrderStatusPaid,
		}); err != nil {
			return fmt.Errorf("update order status to paid: %w", err)
		}
	}

	return nil
}

func (s *paymentService) RefundPayment(ctx context.Context, userID, paymentID int64) (dtos.PaymentResponse, error) {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.PaymentResponse{}, ErrPaymentNotFound
		}
		return dtos.PaymentResponse{}, fmt.Errorf("get payment: %w", err)
	}

	order, err := s.orders.GetOrder(ctx, payment.OrderID)
	if err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("get order: %w", err)
	}

	if order.UserID != userID {
		return dtos.PaymentResponse{}, ErrPaymentNotFound
	}

	if payment.Status != db.PaymentStatusSucceeded {
		return dtos.PaymentResponse{}, ErrRefundFailed
	}

	updated, err := s.repo.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
		ID:     payment.ID,
		Status: db.PaymentStatusRefunded,
	})
	if err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("refund payment: %w", err)
	}

	return dtos.ToPaymentResponse(updated), nil
}
