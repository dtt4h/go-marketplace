package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	telebot "gopkg.in/telebot.v3"
)

// ===================== KEYBOARDS =====================

// replyMainMenu — постоянная клавиатура над полем ввода.
// Даёт быстрый доступ к разделам без команд.
func replyMainMenu() *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{ResizeKeyboard: true}
	rm.Reply(
		rm.Row(
			rm.Text("📊 Статистика"),
			rm.Text("📦 Модерация"),
		),
		rm.Row(
			rm.Text("📋 Заявки"),
			rm.Text("🛒 Заказы"),
		),
	)
	return rm
}

// mainMenuInline — inline-кнопки главного меню (внутри сообщения).
func mainMenuInline() *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}
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
			rm.Data("⬅️ В меню", "menu_back", ""),
		),
	)
	return rm
}

func backKeyboard() *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}
	rm.Inline(rm.Row(rm.Data("⬅️ В меню", "menu_back", "")))
	return rm
}

func statsKeyboard() *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}
	rm.Inline(
		rm.Row(
			rm.Data("🔄 Обновить", "stats_refresh", ""),
			rm.Data("⬅️ В меню", "menu_back", ""),
		),
	)
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

// paginationKeyboard — универсальная пагинация.
// unique — префикс колбэка (mod_page, app_page, orders_page).
// data — строка, передаваемая в callback (формат "page:status" или просто "page").
func paginationKeyboard(unique, data string, page, totalPages int) *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}

	var btns []telebot.Btn
	if page > 1 {
		btns = append(btns, rm.Data("⬅️", unique, fmt.Sprintf("%d:%s", page-1, data)))
	}
	btns = append(btns, rm.Data(fmt.Sprintf("📄 %d/%d", page, totalPages), "noop", ""))
	if page < totalPages {
		btns = append(btns, rm.Data("➡️", unique, fmt.Sprintf("%d:%s", page+1, data)))
	}

	rm.Inline(
		rm.Row(btns...),
		rm.Row(rm.Data("⬅️ В меню", "menu_back", "")),
	)
	return rm
}

func orderDetailKeyboard(orderID int64) *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}
	rm.Inline(
		rm.Row(
			rm.Data("⬅️ К заказам", "menu_orders", ""),
			rm.Data("⬅️ В меню", "menu_back", ""),
		),
	)
	return rm
}

// ===================== HANDLERS =====================

func (b *Bot) handleStart(c telebot.Context) error {
	text := "🤖 *Админ-панель Go Marketplace*\n\n" +
		"Управляйте платформой прямо из Telegram.\n\n" +
		"• 📊 Статистика — метрики в реальном времени\n" +
		"• 📦 Модерация — товары на проверке\n" +
		"• 📋 Заявки — заявки продавцов\n" +
		"• 🛒 Заказы — управление заказами\n\n" +
		"_Кнопки доступны над полем ввода и в сообщениях_"

	return c.Send(text, replyMainMenu(), mainMenuInline(), telebot.ModeMarkdown)
}

func (b *Bot) handleBackToMenu(c telebot.Context) error {
	text := "🤖 *Админ-панель Go Marketplace*\n\nВыберите раздел:"
	return c.Edit(text, mainMenuInline(), telebot.ModeMarkdown)
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
			"━━━ *Пользователи* ━━━\n"+
			"👥 Всего: *%d*\n"+
			"🏪 Продавцов: *%d*\n\n"+
			"━━━ *Контент* ━━━\n"+
			"📦 Товаров: *%d*\n"+
			"🛒 Заказов: *%d*\n\n"+
			"━━━ *Статусы заказов* ━━━\n"+
			"⏳ Ожидают: *%d* %s\n"+
			"💰 Оплачены: *%d* %s\n"+
			"📤 Отправлены: *%d* %s\n"+
			"✅ Доставлены: *%d* %s\n"+
			"❌ Отменены: *%d* %s\n\n"+
			"💰 *Выручка: %s ₽*",
		stats.TotalUsers, stats.TotalSellers,
		stats.TotalProducts, stats.TotalOrders,
		stats.PendingOrders, progressBar(stats.PendingOrders, totalOrders),
		stats.PaidOrders, progressBar(stats.PaidOrders, totalOrders),
		stats.ShippedOrders, progressBar(stats.ShippedOrders, totalOrders),
		stats.DeliveredOrders, progressBar(stats.DeliveredOrders, totalOrders),
		stats.CancelledOrders, progressBar(stats.CancelledOrders, totalOrders),
		stats.TotalRevenue,
	)

	return c.Edit(text, statsKeyboard(), telebot.ModeMarkdown)
}

func (b *Bot) handleModerate(c telebot.Context) error {
	return b.sendModeratePage(c, 1)
}

func (b *Bot) handleModeratePage(c telebot.Context) error {
	parts := strings.SplitN(c.Callback().Data, ":", 2)
	page, err := strconv.Atoi(parts[0])
	if err != nil || page < 1 {
		page = 1
	}
	return b.sendModeratePage(c, page)
}

func (b *Bot) sendModeratePage(c telebot.Context, page int) error {
	items, total, err := b.productSvc.ListProductsByStatus(context.Background(), "pending", page, pageLimit)
	if err != nil {
		b.log.Error("bot: list pending products failed", slog.String("error", err.Error()))
		return c.Edit("❌ Не удалось получить список товаров", backKeyboard())
	}

	if total == 0 {
		return c.Edit("✅ *Нет товаров на модерации*\n\nВсе товары проверены! 🎉", backKeyboard(), telebot.ModeMarkdown)
	}

	totalPages := int((total + int64(pageLimit) - 1) / int64(pageLimit))
	if totalPages < 1 {
		totalPages = 1
	}

	header := fmt.Sprintf("🔍 *Товары на модерации* (%d)\n\n_Страница %d из %d_", total, page, totalPages)
	c.Edit(header, telebot.ModeMarkdown)

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
				"🆔 ID: `%d`",
			escapeMarkdown(item.Title), item.Price, item.Stock, escapeMarkdown(storeName), item.ID,
		)

		if err := c.Send(text, moderationKeyboard(item.ID), telebot.ModeMarkdown); err != nil {
			b.log.Error("bot: send product message failed", slog.String("error", err.Error()))
		}
	}

	return c.Send("Навигация:", paginationKeyboard("mod_page", "", page, totalPages))
}

func (b *Bot) handleApplications(c telebot.Context) error {
	return b.sendApplicationsPage(c, 1)
}

func (b *Bot) handleApplicationsPage(c telebot.Context) error {
	parts := strings.SplitN(c.Callback().Data, ":", 2)
	page, err := strconv.Atoi(parts[0])
	if err != nil || page < 1 {
		page = 1
	}
	return b.sendApplicationsPage(c, page)
}

func (b *Bot) sendApplicationsPage(c telebot.Context, page int) error {
	apps, total, err := b.appSvc.ListPending(context.Background(), page, pageLimit)
	if err != nil {
		b.log.Error("bot: list pending applications failed", slog.String("error", err.Error()))
		return c.Edit("❌ Не удалось получить список заявок", backKeyboard())
	}

	if total == 0 {
		return c.Edit("✅ *Нет заявок продавцов*\n\nВсе заявки рассмотрены! 🎉", backKeyboard(), telebot.ModeMarkdown)
	}

	totalPages := int((total + int64(pageLimit) - 1) / int64(pageLimit))
	if totalPages < 1 {
		totalPages = 1
	}

	header := fmt.Sprintf("📋 *Заявки продавцов* (%d)\n\n_Страница %d из %d_", total, page, totalPages)
	c.Edit(header, telebot.ModeMarkdown)

	for _, app := range apps {
		text := fmt.Sprintf(
			"📋 *Заявка #%d*\n\n"+
				"🏪 Магазин: *%s*\n"+
				"👤 UserID: `%d`",
			app.ID, escapeMarkdown(app.StoreName), app.UserID,
		)
		if app.Description.Valid && app.Description.String != "" {
			text += fmt.Sprintf("\n📄 Описание:\n%s", escapeMarkdown(app.Description.String))
		}

		if err := c.Send(text, applicationKeyboard(app.ID), telebot.ModeMarkdown); err != nil {
			b.log.Error("bot: send application message failed", slog.String("error", err.Error()))
		}
	}

	return c.Send("Навигация:", paginationKeyboard("app_page", "", page, totalPages))
}

func (b *Bot) handleOrdersMenu(c telebot.Context) error {
	text := "🛒 *Заказы*\n\nВыберите фильтр:"
	return c.Edit(text, ordersMenuKeyboard(), telebot.ModeMarkdown)
}

func (b *Bot) handleOrdersAll(c telebot.Context) error {
	return b.sendOrdersList(c, "", 1)
}

func (b *Bot) handleOrdersByStatus(c telebot.Context) error {
	status := strings.TrimPrefix(c.Callback().Data, "orders_")
	return b.sendOrdersList(c, status, 1)
}

func (b *Bot) handleOrdersPage(c telebot.Context) error {
	parts := strings.SplitN(c.Callback().Data, ":", 2)
	if len(parts) < 2 {
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка навигации"})
	}
	page, err := strconv.Atoi(parts[0])
	if err != nil || page < 1 {
		page = 1
	}
	status := parts[1]
	return b.sendOrdersList(c, status, page)
}

func (b *Bot) sendOrdersList(c telebot.Context, status string, page int) error {
	var items []dtos.OrderListItem
	var total int64
	var err error

	if status == "" {
		items, total, err = b.orderSvc.ListAllOrders(context.Background(), page, pageLimit)
	} else {
		items, total, err = b.orderSvc.ListOrdersByStatus(context.Background(), status, page, pageLimit)
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

	totalPages := int((total + int64(pageLimit) - 1) / int64(pageLimit))
	if totalPages < 1 {
		totalPages = 1
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🛒 *%s* (%d)\n\n", title, total))

	for _, item := range items {
		sb.WriteString(fmt.Sprintf(
			"%s *#%d* | %s ₽ | товаров: %d\n",
			orderStatusEmoji(string(item.Status)), item.ID, item.Total, item.ItemsCount,
		))
	}

	sb.WriteString(fmt.Sprintf("\n_Страница %d из %d_", page, totalPages))

	return c.Edit(sb.String(), paginationKeyboard("orders_page", status, page, totalPages), telebot.ModeMarkdown)
}

func (b *Bot) handleOrderDetail(c telebot.Context) error {
	orderID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Некорректный ID", ShowAlert: true})
	}

	order, err := b.orderSvc.GetOrderDetail(context.Background(), orderID)
	if err != nil {
		b.log.Error("bot: get order detail failed", slog.String("error", err.Error()))
		return c.Edit("❌ Заказ не найден", ordersMenuKeyboard())
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🛒 *Заказ #%d*\n\n", order.ID))
	sb.WriteString(fmt.Sprintf("Статус: %s %s\n", orderStatusEmoji(order.Status), orderStatusTitle(order.Status)))
	sb.WriteString(fmt.Sprintf("💰 Сумма: *%s ₽*\n", order.Total))
	sb.WriteString(fmt.Sprintf("📍 Адрес: %s\n", escapeMarkdown(order.Address)))

	if order.TrackingNumber != "" {
		sb.WriteString(fmt.Sprintf("📦 Трек-номер: `%s`\n", order.TrackingNumber))
	}

	if order.Customer != nil {
		sb.WriteString("\n━━━ *Покупатель* ━━━\n")
		if order.Customer.FirstName != "" {
			sb.WriteString(fmt.Sprintf("👤 Имя: %s\n", escapeMarkdown(order.Customer.FirstName)))
		}
		if order.Customer.Email != "" {
			sb.WriteString(fmt.Sprintf("✉️ Email: %s\n", order.Customer.Email))
		}
		if order.Customer.Phone != "" {
			sb.WriteString(fmt.Sprintf("📞 Телефон: %s\n", order.Customer.Phone))
		}
	}

	if len(order.Items) > 0 {
		sb.WriteString("\n━━━ *Состав* ━━━\n")
		for _, item := range order.Items {
			sb.WriteString(fmt.Sprintf("• %s ×%d — %s ₽\n",
				escapeMarkdown(item.ProductTitle), item.Quantity, item.Price))
		}
	}

	sb.WriteString(fmt.Sprintf("\n📅 Создан: %s", order.CreatedAt.Format("02.01.2006 15:04")))

	return c.Edit(sb.String(), orderDetailKeyboard(order.ID), telebot.ModeMarkdown)
}

// ===================== HELPERS =====================

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
		return "Ожидает оплаты"
	case "paid":
		return "Оплачен"
	case "shipped":
		return "Отправлен"
	case "delivered":
		return "Доставлен"
	case "cancelled":
		return "Отменён"
	default:
		return status
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
