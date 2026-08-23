-- name: CreateSellerApplication :one
INSERT INTO seller_applications (user_id, store_name, description)
VALUES ($1, $2, $3)
RETURNING id, user_id, store_name, description, status, created_at, updated_at;

-- name: GetSellerApplicationByID :one
SELECT id, user_id, store_name, description, status, created_at, updated_at
FROM seller_applications
WHERE id = $1;

-- name: GetPendingSellerApplicationByUserID :one
SELECT id, user_id, store_name, description, status, created_at, updated_at
FROM seller_applications
WHERE user_id = $1 AND status = 'pending'
LIMIT 1;

-- name: ListSellerApplicationsByUser :many
SELECT id, user_id, store_name, description, status, created_at, updated_at
FROM seller_applications
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ListPendingSellerApplications :many
SELECT id, user_id, store_name, description, status, created_at, updated_at
FROM seller_applications
WHERE status = 'pending'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPendingSellerApplications :one
SELECT COUNT(*) FROM seller_applications WHERE status = 'pending';

-- name: UpdateSellerApplicationStatus :one
UPDATE seller_applications
SET status = $2,
    updated_at = now()
WHERE id = $1
RETURNING id, user_id, store_name, description, status, created_at, updated_at;
