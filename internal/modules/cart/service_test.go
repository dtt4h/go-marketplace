package cart

import (
	"context"
	"errors"
	"testing"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// --- Mock ---

type mockCartRepository struct {
	getCartItemsByUserID            func(ctx context.Context, userID int64) ([]db.GetCartItemsByUserIDRow, error)
	getCartItemByUserIDAndProductID func(ctx context.Context, userID, productID int64) (db.CartItem, error)
	getCartByID                     func(ctx context.Context, id, userID int64) (db.CartItem, error)
	upsertCartItem                  func(ctx context.Context, arg db.UpsertCartItemParams) (db.CartItem, error)
	updateCartItemQuantity          func(ctx context.Context, arg db.UpdateCartItemQuantityParams) (db.CartItem, error)
	deleteCartItem                  func(ctx context.Context, arg db.DeleteCartItemParams) error
	deleteCartItemsByUserID         func(ctx context.Context, userID int64) error
	getProduct                      func(ctx context.Context, productID int64) (db.GetProductRow, error)
}

func (m *mockCartRepository) GetCartItemsByUserID(ctx context.Context, userID int64) ([]db.GetCartItemsByUserIDRow, error) {
	return m.getCartItemsByUserID(ctx, userID)
}
func (m *mockCartRepository) GetCartItemByUserIDAndProductID(ctx context.Context, userID, productID int64) (db.CartItem, error) {
	return m.getCartItemByUserIDAndProductID(ctx, userID, productID)
}
func (m *mockCartRepository) GetCartByID(ctx context.Context, id, userID int64) (db.CartItem, error) {
	return m.getCartByID(ctx, id, userID)
}
func (m *mockCartRepository) UpsertCartItem(ctx context.Context, arg db.UpsertCartItemParams) (db.CartItem, error) {
	return m.upsertCartItem(ctx, arg)
}
func (m *mockCartRepository) UpdateCartItemQuantity(ctx context.Context, arg db.UpdateCartItemQuantityParams) (db.CartItem, error) {
	return m.updateCartItemQuantity(ctx, arg)
}
func (m *mockCartRepository) DeleteCartItem(ctx context.Context, arg db.DeleteCartItemParams) error {
	return m.deleteCartItem(ctx, arg)
}
func (m *mockCartRepository) DeleteCartItemsByUserID(ctx context.Context, userID int64) error {
	return m.deleteCartItemsByUserID(ctx, userID)
}
func (m *mockCartRepository) GetProduct(ctx context.Context, productID int64) (db.GetProductRow, error) {
	return m.getProduct(ctx, productID)
}

func newTestService(repo CartRepository) CartService {
	return &cartService{repo: repo}
}

func testProduct() db.GetProductRow {
	return db.GetProductRow{
		ID:      1,
		StoreID: 10,
		Title:   "Test Product",
		Stock:   100,
		Status:  db.ProductStatusActive,
	}
}

func testCartItem() db.GetCartItemsByUserIDRow {
	var price pgtype.Numeric
	price.Scan("100.00")
	return db.GetCartItemsByUserIDRow{
		ID:              1,
		ProductID:       1,
		Quantity:        2,
		Title:           "Test Product",
		Price:           price,
		Stock:           100,
		StoreID:         10,
		StoreName:       "Test Store",
		PreviewImageUrl: "http://img.jpg",
	}
}

// --- GetCart ---

func TestGetCart(t *testing.T) {
	ctx := context.Background()

	t.Run("empty cart", func(t *testing.T) {
		repo := &mockCartRepository{
			getCartItemsByUserID: func(ctx context.Context, userID int64) ([]db.GetCartItemsByUserIDRow, error) {
				return []db.GetCartItemsByUserIDRow{}, nil
			},
		}
		svc := newTestService(repo)
		resp, err := svc.GetCart(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Items) != 0 {
			t.Fatalf("expected empty items, got %d", len(resp.Items))
		}
		if resp.UserID != 1 {
			t.Fatalf("expected userID 1, got %d", resp.UserID)
		}
	})

	t.Run("cart with items", func(t *testing.T) {
		item := testCartItem()
		repo := &mockCartRepository{
			getCartItemsByUserID: func(ctx context.Context, userID int64) ([]db.GetCartItemsByUserIDRow, error) {
				return []db.GetCartItemsByUserIDRow{item}, nil
			},
		}
		svc := newTestService(repo)
		resp, err := svc.GetCart(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}
		if resp.Items[0].ID != 1 {
			t.Fatalf("expected item ID 1, got %d", resp.Items[0].ID)
		}
		if resp.Items[0].Quantity != 2 {
			t.Fatalf("expected quantity 2, got %d", resp.Items[0].Quantity)
		}
		if resp.Items[0].PreviewImageURL != "http://img.jpg" {
			t.Fatalf("expected preview image URL, got %s", resp.Items[0].PreviewImageURL)
		}
		if resp.Items[0].Store == nil || resp.Items[0].Store.Name != "Test Store" {
			t.Fatalf("expected store info")
		}
	})

	t.Run("db error", func(t *testing.T) {
		repo := &mockCartRepository{
			getCartItemsByUserID: func(ctx context.Context, userID int64) ([]db.GetCartItemsByUserIDRow, error) {
				return nil, errors.New("db error")
			},
		}
		svc := newTestService(repo)
		_, err := svc.GetCart(ctx, 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// --- AddItem ---

func TestAddItem(t *testing.T) {
	ctx := context.Background()

	t.Run("rejects zero quantity", func(t *testing.T) {
		svc := newTestService(&mockCartRepository{})
		_, err := svc.AddItem(ctx, 1, dtos.AddCartItemRequest{ProductID: 1, Quantity: 0})
		if err != ErrInvalidQuantity {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})

	t.Run("rejects negative quantity", func(t *testing.T) {
		svc := newTestService(&mockCartRepository{})
		_, err := svc.AddItem(ctx, 1, dtos.AddCartItemRequest{ProductID: 1, Quantity: -1})
		if err != ErrInvalidQuantity {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})

	t.Run("product not found", func(t *testing.T) {
		repo := &mockCartRepository{
			getProduct: func(ctx context.Context, productID int64) (db.GetProductRow, error) {
				return db.GetProductRow{}, pgx.ErrNoRows
			},
		}
		svc := newTestService(repo)
		_, err := svc.AddItem(ctx, 1, dtos.AddCartItemRequest{ProductID: 999, Quantity: 1})
		if err != ErrProductNotFound {
			t.Fatalf("expected ErrProductNotFound, got %v", err)
		}
	})

	t.Run("product inactive", func(t *testing.T) {
		product := testProduct()
		product.Status = db.ProductStatusPending
		repo := &mockCartRepository{
			getProduct: func(ctx context.Context, productID int64) (db.GetProductRow, error) {
				return product, nil
			},
		}
		svc := newTestService(repo)
		_, err := svc.AddItem(ctx, 1, dtos.AddCartItemRequest{ProductID: 1, Quantity: 1})
		if err != ErrProductNotFound {
			t.Fatalf("expected ErrProductNotFound for inactive product, got %v", err)
		}
	})

	t.Run("insufficient stock", func(t *testing.T) {
		product := testProduct()
		product.Stock = 5
		repo := &mockCartRepository{
			getProduct: func(ctx context.Context, productID int64) (db.GetProductRow, error) {
				return product, nil
			},
		}
		svc := newTestService(repo)
		_, err := svc.AddItem(ctx, 1, dtos.AddCartItemRequest{ProductID: 1, Quantity: 10})
		if err != ErrInsufficientStock {
			t.Fatalf("expected ErrInsufficientStock, got %v", err)
		}
	})

	t.Run("db error on get product", func(t *testing.T) {
		repo := &mockCartRepository{
			getProduct: func(ctx context.Context, productID int64) (db.GetProductRow, error) {
				return db.GetProductRow{}, errors.New("db error")
			},
		}
		svc := newTestService(repo)
		_, err := svc.AddItem(ctx, 1, dtos.AddCartItemRequest{ProductID: 1, Quantity: 1})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("upsert fails", func(t *testing.T) {
		product := testProduct()
		repo := &mockCartRepository{
			getProduct: func(ctx context.Context, productID int64) (db.GetProductRow, error) {
				return product, nil
			},
			upsertCartItem: func(ctx context.Context, arg db.UpsertCartItemParams) (db.CartItem, error) {
				return db.CartItem{}, errors.New("upsert error")
			},
		}
		svc := newTestService(repo)
		_, err := svc.AddItem(ctx, 1, dtos.AddCartItemRequest{ProductID: 1, Quantity: 1})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		product := testProduct()
		item := testCartItem()
		repo := &mockCartRepository{
			getProduct: func(ctx context.Context, productID int64) (db.GetProductRow, error) {
				return product, nil
			},
			upsertCartItem: func(ctx context.Context, arg db.UpsertCartItemParams) (db.CartItem, error) {
				return db.CartItem{ID: arg.ProductID, Quantity: arg.Quantity}, nil
			},
			getCartItemsByUserID: func(ctx context.Context, userID int64) ([]db.GetCartItemsByUserIDRow, error) {
				return []db.GetCartItemsByUserIDRow{item}, nil
			},
		}
		svc := newTestService(repo)
		resp, err := svc.AddItem(ctx, 1, dtos.AddCartItemRequest{ProductID: 1, Quantity: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}
		if resp.Items[0].Quantity != 2 {
			t.Fatalf("expected quantity 2, got %d", resp.Items[0].Quantity)
		}
	})
}

// --- UpdateItem ---

func TestUpdateItem(t *testing.T) {
	ctx := context.Background()

	t.Run("rejects zero quantity", func(t *testing.T) {
		svc := newTestService(&mockCartRepository{})
		_, err := svc.UpdateItem(ctx, 1, 1, dtos.UpdateCartItemRequest{Quantity: 0})
		if err != ErrInvalidQuantity {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})

	t.Run("cart item not found", func(t *testing.T) {
		repo := &mockCartRepository{
			getCartByID: func(ctx context.Context, id, userID int64) (db.CartItem, error) {
				return db.CartItem{}, pgx.ErrNoRows
			},
		}
		svc := newTestService(repo)
		_, err := svc.UpdateItem(ctx, 1, 999, dtos.UpdateCartItemRequest{Quantity: 1})
		if err != ErrProductNotFound {
			t.Fatalf("expected ErrProductNotFound, got %v", err)
		}
	})

	t.Run("product not found", func(t *testing.T) {
		cartItem := db.CartItem{ID: 1, ProductID: 1}
		repo := &mockCartRepository{
			getCartByID: func(ctx context.Context, id, userID int64) (db.CartItem, error) {
				return cartItem, nil
			},
			getProduct: func(ctx context.Context, productID int64) (db.GetProductRow, error) {
				return db.GetProductRow{}, pgx.ErrNoRows
			},
		}
		svc := newTestService(repo)
		_, err := svc.UpdateItem(ctx, 1, 1, dtos.UpdateCartItemRequest{Quantity: 1})
		if err != ErrProductNotFound {
			t.Fatalf("expected ErrProductNotFound, got %v", err)
		}
	})

	t.Run("insufficient stock on update", func(t *testing.T) {
		cartItem := db.CartItem{ID: 1, ProductID: 1}
		product := testProduct()
		product.Stock = 1
		repo := &mockCartRepository{
			getCartByID: func(ctx context.Context, id, userID int64) (db.CartItem, error) {
				return cartItem, nil
			},
			getProduct: func(ctx context.Context, productID int64) (db.GetProductRow, error) {
				return product, nil
			},
		}
		svc := newTestService(repo)
		_, err := svc.UpdateItem(ctx, 1, 1, dtos.UpdateCartItemRequest{Quantity: 5})
		if err != ErrInsufficientStock {
			t.Fatalf("expected ErrInsufficientStock, got %v", err)
		}
	})

	t.Run("update fails", func(t *testing.T) {
		cartItem := db.CartItem{ID: 1, ProductID: 1}
		product := testProduct()
		repo := &mockCartRepository{
			getCartByID: func(ctx context.Context, id, userID int64) (db.CartItem, error) {
				return cartItem, nil
			},
			getProduct: func(ctx context.Context, productID int64) (db.GetProductRow, error) {
				return product, nil
			},
			updateCartItemQuantity: func(ctx context.Context, arg db.UpdateCartItemQuantityParams) (db.CartItem, error) {
				return db.CartItem{}, errors.New("update error")
			},
		}
		svc := newTestService(repo)
		_, err := svc.UpdateItem(ctx, 1, 1, dtos.UpdateCartItemRequest{Quantity: 2})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		cartItem := db.CartItem{ID: 1, ProductID: 1}
		product := testProduct()
		item := testCartItem()
		item.Quantity = 5
		repo := &mockCartRepository{
			getCartByID: func(ctx context.Context, id, userID int64) (db.CartItem, error) {
				return cartItem, nil
			},
			getProduct: func(ctx context.Context, productID int64) (db.GetProductRow, error) {
				return product, nil
			},
			updateCartItemQuantity: func(ctx context.Context, arg db.UpdateCartItemQuantityParams) (db.CartItem, error) {
				return db.CartItem{ID: arg.ID, Quantity: arg.Quantity}, nil
			},
			getCartItemsByUserID: func(ctx context.Context, userID int64) ([]db.GetCartItemsByUserIDRow, error) {
				return []db.GetCartItemsByUserIDRow{item}, nil
			},
		}
		svc := newTestService(repo)
		resp, err := svc.UpdateItem(ctx, 1, 1, dtos.UpdateCartItemRequest{Quantity: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Items[0].Quantity != 5 {
			t.Fatalf("expected quantity 5, got %d", resp.Items[0].Quantity)
		}
	})
}

// --- RemoveItem ---

func TestRemoveItem(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockCartRepository{
			deleteCartItem: func(ctx context.Context, arg db.DeleteCartItemParams) error {
				return nil
			},
			getCartItemsByUserID: func(ctx context.Context, userID int64) ([]db.GetCartItemsByUserIDRow, error) {
				return []db.GetCartItemsByUserIDRow{}, nil
			},
		}
		svc := newTestService(repo)
		resp, err := svc.RemoveItem(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Items) != 0 {
			t.Fatalf("expected empty cart after remove, got %d items", len(resp.Items))
		}
	})

	t.Run("delete fails", func(t *testing.T) {
		repo := &mockCartRepository{
			deleteCartItem: func(ctx context.Context, arg db.DeleteCartItemParams) error {
				return errors.New("delete error")
			},
		}
		svc := newTestService(repo)
		_, err := svc.RemoveItem(ctx, 1, 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// --- Clear ---

func TestClear(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockCartRepository{
			deleteCartItemsByUserID: func(ctx context.Context, userID int64) error {
				return nil
			},
		}
		svc := newTestService(repo)
		resp, err := svc.Clear(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Items) != 0 {
			t.Fatalf("expected empty cart, got %d items", len(resp.Items))
		}
	})

	t.Run("delete fails", func(t *testing.T) {
		repo := &mockCartRepository{
			deleteCartItemsByUserID: func(ctx context.Context, userID int64) error {
				return errors.New("delete error")
			},
		}
		svc := newTestService(repo)
		_, err := svc.Clear(ctx, 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
