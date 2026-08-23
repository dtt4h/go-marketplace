-- cart_items

-- name: UpsertCartItem :one
INSERT INTO cart_items (user_id, product_id, quantity)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, product_id)
DO UPDATE SET quantity = EXCLUDED.quantity, updated_at = now()
RETURNING id, user_id, product_id, quantity, created_at, updated_at;

-- name: GetCartItemsByUserID :many
SELECT ci.id, ci.user_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
       p.title, p.price, p.stock, p.status,
       s.name AS store_name, s.id AS store_id,
       (
           SELECT COALESCE(pi.url, '') FROM product_images pi
           WHERE pi.product_id = ci.product_id
           ORDER BY pi.position ASC LIMIT 1
       ) AS preview_image_url
FROM cart_items ci
JOIN products p ON p.id = ci.product_id
JOIN stores s ON s.id = p.store_id
WHERE ci.user_id = $1
ORDER BY ci.created_at ASC;

-- name: GetCartItemByUserIDAndProductID :one
SELECT id, user_id, product_id, quantity, created_at, updated_at
FROM cart_items
WHERE user_id = $1 AND product_id = $2;

-- name: GetCartByID :one
SELECT id, user_id, product_id, quantity, created_at, updated_at
FROM cart_items
WHERE id = $1 AND user_id = $2;

-- name: UpdateCartItemQuantity :one
UPDATE cart_items
SET quantity = $3, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, product_id, quantity, created_at, updated_at;

-- name: DeleteCartItem :exec
DELETE FROM cart_items
WHERE id = $1 AND user_id = $2;

-- name: DeleteCartItemsByUserID :exec
DELETE FROM cart_items
WHERE user_id = $1;
