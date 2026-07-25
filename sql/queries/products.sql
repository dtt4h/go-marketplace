-- name: ListCategories :many
WITH recursive category_tree AS (
    SELECT id, parent_id, name, slug, 0 AS depth
    FROM categories
    WHERE parent_id IS NULL
    UNION ALL
    SELECT c.id, c.parent_id, c.name, c.slug, ct.depth + 1
    FROM categories c
    INNER JOIN category_tree ct ON c.parent_id = ct.id
)
SELECT id, parent_id, name, slug
FROM category_tree
ORDER BY id;

-- name: GetProduct :one
SELECT p.id, p.store_id, p.category_id, p.title, p.description,
       p.price, p.stock, p.status, p.created_at, p.updated_at,
       s.name AS store_name, s.description AS store_description,
       c.name AS category_name, c.slug AS category_slug
FROM products p
LEFT JOIN stores s ON p.store_id = s.id
LEFT JOIN categories c ON p.category_id = c.id
WHERE p.id = $1;

-- name: ListProducts :many
SELECT p.id, p.store_id, p.category_id, p.title, p.description,
       p.price, p.stock, p.status, p.created_at, p.updated_at,
       s.name AS store_name, s.description AS store_description,
       c.name AS category_name, c.slug AS category_slug
FROM products p
LEFT JOIN stores s ON p.store_id = s.id
LEFT JOIN categories c ON p.category_id = c.id
WHERE (sqlc.narg(store_id)::bigint IS NULL OR p.store_id = sqlc.narg(store_id))
  AND (sqlc.narg(category_id)::bigint IS NULL OR p.category_id = sqlc.narg(category_id))
  AND (sqlc.narg(min_price)::numeric IS NULL OR p.price >= sqlc.narg(min_price))
  AND (sqlc.narg(max_price)::numeric IS NULL OR p.price <= sqlc.narg(max_price))
  AND (sqlc.narg(search)::text IS NULL OR p.title ILIKE sqlc.narg(search))
  AND p.status = 'active'
ORDER BY
  CASE WHEN sqlc.arg(sort)::text = 'price_asc' THEN p.price END ASC,
  CASE WHEN sqlc.arg(sort)::text = 'price_desc' THEN p.price END DESC,
  CASE WHEN sqlc.arg(sort)::text = 'created_desc' THEN p.created_at END DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: ListProductsCount :one
SELECT COUNT(*)
FROM products p
WHERE (sqlc.narg(store_id)::bigint IS NULL OR p.store_id = sqlc.narg(store_id))
  AND (sqlc.narg(category_id)::bigint IS NULL OR p.category_id = sqlc.narg(category_id))
  AND (sqlc.narg(min_price)::numeric IS NULL OR p.price >= sqlc.narg(min_price))
  AND (sqlc.narg(max_price)::numeric IS NULL OR p.price <= sqlc.narg(max_price))
  AND (sqlc.narg(search)::text IS NULL OR p.title ILIKE sqlc.narg(search))
  AND p.status = 'active';

-- name: CreateProduct :one
INSERT INTO products (store_id, category_id, title, description, price, stock, status)
VALUES ($1, $2, $3, $4, $5, $6, 'pending')
RETURNING id, store_id, category_id, title, description, price, stock, status, created_at, updated_at;

-- name: UpdateProduct :one
UPDATE products
SET title       = COALESCE(sqlc.narg(title), title),
    description = COALESCE(sqlc.narg(description), description),
    price       = COALESCE(sqlc.narg(price), price),
    stock       = COALESCE(sqlc.narg(stock), stock)
WHERE id = sqlc.arg(id)
RETURNING id, store_id, category_id, title, description, price, stock, status, created_at, updated_at;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = $1;

-- name: ModerateProduct :execresult
UPDATE products
SET status = $2,
    updated_at = now()
WHERE id = $1;

-- name: ListProductImages :many
SELECT id, product_id, url, position
FROM product_images
WHERE product_id = $1
ORDER BY position ASC;

-- name: ListProductImagesByProductIDs :many
SELECT id, product_id, url, position
FROM product_images
WHERE product_id = ANY($1::bigint[])
ORDER BY product_id, position ASC;

-- name: CreateProductImage :one
INSERT INTO product_images (product_id, url, position)
VALUES ($1, $2, $3)
RETURNING id, product_id, url, position;

-- name: DeleteProductImages :exec
DELETE FROM product_images
WHERE product_id = $1;

-- name: GetProductStoreOwner :one
SELECT p.id, s.user_id AS store_owner_id
FROM products p
JOIN stores s ON p.store_id = s.id
WHERE p.id = $1;