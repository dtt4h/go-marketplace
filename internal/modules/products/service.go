package products

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidPrice    = errors.New("invalid price format")
	ErrInvalidStock    = errors.New("invalid stock value")
	ErrForbidden       = errors.New("access denied")
	ErrInvalidSort     = errors.New("invalid sort parameter")
	ErrInvalidStatus   = errors.New("invalid moderation status")
	ErrStoreNotFound   = errors.New("store not found for this user")
	ErrTitleRequired   = errors.New("title is required")
	ErrTitleTooLong    = errors.New("title must be at most 300 characters")
)

const maxTitleLength = 300

var validSorts = map[string]bool{
	"price_asc":    true,
	"price_desc":   true,
	"created_desc": true,
}

// ProductService defines business logic for product operations.
type ProductService interface {
	ListCategories(ctx context.Context) ([]dtos.CategoryResponse, error)
	ListProducts(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]dtos.ProductListItem, int64, error)
	GetProduct(ctx context.Context, id int64) (dtos.ProductResponse, error)
	CreateProduct(ctx context.Context, userID int64, req dtos.CreateProductRequest) (dtos.ProductResponse, error)
	UpdateProduct(ctx context.Context, userID, id int64, req dtos.UpdateProductRequest) (dtos.ProductResponse, error)
	DeleteProduct(ctx context.Context, userID, id int64) error
	ModerateProduct(ctx context.Context, id int64, req dtos.ModerateProductRequest) (dtos.ProductResponse, error)
}

type productService struct {
	repo      ProductRepository
	storeRepo StoreResolver
}

// StoreResolver provides store lookup capabilities.
type StoreResolver interface {
	GetStoreByUserID(ctx context.Context, userID int64) (db.Store, error)
}

// NewProductService creates a new ProductService.
func NewProductService(repo ProductRepository, storeRepo StoreResolver) ProductService {
	return &productService{repo: repo, storeRepo: storeRepo}
}

func (s *productService) ListCategories(ctx context.Context) ([]dtos.CategoryResponse, error) {
	rows, err := s.repo.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return dtos.BuildCategoryTree(rows), nil
}

func (s *productService) ListProducts(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]dtos.ProductListItem, int64, error) {
	if sort == "" {
		sort = "created_desc"
	}
	if !validSorts[sort] {
		return nil, 0, ErrInvalidSort
	}

	rows, total, err := s.repo.ListProducts(ctx, storeID, categoryID, minPrice, maxPrice, search, sort, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}

	if len(rows) == 0 {
		return []dtos.ProductListItem{}, 0, nil
	}

	productIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		productIDs = append(productIDs, row.ID)
	}

	allImages, err := s.repo.ListProductImagesByProductIDs(ctx, productIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("list product images batch: %w", err)
	}

	imagesByProduct := make(map[int64][]db.ProductImage)
	for _, img := range allImages {
		imagesByProduct[img.ProductID] = append(imagesByProduct[img.ProductID], img)
	}

	items := make([]dtos.ProductListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dtos.ToProductListItem(row, imagesByProduct[row.ID]))
	}

	return items, total, nil
}

func (s *productService) GetProduct(ctx context.Context, id int64) (dtos.ProductResponse, error) {
	row, err := s.repo.GetProduct(ctx, id)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.ProductResponse{}, ErrProductNotFound
		}
		return dtos.ProductResponse{}, fmt.Errorf("get product: %w", err)
	}

	images, err := s.repo.ListProductImages(ctx, id)
	if err != nil {
		return dtos.ProductResponse{}, fmt.Errorf("list product images: %w", err)
	}

	return dtos.ToProductResponse(row, images), nil
}

func (s *productService) CreateProduct(ctx context.Context, userID int64, req dtos.CreateProductRequest) (dtos.ProductResponse, error) {
	if req.Title == "" {
		return dtos.ProductResponse{}, ErrTitleRequired
	}
	if len(req.Title) > maxTitleLength {
		return dtos.ProductResponse{}, ErrTitleTooLong
	}
	if req.Price == "" {
		return dtos.ProductResponse{}, ErrInvalidPrice
	}

	priceVal, err := strconv.ParseFloat(req.Price, 64)
	if err != nil || priceVal <= 0 {
		return dtos.ProductResponse{}, ErrInvalidPrice
	}
	if req.Stock < 0 {
		return dtos.ProductResponse{}, ErrInvalidStock
	}

	store, err := s.storeRepo.GetStoreByUserID(ctx, userID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.ProductResponse{}, ErrStoreNotFound
		}
		return dtos.ProductResponse{}, fmt.Errorf("get store: %w", err)
	}

	product, err := s.repo.CreateProductWithImages(ctx, store.ID, req.CategoryID, &req.Title, req.Description, req.Price, int32(req.Stock), req.Images)
	if err != nil {
		return dtos.ProductResponse{}, fmt.Errorf("create product with images: %w", err)
	}

	return s.GetProduct(ctx, product.ID)
}

func (s *productService) UpdateProduct(ctx context.Context, userID, id int64, req dtos.UpdateProductRequest) (dtos.ProductResponse, error) {
	if err := s.checkOwnership(ctx, userID, id); err != nil {
		return dtos.ProductResponse{}, err
	}

	if req.Title != nil {
		if *req.Title == "" {
			return dtos.ProductResponse{}, ErrTitleRequired
		}
		if len(*req.Title) > maxTitleLength {
			return dtos.ProductResponse{}, ErrTitleTooLong
		}
	}
	if req.Price != nil {
		priceVal, err := strconv.ParseFloat(*req.Price, 64)
		if err != nil || priceVal <= 0 {
			return dtos.ProductResponse{}, ErrInvalidPrice
		}
	}
	if req.Stock != nil && *req.Stock < 0 {
		return dtos.ProductResponse{}, ErrInvalidStock
	}

	var stockPtr *int32
	if req.Stock != nil {
		v := int32(*req.Stock)
		stockPtr = &v
	}

	if _, err := s.repo.UpdateProduct(ctx, id, req.Title, req.Description, req.Price, stockPtr); err != nil {
		return dtos.ProductResponse{}, fmt.Errorf("update product: %w", err)
	}

	return s.GetProduct(ctx, id)
}

func (s *productService) DeleteProduct(ctx context.Context, userID, id int64) error {
	if err := s.checkOwnership(ctx, userID, id); err != nil {
		return err
	}
	return s.repo.DeleteProduct(ctx, id)
}

func (s *productService) ModerateProduct(ctx context.Context, id int64, req dtos.ModerateProductRequest) (dtos.ProductResponse, error) {
	switch req.Status {
	case db.ProductStatusActive, db.ProductStatusRejected, db.ProductStatusArchived:
	default:
		return dtos.ProductResponse{}, ErrInvalidStatus
	}

	if err := s.repo.ModerateProduct(ctx, id, req.Status); err != nil {
		return dtos.ProductResponse{}, fmt.Errorf("moderate product: %w", err)
	}

	return s.GetProduct(ctx, id)
}

func (s *productService) checkOwnership(ctx context.Context, userID, productID int64) error {
	row, err := s.repo.GetProductStoreOwner(ctx, productID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return ErrProductNotFound
		}
		return fmt.Errorf("check ownership: %w", err)
	}
	if row.StoreOwnerID != userID {
		return ErrForbidden
	}
	return nil
}