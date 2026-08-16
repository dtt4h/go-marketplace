package payments

import (
	"testing"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
)

func TestValidateWebhookRequest(t *testing.T) {
	t.Run("rejects invalid status", func(t *testing.T) {
		req := dtos.WebhookRequest{OrderID: 1, Status: "unknown"}
		_, err := validateWebhookRequest(req)
		if err != ErrInvalidWebhook {
			t.Fatalf("expected ErrInvalidWebhook, got %v", err)
		}
	})

	t.Run("accepts supported status", func(t *testing.T) {
		req := dtos.WebhookRequest{OrderID: 1, Status: "succeeded"}
		status, err := validateWebhookRequest(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if status != "succeeded" {
			t.Fatalf("expected succeeded status, got %v", status)
		}
	})
}
