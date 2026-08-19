package orders

import (
	"context"
	"fmt"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
)

// AdminStatsRepository provides statistics queries.
type AdminStatsRepository interface {
	CountTotalUsers(ctx context.Context) (int64, error)
	CountTotalSellers(ctx context.Context) (int64, error)
	CountTotalProducts(ctx context.Context) (int64, error)
	CountTotalOrders(ctx context.Context) (int64, error)
	CountOrdersByStatus(ctx context.Context, status db.OrderStatus) (int64, error)
	CountTotalRevenue(ctx context.Context) (string, error)
}

type adminStatsRepository struct {
	queries *db.Queries
}

func NewAdminStatsRepository(queries *db.Queries) AdminStatsRepository {
	return &adminStatsRepository{queries: queries}
}

func (r *adminStatsRepository) CountTotalUsers(ctx context.Context) (int64, error) {
	return r.queries.CountUsers(ctx)
}

func (r *adminStatsRepository) CountTotalSellers(ctx context.Context) (int64, error) {
	return r.queries.CountSellers(ctx)
}

func (r *adminStatsRepository) CountTotalProducts(ctx context.Context) (int64, error) {
	return r.queries.CountProducts(ctx)
}

func (r *adminStatsRepository) CountTotalOrders(ctx context.Context) (int64, error) {
	return r.queries.CountOrders(ctx)
}

func (r *adminStatsRepository) CountOrdersByStatus(ctx context.Context, status db.OrderStatus) (int64, error) {
	return r.queries.CountOrdersByStatus(ctx, status)
}

func (r *adminStatsRepository) CountTotalRevenue(ctx context.Context) (string, error) {
	total, err := r.queries.SumOrderTotals(ctx)
	if err != nil {
		return "0", fmt.Errorf("sum order totals: %w", err)
	}
	
	if !total.Valid {
		return "0", nil
	}
	
	return dtos.NumericToStr(total), nil
}
