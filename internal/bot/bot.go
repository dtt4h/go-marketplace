package bot

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dtt4h/go-marketplace/internal/modules/orders"
	"github.com/dtt4h/go-marketplace/internal/modules/products"
	"github.com/dtt4h/go-marketplace/internal/modules/sellerapplications"
	telebot "gopkg.in/telebot.v3"
)

const pageLimit = 5

type Bot struct {
	bot        *telebot.Bot
	log        *slog.Logger
	statsSvc   orders.AdminStatsService
	productSvc products.ProductService
	appSvc     sellerapplications.SellerApplicationService
	orderSvc   orders.AdminOrderService
	adminIDs   map[int64]bool
}

func New(
	token string,
	adminIDs []int64,
	statsSvc orders.AdminStatsService,
	productSvc products.ProductService,
	appSvc sellerapplications.SellerApplicationService,
	orderSvc orders.AdminOrderService,
	log *slog.Logger,
) (*Bot, error) {
	if token == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	adminMap := make(map[int64]bool, len(adminIDs))
	for _, id := range adminIDs {
		adminMap[id] = true
	}

	b, err := telebot.NewBot(telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	instance := &Bot{
		bot:        b,
		log:        log,
		statsSvc:   statsSvc,
		productSvc: productSvc,
		appSvc:     appSvc,
		orderSvc:   orderSvc,
		adminIDs:   adminMap,
	}

	instance.registerHandlers()

	return instance, nil
}

func (b *Bot) registerHandlers() {
	b.bot.Use(func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			sender := c.Sender()
			if sender == nil || !b.adminIDs[sender.ID] {
				return c.Send("⛔ У вас нет доступа к этому боту.")
			}
			return next(c)
		}
	})

	// Commands
	b.bot.Handle("/start", b.handleStart)
	b.bot.Handle("/help", b.handleStart)

	// Reply keyboard buttons (persistent keyboard above input field)
	b.bot.Handle("\xf0\x9f\x93\x8a Статистика", b.handleStats)
	b.bot.Handle("\xf0\x9f\x93\xa6 Модерация", b.handleModerate)
	b.bot.Handle("\xf0\x9f\x93\x8b Заявки", b.handleApplications)
	b.bot.Handle("\xf0\x9f\x9b\x92 Заказы", b.handleOrdersMenu)

	// Fallback: any unrecognized text → show menu
	b.bot.Handle(telebot.OnText, b.handleTextFallback)

	// Main menu inline callbacks
	b.bot.Handle(&telebot.InlineButton{Unique: "menu_stats"}, b.handleStats)
	b.bot.Handle(&telebot.InlineButton{Unique: "menu_moderate"}, b.handleModerate)
	b.bot.Handle(&telebot.InlineButton{Unique: "menu_applications"}, b.handleApplications)
	b.bot.Handle(&telebot.InlineButton{Unique: "menu_orders"}, b.handleOrdersMenu)
	b.bot.Handle(&telebot.InlineButton{Unique: "menu_back"}, b.handleBackToMenu)

	// Stats refresh
	b.bot.Handle(&telebot.InlineButton{Unique: "stats_refresh"}, b.handleStats)

	// Noop (pagination page indicator — non-clickable)
	b.bot.Handle(&telebot.InlineButton{Unique: "noop"}, b.handleNoop)

	// Order filter callbacks
	b.bot.Handle(&telebot.InlineButton{Unique: "orders_all"}, b.handleOrdersAll)
	b.bot.Handle(&telebot.InlineButton{Unique: "orders_pending"}, b.handleOrdersByStatus)
	b.bot.Handle(&telebot.InlineButton{Unique: "orders_paid"}, b.handleOrdersByStatus)
	b.bot.Handle(&telebot.InlineButton{Unique: "orders_shipped"}, b.handleOrdersByStatus)
	b.bot.Handle(&telebot.InlineButton{Unique: "orders_delivered"}, b.handleOrdersByStatus)
	b.bot.Handle(&telebot.InlineButton{Unique: "orders_cancelled"}, b.handleOrdersByStatus)

	// Order detail
	b.bot.Handle(&telebot.InlineButton{Unique: "order_detail"}, b.handleOrderDetail)

	// Pagination
	b.bot.Handle(&telebot.InlineButton{Unique: "mod_page"}, b.handleModeratePage)
	b.bot.Handle(&telebot.InlineButton{Unique: "app_page"}, b.handleApplicationsPage)
	b.bot.Handle(&telebot.InlineButton{Unique: "orders_page"}, b.handleOrdersPage)

	// Moderation callbacks
	b.bot.Handle(&telebot.InlineButton{Unique: "mod_approve"}, b.callbackModerateApprove)
	b.bot.Handle(&telebot.InlineButton{Unique: "mod_reject"}, b.callbackModerateReject)

	// Application callbacks
	b.bot.Handle(&telebot.InlineButton{Unique: "app_approve"}, b.callbackAppApprove)
	b.bot.Handle(&telebot.InlineButton{Unique: "app_reject"}, b.callbackAppReject)
}

func (b *Bot) Start(ctx context.Context) {
	b.log.Info("telegram bot starting")
	go b.bot.Start()

	<-ctx.Done()
	b.log.Info("telegram bot stopping")
	b.bot.Stop()
	b.log.Info("telegram bot stopped")
}

// Stop gracefully stops the bot.
func (b *Bot) Stop() {
	if b.bot != nil {
		b.bot.Stop()
	}
}
