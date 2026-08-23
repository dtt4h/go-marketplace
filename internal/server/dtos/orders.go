package dtos

import (
	"time"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
)

type CustomerInfo struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type DeliveryInfo struct {
	Address string `json:"address"`
	Method  string `json:"method"`
	Cost    string `json:"cost"`
}

type CreateOrderRequest struct {
	// Авторизованный пользователь — user_id берётся из JWT, не из тела.
	// Гостевой заказ — user_id = null, заполняется Customer + Delivery.

	// Для авторизованных (опционально, берётся из токена):
	UserID *int64 `json:"-"` // set by handler from JWT

	// Для гостей:
	Customer *CustomerInfo `json:"customer,omitempty"`

	// Общие поля:
	Items    []OrderItemRequest `json:"items"`
	Delivery DeliveryInfo       `json:"delivery"`
}

type OrderItemRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
}

type UpdateOrderStatusRequest struct {
	Status db.OrderStatus `json:"status"`
}

type UpdateTrackingRequest struct {
	TrackingNumber string `json:"tracking_number"`
}

type OrderItemResponse struct {
	ID           int64  `json:"id"`
	ProductID    int64  `json:"product_id"`
	Quantity     int32  `json:"quantity"`
	Price        string `json:"price"`
	ProductTitle string `json:"product_title"`
}

type OrderResponse struct {
	ID             int64               `json:"id"`
	Status         string              `json:"status"`
	Total          string              `json:"total"`
	Address        string              `json:"address"`
	PublicToken    string              `json:"public_token,omitempty"`
	TrackingNumber string              `json:"tracking_number,omitempty"`
	Items          []OrderItemResponse `json:"items"`
	CreatedAt      time.Time           `json:"created_at"`

	// Guest fields
	Customer       *CustomerInfo `json:"customer,omitempty"`
	DeliveryMethod string        `json:"delivery_method,omitempty"`
	DeliveryCost   string        `json:"delivery_cost,omitempty"`
}

type OrderListItem struct {
	ID             int64     `json:"id"`
	Status         string    `json:"status"`
	Total          string    `json:"total"`
	Address        string    `json:"address"`
	ItemsCount     int32     `json:"items_count"`
	TrackingNumber string    `json:"tracking_number,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
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
	Customer  *CustomerInfo       `json:"customer,omitempty"`
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

	if order.PublicToken.Valid {
		resp.PublicToken = order.PublicToken.String
	}
	if order.TrackingNumber.Valid {
		resp.TrackingNumber = order.TrackingNumber.String
	}
	if order.CustomerFirstName.Valid || order.CustomerEmail.Valid {
		resp.Customer = &CustomerInfo{}
		if order.CustomerFirstName.Valid {
			resp.Customer.FirstName = order.CustomerFirstName.String
		}
		if order.CustomerLastName.Valid {
			resp.Customer.LastName = order.CustomerLastName.String
		}
		if order.CustomerEmail.Valid {
			resp.Customer.Email = order.CustomerEmail.String
		}
		if order.CustomerPhone.Valid {
			resp.Customer.Phone = order.CustomerPhone.String
		}
	}
	if order.DeliveryMethod.Valid {
		resp.DeliveryMethod = order.DeliveryMethod.String
	}
	resp.DeliveryCost = NumericToStr(order.DeliveryCost)

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
	item := OrderListItem{
		ID:         order.ID,
		Status:     string(order.Status),
		Total:      NumericToStr(order.Total),
		Address:    order.Address,
		ItemsCount: itemsCount,
		CreatedAt:  order.CreatedAt.Time,
	}
	if order.TrackingNumber.Valid {
		item.TrackingNumber = order.TrackingNumber.String
	}
	return item
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

	resp := OrderSellerListItem{
		ID:        order.ID,
		Status:    string(order.Status),
		Total:     NumericToStr(order.Total),
		Buyer:     &BuyerInfo{ID: order.UserID.Int64, Username: buyerUsername},
		Items:     itemResponses,
		CreatedAt: order.CreatedAt.Time,
	}

	if order.CustomerFirstName.Valid || order.CustomerEmail.Valid {
		resp.Customer = &CustomerInfo{}
		if order.CustomerFirstName.Valid {
			resp.Customer.FirstName = order.CustomerFirstName.String
		}
		if order.CustomerLastName.Valid {
			resp.Customer.LastName = order.CustomerLastName.String
		}
		if order.CustomerEmail.Valid {
			resp.Customer.Email = order.CustomerEmail.String
		}
		if order.CustomerPhone.Valid {
			resp.Customer.Phone = order.CustomerPhone.String
		}
	}

	return resp
}
