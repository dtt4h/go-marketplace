package orders

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderExpirer handles automatic expiration of pending orders.
type OrderExpirer struct {
	queries *db.Queries
	pool    *pgxpool.Pool
	log     *slog.Logger
	done    chan struct{}
}

// NewOrderExpirer creates a new OrderExpirer.
func NewOrderExpirer(queries *db.Queries, pool *pgxpool.Pool, log *slog.Logger) *OrderExpirer {
	return &OrderExpirer{queries: queries, pool: pool, log: log, done: make(chan struct{})}
}

// Stop signals the expirer to stop.
func (e *OrderExpirer) Stop() {
	select {
	case <-e.done:
	default:
		close(e.done)
	}
}

// ExpirePendingOrders cancels all pending orders whose expires_at has passed,
// and restores their product stock in a single transaction.
func (e *OrderExpirer) ExpirePendingOrders(ctx context.Context) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := e.queries.WithTx(tx)

	// Step 1: Get expired order IDs
	expiredIDs, err := q.ExpirePendingOrders(ctx)
	if err != nil {
		return fmt.Errorf("expire pending orders: %w", err)
	}

	if len(expiredIDs) == 0 {
		return nil
	}

	e.log.Info("expiring orders", slog.Int("count", len(expiredIDs)))

	// Step 2: Restore stock for each expired order
	for _, orderID := range expiredIDs {
		items, err := q.GetOrderItemsByOrderID(ctx, orderID)
		if err != nil {
			e.log.Error("failed to get order items for stock restore", slog.Int64("order_id", orderID), slog.String("error", err.Error()))
			continue
		}

		for _, item := range items {
			if _, err := q.IncrementProductStock(ctx, db.IncrementProductStockParams{
				ID:    item.ProductID,
				Stock: item.Quantity,
			}); err != nil {
				e.log.Error("failed to restore stock", slog.Int64("product_id", item.ProductID), slog.Int("quantity", int(item.Quantity)), slog.String("error", err.Error()))
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	e.log.Info("expired orders processed", slog.Int("count", len(expiredIDs)))
	return nil
}
