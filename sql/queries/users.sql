-- name: GetStoreByUserID :one
SELECT id, user_id, name, description, logo_url, created_at, updated_at
FROM stores
WHERE user_id = $1;

-- name: CreateStore :one
INSERT INTO stores (user_id, name, description, logo_url)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, name, description, logo_url, created_at, updated_at;

-- name: UpdateUserRole :one
UPDATE users
SET role = $2
WHERE id = $1
RETURNING id, email, password_hash, role, username, avatar_url, phone, created_at, updated_at;

-- name: GetStoreOwnerByStoreID :one
SELECT s.user_id
FROM stores s
WHERE s.id = $1;