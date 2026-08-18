package orders

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
)

// AdminOrderService defines admin order operations.
type AdminOrderService interface {
	ListAllOrders(ctx context.Context, page, limit int) ([]dtos.OrderListItem, int64, error)
	ListOrdersByStatus(ctx context.Context, status string, page, limit int) ([]dtos.OrderListItem, int64, error)
	ListOrdersByUser(ctx context.Context, userID int64, page, limit int) ([]dtos.OrderListItem, int64, error)
	GetOrderDetail(ctx context.Context, orderID int64) (dtos.OrderResponse, error)
}

type adminOrderService struct {
	repo AdminOrderRepository
	solver storeResolver
}

func NewAdminOrderService(repo AdminOrderRepository, solver storeResolver) AdminOrderService {
	return &adminOrderService{repo: repo, solver: solver}
}

func (s *adminOrderService) ListAllOrders(ctx context.Context, page, limit int) ([]dtos.OrderListItem, int64, error) {
	orders, total, err := s.repo.ListAllOrders(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}

	if len(orders) == 0 {
		return []dtos.OrderListItem{}, 0, nil
	}

	result := make([]dtos.OrderListItem, 0, len(orders))
	for _, order := range orders {
		count, err := s.repo.CountOrderItems(ctx, order.ID)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, dtos.ToOrderListItem(order, int32(count)))
	}

	return result, total, nil
}

func (s *adminOrderService) ListOrdersByStatus(ctx context.Context, status string, page, limit int) ([]dtos.OrderListItem, int64, error) {
	var orderStatus db.OrderStatus
	switch status {
	case "pending", "paid", "shipped", "delivered", "cancelled":
		orderStatus = db.OrderStatus(status)
	default:
		return nil, 0, ErrInvalidStatus
	}

	orders, total, err := s.repo.ListAllOrdersByStatus(ctx, orderStatus, page, limit)
	if err != nil {
		return nil, 0, err
	}

	if len(orders) == 0 {
		return []dtos.OrderListItem{}, 0, nil
	}

	result := make([]dtos.OrderListItem, 0, len(orders))
	for _, order := range orders {
		count, err := s.repo.CountOrderItems(ctx, order.ID)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, dtos.ToOrderListItem(order, int32(count)))
	}

	return result, total, nil
}

func (s *adminOrderService) ListOrdersByUser(ctx context.Context, userID int64, page, limit int) ([]dtos.OrderListItem, int64, error) {
	orders, total, err := s.repo.ListAllOrdersByUser(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	if len(orders) == 0 {
		return []dtos.OrderListItem{}, 0, nil
	}

	result := make([]dtos.OrderListItem, 0, len(orders))
	for _, order := range orders {
		count, err := s.repo.CountOrderItems(ctx, order.ID)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, dtos.ToOrderListItem(order, int32(count)))
	}

	return result, total, nil
}

func (s *adminOrderService) GetOrderDetail(ctx context.Context, orderID int64) (dtos.OrderResponse, error) {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dtos.OrderResponse{}, ErrOrderNotFound
		}
		return dtos.OrderResponse{}, err
	}

	items, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return dtos.OrderResponse{}, err
	}

	return dtos.ToOrderResponse(order, items), nil
}
