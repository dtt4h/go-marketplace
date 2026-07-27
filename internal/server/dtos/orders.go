package dtos

import (
	"time"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
)

type CreateOrderRequest struct {
	Items   []OrderItemRequest `json:"items"`
	Address string             `json:"address"`
}

type OrderItemRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
}

type UpdateOrderStatusRequest struct {
	Status db.OrderStatus `json:"status"`
}

type OrderItemResponse struct {
	ID           int64  `json:"id"`
	ProductID    int64  `json:"product_id"`
	Quantity     int32  `json:"quantity"`
	Price        string `json:"price"`
	ProductTitle string `json:"product_title"`
}

type OrderResponse struct {
	ID        int64               `json:"id"`
	Status    string              `json:"status"`
	Total     string              `json:"total"`
	Address   string              `json:"address"`
	Items     []OrderItemResponse `json:"items"`
	CreatedAt time.Time           `json:"created_at"`
}

type OrderListItem struct {
	ID         int64     `json:"id"`
	Status     string    `json:"status"`
	Total      string    `json:"total"`
	Address    string    `json:"address"`
	ItemsCount int32     `json:"items_count"`
	CreatedAt  time.Time `json:"created_at"`
}

type BuyerInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type OrderSellerListItem struct {
	ID        int64               `json:"id"`
	Status    string              `json:"status"`
	Total     string              `json:"total"`
	Buyer     *BuyerInfo          `json:"buyer"`
	Items     []OrderItemResponse `json:"items"`
	CreatedAt time.Time           `json:"created_at"`
}

func ToOrderResponse(order db.Order, items []db.GetOrderItemsRow) OrderResponse {
	resp := OrderResponse{
		ID:        order.ID,
		Status:    string(order.Status),
		Total:     NumericToStr(order.Total),
		Address:   order.Address,
		CreatedAt: order.CreatedAt.Time,
	}

	if len(items) > 0 {
		resp.Items = make([]OrderItemResponse, 0, len(items))
		for _, item := range items {
			resp.Items = append(resp.Items, OrderItemResponse{
				ID:           item.ID,
				ProductID:    item.ProductID,
				Quantity:     item.Quantity,
				Price:        NumericToStr(item.Price),
				ProductTitle: item.ProductTitle,
			})
		}
	}

	return resp
}

func ToOrderListItem(order db.Order, itemsCount int32) OrderListItem {
	return OrderListItem{
		ID:         order.ID,
		Status:     string(order.Status),
		Total:      NumericToStr(order.Total),
		Address:    order.Address,
		ItemsCount: itemsCount,
		CreatedAt:  order.CreatedAt.Time,
	}
}

func ToOrderSellerListItem(order db.Order, buyerUsername string, items []db.GetOrderItemsRow) OrderSellerListItem {
	itemResponses := make([]OrderItemResponse, 0, len(items))
	for _, item := range items {
		itemResponses = append(itemResponses, OrderItemResponse{
			ID:           item.ID,
			ProductID:    item.ProductID,
			Quantity:     item.Quantity,
			Price:        NumericToStr(item.Price),
			ProductTitle: item.ProductTitle,
		})
	}

	return OrderSellerListItem{
		ID:        order.ID,
		Status:    string(order.Status),
		Total:     NumericToStr(order.Total),
		Buyer:     &BuyerInfo{ID: order.UserID, Username: buyerUsername},
		Items:     itemResponses,
		CreatedAt: order.CreatedAt.Time,
	}
}
