package orders

import (
	"context"
	"fmt"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
)

// AdminStatsService provides platform statistics.
type AdminStatsService interface {
	GetStats(ctx context.Context) (dtos.AdminStatsResponse, error)
}

type adminStatsService struct {
	repo AdminStatsRepository
}

func NewAdminStatsService(repo AdminStatsRepository) AdminStatsService {
	return &adminStatsService{repo: repo}
}

func (s *adminStatsService) GetStats(ctx context.Context) (dtos.AdminStatsResponse, error) {
	totalUsers, err := s.repo.CountTotalUsers(ctx)
	if err != nil {
		return dtos.AdminStatsResponse{}, fmt.Errorf("count users: %w", err)
	}

	totalSellers, err := s.repo.CountTotalSellers(ctx)
	if err != nil {
		return dtos.AdminStatsResponse{}, fmt.Errorf("count sellers: %w", err)
	}

	totalProducts, err := s.repo.CountTotalProducts(ctx)
	if err != nil {
		return dtos.AdminStatsResponse{}, fmt.Errorf("count products: %w", err)
	}

	totalOrders, err := s.repo.CountTotalOrders(ctx)
	if err != nil {
		return dtos.AdminStatsResponse{}, fmt.Errorf("count orders: %w", err)
	}

	pendingOrders, err := s.repo.CountOrdersByStatus(ctx, db.OrderStatusPending)
	if err != nil {
		return dtos.AdminStatsResponse{}, fmt.Errorf("count pending orders: %w", err)
	}

	paidOrders, err := s.repo.CountOrdersByStatus(ctx, db.OrderStatusPaid)
	if err != nil {
		return dtos.AdminStatsResponse{}, fmt.Errorf("count paid orders: %w", err)
	}

	shippedOrders, err := s.repo.CountOrdersByStatus(ctx, db.OrderStatusShipped)
	if err != nil {
		return dtos.AdminStatsResponse{}, fmt.Errorf("count shipped orders: %w", err)
	}

	deliveredOrders, err := s.repo.CountOrdersByStatus(ctx, db.OrderStatusDelivered)
	if err != nil {
		return dtos.AdminStatsResponse{}, fmt.Errorf("count delivered orders: %w", err)
	}

	cancelledOrders, err := s.repo.CountOrdersByStatus(ctx, db.OrderStatusCancelled)
	if err != nil {
		return dtos.AdminStatsResponse{}, fmt.Errorf("count cancelled orders: %w", err)
	}

	totalRevenue, err := s.repo.CountTotalRevenue(ctx)
	if err != nil {
		return dtos.AdminStatsResponse{}, fmt.Errorf("count revenue: %w", err)
	}

	return dtos.AdminStatsResponse{
		TotalUsers:      totalUsers,
		TotalSellers:    totalSellers,
		TotalProducts:   totalProducts,
		TotalOrders:     totalOrders,
		PendingOrders:   pendingOrders,
		PaidOrders:      paidOrders,
		ShippedOrders:   shippedOrders,
		DeliveredOrders: deliveredOrders,
		CancelledOrders: cancelledOrders,
		TotalRevenue:    totalRevenue,
	}, nil
}
