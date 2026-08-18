package orders

import (
	"context"
	"fmt"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

// OrderRepository defines data access methods for order operations.
type OrderRepository interface {
	CreateOrder(ctx context.Context, arg db.CreateOrderParams) (db.Order, error)
	CreateOrderItem(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error)
	GetOrder(ctx context.Context, id int64) (db.Order, error)
	GetOrderByPublicToken(ctx context.Context, token string) (db.Order, error)
	GetOrderItems(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error)
	ListOrderByUser(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error)
	ListOrdersBySeller(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error)
	UpdateOrderStatus(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error)
	UpdateOrderTracking(ctx context.Context, orderID int64, trackingNumber string) (db.Order, error)
	DecrementProductStock(ctx context.Context, arg db.DecrementProductStockParams) (db.DecrementProductStockRow, error)
	IncrementProductStock(ctx context.Context, arg db.IncrementProductStockParams) (db.IncrementProductStockRow, error)
	GetProductStoreID(ctx context.Context, productID int64) (int64, error)
	GetProduct(ctx context.Context, productID int64) (db.GetProductRow, error)
	CountOrderItems(ctx context.Context, orderID int64) (int64, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int64) ([]db.GetOrderItemsByOrderIDRow, error)
	GetPaymentByOrderID(ctx context.Context, orderID int64) (db.Payment, error)
}

type orderRepository struct {
	queries *db.Queries
}

// NewOrderRepository creates a new OrderRepository.
func NewOrderRepository(queries *db.Queries) OrderRepository {
	return &orderRepository{queries: queries}
}

func (r *orderRepository) CreateOrder(ctx context.Context, arg db.CreateOrderParams) (db.Order, error) {
	return r.queries.CreateOrder(ctx, arg)
}

func (r *orderRepository) CreateOrderItem(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error) {
	return r.queries.CreateOrderItem(ctx, arg)
}

func (r *orderRepository) GetOrder(ctx context.Context, id int64) (db.Order, error) {
	return r.queries.GetOrder(ctx, id)
}

func (r *orderRepository) GetOrderByPublicToken(ctx context.Context, token string) (db.Order, error) {
	return r.queries.GetOrderByPublicToken(ctx, pgtype.Text{String: token, Valid: true})
}

func (r *orderRepository) GetOrderItems(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
	return r.queries.GetOrderItems(ctx, orderID)
}

func (r *orderRepository) ListOrderByUser(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error) {
	offset := int32((page - 1) * limit)

	orders, err := r.queries.ListOrdersByUser(ctx, db.ListOrdersByUserParams{
		UserID: pgtype.Int8{Int64: userID, Valid: true},
		Limit:  int32(limit),
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list orders by user: %w", err)
	}

	count, err := r.queries.ListOrdersByUserCount(ctx, pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return nil, 0, fmt.Errorf("list orders by user count: %w", err)
	}

	return orders, count, nil
}

func (r *orderRepository) ListOrdersBySeller(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error) {
	offset := int32((page - 1) * limit)

	orders, err := r.queries.ListOrdersBySeller(ctx, db.ListOrdersBySellerParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list orders by seller: %w", err)
	}

	count, err := r.queries.ListOrdersBySellerCount(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders by seller count: %w", err)
	}

	return orders, count, nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error) {
	return r.queries.UpdateOrderStatus(ctx, arg)
}

func (r *orderRepository) UpdateOrderTracking(ctx context.Context, orderID int64, trackingNumber string) (db.Order, error) {
	return r.queries.UpdateOrderTracking(ctx, db.UpdateOrderTrackingParams{
		ID:             orderID,
		TrackingNumber: pgtype.Text{String: trackingNumber, Valid: trackingNumber != ""},
	})
}

func (r *orderRepository) DecrementProductStock(ctx context.Context, arg db.DecrementProductStockParams) (db.DecrementProductStockRow, error) {
	return r.queries.DecrementProductStock(ctx, arg)
}

func (r *orderRepository) GetProductStoreID(ctx context.Context, productID int64) (int64, error) {
	return r.queries.GetProductStoreID(ctx, productID)
}

func (r *orderRepository) CountOrderItems(ctx context.Context, orderID int64) (int64, error) {
	return r.queries.CountOrderItems(ctx, orderID)
}

func (r *orderRepository) IncrementProductStock(ctx context.Context, arg db.IncrementProductStockParams) (db.IncrementProductStockRow, error) {
	return r.queries.IncrementProductStock(ctx, arg)
}

func (r *orderRepository) GetOrderItemsByOrderID(ctx context.Context, orderID int64) ([]db.GetOrderItemsByOrderIDRow, error) {
	return r.queries.GetOrderItemsByOrderID(ctx, orderID)
}

func (r *orderRepository) GetPaymentByOrderID(ctx context.Context, orderID int64) (db.Payment, error) {
	return r.queries.GetPaymentByOrderID(ctx, orderID)
}

func (r *orderRepository) GetProduct(ctx context.Context, productID int64) (db.GetProductRow, error) {
	return r.queries.GetProduct(ctx, productID)
}
