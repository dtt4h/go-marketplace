// @title           Go Marketplace API
// @version         1.0
// @description     Marketplace backend API for buyers, sellers and admins.
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT access token.

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/dtt4h/go-marketplace/internal/bot"
	"github.com/dtt4h/go-marketplace/internal/config"
	"github.com/dtt4h/go-marketplace/internal/database"
	sqlcdb "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/logger"
	"github.com/dtt4h/go-marketplace/internal/modules/notifications"
	"github.com/dtt4h/go-marketplace/internal/modules/orders"
	"github.com/dtt4h/go-marketplace/internal/modules/products"
	"github.com/dtt4h/go-marketplace/internal/modules/sellerapplications"
	"github.com/dtt4h/go-marketplace/internal/modules/users"
	"github.com/dtt4h/go-marketplace/internal/server"
	"github.com/dtt4h/go-marketplace/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log := logger.New(cfg.Env)

	ctx := context.Background()

	db, err := database.New(ctx, cfg.Database.DSN(), log)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	if err := runMigrations(cfg.Database.DSN()); err != nil {
		log.Error("failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	store, err := storage.NewObjectStorage(storage.ObjectStorageConfig{
		Endpoint:  cfg.S3.Endpoint,
		Region:    cfg.S3.Region,
		AccessKey: cfg.S3.AccessKey,
		SecretKey: cfg.S3.SecretKey,
		Bucket:    cfg.S3.Bucket,
		Secure:    cfg.S3.Endpoint != "" && strings.HasPrefix(cfg.S3.Endpoint, "https://"),
	})
	if err != nil {
		log.Warn("failed to initialize S3 storage, images will be unavailable", slog.String("error", err.Error()))
	}

	srv := server.New(cfg, log, db, store)

	// Periodically expire unpaid pending orders and restore their stock.
	expirer := orders.NewOrderExpirer(sqlcdb.New(db), db, log)
	go runOrderExpirer(expirer, log, ctx)

	// Start Telegram admin bot (optional, requires TELEGRAM_BOT_TOKEN).
	if cfg.Telegram.BotToken != "" {
		queries := sqlcdb.New(db)

		statsRepo := orders.NewAdminStatsRepository(queries)
		statsSvc := orders.NewAdminStatsService(statsRepo)

		productRepo := products.NewProductRepository(queries, db)
		storeResolver := users.NewUserRepository(queries, db)
		productSvc := products.NewProductService(productRepo, storeResolver)

		appRepo := sellerapplications.NewSellerApplicationRepository(queries)
		userRepo := users.NewUserRepository(queries, db)
		emailSvc := notifications.NewEmailService(cfg.SMTP, log)
		receiptSvc := notifications.NewReceiptService()
		notifSvc := notifications.NewNotificationService(emailSvc, receiptSvc, log)
		appSvc := sellerapplications.NewSellerApplicationService(appRepo, userRepo, notifSvc)

		adminOrderRepo := orders.NewAdminOrderRepository(queries)
		adminOrderSvc := orders.NewAdminOrderService(adminOrderRepo, userRepo)

		adminBot, err := bot.New(
			cfg.Telegram.BotToken,
			cfg.Telegram.AdminIDs,
			statsSvc,
			productSvc,
			appSvc,
			adminOrderSvc,
			log,
		)
		if err != nil {
			log.Error("failed to create telegram bot", slog.String("error", err.Error()))
		} else {
			go adminBot.Start(ctx)
		}
	} else {
		log.Info("telegram bot token not set, admin bot disabled")
	}

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("server stopped gracefully")
}

func runMigrations(dsn string) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func runOrderExpirer(expirer *orders.OrderExpirer, log *slog.Logger, ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := expirer.ExpirePendingOrders(context.Background()); err != nil {
				log.Error("order expiration failed", slog.String("error", err.Error()))
			}
		}
	}
}
