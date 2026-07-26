package products

import (
	"context"
	"fmt"
	"strings"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository interface {
	ListCategories(ctx context.Context) ([]db.ListCategoriesRow, error)
	ListProducts(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]db.ListProductsRow, int64, error)
	GetProduct(ctx context.Context, id int64) (db.GetProductRow, error)
	CreateProductWithImages(ctx context.Context, storeID int64, categoryID *int64, title, description *string, price string, stock int32, images []string) (db.Product, error)
	UpdateProduct(ctx context.Context, id int64, title, description *string, price *string, stock *int32) (db.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
	ModerateProduct(ctx context.Context, id int64, status db.ProductStatus) error
	ListProductImages(ctx context.Context, productID int64) ([]db.ProductImage, error)
	ListProductImagesByProductIDs(ctx context.Context, productIDs []int64) ([]db.ProductImage, error)
	GetProductStoreOwner(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error)
}

type productRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewProductRepository(queries *db.Queries, pool *pgxpool.Pool) ProductRepository {
	return &productRepository{queries: queries, pool: pool}
}

func (r *productRepository) ListCategories(ctx context.Context) ([]db.ListCategoriesRow, error) {
	return r.queries.ListCategories(ctx)
}

func (r *productRepository) ListProducts(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]db.ListProductsRow, int64, error) {
	if sort == "" {
		sort = "created_desc"
	}

	minP, err := toNumeric(minPrice)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid min_price: %w", err)
	}
	maxP, err := toNumeric(maxPrice)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid max_price: %w", err)
	}

	offset := int32((page - 1) * limit)

	products, err := r.queries.ListProducts(ctx, db.ListProductsParams{
		StoreID:     pgutil.NullInt8(storeID),
		CategoryID:  pgutil.NullInt8(categoryID),
		MinPrice:    minP,
		MaxPrice:    maxP,
		Search:      toSearchText(search),
		Sort:        sort,
		LimitCount:  int32(limit),
		OffsetCount: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}

	count, err := r.queries.ListProductsCount(ctx, db.ListProductsCountParams{
		StoreID:    pgutil.NullInt8(storeID),
		CategoryID: pgutil.NullInt8(categoryID),
		MinPrice:   minP,
		MaxPrice:   maxP,
		Search:     toSearchText(search),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list products count: %w", err)
	}

	return products, count, nil
}

func (r *productRepository) GetProduct(ctx context.Context, id int64) (db.GetProductRow, error) {
	return r.queries.GetProduct(ctx, id)
}

func (r *productRepository) CreateProductWithImages(ctx context.Context, storeID int64, categoryID *int64, title, description *string, price string, stock int32, images []string) (db.Product, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.Product{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := r.queries.WithTx(tx)

	var catID pgtype.Int8
	if categoryID != nil {
		catID = pgtype.Int8{Int64: *categoryID, Valid: true}
	}

	desc := pgutil.NullText(description)

	var priceNum pgtype.Numeric
	if err := priceNum.Scan(price); err != nil {
		return db.Product{}, fmt.Errorf("scan price: %w", err)
	}

	product, err := q.CreateProduct(ctx, db.CreateProductParams{
		StoreID:     storeID,
		CategoryID:  catID,
		Title:       *title,
		Description: desc,
		Price:       priceNum,
		Stock:       stock,
	})
	if err != nil {
		return db.Product{}, fmt.Errorf("create product: %w", err)
	}

	for i, url := range images {
		if _, err := q.CreateProductImage(ctx, db.CreateProductImageParams{
			ProductID: product.ID,
			Url:       url,
			Position:  int32(i),
		}); err != nil {
			return db.Product{}, fmt.Errorf("create product image: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Product{}, fmt.Errorf("commit tx: %w", err)
	}

	return product, nil
}

func (r *productRepository) UpdateProduct(ctx context.Context, id int64, title, description *string, price *string, stock *int32) (db.Product, error) {
	params := db.UpdateProductParams{
		ID: id,
	}

	if title != nil {
		params.Title = pgutil.NullText(title)
	}
	if description != nil {
		params.Description = pgutil.NullText(description)
	}
	if price != nil {
		if err := params.Price.Scan(*price); err != nil {
			return db.Product{}, fmt.Errorf("scan price: %w", err)
		}
	}
	if stock != nil {
		params.Stock = pgtype.Int4{Int32: *stock, Valid: true}
	}

	return r.queries.UpdateProduct(ctx, params)
}

func (r *productRepository) DeleteProduct(ctx context.Context, id int64) error {
	return r.queries.DeleteProduct(ctx, id)
}

func (r *productRepository) ModerateProduct(ctx context.Context, id int64, status db.ProductStatus) error {
	_, err := r.queries.ModerateProduct(ctx, db.ModerateProductParams{
		ID:     id,
		Status: status,
	})
	return err
}

func (r *productRepository) ListProductImages(ctx context.Context, productID int64) ([]db.ProductImage, error) {
	return r.queries.ListProductImages(ctx, productID)
}

func (r *productRepository) ListProductImagesByProductIDs(ctx context.Context, productIDs []int64) ([]db.ProductImage, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}
	return r.queries.ListProductImagesByProductIDs(ctx, productIDs)
}

func (r *productRepository) GetProductStoreOwner(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error) {
	return r.queries.GetProductStoreOwner(ctx, productID)
}

func toNumeric(v *string) (pgtype.Numeric, error) {
	if v == nil {
		return pgtype.Numeric{Valid: false}, nil
	}
	var n pgtype.Numeric
	if err := n.Scan(*v); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}

func toSearchText(v *string) pgtype.Text {
	if v == nil || *v == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: "%" + escapeLike(*v) + "%", Valid: true}
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}
