package products

import (
	"context"
	"errors"
	"testing"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// --- Mocks ---

type mockProductRepository struct {
	listCategories                func(ctx context.Context) ([]db.ListCategoriesRow, error)
	listProducts                  func(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]db.ListProductsRow, int64, error)
	listProductsByStoreID         func(ctx context.Context, storeID int64, page, limit int) ([]db.ListProductsByStoreIDRow, int64, error)
	getProduct                    func(ctx context.Context, id int64) (db.GetProductRow, error)
	createProductWithImages       func(ctx context.Context, storeID int64, categoryID *int64, title, description *string, price string, stock int32, images []string) (db.Product, error)
	updateProduct                 func(ctx context.Context, id int64, title, description *string, price *string, stock *int32) (db.Product, error)
	deleteProduct                 func(ctx context.Context, id int64) error
	moderateProduct               func(ctx context.Context, id int64, status db.ProductStatus) error
	moderateProductWithReason     func(ctx context.Context, id int64, status db.ProductStatus, reason pgtype.Text) error
	listProductImages             func(ctx context.Context, productID int64) ([]db.ProductImage, error)
	listProductImagesByProductIDs func(ctx context.Context, productIDs []int64) ([]db.ProductImage, error)
	getProductStoreOwner          func(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error)
	getImageByID                  func(ctx context.Context, id int64) (db.ProductImage, error)
	deleteImageByID               func(ctx context.Context, id int64) error
	createProductImage            func(ctx context.Context, productID int64, url string, position int32) (db.ProductImage, error)
	listProductsByStatus          func(ctx context.Context, status db.ProductStatus, page, limit int) ([]db.ListProductsByStatusRow, int64, error)
}

func (m *mockProductRepository) ListCategories(ctx context.Context) ([]db.ListCategoriesRow, error) {
	return m.listCategories(ctx)
}
func (m *mockProductRepository) ListProducts(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]db.ListProductsRow, int64, error) {
	return m.listProducts(ctx, storeID, categoryID, minPrice, maxPrice, search, sort, page, limit)
}
func (m *mockProductRepository) ListProductsByStoreID(ctx context.Context, storeID int64, page, limit int) ([]db.ListProductsByStoreIDRow, int64, error) {
	return m.listProductsByStoreID(ctx, storeID, page, limit)
}
func (m *mockProductRepository) GetProduct(ctx context.Context, id int64) (db.GetProductRow, error) {
	return m.getProduct(ctx, id)
}
func (m *mockProductRepository) CreateProductWithImages(ctx context.Context, storeID int64, categoryID *int64, title, description *string, price string, stock int32, images []string) (db.Product, error) {
	return m.createProductWithImages(ctx, storeID, categoryID, title, description, price, stock, images)
}
func (m *mockProductRepository) UpdateProduct(ctx context.Context, id int64, title, description *string, price *string, stock *int32) (db.Product, error) {
	return m.updateProduct(ctx, id, title, description, price, stock)
}
func (m *mockProductRepository) DeleteProduct(ctx context.Context, id int64) error {
	return m.deleteProduct(ctx, id)
}
func (m *mockProductRepository) ModerateProduct(ctx context.Context, id int64, status db.ProductStatus) error {
	return m.moderateProduct(ctx, id, status)
}
func (m *mockProductRepository) ModerateProductWithReason(ctx context.Context, id int64, status db.ProductStatus, reason pgtype.Text) error {
	return m.moderateProductWithReason(ctx, id, status, reason)
}
func (m *mockProductRepository) ListProductImages(ctx context.Context, productID int64) ([]db.ProductImage, error) {
	return m.listProductImages(ctx, productID)
}
func (m *mockProductRepository) ListProductImagesByProductIDs(ctx context.Context, productIDs []int64) ([]db.ProductImage, error) {
	return m.listProductImagesByProductIDs(ctx, productIDs)
}
func (m *mockProductRepository) GetProductStoreOwner(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error) {
	return m.getProductStoreOwner(ctx, productID)
}
func (m *mockProductRepository) GetImageByID(ctx context.Context, id int64) (db.ProductImage, error) {
	return m.getImageByID(ctx, id)
}
func (m *mockProductRepository) DeleteImageByID(ctx context.Context, id int64) error {
	return m.deleteImageByID(ctx, id)
}
func (m *mockProductRepository) CreateProductImage(ctx context.Context, productID int64, url string, position int32) (db.ProductImage, error) {
	return m.createProductImage(ctx, productID, url, position)
}
func (m *mockProductRepository) ListProductsByStatus(ctx context.Context, status db.ProductStatus, page, limit int) ([]db.ListProductsByStatusRow, int64, error) {
	return m.listProductsByStatus(ctx, status, page, limit)
}

type mockStoreResolver struct {
	getStoreByUserID func(ctx context.Context, userID int64) (db.Store, error)
}

func (m *mockStoreResolver) GetStoreByUserID(ctx context.Context, userID int64) (db.Store, error) {
	return m.getStoreByUserID(ctx, userID)
}

func newTestService(repo ProductRepository, storeRepo StoreResolver) ProductService {
	return &productService{repo: repo, storeRepo: storeRepo, cache: nil}
}

func testProductRow() db.GetProductRow {
	return db.GetProductRow{
		ID:           1,
		StoreID:      10,
		Title:        "Test Product",
		Stock:        100,
		Status:       db.ProductStatusActive,
		StoreName:    pgtype.Text{String: "Test Store", Valid: true},
		CategoryName: pgtype.Text{String: "Electronics", Valid: true},
		CategoryID:   pgtype.Int8{Int64: 1, Valid: true},
		CategorySlug: pgtype.Text{String: "electronics", Valid: true},
	}
}

func testStore() db.Store {
	return db.Store{ID: 10, UserID: 1, Name: "Test Store"}
}

// --- ListCategories ---

func TestListCategories(t *testing.T) {
	ctx := context.Background()

	t.Run("empty categories", func(t *testing.T) {
		repo := &mockProductRepository{
			listCategories: func(ctx context.Context) ([]db.ListCategoriesRow, error) {
				return []db.ListCategoriesRow{}, nil
			},
		}
		svc := newTestService(repo, nil)
		resp, err := svc.ListCategories(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp) != 0 {
			t.Fatalf("expected empty categories, got %d", len(resp))
		}
	})

	t.Run("with categories", func(t *testing.T) {
		rows := []db.ListCategoriesRow{
			{ID: 1, Name: "Electronics", Slug: "electronics"},
			{ID: 2, Name: "Phones", Slug: "phones", ParentID: pgtype.Int8{Int64: 1, Valid: true}},
		}
		repo := &mockProductRepository{
			listCategories: func(ctx context.Context) ([]db.ListCategoriesRow, error) {
				return rows, nil
			},
		}
		svc := newTestService(repo, nil)
		resp, err := svc.ListCategories(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp) != 1 {
			t.Fatalf("expected 1 root category, got %d", len(resp))
		}
		if resp[0].Name != "Electronics" {
			t.Fatalf("expected Electronics, got %s", resp[0].Name)
		}
		if len(resp[0].Children) != 1 {
			t.Fatalf("expected 1 child, got %d", len(resp[0].Children))
		}
		if resp[0].Children[0].Name != "Phones" {
			t.Fatalf("expected child Phones, got %s", resp[0].Children[0].Name)
		}
	})

	t.Run("db error", func(t *testing.T) {
		repo := &mockProductRepository{
			listCategories: func(ctx context.Context) ([]db.ListCategoriesRow, error) {
				return nil, errors.New("db error")
			},
		}
		svc := newTestService(repo, nil)
		_, err := svc.ListCategories(ctx)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// --- ListProducts ---

func TestListProducts(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid sort", func(t *testing.T) {
		svc := newTestService(&mockProductRepository{}, nil)
		_, _, err := svc.ListProducts(ctx, nil, nil, nil, nil, nil, "invalid_sort", 1, 20)
		if err != ErrInvalidSort {
			t.Fatalf("expected ErrInvalidSort, got %v", err)
		}
	})

	t.Run("default sort", func(t *testing.T) {
		repo := &mockProductRepository{
			listProducts: func(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]db.ListProductsRow, int64, error) {
				if sort != "created_desc" {
					t.Fatalf("expected default sort created_desc, got %s", sort)
				}
				return []db.ListProductsRow{}, 0, nil
			},
		}
		svc := newTestService(repo, nil)
		_, _, err := svc.ListProducts(ctx, nil, nil, nil, nil, nil, "", 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("valid sort price_asc", func(t *testing.T) {
		repo := &mockProductRepository{
			listProducts: func(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]db.ListProductsRow, int64, error) {
				if sort != "price_asc" {
					t.Fatalf("expected price_asc, got %s", sort)
				}
				return []db.ListProductsRow{}, 0, nil
			},
		}
		svc := newTestService(repo, nil)
		_, _, err := svc.ListProducts(ctx, nil, nil, nil, nil, nil, "price_asc", 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("valid sort price_desc", func(t *testing.T) {
		repo := &mockProductRepository{
			listProducts: func(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]db.ListProductsRow, int64, error) {
				if sort != "price_desc" {
					t.Fatalf("expected price_desc, got %s", sort)
				}
				return []db.ListProductsRow{}, 0, nil
			},
		}
		svc := newTestService(repo, nil)
		_, _, err := svc.ListProducts(ctx, nil, nil, nil, nil, nil, "price_desc", 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty results", func(t *testing.T) {
		repo := &mockProductRepository{
			listProducts: func(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]db.ListProductsRow, int64, error) {
				return []db.ListProductsRow{}, 0, nil
			},
		}
		svc := newTestService(repo, nil)
		items, _, err := svc.ListProducts(ctx, nil, nil, nil, nil, nil, "created_desc", 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("expected empty items, got %d", len(items))
		}
	})

	t.Run("db error on list", func(t *testing.T) {
		repo := &mockProductRepository{
			listProducts: func(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]db.ListProductsRow, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}
		svc := newTestService(repo, nil)
		_, _, err := svc.ListProducts(ctx, nil, nil, nil, nil, nil, "created_desc", 1, 20)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success with items", func(t *testing.T) {
		row := db.ListProductsRow{
			ID:        1,
			Title:     "Test Product",
			Stock:     100,
			StoreID:   10,
			StoreName: pgtype.Text{String: "Test Store", Valid: true},
		}
		repo := &mockProductRepository{
			listProducts: func(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]db.ListProductsRow, int64, error) {
				return []db.ListProductsRow{row}, 1, nil
			},
			listProductImagesByProductIDs: func(ctx context.Context, productIDs []int64) ([]db.ProductImage, error) {
				return []db.ProductImage{}, nil
			},
		}
		svc := newTestService(repo, nil)
		items, total, err := svc.ListProducts(ctx, nil, nil, nil, nil, nil, "created_desc", 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if total != 1 {
			t.Fatalf("expected total 1, got %d", total)
		}
	})
}

// --- ListProductsByStore ---

func TestListProductsByStore(t *testing.T) {
	ctx := context.Background()

	t.Run("empty results", func(t *testing.T) {
		repo := &mockProductRepository{
			listProductsByStoreID: func(ctx context.Context, storeID int64, page, limit int) ([]db.ListProductsByStoreIDRow, int64, error) {
				return []db.ListProductsByStoreIDRow{}, 0, nil
			},
		}
		svc := newTestService(repo, nil)
		items, _, err := svc.ListProductsByStore(ctx, 10, 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("expected empty items, got %d", len(items))
		}
	})

	t.Run("db error", func(t *testing.T) {
		repo := &mockProductRepository{
			listProductsByStoreID: func(ctx context.Context, storeID int64, page, limit int) ([]db.ListProductsByStoreIDRow, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}
		svc := newTestService(repo, nil)
		_, _, err := svc.ListProductsByStore(ctx, 10, 1, 20)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		row := db.ListProductsByStoreIDRow{
			ID:        1,
			Title:     "Test Product",
			Stock:     100,
			StoreID:   10,
			StoreName: pgtype.Text{String: "Test Store", Valid: true},
		}
		repo := &mockProductRepository{
			listProductsByStoreID: func(ctx context.Context, storeID int64, page, limit int) ([]db.ListProductsByStoreIDRow, int64, error) {
				return []db.ListProductsByStoreIDRow{row}, 1, nil
			},
			listProductImagesByProductIDs: func(ctx context.Context, productIDs []int64) ([]db.ProductImage, error) {
				return []db.ProductImage{}, nil
			},
		}
		svc := newTestService(repo, nil)
		items, total, err := svc.ListProductsByStore(ctx, 10, 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if total != 1 {
			t.Fatalf("expected total 1, got %d", total)
		}
	})
}

// --- GetProduct ---

func TestGetProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("not found", func(t *testing.T) {
		repo := &mockProductRepository{
			getProduct: func(ctx context.Context, id int64) (db.GetProductRow, error) {
				return db.GetProductRow{}, pgx.ErrNoRows
			},
		}
		svc := newTestService(repo, nil)
		_, err := svc.GetProduct(ctx, 999)
		if err != ErrProductNotFound {
			t.Fatalf("expected ErrProductNotFound, got %v", err)
		}
	})

	t.Run("db error", func(t *testing.T) {
		repo := &mockProductRepository{
			getProduct: func(ctx context.Context, id int64) (db.GetProductRow, error) {
				return db.GetProductRow{}, errors.New("db error")
			},
		}
		svc := newTestService(repo, nil)
		_, err := svc.GetProduct(ctx, 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("images error", func(t *testing.T) {
		row := testProductRow()
		repo := &mockProductRepository{
			getProduct: func(ctx context.Context, id int64) (db.GetProductRow, error) {
				return row, nil
			},
			listProductImages: func(ctx context.Context, productID int64) ([]db.ProductImage, error) {
				return nil, errors.New("images db error")
			},
		}
		svc := newTestService(repo, nil)
		_, err := svc.GetProduct(ctx, 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		row := testProductRow()
		repo := &mockProductRepository{
			getProduct: func(ctx context.Context, id int64) (db.GetProductRow, error) {
				return row, nil
			},
			listProductImages: func(ctx context.Context, productID int64) ([]db.ProductImage, error) {
				return []db.ProductImage{}, nil
			},
		}
		svc := newTestService(repo, nil)
		resp, err := svc.GetProduct(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ID != 1 {
			t.Fatalf("expected ID 1, got %d", resp.ID)
		}
		if resp.Title != "Test Product" {
			t.Fatalf("expected title Test Product, got %s", resp.Title)
		}
		if resp.Store == nil || resp.Store.Name != "Test Store" {
			t.Fatalf("expected store info")
		}
	})
}

// --- CreateProduct ---

func TestCreateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("empty title", func(t *testing.T) {
		svc := newTestService(&mockProductRepository{}, nil)
		_, err := svc.CreateProduct(ctx, 1, dtos.CreateProductRequest{Price: "100", Stock: 10})
		if err != ErrTitleRequired {
			t.Fatalf("expected ErrTitleRequired, got %v", err)
		}
	})

	t.Run("title too long", func(t *testing.T) {
		title := ""
		for i := 0; i < 301; i++ {
			title += "a"
		}
		svc := newTestService(&mockProductRepository{}, nil)
		_, err := svc.CreateProduct(ctx, 1, dtos.CreateProductRequest{Title: title, Price: "100", Stock: 10})
		if err != ErrTitleTooLong {
			t.Fatalf("expected ErrTitleTooLong, got %v", err)
		}
	})

	t.Run("empty price", func(t *testing.T) {
		svc := newTestService(&mockProductRepository{}, nil)
		_, err := svc.CreateProduct(ctx, 1, dtos.CreateProductRequest{Title: "Test", Stock: 10})
		if err != ErrInvalidPrice {
			t.Fatalf("expected ErrInvalidPrice, got %v", err)
		}
	})

	t.Run("invalid price", func(t *testing.T) {
		svc := newTestService(&mockProductRepository{}, nil)
		_, err := svc.CreateProduct(ctx, 1, dtos.CreateProductRequest{Title: "Test", Price: "abc", Stock: 10})
		if err != ErrInvalidPrice {
			t.Fatalf("expected ErrInvalidPrice, got %v", err)
		}
	})

	t.Run("negative price", func(t *testing.T) {
		svc := newTestService(&mockProductRepository{}, nil)
		_, err := svc.CreateProduct(ctx, 1, dtos.CreateProductRequest{Title: "Test", Price: "-10", Stock: 10})
		if err != ErrInvalidPrice {
			t.Fatalf("expected ErrInvalidPrice, got %v", err)
		}
	})

	t.Run("negative stock", func(t *testing.T) {
		svc := newTestService(&mockProductRepository{}, nil)
		_, err := svc.CreateProduct(ctx, 1, dtos.CreateProductRequest{Title: "Test", Price: "100", Stock: -1})
		if err != ErrInvalidStock {
			t.Fatalf("expected ErrInvalidStock, got %v", err)
		}
	})

	t.Run("store not found", func(t *testing.T) {
		storeRepo := &mockStoreResolver{
			getStoreByUserID: func(ctx context.Context, userID int64) (db.Store, error) {
				return db.Store{}, pgx.ErrNoRows
			},
		}
		svc := newTestService(&mockProductRepository{}, storeRepo)
		_, err := svc.CreateProduct(ctx, 1, dtos.CreateProductRequest{Title: "Test", Price: "100", Stock: 10})
		if err != ErrStoreNotFound {
			t.Fatalf("expected ErrStoreNotFound, got %v", err)
		}
	})

	t.Run("create fails", func(t *testing.T) {
		store := testStore()
		storeRepo := &mockStoreResolver{
			getStoreByUserID: func(ctx context.Context, userID int64) (db.Store, error) {
				return store, nil
			},
		}
		repo := &mockProductRepository{
			createProductWithImages: func(ctx context.Context, storeID int64, categoryID *int64, title, description *string, price string, stock int32, images []string) (db.Product, error) {
				return db.Product{}, errors.New("create error")
			},
		}
		svc := newTestService(repo, storeRepo)
		_, err := svc.CreateProduct(ctx, 1, dtos.CreateProductRequest{Title: "Test", Price: "100", Stock: 10})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		store := testStore()
		product := db.Product{ID: 1, StoreID: 10, Title: "Test", Stock: 10}
		productRow := testProductRow()
		productRow.ID = 1
		storeRepo := &mockStoreResolver{
			getStoreByUserID: func(ctx context.Context, userID int64) (db.Store, error) {
				return store, nil
			},
		}
		repo := &mockProductRepository{
			createProductWithImages: func(ctx context.Context, storeID int64, categoryID *int64, title, description *string, price string, stock int32, images []string) (db.Product, error) {
				return product, nil
			},
			getProduct: func(ctx context.Context, id int64) (db.GetProductRow, error) {
				return productRow, nil
			},
			listProductImages: func(ctx context.Context, productID int64) ([]db.ProductImage, error) {
				return []db.ProductImage{}, nil
			},
		}
		svc := newTestService(repo, storeRepo)
		resp, err := svc.CreateProduct(ctx, 1, dtos.CreateProductRequest{Title: "Test", Price: "100", Stock: 10})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ID != 1 {
			t.Fatalf("expected ID 1, got %d", resp.ID)
		}
	})
}

// --- UpdateProduct ---

func TestUpdateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("forbidden - wrong owner", func(t *testing.T) {
		storeRepo := &mockStoreResolver{}
		repo := &mockProductRepository{
			getProductStoreOwner: func(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error) {
				return db.GetProductStoreOwnerRow{StoreOwnerID: 999}, nil
			},
		}
		svc := newTestService(repo, storeRepo)
		_, err := svc.UpdateProduct(ctx, 1, 1, dtos.UpdateProductRequest{Title: ptrStr("New Title")})
		if err != ErrForbidden {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("product not found on ownership check", func(t *testing.T) {
		storeRepo := &mockStoreResolver{}
		repo := &mockProductRepository{
			getProductStoreOwner: func(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error) {
				return db.GetProductStoreOwnerRow{}, pgx.ErrNoRows
			},
		}
		svc := newTestService(repo, storeRepo)
		_, err := svc.UpdateProduct(ctx, 1, 1, dtos.UpdateProductRequest{Title: ptrStr("New Title")})
		if err != ErrProductNotFound {
			t.Fatalf("expected ErrProductNotFound, got %v", err)
		}
	})

	t.Run("empty title on update", func(t *testing.T) {
		storeRepo := &mockStoreResolver{}
		repo := &mockProductRepository{
			getProductStoreOwner: func(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error) {
				return db.GetProductStoreOwnerRow{StoreOwnerID: 1}, nil
			},
		}
		svc := newTestService(repo, storeRepo)
		_, err := svc.UpdateProduct(ctx, 1, 1, dtos.UpdateProductRequest{Title: ptrStr("")})
		if err != ErrTitleRequired {
			t.Fatalf("expected ErrTitleRequired, got %v", err)
		}
	})

	t.Run("title too long on update", func(t *testing.T) {
		storeRepo := &mockStoreResolver{}
		repo := &mockProductRepository{
			getProductStoreOwner: func(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error) {
				return db.GetProductStoreOwnerRow{StoreOwnerID: 1}, nil
			},
		}
		title := ""
		for i := 0; i < 301; i++ {
			title += "a"
		}
		svc := newTestService(repo, storeRepo)
		_, err := svc.UpdateProduct(ctx, 1, 1, dtos.UpdateProductRequest{Title: &title})
		if err != ErrTitleTooLong {
			t.Fatalf("expected ErrTitleTooLong, got %v", err)
		}
	})

	t.Run("invalid price on update", func(t *testing.T) {
		storeRepo := &mockStoreResolver{}
		repo := &mockProductRepository{
			getProductStoreOwner: func(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error) {
				return db.GetProductStoreOwnerRow{StoreOwnerID: 1}, nil
			},
		}
		svc := newTestService(repo, storeRepo)
		_, err := svc.UpdateProduct(ctx, 1, 1, dtos.UpdateProductRequest{Price: ptrStr("abc")})
		if err != ErrInvalidPrice {
			t.Fatalf("expected ErrInvalidPrice, got %v", err)
		}
	})

	t.Run("negative stock on update", func(t *testing.T) {
		storeRepo := &mockStoreResolver{}
		repo := &mockProductRepository{
			getProductStoreOwner: func(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error) {
				return db.GetProductStoreOwnerRow{StoreOwnerID: 1}, nil
			},
		}
		neg := -1
		svc := newTestService(repo, storeRepo)
		_, err := svc.UpdateProduct(ctx, 1, 1, dtos.UpdateProductRequest{Stock: &neg})
		if err != ErrInvalidStock {
			t.Fatalf("expected ErrInvalidStock, got %v", err)
		}
	})

	t.Run("update fails", func(t *testing.T) {
		storeRepo := &mockStoreResolver{}
		repo := &mockProductRepository{
			getProductStoreOwner: func(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error) {
				return db.GetProductStoreOwnerRow{StoreOwnerID: 1}, nil
			},
			updateProduct: func(ctx context.Context, id int64, title, description *string, price *string, stock *int32) (db.Product, error) {
				return db.Product{}, errors.New("update error")
			},
		}
		svc := newTestService(repo, storeRepo)
		_, err := svc.UpdateProduct(ctx, 1, 1, dtos.UpdateProductRequest{Title: ptrStr("New Title")})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		storeRepo := &mockStoreResolver{}
		product := db.Product{ID: 1, StoreID: 10, Title: "Updated", Stock: 50}
		productRow := testProductRow()
		productRow.ID = 1
		productRow.Title = "Updated"
		repo := &mockProductRepository{
			getProductStoreOwner: func(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error) {
				return db.GetProductStoreOwnerRow{StoreOwnerID: 1}, nil
			},
			updateProduct: func(ctx context.Context, id int64, title, description *string, price *string, stock *int32) (db.Product, error) {
				return product, nil
			},
			getProduct: func(ctx context.Context, id int64) (db.GetProductRow, error) {
				return productRow, nil
			},
			listProductImages: func(ctx context.Context, productID int64) ([]db.ProductImage, error) {
				return []db.ProductImage{}, nil
			},
		}
		svc := newTestService(repo, storeRepo)
		resp, err := svc.UpdateProduct(ctx, 1, 1, dtos.UpdateProductRequest{Title: ptrStr("Updated")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Title != "Updated" {
			t.Fatalf("expected title Updated, got %s", resp.Title)
		}
	})
}

// --- DeleteProduct ---

func TestDeleteProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("forbidden", func(t *testing.T) {
		storeRepo := &mockStoreResolver{}
		repo := &mockProductRepository{
			getProductStoreOwner: func(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error) {
				return db.GetProductStoreOwnerRow{StoreOwnerID: 999}, nil
			},
		}
		svc := newTestService(repo, storeRepo)
		err := svc.DeleteProduct(ctx, 1, 1)
		if err != ErrForbidden {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		storeRepo := &mockStoreResolver{}
		repo := &mockProductRepository{
			getProductStoreOwner: func(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error) {
				return db.GetProductStoreOwnerRow{StoreOwnerID: 1}, nil
			},
			deleteProduct: func(ctx context.Context, id int64) error {
				return nil
			},
		}
		svc := newTestService(repo, storeRepo)
		err := svc.DeleteProduct(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("delete fails", func(t *testing.T) {
		storeRepo := &mockStoreResolver{}
		repo := &mockProductRepository{
			getProductStoreOwner: func(ctx context.Context, productID int64) (db.GetProductStoreOwnerRow, error) {
				return db.GetProductStoreOwnerRow{StoreOwnerID: 1}, nil
			},
			deleteProduct: func(ctx context.Context, id int64) error {
				return errors.New("delete error")
			},
		}
		svc := newTestService(repo, storeRepo)
		err := svc.DeleteProduct(ctx, 1, 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// --- ModerateProduct ---

func TestModerateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid status", func(t *testing.T) {
		svc := newTestService(&mockProductRepository{}, nil)
		_, err := svc.ModerateProduct(ctx, 1, dtos.ModerateProductRequest{Status: "invalid"})
		if err != ErrInvalidStatus {
			t.Fatalf("expected ErrInvalidStatus, got %v", err)
		}
	})

	t.Run("active status", func(t *testing.T) {
		productRow := testProductRow()
		repo := &mockProductRepository{
			moderateProductWithReason: func(ctx context.Context, id int64, status db.ProductStatus, reason pgtype.Text) error {
				return nil
			},
			getProduct: func(ctx context.Context, id int64) (db.GetProductRow, error) {
				return productRow, nil
			},
			listProductImages: func(ctx context.Context, productID int64) ([]db.ProductImage, error) {
				return []db.ProductImage{}, nil
			},
		}
		svc := newTestService(repo, nil)
		resp, err := svc.ModerateProduct(ctx, 1, dtos.ModerateProductRequest{Status: db.ProductStatusActive})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Status != string(db.ProductStatusActive) {
			t.Fatalf("expected status active, got %s", resp.Status)
		}
	})

	t.Run("rejected status", func(t *testing.T) {
		productRow := testProductRow()
		productRow.Status = db.ProductStatusRejected
		repo := &mockProductRepository{
			moderateProductWithReason: func(ctx context.Context, id int64, status db.ProductStatus, reason pgtype.Text) error {
				return nil
			},
			getProduct: func(ctx context.Context, id int64) (db.GetProductRow, error) {
				return productRow, nil
			},
			listProductImages: func(ctx context.Context, productID int64) ([]db.ProductImage, error) {
				return []db.ProductImage{}, nil
			},
		}
		svc := newTestService(repo, nil)
		_, err := svc.ModerateProduct(ctx, 1, dtos.ModerateProductRequest{Status: db.ProductStatusRejected})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("archived status", func(t *testing.T) {
		productRow := testProductRow()
		productRow.Status = db.ProductStatusArchived
		repo := &mockProductRepository{
			moderateProductWithReason: func(ctx context.Context, id int64, status db.ProductStatus, reason pgtype.Text) error {
				return nil
			},
			getProduct: func(ctx context.Context, id int64) (db.GetProductRow, error) {
				return productRow, nil
			},
			listProductImages: func(ctx context.Context, productID int64) ([]db.ProductImage, error) {
				return []db.ProductImage{}, nil
			},
		}
		svc := newTestService(repo, nil)
		_, err := svc.ModerateProduct(ctx, 1, dtos.ModerateProductRequest{Status: db.ProductStatusArchived})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("moderate fails", func(t *testing.T) {
		repo := &mockProductRepository{
			moderateProductWithReason: func(ctx context.Context, id int64, status db.ProductStatus, reason pgtype.Text) error {
				return errors.New("moderate error")
			},
		}
		svc := newTestService(repo, nil)
		_, err := svc.ModerateProduct(ctx, 1, dtos.ModerateProductRequest{Status: db.ProductStatusActive})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func ptrStr(s string) *string {
	return &s
}
