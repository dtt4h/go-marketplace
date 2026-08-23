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

	b.bot.Handle("/start", b.handleStart)
	b.bot.Handle("/help", b.handleHelp)
	b.bot.Handle("/stats", b.handleStats)
	b.bot.Handle("/moderate", b.handleModerate)
	b.bot.Handle("/applications", b.handleApplications)
	b.bot.Handle("/orders", b.handleOrders)

	b.bot.Handle(&telebot.InlineButton{Unique: "mod_approve"}, b.callbackModerateApprove)
	b.bot.Handle(&telebot.InlineButton{Unique: "mod_reject"}, b.callbackModerateReject)
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
