package payments

import (
	"context"
	"fmt"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, arg db.CreatePaymentParams) (db.Payment, error)
	GetPayment(ctx context.Context, id int64) (db.Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID int64) (db.Payment, error)
	UpdatePaymentStatus(ctx context.Context, arg db.UpdatePaymentStatusParams) (db.Payment, error)
}

type paymentRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

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

func toNumeric(s string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(s); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("scan numeric: %w", err)
	}
	return n, nil
}
