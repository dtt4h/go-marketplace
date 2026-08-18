package notifications

import (
	"fmt"
	"strings"
)

// ReceiptService generates HTML receipts for orders.
type ReceiptService interface {
	GenerateOrderReceiptHTML(orderID, customerName, customerEmail, customerPhone, address, total, itemsHTML string) string
	GenerateOrderSummaryHTML(items []OrderItem) string
}

type OrderItem struct {
	Title    string
	Quantity int32
	Price    string
	Total    string
}

type receiptService struct{}

// NewReceiptService creates a new receipt service.
func NewReceiptService() ReceiptService {
	return &receiptService{}
}

func (s *receiptService) GenerateOrderSummaryHTML(items []OrderItem) string {
	if len(items) == 0 {
		return "<p>Состав заказа отсутствует.</p>"
	}

	var sb strings.Builder
	sb.WriteString("<table style='border-collapse: collapse; width: 100%; margin: 10px 0;'>")
	sb.WriteString("<thead><tr style='background: #f5f5f5;'>")
	sb.WriteString("<th style='border: 1px solid #ddd; padding: 8px; text-align: left;'>Товар</th>")
	sb.WriteString("<th style='border: 1px solid #ddd; padding: 8px; text-align: center;'>Кол-во</th>")
	sb.WriteString("<th style='border: 1px solid #ddd; padding: 8px; text-align: right;'>Цена</th>")
	sb.WriteString("<th style='border: 1px solid #ddd; padding: 8px; text-align: right;'>Итого</th>")
	sb.WriteString("</tr></thead><tbody>")

	for _, item := range items {
		sb.WriteString("<tr>")
		sb.WriteString(fmt.Sprintf("<td style='border: 1px solid #ddd; padding: 8px;'>%s</td>", item.Title))
		sb.WriteString(fmt.Sprintf("<td style='border: 1px solid #ddd; padding: 8px; text-align: center;'>%d</td>", item.Quantity))
		sb.WriteString(fmt.Sprintf("<td style='border: 1px solid #ddd; padding: 8px; text-align: right;'>%s ₽</td>", item.Price))
		sb.WriteString(fmt.Sprintf("<td style='border: 1px solid #ddd; padding: 8px; text-align: right;'><strong>%s ₽</strong></td>", item.Total))
		sb.WriteString("</tr>")
	}

	sb.WriteString("</tbody></table>")
	return sb.String()
}

func (s *receiptService) GenerateOrderReceiptHTML(orderID, customerName, customerEmail, customerPhone, address, total, itemsHTML string) string {
	return `
		<div style="max-width: 600px; margin: 0 auto; font-family: Arial, sans-serif;">
			<div style="background: #f8f9fa; padding: 20px; border-radius: 8px; margin-bottom: 20px;">
				<h1 style="margin: 0 0 10px 0; color: #333;">Чек</h1>
				<p style="margin: 5px 0;"><strong>Заказ №:</strong> ` + orderID + `</p>
				<p style="margin: 5px 0;"><strong>Дата:</strong> ` + fmt.Sprintf("%d", 0) + `</p>
			</div>
			<div style="margin-bottom: 20px;">
				<h3 style="color: #555;">Покупатель</h3>
				<p style="margin: 5px 0;"><strong>Имя:</strong> ` + customerName + `</p>
				<p style="margin: 5px 0;"><strong>Email:</strong> ` + customerEmail + `</p>
				` + func() string {
					if customerPhone != "" {
						return `<p style="margin: 5px 0;"><strong>Телефон:</strong> ` + customerPhone + `</p>`
					}
					return ""
				}() + `
				<p style="margin: 5px 0;"><strong>Адрес доставки:</strong> ` + address + `</p>
			</div>
			<div style="margin-bottom: 20px;">
				<h3 style="color: #555;">Состав заказа</h3>
				` + itemsHTML + `
			</div>
			<div style="background: #e8f5e9; padding: 15px; border-radius: 8px; text-align: right;">
				<p style="margin: 0; font-size: 18px;"><strong>Итого: ` + total + ` ₽</strong></p>
			</div>
			<div style="margin-top: 20px; padding: 15px; background: #fff3e0; border-radius: 8px;">
				<p style="margin: 0; font-size: 12px; color: #666;">
					Данный документ является электронным чеком и имеет силу бумажного чека.
					Сохраните его для подтверждения покупки.
				</p>
			</div>
		</div>
	`
}
