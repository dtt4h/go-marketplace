-- name: CreateOrder :one
INSERT INTO orders (
    user_id, status, total, address,
    public_token, customer_first_name, customer_last_name,
    customer_email, customer_phone, delivery_method, delivery_cost
)
VALUES ($1, 'pending', $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, quantity, price)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrder :one
SELECT * FROM orders WHERE id = $1;

-- name: GetOrderByPublicToken :one
SELECT * FROM orders WHERE public_token = $1;

-- name: ListOrdersByUser :many
SELECT * FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOrdersByUserCount :one
SELECT COUNT(*) FROM orders WHERE user_id = $1;

-- name: ListOrdersBySeller :many
SELECT DISTINCT o.*
FROM orders o
JOIN order_items oi ON oi.order_id = o.id
JOIN products p ON p.id = oi.product_id
JOIN stores s ON s.id = p.store_id
WHERE s.user_id = $1
ORDER BY o.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOrdersBySellerCount :one
SELECT COUNT(DISTINCT o.id)
FROM orders o
JOIN order_items oi ON oi.order_id = o.id
JOIN products p ON p.id = oi.product_id
JOIN stores s ON s.id = p.store_id
WHERE s.user_id = $1;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateOrderTracking :one
UPDATE orders
SET tracking_number = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetOrderItems :many
SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price,
       p.title AS product_title, p.store_id
FROM order_items oi
JOIN products p ON p.id = oi.product_id
WHERE oi.order_id = $1;

-- name: DecrementProductStock :one
UPDATE products
SET stock = stock - $2
WHERE id = $1 AND stock >= $2
RETURNING id, store_id, title, price, stock;

-- name: GetProductStoreID :one
SELECT store_id FROM products WHERE id = $1;

-- name: CountOrderItems :one
SELECT COUNT(*) FROM order_items WHERE order_id = $1;

-- name: IncrementProductStock :one
UPDATE products
SET stock = stock + $2
WHERE id = $1
RETURNING id, store_id, title, price, stock;

-- name: GetOrderItemsByOrderID :many
SELECT product_id, quantity FROM order_items WHERE order_id = $1;

-- Admin queries

-- name: ListAllOrders :many
SELECT * FROM orders
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAllOrdersCount :one
SELECT COUNT(*) FROM orders;

-- name: ListAllOrdersByStatus :many
SELECT * FROM orders
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllOrdersByStatusCount :one
SELECT COUNT(*) FROM orders WHERE status = $1;

-- name: ListAllOrdersByUser :many
SELECT * FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllOrdersByUserCount :one
SELECT COUNT(*) FROM orders WHERE user_id = $1;

-- Admin statistics queries

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CountSellers :one
SELECT COUNT(*) FROM users WHERE role = 'seller';

-- name: CountProducts :one
SELECT COUNT(*) FROM products;

-- name: CountOrders :one
SELECT COUNT(*) FROM orders;

-- name: CountOrdersByStatus :one
SELECT COUNT(*) FROM orders WHERE status = $1;

-- name: SumOrderTotals :one
SELECT COALESCE(SUM(total), 0) FROM orders WHERE status != 'cancelled';
