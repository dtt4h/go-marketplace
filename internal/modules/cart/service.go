package cart

import (
	"context"
	"errors"
	"fmt"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
)

var (
	ErrCartEmpty         = errors.New("cart is empty")
	ErrInvalidQuantity   = errors.New("quantity must be greater than zero")
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type CartRepository interface {
	GetCartItemsByUserID(ctx context.Context, userID int64) ([]db.GetCartItemsByUserIDRow, error)
	GetCartItemByUserIDAndProductID(ctx context.Context, userID, productID int64) (db.CartItem, error)
	GetCartByID(ctx context.Context, id, userID int64) (db.CartItem, error)
	UpsertCartItem(ctx context.Context, arg db.UpsertCartItemParams) (db.CartItem, error)
	UpdateCartItemQuantity(ctx context.Context, arg db.UpdateCartItemQuantityParams) (db.CartItem, error)
	DeleteCartItem(ctx context.Context, arg db.DeleteCartItemParams) error
	DeleteCartItemsByUserID(ctx context.Context, userID int64) error
	GetProduct(ctx context.Context, productID int64) (db.GetProductRow, error)
}

type cartRepository struct {
	queries *db.Queries
}

func NewCartRepository(queries *db.Queries) CartRepository {
	return &cartRepository{queries: queries}
}

func (r *cartRepository) GetCartItemsByUserID(ctx context.Context, userID int64) ([]db.GetCartItemsByUserIDRow, error) {
	return r.queries.GetCartItemsByUserID(ctx, userID)
}

func (r *cartRepository) GetCartItemByUserIDAndProductID(ctx context.Context, userID, productID int64) (db.CartItem, error) {
	return r.queries.GetCartItemByUserIDAndProductID(ctx, db.GetCartItemByUserIDAndProductIDParams{
		UserID:    userID,
		ProductID: productID,
	})
}

func (r *cartRepository) GetCartByID(ctx context.Context, id, userID int64) (db.CartItem, error) {
	return r.queries.GetCartByID(ctx, db.GetCartByIDParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *cartRepository) UpsertCartItem(ctx context.Context, arg db.UpsertCartItemParams) (db.CartItem, error) {
	return r.queries.UpsertCartItem(ctx, arg)
}

func (r *cartRepository) UpdateCartItemQuantity(ctx context.Context, arg db.UpdateCartItemQuantityParams) (db.CartItem, error) {
	return r.queries.UpdateCartItemQuantity(ctx, arg)
}

func (r *cartRepository) DeleteCartItem(ctx context.Context, arg db.DeleteCartItemParams) error {
	return r.queries.DeleteCartItem(ctx, arg)
}

func (r *cartRepository) DeleteCartItemsByUserID(ctx context.Context, userID int64) error {
	return r.queries.DeleteCartItemsByUserID(ctx, userID)
}

func (r *cartRepository) GetProduct(ctx context.Context, productID int64) (db.GetProductRow, error) {
	return r.queries.GetProduct(ctx, productID)
}

type CartService interface {
	GetCart(ctx context.Context, userID int64) (dtos.CartResponse, error)
	AddItem(ctx context.Context, userID int64, req dtos.AddCartItemRequest) (dtos.CartResponse, error)
	UpdateItem(ctx context.Context, userID int64, itemID int64, req dtos.UpdateCartItemRequest) (dtos.CartResponse, error)
	RemoveItem(ctx context.Context, userID int64, itemID int64) (dtos.CartResponse, error)
	Clear(ctx context.Context, userID int64) (dtos.CartResponse, error)
}

type cartService struct {
	repo CartRepository
}

func NewCartService(repo CartRepository) CartService {
	return &cartService{repo: repo}
}

func (s *cartService) GetCart(ctx context.Context, userID int64) (dtos.CartResponse, error) {
	rows, err := s.repo.GetCartItemsByUserID(ctx, userID)
	if err != nil {
		return dtos.CartResponse{}, fmt.Errorf("get cart items: %w", err)
	}

	items := make([]dtos.CartItemResponse, 0, len(rows))
	for _, row := range rows {
		item := dtos.CartItemResponse{
			ID:        row.ID,
			ProductID: row.ProductID,
			Quantity:  row.Quantity,
			Title:     row.Title,
			Price:     dtos.NumericToStr(row.Price),
			Stock:     row.Stock,
			Store: &dtos.CartStoreInfo{
				ID:   row.StoreID,
				Name: row.StoreName,
			},
		}
		if row.PreviewImageUrl != "" {
			item.PreviewImageURL = row.PreviewImageUrl
		}
		items = append(items, item)
	}

	return dtos.CartResponse{UserID: userID, Items: items}, nil
}

func (s *cartService) AddItem(ctx context.Context, userID int64, req dtos.AddCartItemRequest) (dtos.CartResponse, error) {
	if req.Quantity <= 0 {
		return dtos.CartResponse{}, ErrInvalidQuantity
	}

	product, err := s.repo.GetProduct(ctx, req.ProductID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.CartResponse{}, ErrProductNotFound
		}
		return dtos.CartResponse{}, fmt.Errorf("get product: %w", err)
	}

	if product.Status != db.ProductStatusActive {
		return dtos.CartResponse{}, ErrProductNotFound
	}

	if product.Stock < int32(req.Quantity) {
		return dtos.CartResponse{}, ErrInsufficientStock
	}

	_, err = s.repo.UpsertCartItem(ctx, db.UpsertCartItemParams{
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	})
	if err != nil {
		return dtos.CartResponse{}, fmt.Errorf("upsert cart item: %w", err)
	}

	return s.GetCart(ctx, userID)
}

func (s *cartService) UpdateItem(ctx context.Context, userID int64, itemID int64, req dtos.UpdateCartItemRequest) (dtos.CartResponse, error) {
	if req.Quantity <= 0 {
		return dtos.CartResponse{}, ErrInvalidQuantity
	}

	existing, err := s.repo.GetCartByID(ctx, itemID, userID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.CartResponse{}, ErrProductNotFound
		}
		return dtos.CartResponse{}, fmt.Errorf("get cart item: %w", err)
	}

	product, err := s.repo.GetProduct(ctx, existing.ProductID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.CartResponse{}, ErrProductNotFound
		}
		return dtos.CartResponse{}, fmt.Errorf("get product: %w", err)
	}

	if product.Stock < req.Quantity {
		return dtos.CartResponse{}, ErrInsufficientStock
	}

	_, err = s.repo.UpdateCartItemQuantity(ctx, db.UpdateCartItemQuantityParams{
		ID:       itemID,
		UserID:   userID,
		Quantity: req.Quantity,
	})
	if err != nil {
		return dtos.CartResponse{}, fmt.Errorf("update cart item: %w", err)
	}

	return s.GetCart(ctx, userID)
}

func (s *cartService) RemoveItem(ctx context.Context, userID int64, itemID int64) (dtos.CartResponse, error) {
	err := s.repo.DeleteCartItem(ctx, db.DeleteCartItemParams{
		ID:     itemID,
		UserID: userID,
	})
	if err != nil {
		return dtos.CartResponse{}, fmt.Errorf("delete cart item: %w", err)
	}

	return s.GetCart(ctx, userID)
}

func (s *cartService) Clear(ctx context.Context, userID int64) (dtos.CartResponse, error) {
	err := s.repo.DeleteCartItemsByUserID(ctx, userID)
	if err != nil {
		return dtos.CartResponse{}, fmt.Errorf("clear cart: %w", err)
	}

	return dtos.CartResponse{UserID: userID, Items: []dtos.CartItemResponse{}}, nil
}
