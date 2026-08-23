package orders

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/modules/notifications"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v3"
)

// --- Mocks ---

type mockOrderRepository struct {
	createOrder            func(ctx context.Context, arg db.CreateOrderParams) (db.Order, error)
	createOrderItem        func(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error)
	getOrder               func(ctx context.Context, id int64) (db.Order, error)
	getOrderByPublicToken  func(ctx context.Context, token string) (db.Order, error)
	getOrderItems          func(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error)
	listOrderByUser        func(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error)
	listOrdersBySeller     func(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error)
	updateOrderStatus      func(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error)
	updateOrderTracking    func(ctx context.Context, orderID int64, trackingNumber string) (db.Order, error)
	decrementProductStock  func(ctx context.Context, arg db.DecrementProductStockParams) (db.DecrementProductStockRow, error)
	incrementProductStock  func(ctx context.Context, arg db.IncrementProductStockParams) (db.IncrementProductStockRow, error)
	getProductStoreID      func(ctx context.Context, productID int64) (int64, error)
	getProduct             func(ctx context.Context, productID int64) (db.GetProductRow, error)
	countOrderItems        func(ctx context.Context, orderID int64) (int64, error)
	getOrderItemsByOrderID func(ctx context.Context, orderID int64) ([]db.GetOrderItemsByOrderIDRow, error)
	getPaymentByOrderID    func(ctx context.Context, orderID int64) (db.Payment, error)
}

func (m *mockOrderRepository) CreateOrder(ctx context.Context, arg db.CreateOrderParams) (db.Order, error) {
	return m.createOrder(ctx, arg)
}
func (m *mockOrderRepository) CreateOrderItem(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error) {
	return m.createOrderItem(ctx, arg)
}
func (m *mockOrderRepository) GetOrder(ctx context.Context, id int64) (db.Order, error) {
	return m.getOrder(ctx, id)
}
func (m *mockOrderRepository) GetOrderByPublicToken(ctx context.Context, token string) (db.Order, error) {
	return m.getOrderByPublicToken(ctx, token)
}
func (m *mockOrderRepository) GetOrderItems(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
	return m.getOrderItems(ctx, orderID)
}
func (m *mockOrderRepository) ListOrderByUser(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error) {
	return m.listOrderByUser(ctx, userID, page, limit)
}
func (m *mockOrderRepository) ListOrdersBySeller(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error) {
	return m.listOrdersBySeller(ctx, userID, page, limit)
}
func (m *mockOrderRepository) UpdateOrderStatus(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error) {
	return m.updateOrderStatus(ctx, arg)
}
func (m *mockOrderRepository) UpdateOrderTracking(ctx context.Context, orderID int64, trackingNumber string) (db.Order, error) {
	return m.updateOrderTracking(ctx, orderID, trackingNumber)
}
func (m *mockOrderRepository) DecrementProductStock(ctx context.Context, arg db.DecrementProductStockParams) (db.DecrementProductStockRow, error) {
	return m.decrementProductStock(ctx, arg)
}
func (m *mockOrderRepository) IncrementProductStock(ctx context.Context, arg db.IncrementProductStockParams) (db.IncrementProductStockRow, error) {
	return m.incrementProductStock(ctx, arg)
}
func (m *mockOrderRepository) GetProductStoreID(ctx context.Context, productID int64) (int64, error) {
	return m.getProductStoreID(ctx, productID)
}
func (m *mockOrderRepository) GetProduct(ctx context.Context, productID int64) (db.GetProductRow, error) {
	return m.getProduct(ctx, productID)
}
func (m *mockOrderRepository) CountOrderItems(ctx context.Context, orderID int64) (int64, error) {
	return m.countOrderItems(ctx, orderID)
}
func (m *mockOrderRepository) GetOrderItemsByOrderID(ctx context.Context, orderID int64) ([]db.GetOrderItemsByOrderIDRow, error) {
	return m.getOrderItemsByOrderID(ctx, orderID)
}
func (m *mockOrderRepository) GetPaymentByOrderID(ctx context.Context, orderID int64) (db.Payment, error) {
	return m.getPaymentByOrderID(ctx, orderID)
}

type mockStoreResolver struct {
	getStoreOwnerByStoreID func(ctx context.Context, storeID int64) (int64, error)
	getUserByID            func(ctx context.Context, id int64) (db.User, error)
}

func (m *mockStoreResolver) GetStoreOwnerByStoreID(ctx context.Context, storeID int64) (int64, error) {
	return m.getStoreOwnerByStoreID(ctx, storeID)
}
func (m *mockStoreResolver) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	return m.getUserByID(ctx, id)
}

type mockPaymentRefunder struct {
	refundPayment func(ctx context.Context, userID, paymentID int64) (dtos.PaymentResponse, error)
	getPayment    func(ctx context.Context, userID, paymentID int64) (dtos.PaymentResponse, error)
}

func (m *mockPaymentRefunder) RefundPayment(ctx context.Context, userID, paymentID int64) (dtos.PaymentResponse, error) {
	return m.refundPayment(ctx, userID, paymentID)
}
func (m *mockPaymentRefunder) GetPayment(ctx context.Context, userID, paymentID int64) (dtos.PaymentResponse, error) {
	return m.getPayment(ctx, userID, paymentID)
}

type mockNotificationSender struct {
	sendOrderConfirmation func(orderID int64, customerEmail, customerName string, items []notifications.OrderItem, total string) error
	sendOrderStatusUpdate func(orderID int64, customerEmail, customerName, status string) error
	sendSellerNewOrder    func(sellerEmail, sellerName, orderID, total, customerEmail, customerName string, items []notifications.OrderItem) error
}

func (m *mockNotificationSender) SendOrderConfirmation(orderID int64, customerEmail, customerName string, items []notifications.OrderItem, total string) error {
	if m.sendOrderConfirmation != nil {
		return m.sendOrderConfirmation(orderID, customerEmail, customerName, items, total)
	}
	return nil
}
func (m *mockNotificationSender) SendOrderStatusUpdate(orderID int64, customerEmail, customerName, status string) error {
	if m.sendOrderStatusUpdate != nil {
		return m.sendOrderStatusUpdate(orderID, customerEmail, customerName, status)
	}
	return nil
}
func (m *mockNotificationSender) SendSellerNewOrder(sellerEmail, sellerName, orderID, total, customerEmail, customerName string, items []notifications.OrderItem) error {
	if m.sendSellerNewOrder != nil {
		return m.sendSellerNewOrder(sellerEmail, sellerName, orderID, total, customerEmail, customerName, items)
	}
	return nil
}

func newTestService(repo OrderRepository, solver storeResolver, paymentRef paymentRefunder, notifications notificationSender, pool db.DBPool) OrderService {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return NewOrderService(repo, solver, paymentRef, notifications, pool, logger)
}

func testOrder() db.Order {
	return db.Order{
		ID:        1,
		UserID:    pgtype.Int8{Int64: 1, Valid: true},
		Status:    db.OrderStatusPending,
		Total:     pgtype.Numeric{Valid: true},
		Address:   "123 Main St",
		CreatedAt: pgtype.Timestamptz{Valid: true},
	}
}

func testOrderItems() []db.GetOrderItemsRow {
	return []db.GetOrderItemsRow{
		{ProductID: 1, Quantity: 2, Price: pgtype.Numeric{Valid: true}},
	}
}

// --- validateCreateOrderRequest ---

func TestValidateCreateOrderRequest(t *testing.T) {
	t.Run("empty items", func(t *testing.T) {
		req := dtos.CreateOrderRequest{
			Delivery: dtos.DeliveryInfo{Address: "123 Main St"},
		}
		err := validateCreateOrderRequest(req, false)
		if err != ErrOrderEmpty {
			t.Fatalf("expected ErrOrderEmpty, got %v", err)
		}
	})

	t.Run("empty address", func(t *testing.T) {
		req := dtos.CreateOrderRequest{
			Delivery: dtos.DeliveryInfo{},
			Items:    []dtos.OrderItemRequest{{ProductID: 1, Quantity: 1}},
		}
		err := validateCreateOrderRequest(req, false)
		if err != ErrAddressRequired {
			t.Fatalf("expected ErrAddressRequired, got %v", err)
		}
	})

	t.Run("guest without customer info", func(t *testing.T) {
		req := dtos.CreateOrderRequest{
			Delivery: dtos.DeliveryInfo{Address: "123 Main St"},
			Items:    []dtos.OrderItemRequest{{ProductID: 1, Quantity: 1}},
		}
		err := validateCreateOrderRequest(req, true)
		if err != ErrGuestCustomerReq {
			t.Fatalf("expected ErrGuestCustomerReq, got %v", err)
		}
	})

	t.Run("guest with partial customer info", func(t *testing.T) {
		req := dtos.CreateOrderRequest{
			Delivery: dtos.DeliveryInfo{Address: "123 Main St"},
			Customer: &dtos.CustomerInfo{FirstName: "Ivan"},
			Items:    []dtos.OrderItemRequest{{ProductID: 1, Quantity: 1}},
		}
		err := validateCreateOrderRequest(req, true)
		if err != ErrGuestCustomerReq {
			t.Fatalf("expected ErrGuestCustomerReq, got %v", err)
		}
	})

	t.Run("invalid product ID", func(t *testing.T) {
		req := dtos.CreateOrderRequest{
			Delivery: dtos.DeliveryInfo{Address: "123 Main St"},
			Items:    []dtos.OrderItemRequest{{ProductID: 0, Quantity: 1}},
		}
		err := validateCreateOrderRequest(req, false)
		if err != ErrInvalidProductID {
			t.Fatalf("expected ErrInvalidProductID, got %v", err)
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		req := dtos.CreateOrderRequest{
			Delivery: dtos.DeliveryInfo{Address: "123 Main St"},
			Items:    []dtos.OrderItemRequest{{ProductID: 1, Quantity: 0}},
		}
		err := validateCreateOrderRequest(req, false)
		if err != ErrInvalidQuantity {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})

	t.Run("valid authenticated order", func(t *testing.T) {
		req := dtos.CreateOrderRequest{
			Delivery: dtos.DeliveryInfo{Address: "123 Main St"},
			Items:    []dtos.OrderItemRequest{{ProductID: 1, Quantity: 2}},
		}
		err := validateCreateOrderRequest(req, false)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("valid guest order", func(t *testing.T) {
		req := dtos.CreateOrderRequest{
			Delivery: dtos.DeliveryInfo{Address: "123 Main St"},
			Customer: &dtos.CustomerInfo{FirstName: "Ivan", Email: "ivan@example.com", Phone: "+79990000000"},
			Items:    []dtos.OrderItemRequest{{ProductID: 1, Quantity: 2}},
		}
		err := validateCreateOrderRequest(req, true)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

// --- GetOrder ---

func TestGetOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("order not found", func(t *testing.T) {
		repo := &mockOrderRepository{
			getOrder: func(ctx context.Context, orderID int64) (db.Order, error) {
				return db.Order{}, pgx.ErrNoRows
			},
		}
		svc := newTestService(repo, nil, nil, nil, nil)
		_, err := svc.GetOrder(ctx, 1, 999)
		if err != ErrOrderNotFound {
			t.Fatalf("expected ErrOrderNotFound, got %v", err)
		}
	})

	t.Run("db error on get order", func(t *testing.T) {
		repo := &mockOrderRepository{
			getOrder: func(ctx context.Context, orderID int64) (db.Order, error) {
				return db.Order{}, errors.New("db error")
			},
		}
		svc := newTestService(repo, nil, nil, nil, nil)
		_, err := svc.GetOrder(ctx, 1, 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("forbidden - not owner, not seller", func(t *testing.T) {
		order := testOrder()
		order.UserID = pgtype.Int8{Int64: 5, Valid: true}
		repo := &mockOrderRepository{
			getOrder: func(ctx context.Context, orderID int64) (db.Order, error) {
				return order, nil
			},
			getOrderItems: func(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
				return testOrderItems(), nil
			},
			getProductStoreID: func(ctx context.Context, productID int64) (int64, error) {
				return 10, nil
			},
		}
		solver := &mockStoreResolver{
			getStoreOwnerByStoreID: func(ctx context.Context, storeID int64) (int64, error) {
				return 999, nil
			},
		}
		svc := newTestService(repo, solver, nil, nil, nil)
		_, err := svc.GetOrder(ctx, 1, 1)
		if err != ErrForbidden {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("success - buyer", func(t *testing.T) {
		order := testOrder()
		repo := &mockOrderRepository{
			getOrder: func(ctx context.Context, orderID int64) (db.Order, error) {
				return order, nil
			},
			getOrderItems: func(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
				return testOrderItems(), nil
			},
		}
		svc := newTestService(repo, nil, nil, nil, nil)
		resp, err := svc.GetOrder(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ID != 1 {
			t.Fatalf("expected ID 1, got %d", resp.ID)
		}
	})

	t.Run("success - seller", func(t *testing.T) {
		order := testOrder()
		repo := &mockOrderRepository{
			getOrder: func(ctx context.Context, orderID int64) (db.Order, error) {
				return order, nil
			},
			getOrderItems: func(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
				return testOrderItems(), nil
			},
			getProductStoreID: func(ctx context.Context, productID int64) (int64, error) {
				return 10, nil
			},
		}
		solver := &mockStoreResolver{
			getStoreOwnerByStoreID: func(ctx context.Context, storeID int64) (int64, error) {
				return 1, nil
			},
		}
		svc := newTestService(repo, solver, nil, nil, nil)
		resp, err := svc.GetOrder(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ID != 1 {
			t.Fatalf("expected ID 1, got %d", resp.ID)
		}
	})
}

// --- GetOrderByPublicToken ---

func TestGetOrderByPublicToken(t *testing.T) {
	ctx := context.Background()

	t.Run("not found", func(t *testing.T) {
		repo := &mockOrderRepository{
			getOrderByPublicToken: func(ctx context.Context, token string) (db.Order, error) {
				return db.Order{}, pgx.ErrNoRows
			},
		}
		svc := newTestService(repo, nil, nil, nil, nil)
		_, err := svc.GetOrderByPublicToken(ctx, "invalid-token")
		if err != ErrOrderNotFound {
			t.Fatalf("expected ErrOrderNotFound, got %v", err)
		}
	})

	t.Run("db error", func(t *testing.T) {
		repo := &mockOrderRepository{
			getOrderByPublicToken: func(ctx context.Context, token string) (db.Order, error) {
				return db.Order{}, errors.New("db error")
			},
		}
		svc := newTestService(repo, nil, nil, nil, nil)
		_, err := svc.GetOrderByPublicToken(ctx, "token")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		order := testOrder()
		repo := &mockOrderRepository{
			getOrderByPublicToken: func(ctx context.Context, token string) (db.Order, error) {
				return order, nil
			},
			getOrderItems: func(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
				return testOrderItems(), nil
			},
		}
		svc := newTestService(repo, nil, nil, nil, nil)
		resp, err := svc.GetOrderByPublicToken(ctx, "valid-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ID != 1 {
			t.Fatalf("expected ID 1, got %d", resp.ID)
		}
	})
}

// --- ListOrdersByUser ---

func TestListOrdersByUser(t *testing.T) {
	ctx := context.Background()

	t.Run("empty list", func(t *testing.T) {
		repo := &mockOrderRepository{
			listOrderByUser: func(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error) {
				return []db.Order{}, 0, nil
			},
		}
		svc := newTestService(repo, nil, nil, nil, nil)
		items, total, err := svc.ListOrdersByUser(ctx, 1, 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("expected empty items, got %d", len(items))
		}
		if total != 0 {
			t.Fatalf("expected total 0, got %d", total)
		}
	})

	t.Run("success", func(t *testing.T) {
		order := testOrder()
		repo := &mockOrderRepository{
			listOrderByUser: func(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error) {
				return []db.Order{order}, 1, nil
			},
			countOrderItems: func(ctx context.Context, orderID int64) (int64, error) {
				return 2, nil
			},
		}
		svc := newTestService(repo, nil, nil, nil, nil)
		items, total, err := svc.ListOrdersByUser(ctx, 1, 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if total != 1 {
			t.Fatalf("expected total 1, got %d", total)
		}
	})

	t.Run("db error", func(t *testing.T) {
		repo := &mockOrderRepository{
			listOrderByUser: func(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}
		svc := newTestService(repo, nil, nil, nil, nil)
		_, _, err := svc.ListOrdersByUser(ctx, 1, 1, 20)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// --- ListOrdersBySeller ---

func TestListOrdersBySeller(t *testing.T) {
	ctx := context.Background()

	t.Run("empty list", func(t *testing.T) {
		repo := &mockOrderRepository{
			listOrdersBySeller: func(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error) {
				return []db.Order{}, 0, nil
			},
		}
		svc := newTestService(repo, nil, nil, nil, nil)
		items, _, err := svc.ListOrdersBySeller(ctx, 1, 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("expected empty items, got %d", len(items))
		}
	})

	t.Run("success", func(t *testing.T) {
		order := testOrder()
		repo := &mockOrderRepository{
			listOrdersBySeller: func(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error) {
				return []db.Order{order}, 1, nil
			},
			getOrderItems: func(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
				return testOrderItems(), nil
			},
		}
		solver := &mockStoreResolver{
			getUserByID: func(ctx context.Context, id int64) (db.User, error) {
				return db.User{Username: "buyer"}, nil
			},
		}
		svc := newTestService(repo, solver, nil, nil, nil)
		items, total, err := svc.ListOrdersBySeller(ctx, 1, 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if total != 1 {
			t.Fatalf("expected total 1, got %d", total)
		}
	})

	t.Run("db error", func(t *testing.T) {
		repo := &mockOrderRepository{
			listOrdersBySeller: func(ctx context.Context, userID int64, page, limit int) ([]db.Order, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}
		svc := newTestService(repo, nil, nil, nil, nil)
		_, _, err := svc.ListOrdersBySeller(ctx, 1, 1, 20)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// --- UpdateOrderStatus ---

func TestUpdateOrderStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid status", func(t *testing.T) {
		svc := newTestService(&mockOrderRepository{}, nil, nil, nil, nil)
		_, err := svc.UpdateOrderStatus(ctx, 1, 1, dtos.UpdateOrderStatusRequest{Status: "invalid"})
		if err != ErrInvalidStatus {
			t.Fatalf("expected ErrInvalidStatus, got %v", err)
		}
	})

	t.Run("order not found", func(t *testing.T) {
		repo := &mockOrderRepository{
			getOrder: func(ctx context.Context, orderID int64) (db.Order, error) {
				return db.Order{}, pgx.ErrNoRows
			},
		}
		svc := newTestService(repo, nil, nil, nil, nil)
		_, err := svc.UpdateOrderStatus(ctx, 1, 999, dtos.UpdateOrderStatusRequest{Status: db.OrderStatusShipped})
		if err != ErrOrderNotFound {
			t.Fatalf("expected ErrOrderNotFound, got %v", err)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		order := testOrder()
		order.UserID = pgtype.Int8{Int64: 999, Valid: true}
		repo := &mockOrderRepository{
			getOrder: func(ctx context.Context, orderID int64) (db.Order, error) {
				return order, nil
			},
			getOrderItems: func(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
				return testOrderItems(), nil
			},
			getProductStoreID: func(ctx context.Context, productID int64) (int64, error) {
				return 10, nil
			},
		}
		solver := &mockStoreResolver{
			getStoreOwnerByStoreID: func(ctx context.Context, storeID int64) (int64, error) {
				return 999, nil
			},
		}
		svc := newTestService(repo, solver, nil, nil, nil)
		_, err := svc.UpdateOrderStatus(ctx, 1, 1, dtos.UpdateOrderStatusRequest{Status: db.OrderStatusShipped})
		if err != ErrForbidden {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("invalid status transition - cancelling shipped order", func(t *testing.T) {
		order := testOrder()
		order.Status = db.OrderStatusShipped
		repo := &mockOrderRepository{
			getOrder: func(ctx context.Context, orderID int64) (db.Order, error) {
				return order, nil
			},
			getOrderItems: func(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
				return testOrderItems(), nil
			},
			getProductStoreID: func(ctx context.Context, productID int64) (int64, error) {
				return 10, nil
			},
		}
		solver := &mockStoreResolver{
			getStoreOwnerByStoreID: func(ctx context.Context, storeID int64) (int64, error) {
				return 999, nil
			},
		}
		svc := newTestService(repo, solver, nil, nil, nil)
		_, err := svc.UpdateOrderStatus(ctx, 1, 1, dtos.UpdateOrderStatusRequest{Status: db.OrderStatusCancelled})
		if err != ErrInvalidStatusTrans {
			t.Fatalf("expected ErrInvalidStatusTrans, got %v", err)
		}
	})

	t.Run("success - seller shipping", func(t *testing.T) {
		order := testOrder()
		order.Status = db.OrderStatusPaid
		order.UserID = pgtype.Int8{Valid: false}
		repo := &mockOrderRepository{
			getOrder: func(ctx context.Context, orderID int64) (db.Order, error) {
				return order, nil
			},
			getOrderItems: func(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
				return testOrderItems(), nil
			},
			getProductStoreID: func(ctx context.Context, productID int64) (int64, error) {
				return 10, nil
			},
			updateOrderStatus: func(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error) {
				if arg.Status != db.OrderStatusShipped {
					t.Fatalf("expected status shipped, got %s", arg.Status)
				}
				return order, nil
			},
		}
		solver := &mockStoreResolver{
			getStoreOwnerByStoreID: func(ctx context.Context, storeID int64) (int64, error) {
				return 1, nil
			},
		}
		svc := newTestService(repo, solver, nil, nil, nil)
		resp, err := svc.UpdateOrderStatus(ctx, 1, 1, dtos.UpdateOrderStatusRequest{Status: db.OrderStatusShipped})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ID != 1 {
			t.Fatalf("expected ID 1, got %d", resp.ID)
		}
	})

	t.Run("success - cancel pending order", func(t *testing.T) {
		order := testOrder()
		order.Status = db.OrderStatusPending
		repo := &mockOrderRepository{
			getOrder: func(ctx context.Context, orderID int64) (db.Order, error) {
				return order, nil
			},
			getOrderItems: func(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
				return testOrderItems(), nil
			},
			getProductStoreID: func(ctx context.Context, productID int64) (int64, error) {
				return 10, nil
			},
			getOrderItemsByOrderID: func(ctx context.Context, orderID int64) ([]db.GetOrderItemsByOrderIDRow, error) {
				return []db.GetOrderItemsByOrderIDRow{
					{ProductID: 1, Quantity: 2},
				}, nil
			},
			incrementProductStock: func(ctx context.Context, arg db.IncrementProductStockParams) (db.IncrementProductStockRow, error) {
				return db.IncrementProductStockRow{}, nil
			},
			updateOrderStatus: func(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error) {
				return order, nil
			},
		}
		solver := &mockStoreResolver{
			getStoreOwnerByStoreID: func(ctx context.Context, storeID int64) (int64, error) {
				return 1, nil
			},
		}
		pool, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("failed to create mock pool: %v", err)
		}
		pool.ExpectBegin()
		pool.ExpectQuery("SELECT product_id, quantity FROM order_items WHERE order_id").
			WithArgs(int64(1)).
			WillReturnRows(pgxmock.NewRows([]string{"product_id", "quantity"}).AddRow(int64(1), int32(2)))
		pool.ExpectQuery("UPDATE products SET stock = stock \\+ \\$2").
			WithArgs(int64(1), int32(2)).
			WillReturnRows(pgxmock.NewRows([]string{"id", "store_id", "title", "price", "stock"}).AddRow(int64(1), int64(10), "Test", "100.00", int32(100)))
		pool.ExpectQuery("UPDATE orders SET status").
			WithArgs(int64(1), db.OrderStatus("cancelled")).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "status", "total", "address", "created_at", "updated_at", "public_token", "customer_first_name", "customer_last_name", "customer_email", "customer_phone", "delivery_method", "delivery_cost", "tracking_number", "expires_at"}).
				AddRow(int64(1), int64(1), "cancelled", "100.00", "123 Main St", pgtype.Timestamptz{}, pgtype.Timestamptz{}, nil, nil, nil, nil, nil, nil, "0", nil, nil))
		pool.ExpectCommit()
		svc := newTestService(repo, solver, nil, nil, pool)
		resp, err := svc.UpdateOrderStatus(ctx, 1, 1, dtos.UpdateOrderStatusRequest{Status: db.OrderStatusCancelled})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Status != string(db.OrderStatusCancelled) {
			t.Fatalf("expected status cancelled, got %s", resp.Status)
		}
	})
}

// --- UpdateOrderTracking ---

func TestUpdateOrderTracking(t *testing.T) {
	ctx := context.Background()

	t.Run("order not found", func(t *testing.T) {
		repo := &mockOrderRepository{
			getOrder: func(ctx context.Context, orderID int64) (db.Order, error) {
				return db.Order{}, pgx.ErrNoRows
			},
		}
		svc := newTestService(repo, nil, nil, nil, nil)
		_, err := svc.UpdateOrderTracking(ctx, 1, 999, dtos.UpdateTrackingRequest{TrackingNumber: "TRK123"})
		if err != ErrOrderNotFound {
			t.Fatalf("expected ErrOrderNotFound, got %v", err)
		}
	})

	t.Run("not seller", func(t *testing.T) {
		order := testOrder()
		repo := &mockOrderRepository{
			getOrder: func(ctx context.Context, orderID int64) (db.Order, error) {
				return order, nil
			},
			getOrderItems: func(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
				return testOrderItems(), nil
			},
			getProductStoreID: func(ctx context.Context, productID int64) (int64, error) {
				return 10, nil
			},
		}
		solver := &mockStoreResolver{
			getStoreOwnerByStoreID: func(ctx context.Context, storeID int64) (int64, error) {
				return 999, nil
			},
		}
		svc := newTestService(repo, solver, nil, nil, nil)
		_, err := svc.UpdateOrderTracking(ctx, 1, 1, dtos.UpdateTrackingRequest{TrackingNumber: "TRK123"})
		if err != ErrTrackingOnlySeller {
			t.Fatalf("expected ErrTrackingOnlySeller, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		order := testOrder()
		repo := &mockOrderRepository{
			getOrder: func(ctx context.Context, orderID int64) (db.Order, error) {
				return order, nil
			},
			getOrderItems: func(ctx context.Context, orderID int64) ([]db.GetOrderItemsRow, error) {
				return testOrderItems(), nil
			},
			getProductStoreID: func(ctx context.Context, productID int64) (int64, error) {
				return 10, nil
			},
			updateOrderTracking: func(ctx context.Context, orderID int64, trackingNumber string) (db.Order, error) {
				if trackingNumber != "TRK123" {
					t.Fatalf("expected tracking TRK123, got %s", trackingNumber)
				}
				return order, nil
			},
		}
		solver := &mockStoreResolver{
			getStoreOwnerByStoreID: func(ctx context.Context, storeID int64) (int64, error) {
				return 1, nil
			},
		}
		svc := newTestService(repo, solver, nil, nil, nil)
		resp, err := svc.UpdateOrderTracking(ctx, 1, 1, dtos.UpdateTrackingRequest{TrackingNumber: "TRK123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ID != 1 {
			t.Fatalf("expected ID 1, got %d", resp.ID)
		}
	})
}

// --- Numeric helpers ---

func TestMulNumeric(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	svc := newTestService(&mockOrderRepository{}, nil, nil, nil, nil)

	t.Run("invalid numeric", func(t *testing.T) {
		result, err := svc.(*orderService).mulNumeric(pgtype.Numeric{Valid: false}, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Valid {
			t.Fatal("expected invalid result")
		}
	})

	t.Run("multiply 10.50 * 3", func(t *testing.T) {
		var n pgtype.Numeric
		if err := n.Scan("10.50"); err != nil {
			t.Fatalf("failed to scan: %v", err)
		}
		result, err := svc.(*orderService).mulNumeric(n, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Fatal("expected valid result")
		}
		str := dtos.NumericToStr(result)
		if str != "31.5" {
			t.Fatalf("expected 31.5, got %s", str)
		}
	})
}

func TestAddNumeric(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	svc := newTestService(&mockOrderRepository{}, nil, nil, nil, nil)

	t.Run("both invalid", func(t *testing.T) {
		result, err := svc.(*orderService).addNumeric(
			pgtype.Numeric{Valid: false},
			pgtype.Numeric{Valid: false},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Valid {
			t.Fatal("expected invalid result")
		}
	})

	t.Run("add 10.00 + 5.50", func(t *testing.T) {
		var a, b pgtype.Numeric
		a.Scan("10.00")
		b.Scan("5.50")
		result, err := svc.(*orderService).addNumeric(a, b)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Fatal("expected valid result")
		}
		str := dtos.NumericToStr(result)
		if str != "15.5" {
			t.Fatalf("expected 15.5, got %s", str)
		}
	})

	t.Run("one invalid", func(t *testing.T) {
		var a pgtype.Numeric
		a.Scan("10.00")
		result, err := svc.(*orderService).addNumeric(a, pgtype.Numeric{Valid: false})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Fatal("expected valid result")
		}
		str := dtos.NumericToStr(result)
		if str != "10" {
			t.Fatalf("expected 10, got %s", str)
		}
	})
}
