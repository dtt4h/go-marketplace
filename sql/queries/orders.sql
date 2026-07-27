-- name: CreateOrder :one
INSERT INTO orders (user_id, status, total, address)
VALUES ($1, 'pending', $2, $3)
RETURNING id, user_id, status, total, address, created_at, updated_at;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, quantity, price)
VALUES ($1, $2, $3, $4)
RETURNING id, order_id, product_id, quantity, price;

-- name: GetOrder :one
SELECT o.id, o.user_id, o.status, o.total, o.address, o.created_at, o.updated_at
FROM orders o
WHERE o.id = $1;

-- name: GetOrderWithItems :one
SELECT o.id, o.user_id, o.status, o.total, o.address, o.created_at, o.updated_at,
       oi.id AS item_id, oi.product_id, oi.quantity, oi.price,
       p.title AS product_title, p.store_id
FROM orders o
JOIN order_items oi ON oi.order_id = o.id
JOIN products p ON p.id = oi.product_id
WHERE o.id = $1;

-- name: ListOrdersByUser :many
SELECT o.id, o.user_id, o.status, o.total, o.address, o.created_at, o.updated_at
FROM orders o
WHERE o.user_id = $1
ORDER BY o.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOrdersByUserCount :one
SELECT COUNT(*)
FROM orders o
WHERE o.user_id = $1;

-- name: ListOrdersBySeller :many
SELECT DISTINCT o.id, o.user_id, o.status, o.total, o.address, o.created_at, o.updated_at
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
SET status = $2,
    updated_at = now()
WHERE id = $1
RETURNING id, user_id, status, total, address, created_at, updated_at;

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
