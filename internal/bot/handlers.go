package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	telebot "gopkg.in/telebot.v3"
)

// --- Keyboards ---

func mainMenuKeyboard() *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{ResizeKeyboard: true}
	rm.Inline(
		rm.Row(
			rm.Data("📊 Статистика", "menu_stats", ""),
			rm.Data("📦 Модерация", "menu_moderate", ""),
		),
		rm.Row(
			rm.Data("📋 Заявки", "menu_applications", ""),
			rm.Data("🛒 Заказы", "menu_orders", ""),
		),
	)
	return rm
}

func ordersMenuKeyboard() *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}
	rm.Inline(
		rm.Row(
			rm.Data("📋 Все", "orders_all", ""),
			rm.Data("⏳ Ожидают", "orders_pending", ""),
			rm.Data("💰 Оплачены", "orders_paid", ""),
		),
		rm.Row(
			rm.Data("📤 Отправлены", "orders_shipped", ""),
			rm.Data("✅ Доставлены", "orders_delivered", ""),
			rm.Data("❌ Отменены", "orders_cancelled", ""),
		),
		rm.Row(
			rm.Data("⬅️ Назад", "menu_back", ""),
		),
	)
	return rm
}

func backKeyboard() *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}
	rm.Inline(rm.Row(rm.Data("⬅️ В меню", "menu_back", "")))
	return rm
}

func moderationKeyboard(productID int64) *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}
	rm.Inline(
		rm.Row(
			rm.Data("✅ Одобрить", "mod_approve", fmt.Sprintf("%d", productID)),
			rm.Data("❌ Отклонить", "mod_reject", fmt.Sprintf("%d", productID)),
		),
	)
	return rm
}

func applicationKeyboard(appID int64) *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}
	rm.Inline(
		rm.Row(
			rm.Data("✅ Одобрить", "app_approve", fmt.Sprintf("%d", appID)),
			rm.Data("❌ Отклонить", "app_reject", fmt.Sprintf("%d", appID)),
		),
	)
	return rm
}

// --- Handlers ---

func (b *Bot) handleStart(c telebot.Context) error {
	text := "🤖 *Админ-панель Go Marketplace*\n\n" +
		"Выберите раздел с помощью кнопок ниже:"

	return c.Send(text, mainMenuKeyboard(), telebot.ModeMarkdown)
}

func (b *Bot) handleBackToMenu(c telebot.Context) error {
	text := "🤖 *Админ-панель Go Marketplace*\n\nВыберите раздел:"
	return c.Edit(text, mainMenuKeyboard(), telebot.ModeMarkdown)
}

func (b *Bot) handleStats(c telebot.Context) error {
	stats, err := b.statsSvc.GetStats(context.Background())
	if err != nil {
		b.log.Error("bot: get stats failed", slog.String("error", err.Error()))
		return c.Edit("❌ Не удалось получить статистику", backKeyboard())
	}

	totalOrders := stats.TotalOrders
	if totalOrders == 0 {
		totalOrders = 1
	}

	text := fmt.Sprintf(
		"📊 *Статистика платформы*\n\n"+
			"👥 Пользователи: *%d*\n"+
			"🏪 Продавцы: *%d*\n"+
			"📦 Товары: *%d*\n"+
			"🛒 Заказы: *%d*\n\n"+
			"━━━ *Статусы заказов* ━━━\n"+
			"⏳ Ожидают: *%d* %s\n"+
			"💰 Оплачены: *%d* %s\n"+
			"📤 Отправлены: *%d* %s\n"+
			"✅ Доставлены: *%d* %s\n"+
			"❌ Отменены: *%d* %s\n\n"+
			"💰 *Выручка: %s ₽*",
		stats.TotalUsers, stats.TotalSellers, stats.TotalProducts, stats.TotalOrders,
		stats.PendingOrders, progressBar(stats.PendingOrders, totalOrders),
		stats.PaidOrders, progressBar(stats.PaidOrders, totalOrders),
		stats.ShippedOrders, progressBar(stats.ShippedOrders, totalOrders),
		stats.DeliveredOrders, progressBar(stats.DeliveredOrders, totalOrders),
		stats.CancelledOrders, progressBar(stats.CancelledOrders, totalOrders),
		stats.TotalRevenue,
	)

	return c.Edit(text, backKeyboard(), telebot.ModeMarkdown)
}

func (b *Bot) handleModerate(c telebot.Context) error {
	items, total, err := b.productSvc.ListProductsByStatus(context.Background(), "pending", 1, 10)
	if err != nil {
		b.log.Error("bot: list pending products failed", slog.String("error", err.Error()))
		return c.Edit("❌ Не удалось получить список товаров", backKeyboard())
	}

	if total == 0 {
		return c.Edit("✅ *Нет товаров на модерации*\n\nВсе товары проверены!", backKeyboard(), telebot.ModeMarkdown)
	}

	c.Edit(fmt.Sprintf("🔍 *Товары на модерации* (%d)\n\nЗагружаю карточки...", total), telebot.ModeMarkdown)

	for _, item := range items {
		storeName := "Неизвестно"
		if item.Store != nil {
			storeName = item.Store.Name
		}

		text := fmt.Sprintf(
			"📦 *%s*\n\n"+
				"💰 Цена: *%s ₽*\n"+
				"📦 Остаток: *%d*\n"+
				"🏪 Магазин: %s\n"+
				"🆔 ID: %d",
			escapeMarkdown(item.Title), item.Price, item.Stock, escapeMarkdown(storeName), item.ID,
		)

		if err := c.Send(text, moderationKeyboard(item.ID), telebot.ModeMarkdown); err != nil {
			b.log.Error("bot: send product message failed", slog.String("error", err.Error()))
		}
	}

	return c.Send("⬅️", backKeyboard())
}

func (b *Bot) handleApplications(c telebot.Context) error {
	apps, total, err := b.appSvc.ListPending(context.Background(), 1, 10)
	if err != nil {
		b.log.Error("bot: list pending applications failed", slog.String("error", err.Error()))
		return c.Edit("❌ Не удалось получить список заявок", backKeyboard())
	}

	if total == 0 {
		return c.Edit("✅ *Нет заявок продавцов*\n\nВсе заявки рассмотрены!", backKeyboard(), telebot.ModeMarkdown)
	}

	c.Edit(fmt.Sprintf("📋 *Заявки продавцов* (%d)\n\nЗагружаю...", total), telebot.ModeMarkdown)

	for _, app := range apps {
		text := fmt.Sprintf(
			"📋 *Заявка #%d*\n\n"+
				"🏪 Магазин: *%s*\n"+
				"👤 UserID: %d",
			app.ID, escapeMarkdown(app.StoreName), app.UserID,
		)
		if app.Description.Valid {
			text += fmt.Sprintf("\n📄 Описание:\n%s", escapeMarkdown(app.Description.String))
		}

		if err := c.Send(text, applicationKeyboard(app.ID), telebot.ModeMarkdown); err != nil {
			b.log.Error("bot: send application message failed", slog.String("error", err.Error()))
		}
	}

	return c.Send("⬅️", backKeyboard())
}

func (b *Bot) handleOrdersMenu(c telebot.Context) error {
	text := "🛒 *Заказы*\n\nВыберите фильтр:"
	return c.Edit(text, ordersMenuKeyboard(), telebot.ModeMarkdown)
}

func (b *Bot) handleOrdersAll(c telebot.Context) error {
	return b.sendOrdersList(c, "")
}

func (b *Bot) handleOrdersByStatus(c telebot.Context) error {
	status := strings.TrimPrefix(c.Callback().Data, "orders_")
	return b.sendOrdersList(c, status)
}

func (b *Bot) sendOrdersList(c telebot.Context, status string) error {
	var items []dtos.OrderListItem
	var total int64
	var err error

	if status == "" {
		items, total, err = b.orderSvc.ListAllOrders(context.Background(), 1, 10)
	} else {
		items, total, err = b.orderSvc.ListOrdersByStatus(context.Background(), status, 1, 10)
	}

	if err != nil {
		b.log.Error("bot: list orders failed", slog.String("error", err.Error()))
		return c.Edit("❌ Не удалось получить заказы", ordersMenuKeyboard())
	}

	if total == 0 {
		title := "Все заказы"
		if status != "" {
			title = orderStatusTitle(status)
		}
		return c.Edit(fmt.Sprintf("📦 *%s*\n\nЗаказов нет", title), ordersMenuKeyboard(), telebot.ModeMarkdown)
	}

	title := "Все заказы"
	if status != "" {
		title = orderStatusTitle(status)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🛒 *%s* (%d)\n\n", title, total))

	for _, item := range items {
		sb.WriteString(fmt.Sprintf(
			"%s #%d | %s ₽ | товаров: %d\n",
			orderStatusEmoji(string(item.Status)), item.ID, item.Total, item.ItemsCount,
		))
	}

	sb.WriteString("\n_Показано последние 10_")

	return c.Edit(sb.String(), ordersMenuKeyboard(), telebot.ModeMarkdown)
}

// --- Helpers ---

func progressBar(value, total int64) string {
	if total == 0 {
		return ""
	}
	pct := float64(value) / float64(total)
	filled := int(pct * 10)
	if filled > 10 {
		filled = 10
	}
	return strings.Repeat("🟩", filled) + strings.Repeat("⬜", 10-filled)
}

func orderStatusEmoji(status string) string {
	switch status {
	case "pending":
		return "⏳"
	case "paid":
		return "💰"
	case "shipped":
		return "📤"
	case "delivered":
		return "✅"
	case "cancelled":
		return "❌"
	default:
		return "📦"
	}
}

func orderStatusTitle(status string) string {
	switch status {
	case "pending":
		return "Ожидают оплаты"
	case "paid":
		return "Оплаченные"
	case "shipped":
		return "Отправленные"
	case "delivered":
		return "Доставленные"
	case "cancelled":
		return "Отменённые"
	default:
		return "Заказы"
	}
}

func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"*", "\\*",
		"_", "\\_",
		"[", "\\[",
		"]", "\\]",
		"`", "\\`",
	)
	return replacer.Replace(s)
}