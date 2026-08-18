package payments

import (
	"context"
	"fmt"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PaymentRepository defines data access methods for payment operations.
type PaymentRepository interface {
	CreatePayment(ctx context.Context, arg db.CreatePaymentParams) (db.Payment, error)
	GetPayment(ctx context.Context, id int64) (db.Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID int64) (db.Payment, error)
	UpdatePaymentStatus(ctx context.Context, arg db.UpdatePaymentStatusParams) (db.Payment, error)
	ListPaymentsByUserID(ctx context.Context, userID int64, page, limit int) ([]db.Payment, int64, error)
}

type paymentRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

// NewPaymentRepository creates a new PaymentRepository.
func NewPaymentRepository(queries *db.Queries, pool *pgxpool.Pool) PaymentRepository {
	return &paymentRepository{queries: queries, pool: pool}
}

func (r *paymentRepository) CreatePayment(ctx context.Context, arg db.CreatePaymentParams) (db.Payment, error) {
	return r.queries.CreatePayment(ctx, arg)
}

func (r *paymentRepository) GetPayment(ctx context.Context, id int64) (db.Payment, error) {
	return r.queries.GetPayment(ctx, id)
}

func (r *paymentRepository) GetPaymentByOrderID(ctx context.Context, orderID int64) (db.Payment, error) {
	return r.queries.GetPaymentByOrderID(ctx, orderID)
}

func (r *paymentRepository) UpdatePaymentStatus(ctx context.Context, arg db.UpdatePaymentStatusParams) (db.Payment, error) {
	return r.queries.UpdatePaymentStatus(ctx, arg)
}

func (r *paymentRepository) ListPaymentsByUserID(ctx context.Context, userID int64, page, limit int) ([]db.Payment, int64, error) {
	offset := int32((page - 1) * limit)

	payments, err := r.queries.ListPaymentsByUserID(ctx, db.ListPaymentsByUserIDParams{
		Column1: userID,
		Limit:   int32(limit),
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list payments by user: %w", err)
	}

	count, err := r.queries.ListPaymentsByUserIDCount(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("list payments by user count: %w", err)
	}

	return payments, count, nil
}
