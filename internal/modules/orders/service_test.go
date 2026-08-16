package orders

import (
	"testing"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
)

func TestValidateCreateOrderRequest(t *testing.T) {
	t.Run("rejects empty address", func(t *testing.T) {
		req := dtos.CreateOrderRequest{
			Items: []dtos.OrderItemRequest{{ProductID: 1, Quantity: 2}},
		}

		err := validateCreateOrderRequest(req)
		if err != ErrAddressRequired {
			t.Fatalf("expected ErrAddressRequired, got %v", err)
		}
	})

	t.Run("rejects invalid product id", func(t *testing.T) {
		req := dtos.CreateOrderRequest{
			Address: "123 Main St",
			Items:   []dtos.OrderItemRequest{{ProductID: 0, Quantity: 1}},
		}

		err := validateCreateOrderRequest(req)
		if err != ErrInvalidProductID {
			t.Fatalf("expected ErrInvalidProductID, got %v", err)
		}
	})

	t.Run("accepts valid request", func(t *testing.T) {
		req := dtos.CreateOrderRequest{
			Address: "123 Main St",
			Items:   []dtos.OrderItemRequest{{ProductID: 1, Quantity: 2}},
		}

		if err := validateCreateOrderRequest(req); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
