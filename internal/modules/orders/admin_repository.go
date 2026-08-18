package orders

import (
	"context"
	"fmt"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

// AdminOrderRepository extends OrderRepository with admin capabilities.
type AdminOrderRepository interface {
	ListAllOrders(ctx context.Context, page, limit int) ([]db.Order, int64, error)
	ListAllOrdersByStatus(ctx context.Context, status db.OrderStatus, page, limit int) ([]db.Order, int64, error)
	ListAllOrdersByUser(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error)
	GetOrder(ctx context.Context, id int64) (db.Order, error)
	GetOrderItems(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error)
	GetProductStoreID(ctx context.Context, productID int64) (int64, error)
	CountOrderItems(ctx context.Context, orderID int64) (int64, error)
}

type adminOrderRepository struct {
	queries *db.Queries
}

func NewAdminOrderRepository(queries *db.Queries) AdminOrderRepository {
	return &adminOrderRepository{queries: queries}
}

func (r *adminOrderRepository) ListAllOrders(ctx context.Context, page, limit int) ([]db.Order, int64, error) {
	offset := int32((page - 1) * limit)

	orders, err := r.queries.ListAllOrders(ctx, db.ListAllOrdersParams{
		Limit:  int32(limit),
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list all orders: %w", err)
	}

	count, err := r.queries.ListAllOrdersCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list all orders count: %w", err)
	}

	return orders, count, nil
}

func (r *adminOrderRepository) ListAllOrdersByStatus(ctx context.Context, status db.OrderStatus, page, limit int) ([]db.Order, int64, error) {
	offset := int32((page - 1) * limit)

	orders, err := r.queries.ListAllOrdersByStatus(ctx, db.ListAllOrdersByStatusParams{
		Status: status,
		Limit:  int32(limit),
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list all orders by status: %w", err)
	}

	count, err := r.queries.ListAllOrdersByStatusCount(ctx, status)
	if err != nil {
		return nil, 0, fmt.Errorf("list all orders by status count: %w", err)
	}

	return orders, count, nil
}

func (r *adminOrderRepository) ListAllOrdersByUser(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error) {
	offset := int32((page - 1) * limit)

	orders, err := r.queries.ListAllOrdersByUser(ctx, db.ListAllOrdersByUserParams{
		UserID: pgtype.Int8{Int64: userID, Valid: true},
		Limit:  int32(limit),
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list all orders by user: %w", err)
	}

	count, err := r.queries.ListAllOrdersByUserCount(ctx, pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return nil, 0, fmt.Errorf("list all orders by user count: %w", err)
	}

	return orders, count, nil
}

func (r *adminOrderRepository) GetOrderItems(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
	return r.queries.GetOrderItems(ctx, orderID)
}

func (r *adminOrderRepository) GetProductStoreID(ctx context.Context, productID int64) (int64, error) {
	return r.queries.GetProductStoreID(ctx, productID)
}

func (r *adminOrderRepository) GetOrder(ctx context.Context, id int64) (db.Order, error) {
	return r.queries.GetOrder(ctx, id)
}

func (r *adminOrderRepository) CountOrderItems(ctx context.Context, orderID int64) (int64, error) {
	return r.queries.CountOrderItems(ctx, orderID)
}
