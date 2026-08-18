package orders

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/modules/notifications"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrForbidden          = errors.New("access denied")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrOrderEmpty         = errors.New("order must have at least one item")
	ErrInvalidStatus      = errors.New("invalid order status")
	ErrInvalidStatusTrans = errors.New("invalid status transition")
	ErrInvalidQuantity    = errors.New("quantity must be greater than zero")
	ErrProductNotFound    = errors.New("product not found")
	ErrAddressRequired    = errors.New("address is required")
	ErrInvalidProductID   = errors.New("invalid product id")
	ErrOrderAlreadyPaid   = errors.New("order is already paid, cannot cancel")
	ErrGuestCustomerReq   = errors.New("customer info is required for guest orders")
	ErrTrackingOnlySeller = errors.New("only seller or admin can set tracking number")
)

type storeResolver interface {
	GetStoreOwnerByStoreID(ctx context.Context, storeID int64) (int64, error)
	GetUserByID(ctx context.Context, id int64) (db.User, error)
}

type paymentRefunder interface {
	RefundPayment(ctx context.Context, userID, paymentID int64) (dtos.PaymentResponse, error)
	GetPayment(ctx context.Context, userID, paymentID int64) (dtos.PaymentResponse, error)
}

type notificationSender interface {
	SendOrderConfirmation(orderID int64, customerEmail, customerName string, items []notifications.OrderItem, total string) error
	SendOrderStatusUpdate(orderID int64, customerEmail, customerName, status string) error
	SendSellerNewOrder(sellerEmail string, sellerName string, orderID int64, total string, customerEmail, customerName string, items []notifications.OrderItem) error
}

// OrderService defines business logic for order operations.
type OrderService interface {
	CreateOrder(ctx context.Context, userID *int64, req dtos.CreateOrderRequest) (dtos.OrderResponse, error)
	GetOrder(ctx context.Context, userID int64, orderID int64) (dtos.OrderResponse, error)
	GetOrderByPublicToken(ctx context.Context, token string) (dtos.OrderResponse, error)
	ListOrdersByUser(ctx context.Context, userID int64, page, limit int) ([]dtos.OrderListItem, int64, error)
	ListOrdersBySeller(ctx context.Context, userID int64, page, limit int) ([]dtos.OrderSellerListItem, int64, error)
	UpdateOrderStatus(ctx context.Context, userID int64, orderID int64, req dtos.UpdateOrderStatusRequest) (dtos.OrderResponse, error)
	UpdateOrderTracking(ctx context.Context, userID int64, orderID int64, req dtos.UpdateTrackingRequest) (dtos.OrderResponse, error)
}

type orderService struct {
	repo          OrderRepository
	solver        storeResolver
	paymentRef    paymentRefunder
	notifications notificationSender
	pool          *pgxpool.Pool
	log           *slog.Logger
}

// NewOrderService creates a new OrderService.
func NewOrderService(repo OrderRepository, solver storeResolver, paymentRef paymentRefunder, notifications notificationSender, pool *pgxpool.Pool, log *slog.Logger) OrderService {
	return &orderService{repo: repo, solver: solver, paymentRef: paymentRef, notifications: notifications, pool: pool, log: log}
}

type orderItem struct {
	productID int64
	quantity  int32
	price     pgtype.Numeric
}

func validateCreateOrderRequest(req dtos.CreateOrderRequest, isGuest bool) error {
	if len(req.Items) == 0 {
		return ErrOrderEmpty
	}
	if req.Delivery.Address == "" {
		return ErrAddressRequired
	}
	if isGuest && req.Customer == nil {
		return ErrGuestCustomerReq
	}
	if isGuest {
		if req.Customer.FirstName == "" || req.Customer.Email == "" || req.Customer.Phone == "" {
			return ErrGuestCustomerReq
		}
	}
	for _, item := range req.Items {
		if item.ProductID <= 0 {
			return ErrInvalidProductID
		}
		if item.Quantity <= 0 {
			return ErrInvalidQuantity
		}
	}
	return nil
}

func generatePublicToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func parseNumeric(s string) (pgtype.Numeric, error) {
	if s == "" {
		return pgtype.Numeric{Valid: false}, nil
	}
	var n pgtype.Numeric
	if err := n.Scan(s); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}

func (s *orderService) CreateOrder(ctx context.Context, userID *int64, req dtos.CreateOrderRequest) (dtos.OrderResponse, error) {
	isGuest := userID == nil
	if err := validateCreateOrderRequest(req, isGuest); err != nil {
		return dtos.OrderResponse{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return dtos.OrderResponse{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)

	var total pgtype.Numeric
	var items []orderItem

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return dtos.OrderResponse{}, ErrInvalidQuantity
		}

		stockRow, err := q.DecrementProductStock(ctx, db.DecrementProductStockParams{
			ID:    item.ProductID,
			Stock: item.Quantity,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return dtos.OrderResponse{}, ErrProductNotFound
			}
			return dtos.OrderResponse{}, fmt.Errorf("decrement stock for product %d: %w", item.ProductID, err)
		}

		itemTotal, err := s.mulNumeric(stockRow.Price, item.Quantity)
		if err != nil {
			return dtos.OrderResponse{}, fmt.Errorf("calc item total: %w", err)
		}

		total, err = s.addNumeric(total, itemTotal)
		if err != nil {
			return dtos.OrderResponse{}, fmt.Errorf("acc total: %w", err)
		}

		items = append(items, orderItem{
			productID: item.ProductID,
			quantity:  item.Quantity,
			price:     stockRow.Price,
		})
	}

	// Add delivery cost to total
	deliveryCost, err := parseNumeric(req.Delivery.Cost)
	if err != nil {
		return dtos.OrderResponse{}, fmt.Errorf("parse delivery cost: %w", err)
	}
	total, err = s.addNumeric(total, deliveryCost)
	if err != nil {
		return dtos.OrderResponse{}, fmt.Errorf("add delivery cost: %w", err)
	}

	// Build params
	params := db.CreateOrderParams{
		Total:         total,
		Address:       req.Delivery.Address,
		DeliveryMethod: pgtype.Text{String: req.Delivery.Method, Valid: req.Delivery.Method != ""},
		DeliveryCost:  deliveryCost,
	}

	if userID != nil {
		params.UserID = pgtype.Int8{Int64: *userID, Valid: true}
	} else {
		params.UserID = pgtype.Int8{Valid: false}
		params.PublicToken = pgtype.Text{String: generatePublicToken(), Valid: true}
		if req.Customer != nil {
			params.CustomerFirstName = pgtype.Text{String: req.Customer.FirstName, Valid: req.Customer.FirstName != ""}
			params.CustomerLastName = pgtype.Text{String: req.Customer.LastName, Valid: req.Customer.LastName != ""}
			params.CustomerEmail = pgtype.Text{String: req.Customer.Email, Valid: req.Customer.Email != ""}
			params.CustomerPhone = pgtype.Text{String: req.Customer.Phone, Valid: req.Customer.Phone != ""}
		}
	}

	order, err := q.CreateOrder(ctx, params)
	if err != nil {
		return dtos.OrderResponse{}, fmt.Errorf("create order: %w", err)
	}

	for _, item := range items {
		_, err := q.CreateOrderItem(ctx, db.CreateOrderItemParams{
			OrderID:   order.ID,
			ProductID: item.productID,
			Quantity:  item.quantity,
			Price:     item.price,
		})
		if err != nil {
			return dtos.OrderResponse{}, fmt.Errorf("create order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return dtos.OrderResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	// Send notifications asynchronously (non-blocking)
	go func() {
		customerEmail := ""
		customerName := "Покупатель"
		if req.Customer != nil {
			customerEmail = req.Customer.Email
			customerName = req.Customer.FirstName
		}
		if userID != nil {
			user, err := s.solver.GetUserByID(ctx, *userID)
			if err == nil {
				customerEmail = user.Email
				customerName = user.Username
			}
		}

		// Build items for notification (fetch product titles)
		notifItems := make([]notifications.OrderItem, 0, len(items))
		for _, oi := range items {
			product, err := s.repo.GetProduct(ctx, oi.productID)
			if err != nil {
				notifItems = append(notifItems, notifications.OrderItem{
					Title:    "Товар",
					Quantity: oi.quantity,
					Price:    dtos.NumericToStr(oi.price),
					Total:    dtos.NumericToStr(oi.price),
				})
				continue
			}
			notifItems = append(notifItems, notifications.OrderItem{
				Title:    product.Title,
				Quantity: oi.quantity,
				Price:    dtos.NumericToStr(oi.price),
				Total:    dtos.NumericToStr(oi.price),
			})
		}

		// Send confirmation to buyer
		if customerEmail != "" {
			if err := s.notifications.SendOrderConfirmation(order.ID, customerEmail, customerName, notifItems, dtos.NumericToStr(total)); err != nil {
				s.log.Error("failed to send order confirmation email", slog.String("error", err.Error()), slog.Int64("order_id", order.ID))
			}
		}

		// Send notification to sellers
		for _, oi := range items {
			storeID, err := s.repo.GetProductStoreID(ctx, oi.productID)
			if err != nil {
				continue
			}
			storeOwner, err := s.solver.GetStoreOwnerByStoreID(ctx, storeID)
			if err != nil {
				continue
			}
			storeOwnerUser, err := s.solver.GetUserByID(ctx, storeOwner)
			if err != nil {
				continue
			}
			if err := s.notifications.SendSellerNewOrder(storeOwnerUser.Email, storeOwnerUser.Username, order.ID, dtos.NumericToStr(total), customerEmail, customerName, notifItems); err != nil {
				s.log.Error("failed to send seller notification", slog.String("error", err.Error()), slog.Int64("order_id", order.ID))
			}
		}
	}()

	if userID != nil {
		return s.GetOrder(ctx, *userID, order.ID)
	}
	return s.GetOrderByPublicToken(ctx, order.PublicToken.String)
}

func (s *orderService) isSellerOfOrder(ctx context.Context, userID, orderID int64, items []db.GetOrderItemsRow) (bool, error) {
	for _, item := range items {
		storeID, err := s.repo.GetProductStoreID(ctx, item.ProductID)
		if err != nil {
			return false, fmt.Errorf("get product store: %w", err)
		}
		storeOwner, err := s.solver.GetStoreOwnerByStoreID(ctx, storeID)
		if err != nil {
			return false, fmt.Errorf("get store owner: %w", err)
		}
		if storeOwner == userID {
			return true, nil
		}
	}
	return false, nil
}

func (s *orderService) GetOrder(ctx context.Context, userID, orderID int64) (dtos.OrderResponse, error) {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dtos.OrderResponse{}, ErrOrderNotFound
		}
		return dtos.OrderResponse{}, fmt.Errorf("get order: %w", err)
	}

	items, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return dtos.OrderResponse{}, fmt.Errorf("get order items: %w", err)
	}

	if !order.UserID.Valid || order.UserID.Int64 != userID {
		isSeller, err := s.isSellerOfOrder(ctx, userID, orderID, items)
		if err != nil {
			return dtos.OrderResponse{}, err
		}
		if !isSeller {
			return dtos.OrderResponse{}, ErrForbidden
		}
	}

	return dtos.ToOrderResponse(order, items), nil
}

func (s *orderService) GetOrderByPublicToken(ctx context.Context, token string) (dtos.OrderResponse, error) {
	order, err := s.repo.GetOrderByPublicToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dtos.OrderResponse{}, ErrOrderNotFound
		}
		return dtos.OrderResponse{}, fmt.Errorf("get order by public token: %w", err)
	}

	items, err := s.repo.GetOrderItems(ctx, order.ID)
	if err != nil {
		return dtos.OrderResponse{}, fmt.Errorf("get order items: %w", err)
	}

	return dtos.ToOrderResponse(order, items), nil
}

func (s *orderService) ListOrdersByUser(ctx context.Context, userID int64, page, limit int) ([]dtos.OrderListItem, int64, error) {
	orders, total, err := s.repo.ListOrderByUser(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders by user: %w", err)
	}

	if len(orders) == 0 {
		return []dtos.OrderListItem{}, 0, nil
	}

	result := make([]dtos.OrderListItem, 0, len(orders))
	for _, order := range orders {
		count, err := s.repo.CountOrderItems(ctx, order.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("count order items for %d: %w", order.ID, err)
		}
		result = append(result, dtos.ToOrderListItem(order, int32(count)))
	}

	return result, total, nil
}

func (s *orderService) ListOrdersBySeller(ctx context.Context, userID int64, page, limit int) ([]dtos.OrderSellerListItem, int64, error) {
	orders, total, err := s.repo.ListOrdersBySeller(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders by seller: %w", err)
	}

	if len(orders) == 0 {
		return []dtos.OrderSellerListItem{}, 0, nil
	}

	result := make([]dtos.OrderSellerListItem, 0, len(orders))
	for _, order := range orders {
		orderItems, err := s.repo.GetOrderItems(ctx, order.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("get order items for %d: %w", order.ID, err)
		}

		buyerUsername := ""
		if order.UserID.Valid {
			user, err := s.solver.GetUserByID(ctx, order.UserID.Int64)
			if err == nil {
				buyerUsername = user.Username
			}
		}

		result = append(result, dtos.ToOrderSellerListItem(order, buyerUsername, orderItems))
	}

	return result, total, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, userID int64, orderID int64, req dtos.UpdateOrderStatusRequest) (dtos.OrderResponse, error) {
	switch req.Status {
	case db.OrderStatusPending, db.OrderStatusPaid, db.OrderStatusShipped, db.OrderStatusDelivered, db.OrderStatusCancelled:
	default:
		return dtos.OrderResponse{}, ErrInvalidStatus
	}

	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dtos.OrderResponse{}, ErrOrderNotFound
		}
		return dtos.OrderResponse{}, fmt.Errorf("get order: %w", err)
	}

	items, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return dtos.OrderResponse{}, fmt.Errorf("get order items: %w", err)
	}

	isSeller, err := s.isSellerOfOrder(ctx, userID, orderID, items)
	if err != nil {
		return dtos.OrderResponse{}, err
	}

	isBuyer := order.UserID.Valid && order.UserID.Int64 == userID

	if !isSeller && !isBuyer {
		return dtos.OrderResponse{}, ErrForbidden
	}

	if req.Status == db.OrderStatusCancelled {
		if order.Status == db.OrderStatusPending {
			// pending: покупатель или продавец может отменить
		} else if order.Status == db.OrderStatusPaid && isBuyer && !isSeller {
			// paid: только покупатель может отменить (продавец может shipped/cancelled)
		} else if order.Status == db.OrderStatusPaid && isSeller && !isBuyer {
			// paid: продавец может отменить
		} else {
			return dtos.OrderResponse{}, ErrInvalidStatusTrans
		}
	} else {
		// Не отмена — проверяем обычные переходы
		var validTransitions map[db.OrderStatus][]db.OrderStatus
		if isSeller && !isBuyer {
			validTransitions = map[db.OrderStatus][]db.OrderStatus{
				db.OrderStatusPending: {db.OrderStatusCancelled},
				db.OrderStatusPaid:    {db.OrderStatusShipped, db.OrderStatusCancelled},
				db.OrderStatusShipped: {db.OrderStatusDelivered},
			}
		} else {
			validTransitions = map[db.OrderStatus][]db.OrderStatus{
				db.OrderStatusPending: {db.OrderStatusCancelled},
				db.OrderStatusPaid:    {db.OrderStatusCancelled},
			}
		}

		prevTransitions := validTransitions[order.Status]
		allowed := false
		for _, st := range prevTransitions {
			if st == req.Status {
				allowed = true
				break
			}
		}
		if !allowed {
			return dtos.OrderResponse{}, ErrInvalidStatusTrans
		}
	}

	// Отмена заказа — возврат товара и средств в одной транзакции
	if req.Status == db.OrderStatusCancelled {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return dtos.OrderResponse{}, fmt.Errorf("begin tx: %w", err)
		}
		defer tx.Rollback(ctx)

		q := db.New(tx)

		orderItems, err := q.GetOrderItemsByOrderID(ctx, orderID)
		if err != nil {
			return dtos.OrderResponse{}, fmt.Errorf("get order items for stock return: %w", err)
		}

		for _, oi := range orderItems {
			if _, err := q.IncrementProductStock(ctx, db.IncrementProductStockParams{
				ID:    oi.ProductID,
				Stock: oi.Quantity,
			}); err != nil {
				return dtos.OrderResponse{}, fmt.Errorf("increment stock for product %d: %w", oi.ProductID, err)
			}
		}

		updated, err := q.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
			ID:     orderID,
			Status: req.Status,
		})
		if err != nil {
			return dtos.OrderResponse{}, fmt.Errorf("update order status: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return dtos.OrderResponse{}, fmt.Errorf("commit tx: %w", err)
		}

		// Возврат средств после коммита — не блокирует транзакцию
		if order.Status == db.OrderStatusPaid && order.UserID.Valid {
			payment, err := s.repo.GetPaymentByOrderID(ctx, orderID)
			if err == nil {
				_, refundErr := s.paymentRef.RefundPayment(ctx, order.UserID.Int64, payment.ID)
				if refundErr != nil {
					s.log.Warn("refund failed for cancelled order", "order_id", orderID, "error", refundErr)
				}
			}
		}

		return dtos.ToOrderResponse(updated, items), nil
	}

	updated, err := s.repo.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     orderID,
		Status: req.Status,
	})
	if err != nil {
		return dtos.OrderResponse{}, fmt.Errorf("update order status: %w", err)
	}

	// Send status update notification to buyer
	if order.UserID.Valid {
		go func() {
			user, err := s.solver.GetUserByID(ctx, order.UserID.Int64)
			if err != nil {
				return
			}
			_ = s.notifications.SendOrderStatusUpdate(order.ID, user.Email, user.Username, string(req.Status))
		}()
	}

	return dtos.ToOrderResponse(updated, items), nil
}

func (s *orderService) mulNumeric(n pgtype.Numeric, factor int32) (pgtype.Numeric, error) {
	if !n.Valid {
		return pgtype.Numeric{Valid: false}, nil
	}

	str := dtos.NumericToStr(n)
	f, _, err := big.ParseFloat(str, 10, 256, big.ToNearestEven)
	if err != nil {
		return pgtype.Numeric{}, fmt.Errorf("parse numeric %s: %w", str, err)
	}

	multiplier := big.NewFloat(float64(factor))
	result := new(big.Float).Mul(f, multiplier)

	resultStr := result.Text('f', 2)
	var res pgtype.Numeric
	if err := res.Scan(resultStr); err != nil {
		return pgtype.Numeric{}, err
	}
	return res, nil
}

func (s *orderService) addNumeric(a, b pgtype.Numeric) (pgtype.Numeric, error) {
	if !a.Valid && !b.Valid {
		return pgtype.Numeric{Valid: false}, nil
	}
	if !a.Valid {
		return b, nil
	}
	if !b.Valid {
		return a, nil
	}

	aStr := dtos.NumericToStr(a)
	bStr := dtos.NumericToStr(b)

	var aFloat, bFloat big.Float
	aFloat.SetString(aStr)
	bFloat.SetString(bStr)

	result := new(big.Float).Add(&aFloat, &bFloat)

	resultStr := result.Text('f', 2)
	var res pgtype.Numeric
	if err := res.Scan(resultStr); err != nil {
		return pgtype.Numeric{}, err
	}
	return res, nil
}

func (s *orderService) UpdateOrderTracking(ctx context.Context, userID int64, orderID int64, req dtos.UpdateTrackingRequest) (dtos.OrderResponse, error) {
	_, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dtos.OrderResponse{}, ErrOrderNotFound
		}
		return dtos.OrderResponse{}, fmt.Errorf("get order: %w", err)
	}

	items, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return dtos.OrderResponse{}, fmt.Errorf("get order items: %w", err)
	}

	isSeller, err := s.isSellerOfOrder(ctx, userID, orderID, items)
	if err != nil {
		return dtos.OrderResponse{}, err
	}
	if !isSeller {
		return dtos.OrderResponse{}, ErrTrackingOnlySeller
	}

	updated, err := s.repo.UpdateOrderTracking(ctx, orderID, req.TrackingNumber)
	if err != nil {
		return dtos.OrderResponse{}, fmt.Errorf("update tracking: %w", err)
	}

	return dtos.ToOrderResponse(updated, items), nil
}
