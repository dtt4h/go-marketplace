-- users

-- name: CreateUser :one
INSERT INTO users (email, password_hash, role, username)
VALUES ($1, $2, $3, $4)
RETURNING id, email, password_hash, role, username, avatar_url, phone, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, email, password_hash, role, username, avatar_url, phone, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, role, username, avatar_url, phone, created_at, updated_at
FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET username   = COALESCE(sqlc.narg(username), username),
    avatar_url = COALESCE(sqlc.narg(avatar_url), avatar_url),
    phone      = COALESCE(sqlc.narg(phone), phone)
WHERE id = sqlc.arg(id)
RETURNING id, email, password_hash, role, username, avatar_url, phone, created_at, updated_at;

-- refresh_tokens

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token, expires_at, created_at;

-- name: GetRefreshToken :one
SELECT id, user_id, token, expires_at, created_at
FROM refresh_tokens
WHERE token = $1;

-- name: DeleteRefreshToken :one
DELETE FROM refresh_tokens
WHERE token = $1
RETURNING id, user_id, token, expires_at, created_at;

-- name: DeleteUserRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE user_id = $1;

-- name: GenerateResetToken :one
UPDATE users
SET reset_token = $2,
    reset_token_expires_at = $3
WHERE id = $1
RETURNING id, email, password_hash, role, username, avatar_url, phone, created_at, updated_at;

-- name: GetUserByResetToken :one
SELECT id, email, password_hash, role, username, avatar_url, phone, created_at, updated_at
FROM users
WHERE reset_token = $1
    AND reset_token_expires_at > now();

-- name: ResetPassword :one 
UPDATE users
SET password_hash = $2,
    reset_token = NULL,
    reset_token_expires_at = NULL
WHERE id = $1
RETURNING id, email, password_hash, role, username, avatar_url, phone, created_at, updated_at;

-- name: DeleteResetToken :exec
UPDATE users
SET reset_token = NULL,
    reset_token_expires_at = NULL
WHERE id = $1;