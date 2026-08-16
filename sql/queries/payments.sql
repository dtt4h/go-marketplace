-- name: CreatePayment :one
INSERT INTO payments (order_id, amount, currency, status, provider, provider_payment_id)
VALUES ($1, $2, $3, 'pending', $4, $5)
RETURNING id, order_id, amount, currency, status, provider, provider_payment_id, created_at, updated_at;

-- name: GetPayment :one
SELECT id, order_id, amount, currency, status, provider, provider_payment_id, created_at, updated_at
FROM payments
WHERE id = $1;

-- name: GetPaymentByOrderID :one
SELECT id, order_id, amount, currency, status, provider, provider_payment_id, created_at, updated_at
FROM payments
WHERE order_id = $1;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $2,
    provider_payment_id = COALESCE($3, provider_payment_id),
    updated_at = now()
WHERE id = $1
RETURNING id, order_id, amount, currency, status, provider, provider_payment_id, created_at, updated_at;

-- name: ListPaymentsByUserID :many
SELECT p.id, p.order_id, p.amount, p.currency, p.status, p.provider, p.provider_payment_id, p.created_at, p.updated_at
FROM payments p
JOIN orders o ON o.id = p.order_id
WHERE o.user_id = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPaymentsByUserIDCount :one
SELECT COUNT(*)
FROM payments p
JOIN orders o ON o.id = p.order_id
WHERE o.user_id = $1;
