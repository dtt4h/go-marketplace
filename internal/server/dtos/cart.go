package dtos

type AddCartItemRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
}

type UpdateCartItemRequest struct {
	Quantity int32 `json:"quantity"`
}

type CartStoreInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CartItemResponse struct {
	ID              int64          `json:"id"`
	ProductID       int64          `json:"product_id"`
	Quantity        int32          `json:"quantity"`
	Title           string         `json:"title,omitempty"`
	Price           string         `json:"price,omitempty"`
	Stock           int32          `json:"stock"`
	PreviewImageURL string         `json:"preview_image_url,omitempty"`
	Store           *CartStoreInfo `json:"store,omitempty"`
}

type CartResponse struct {
	UserID int64              `json:"user_id"`
	Items  []CartItemResponse `json:"items"`
}
