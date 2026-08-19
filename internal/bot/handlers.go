package bot

import (
	"context"
	"fmt"
	"log/slog"

	telebot "gopkg.in/telebot.v3"
)

func (b *Bot) handleStart(c telebot.Context) error {
	return c.Send(
		"🤖 Админ-панель Go Marketplace\n\n" +
			"Доступные команды:\n" +
			"/stats — статистика платформы\n" +
			"/moderate — товары на модерации\n" +
			"/applications — заявки продавцов\n" +
			"/orders — последние заказы\n" +
			"/help — справка",
	)
}

func (b *Bot) handleHelp(c telebot.Context) error {
	return b.handleStart(c)
}

func (b *Bot) handleStats(c telebot.Context) error {
	stats, err := b.statsSvc.GetStats(context.Background())
	if err != nil {
		b.log.Error("bot: get stats failed", slog.String("error", err.Error()))
		return c.Send("❌ Не удалось получить статистику")
	}

	msg := fmt.Sprintf(
		"📊 Статистика платформы\n\n"+
			"👥 Пользователи: %d\n"+
			"🏪 Продавцы: %d\n"+
			"📦 Товары: %d\n"+
			"🛒 Заказы: %d\n\n"+
			"Статусы заказов:\n"+
			"  ⏳ Ожидают: %d\n"+
			"  💰 Оплачены: %d\n"+
			"  📤 Отправлены: %d\n"+
			"  ✅ Доставлены: %d\n"+
			"  ❌ Отменены: %d\n\n"+
			"💰 Выручка: %s ₽",
		stats.TotalUsers, stats.TotalSellers, stats.TotalProducts, stats.TotalOrders,
		stats.PendingOrders, stats.PaidOrders, stats.ShippedOrders,
		stats.DeliveredOrders, stats.CancelledOrders,
		stats.TotalRevenue,
	)

	return c.Send(msg)
}

func (b *Bot) handleModerate(c telebot.Context) error {
	items, total, err := b.productSvc.ListProductsByStatus(context.Background(), "pending", 1, 10)
	if err != nil {
		b.log.Error("bot: list pending products failed", slog.String("error", err.Error()))
		return c.Send("❌ Не удалось получить список товаров")
	}

	if total == 0 {
		return c.Send("✅ Нет товаров на модерации")
	}

	c.Send(fmt.Sprintf("🔍 Товары на модерации (%d):", total))

	for _, item := range items {
		storeName := "Неизвестно"
		if item.Store != nil {
			storeName = item.Store.Name
		}

		msg := fmt.Sprintf(
			"📦 %s\n💰 %s ₽ | Остаток: %d\n🏪 %s",
			item.Title, item.Price, item.Stock, storeName,
		)
		if err := c.Send(msg, moderationKeyboard(item.ID)); err != nil {
			b.log.Error("bot: send product message failed", slog.String("error", err.Error()))
		}
	}

	return nil
}

func (b *Bot) handleApplications(c telebot.Context) error {
	apps, total, err := b.appSvc.ListPending(context.Background(), 1, 10)
	if err != nil {
		b.log.Error("bot: list pending applications failed", slog.String("error", err.Error()))
		return c.Send("❌ Не удалось получить список заявок")
	}

	if total == 0 {
		return c.Send("✅ Нет заявок продавцов на рассмотрении")
	}

	c.Send(fmt.Sprintf("📋 Заявки продавцов (%d):", total))

	for _, app := range apps {
		msg := fmt.Sprintf("📝 #%d\n🏪 %s\n👤 UserID: %d", app.ID, app.StoreName, app.UserID)
		if app.Description.Valid {
			msg += fmt.Sprintf("\n📄 %s", app.Description.String)
		}
		if err := c.Send(msg, applicationKeyboard(app.ID)); err != nil {
			b.log.Error("bot: send application message failed", slog.String("error", err.Error()))
		}
	}

	return nil
}

func (b *Bot) handleOrders(c telebot.Context) error {
	items, total, err := b.orderSvc.ListAllOrders(context.Background(), 1, 10)
	if err != nil {
		b.log.Error("bot: list orders failed", slog.String("error", err.Error()))
		return c.Send("❌ Не удалось получить список заказов")
	}

	if total == 0 {
		return c.Send("📦 Заказов пока нет")
	}

	msg := fmt.Sprintf("📦 Последние заказы (показано %d из %d):\n\n", len(items), total)
	for _, item := range items {
		msg += fmt.Sprintf("#%d | %s | %s ₽ | товаров: %d\n",
			item.ID, orderStatusEmoji(item.Status), item.Total, item.ItemsCount)
	}

	return c.Send(msg)
}

func orderStatusEmoji(status string) string {
	switch status {
	case "pending":
		return "⏳ pending"
	case "paid":
		return "💰 paid"
	case "shipped":
		return "📤 shipped"
	case "delivered":
		return "✅ delivered"
	case "cancelled":
		return "❌ cancelled"
	default:
		return status
	}
}