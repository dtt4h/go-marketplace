package notifications

import (
	"log/slog"
	"time"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
)

// NotificationService handles all notification logic.
type NotificationService interface {
	SendOrderConfirmation(orderID int64, customerEmail, customerName string, items []OrderItem, total string) error
	SendOrderStatusUpdate(orderID int64, customerEmail, customerName, status string) error
	SendSellerNewOrder(sellerEmail string, sellerName string, orderID int64, total string, customerEmail, customerName string, items []OrderItem) error
	SendReceipt(to, customerName, orderID, total, items, receiptHTML string) error
	GenerateReceiptHTML(orderID int64, customerName, customerEmail, customerPhone, address, total string, items []OrderItem) string
	SendSellerApplicationApproved(to, username, storeName string) error
	SendSellerApplicationRejected(to, username, reason string) error
	SendNewApplicationNotification(to, username, storeName, description string) error
}

type notificationService struct {
	email   EmailService
	receipt ReceiptService
	log     *slog.Logger
}

// NewNotificationService creates a new notification service.
func NewNotificationService(email EmailService, receipt ReceiptService, log *slog.Logger) NotificationService {
	return &notificationService{email: email, receipt: receipt, log: log}
}

func (s *notificationService) SendOrderConfirmation(orderID int64, customerEmail, customerName string, items []OrderItem, total string) error {
	itemsHTML := s.receipt.GenerateOrderSummaryHTML(items)
	return s.email.SendOrderConfirmation(customerEmail, customerName, formatOrderID(orderID), total, itemsHTML)
}

func (s *notificationService) SendOrderStatusUpdate(orderID int64, customerEmail, customerName, status string) error {
	return s.email.SendOrderStatusUpdate(customerEmail, customerName, formatOrderID(orderID), status)
}

func (s *notificationService) SendSellerNewOrder(sellerEmail string, sellerName string, orderID int64, total string, customerEmail, customerName string, items []OrderItem) error {
	return s.email.SendSellerNewOrder(sellerEmail, sellerName, formatOrderID(orderID), total, customerEmail, customerName)
}

func (s *notificationService) SendReceipt(to, customerName, orderID, total, items, receiptHTML string) error {
	return s.email.SendReceipt(to, customerName, orderID, total, items, receiptHTML)
}

func (s *notificationService) GenerateReceiptHTML(orderID int64, customerName, customerEmail, customerPhone, address, total string, items []OrderItem) string {
	itemsHTML := s.receipt.GenerateOrderSummaryHTML(items)
	return s.receipt.GenerateOrderReceiptHTML(formatOrderID(orderID), customerName, customerEmail, customerPhone, address, total, itemsHTML)
}

func (s *notificationService) SendSellerApplicationApproved(to, username, storeName string) error {
	return s.email.SendSellerApplicationApproved(to, username, storeName)
}

func (s *notificationService) SendSellerApplicationRejected(to, username, reason string) error {
	return s.email.SendSellerApplicationRejected(to, username, reason)
}

func (s *notificationService) SendNewApplicationNotification(to, username, storeName, description string) error {
	return s.email.SendNewApplicationNotification(to, username, storeName, description)
}

// BuildOrderItems converts order items to OrderItem for notifications.
func BuildOrderItems(items []db.GetOrderItemsRow) []OrderItem {
	result := make([]OrderItem, 0, len(items))
	for _, item := range items {
		result = append(result, OrderItem{
			Title:    item.ProductTitle,
			Quantity: item.Quantity,
			Price:    dtos.NumericToStr(item.Price),
			Total:    dtos.NumericToStr(item.Price) + " ₽",
		})
	}
	return result
}

func formatOrderID(id int64) string {
	now := time.Now()
	return now.Format("20060102") + "-" + string(rune('0'+id%10)) + string(rune('A'+id%26))
}
