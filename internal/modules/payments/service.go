package payments

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/modules/notifications"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
)

var (
	ErrOrderNotFound   = errors.New("order not found")
	ErrPaymentExists   = errors.New("payment already exists for this order")
	ErrPaymentNotFound = errors.New("payment not found")
	ErrAlreadyPaid     = errors.New("order is already paid")
	ErrInvalidWebhook  = errors.New("invalid webhook payload")
	ErrRefundFailed    = errors.New("refund failed")
)

// OrderFetcher provides order lookup capabilities.
type OrderFetcher interface {
	GetOrder(ctx context.Context, orderID int64) (db.Order, error)
	GetOrderByPublicToken(ctx context.Context, token string) (db.Order, error)
	UpdateOrderStatus(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error)
	GetOrderItems(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error)
}

// UserFetcher provides user lookup.
type UserFetcher interface {
	GetUserByID(ctx context.Context, id int64) (db.User, error)
}

// PaymentService defines business logic for payment operations.
type PaymentService interface {
	CreatePayment(ctx context.Context, userID int64, req dtos.CreatePaymentRequest) (dtos.PaymentResponse, error)
	CreateGuestPayment(ctx context.Context, publicToken string) (dtos.PaymentResponse, error)
	GetPayment(ctx context.Context, userID, paymentID int64) (dtos.PaymentResponse, error)
	ListPaymentsByUser(ctx context.Context, userID int64, page, limit int) ([]dtos.PaymentListItem, int64, error)
	HandleWebhook(ctx context.Context, req dtos.WebhookRequest) error
	RefundPayment(ctx context.Context, userID, paymentID int64) (dtos.PaymentResponse, error)
}

type paymentService struct {
	repo          PaymentRepository
	orders        OrderFetcher
	users         UserFetcher
	pool          *pgxpool.Pool
	notifications notifications.NotificationService
	log           *slog.Logger
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(repo PaymentRepository, orders OrderFetcher, users UserFetcher, pool *pgxpool.Pool, notifications notifications.NotificationService, log *slog.Logger) PaymentService {
	return &paymentService{repo: repo, orders: orders, users: users, pool: pool, notifications: notifications, log: log}
}

func (s *paymentService) CreatePayment(ctx context.Context, userID int64, req dtos.CreatePaymentRequest) (dtos.PaymentResponse, error) {
	order, err := s.orders.GetOrder(ctx, req.OrderID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.PaymentResponse{}, ErrOrderNotFound
		}
		return dtos.PaymentResponse{}, fmt.Errorf("get order: %w", err)
	}

	if !order.UserID.Valid || order.UserID.Int64 != userID {
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
	amount, err := pgutil.ToNumeric(&amountStr)
	if err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("parse amount: %w", err)
	}

	payment, err := s.repo.CreatePayment(ctx, db.CreatePaymentParams{
		OrderID:           req.OrderID,
		Amount:            amount,
		Currency:          "RUB",
		Provider:          "mock",
		ProviderPaymentID: pgutil.NullText(nil),
	})
	if err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("create payment: %w", err)
	}

	resp := dtos.ToPaymentResponse(payment)
	confirmationURL := "https://pay.mock-gateway.ru/checkout/" + fmt.Sprintf("%d", payment.ID)
	resp.ConfirmationURL = &confirmationURL

	return resp, nil
}

func (s *paymentService) CreateGuestPayment(ctx context.Context, publicToken string) (dtos.PaymentResponse, error) {
	order, err := s.orders.GetOrderByPublicToken(ctx, publicToken)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.PaymentResponse{}, ErrOrderNotFound
		}
		return dtos.PaymentResponse{}, fmt.Errorf("get order by public token: %w", err)
	}

	if order.Status != db.OrderStatusPending {
		return dtos.PaymentResponse{}, ErrAlreadyPaid
	}

	existing, err := s.repo.GetPaymentByOrderID(ctx, order.ID)
	if err == nil && existing.Status != db.PaymentStatusFailed {
		return dtos.PaymentResponse{}, ErrPaymentExists
	}

	amountStr := dtos.NumericToStr(order.Total)
	amount, err := pgutil.ToNumeric(&amountStr)
	if err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("parse amount: %w", err)
	}

	payment, err := s.repo.CreatePayment(ctx, db.CreatePaymentParams{
		OrderID:           order.ID,
		Amount:            amount,
		Currency:          "RUB",
		Provider:          "mock",
		ProviderPaymentID: pgutil.NullText(nil),
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

	if !order.UserID.Valid || order.UserID.Int64 != userID {
		return dtos.PaymentResponse{}, ErrPaymentNotFound
	}

	return dtos.ToPaymentResponse(payment), nil
}

func validateWebhookRequest(req dtos.WebhookRequest) (db.PaymentStatus, error) {
	switch req.Status {
	case "succeeded":
		return db.PaymentStatusSucceeded, nil
	case "failed":
		return db.PaymentStatusFailed, nil
	case "refunded":
		return db.PaymentStatusRefunded, nil
	default:
		return "", ErrInvalidWebhook
	}
}

func formatOrderID(id int64) string {
	return fmt.Sprintf("%d", id)
}

func (s *paymentService) HandleWebhook(ctx context.Context, req dtos.WebhookRequest) error {
	status, err := validateWebhookRequest(req)
	if err != nil {
		return err
	}

	payment, err := s.repo.GetPaymentByOrderID(ctx, req.OrderID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return ErrPaymentNotFound
		}
		return fmt.Errorf("get payment by order: %w", err)
	}

	_, err = s.repo.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
		ID:                payment.ID,
		Status:            status,
		ProviderPaymentID: pgutil.NullText(&req.ProviderPaymentID),
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

		// Send receipt and status update to buyer
		go func() {
			order, orderErr := s.orders.GetOrder(ctx, req.OrderID)
			if orderErr != nil {
				s.log.Error("webhook: failed to get order for receipt", slog.Int64("order_id", req.OrderID), slog.String("error", orderErr.Error()))
				return
			}

			if !order.UserID.Valid {
				return
			}

			user, userErr := s.users.GetUserByID(ctx, order.UserID.Int64)
			if userErr != nil {
				s.log.Error("webhook: failed to get user for receipt", slog.Int64("user_id", order.UserID.Int64), slog.String("error", userErr.Error()))
				return
			}

			orderItems, itemsErr := s.orders.GetOrderItems(ctx, req.OrderID)
			if itemsErr != nil {
				s.log.Error("webhook: failed to get order items", slog.Int64("order_id", req.OrderID), slog.String("error", itemsErr.Error()))
				return
			}

			notifItems := notifications.BuildOrderItems(orderItems)

			// Send status update
			if sendErr := s.notifications.SendOrderStatusUpdate(order.ID, user.Email, user.Username, "paid"); sendErr != nil {
				s.log.Error("webhook: failed to send status update", slog.String("error", sendErr.Error()))
			}

			// Send receipt
			receiptHTML := s.notifications.GenerateReceiptHTML(order.ID, user.Username, user.Email, "", order.Address, dtos.NumericToStr(order.Total), notifItems)
			if sendErr := s.notifications.SendReceipt(user.Email, user.Username, formatOrderID(order.ID), dtos.NumericToStr(order.Total), "", receiptHTML); sendErr != nil {
				s.log.Error("webhook: failed to send receipt", slog.String("error", sendErr.Error()))
			}
		}()
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

	if !order.UserID.Valid || order.UserID.Int64 != userID {
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

func (s *paymentService) ListPaymentsByUser(ctx context.Context, userID int64, page, limit int) ([]dtos.PaymentListItem, int64, error) {
	payments, total, err := s.repo.ListPaymentsByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("list payments by user: %w", err)
	}

	if len(payments) == 0 {
		return []dtos.PaymentListItem{}, 0, nil
	}

	result := make([]dtos.PaymentListItem, 0, len(payments))
	for _, p := range payments {
		result = append(result, dtos.ToPaymentListItem(p))
	}

	return result, total, nil
}
