package server

import (
	"time"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/dtt4h/go-marketplace/docs"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/modules/auth"
	"github.com/dtt4h/go-marketplace/internal/modules/cart"
	"github.com/dtt4h/go-marketplace/internal/modules/notifications"
	"github.com/dtt4h/go-marketplace/internal/modules/orders"
	"github.com/dtt4h/go-marketplace/internal/modules/payments"
	"github.com/dtt4h/go-marketplace/internal/modules/products"
	"github.com/dtt4h/go-marketplace/internal/modules/sellerapplications"
	"github.com/dtt4h/go-marketplace/internal/modules/users"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
)

func (s *Server) setupRoutes() {
	s.router.Get("/health", s.healthReady)
	s.router.Get("/health/live", s.healthLive)
	s.router.Get("/health/ready", s.healthReady)

	s.router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Prometheus metrics endpoint
	s.router.Handle("/metrics", s.metrics.Handler())

	s.router.Route("/api/v1", func(r chi.Router) {
		s.registerAuthRoutes(r)
		s.registerUserRoutes(r)
		s.registerProductRoutes(r)
		s.registerCartRoutes(r)
		s.registerOrderRoutes(r)
		s.registerPaymentRoutes(r)
		s.registerSellerApplicationRoutes(r)
	})
}

func (s *Server) registerAuthRoutes(r chi.Router) {
	queries := db.New(s.db)

	authRepo := auth.NewAuthRepository(queries)
	authSvc := auth.NewAuthService(authRepo, s.cfg)
	authHandler := auth.NewAuthHandler(authSvc, s.cfg.DevMode)

	authRL := mw.NewRateLimiterWithLog(s.rdb, s.log, 10, time.Minute)

	r.With(mw.RateLimit(authRL)).Post("/auth/register", authHandler.Register)
	r.With(mw.RateLimit(authRL)).Post("/auth/login", authHandler.Login)
	r.Post("/auth/refresh", authHandler.Refresh)
	r.With(mw.JWTAuth(s.cfg)).Post("/auth/logout", authHandler.Logout)

	r.With(mw.RateLimit(authRL)).Post("/auth/forgot-password", authHandler.ForgotPassword)
	r.With(mw.RateLimit(authRL)).Post("/auth/reset-password", authHandler.ResetPassword)
}

func (s *Server) registerUserRoutes(r chi.Router) {
	queries := db.New(s.db)

	userRepo := users.NewUserRepository(queries, s.db)
	userSvc := users.NewUserService(userRepo)
	userHandler := users.NewUserHandler(userSvc)

	r.With(mw.JWTAuth(s.cfg)).Get("/users/me", userHandler.GetProfile)
	r.With(mw.JWTAuth(s.cfg)).Get("/users/me/store", userHandler.GetStore)
	r.With(mw.JWTAuth(s.cfg)).Patch("/users/me", userHandler.UpdateProfile)
	r.With(mw.JWTAuth(s.cfg)).Patch("/users/me/store", userHandler.UpdateStore)
}

func (s *Server) registerProductRoutes(r chi.Router) {
	queries := db.New(s.db)

	productRepo := products.NewProductRepository(queries, s.db)
	storeResolver := users.NewUserRepository(queries, s.db)
	productSvc := products.NewProductService(productRepo, storeResolver, s.cache)
	productHandler := products.NewProductHandler(productSvc, s.cache)

	r.Get("/products", productHandler.ListProducts)
	r.Get("/products/categories", productHandler.ListCategories)
	r.Get("/products/{id}", productHandler.GetProduct)
	r.Get("/stores/{id}/products", productHandler.ListProductsByStore)

	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleSeller))).
		Post("/products", productHandler.CreateProduct)
	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleSeller))).
		Patch("/products/{id}", productHandler.UpdateProduct)
	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleSeller))).
		Delete("/products/{id}", productHandler.DeleteProduct)

	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleAdmin))).
		Patch("/products/{id}/moderate", productHandler.ModerateProduct)

	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleAdmin))).
		Get("/admin/products", productHandler.ListProductsByStatus)

	if s.store != nil {
		imageSvc := products.NewImageUploadService(productRepo, s.store)
		imageHandler := products.NewImageHandler(imageSvc)

		r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleSeller))).
			Post("/products/{id}/images", imageHandler.UploadImage)
		r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleSeller))).
			Delete("/products/images/{id}", imageHandler.DeleteImage)
		r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleSeller))).
			Get("/products/{id}/images/presigned", imageHandler.GetPresignedUploadURL)
	}
}

func (s *Server) registerCartRoutes(r chi.Router) {
	queries := db.New(s.db)

	cartRepo := cart.NewCartRepository(queries)
	cartSvc := cart.NewCartService(cartRepo)
	cartHandler := cart.NewCartHandler(cartSvc)

	r.With(mw.JWTAuth(s.cfg)).Get("/cart", cartHandler.GetCart)
	r.With(mw.JWTAuth(s.cfg)).Post("/cart/items", cartHandler.AddItem)
	r.With(mw.JWTAuth(s.cfg)).Patch("/cart/items/{itemID}", cartHandler.UpdateItem)
	r.With(mw.JWTAuth(s.cfg)).Delete("/cart/items/{itemID}", cartHandler.RemoveItem)
	r.With(mw.JWTAuth(s.cfg)).Delete("/cart", cartHandler.Clear)
}

func (s *Server) registerOrderRoutes(r chi.Router) {
	queries := db.New(s.db)

	orderRepo := orders.NewOrderRepository(queries)
	userRepo := users.NewUserRepository(queries, s.db)

	// Create shared notification service
	emailSvc := notifications.NewEmailService(s.cfg.SMTP, s.log)
	receiptSvc := notifications.NewReceiptService()
	notifSvc := notifications.NewNotificationService(emailSvc, receiptSvc, s.log)

	// Create payment service once and share it
	paymentRepo := payments.NewPaymentRepository(queries, s.db)
	paymentSvc := payments.NewPaymentService(paymentRepo, orderRepo, userRepo, s.db, notifSvc, s.log)

	orderSvc := orders.NewOrderService(orderRepo, userRepo, paymentSvc, notifSvc, s.db, s.log)
	orderHandler := orders.NewOrderHandler(orderSvc)
	paymentHandler := payments.NewPaymentHandler(paymentSvc, s.cfg.Payments.WebhookSecret)

	// Admin endpoints
	adminOrderRepo := orders.NewAdminOrderRepository(queries)
	adminOrderSvc := orders.NewAdminOrderService(adminOrderRepo, userRepo)
	adminOrderHandler := orders.NewAdminOrderHandler(adminOrderSvc)

	// Admin stats
	adminStatsRepo := orders.NewAdminStatsRepository(queries)
	adminStatsSvc := orders.NewAdminStatsService(adminStatsRepo)
	adminStatsHandler := orders.NewAdminStatsHandler(adminStatsSvc)

	// Rate limiters
	orderRL := mw.NewRateLimiterWithLog(s.rdb, s.log, 20, time.Minute)
	paymentRL := mw.NewRateLimiterWithLog(s.rdb, s.log, 10, time.Minute)

	// Guest checkout — no auth required
	r.With(mw.RateLimit(orderRL)).Post("/orders", orderHandler.CreateOrder)
	r.Get("/orders/public/{public_token}", orderHandler.GetPublicOrder)
	r.With(mw.RateLimit(paymentRL)).Post("/orders/public/{public_token}/payment", paymentHandler.CreateGuestPayment)

	// Authenticated user orders
	r.With(mw.JWTAuth(s.cfg)).Get("/orders/me", orderHandler.ListOrdersByUser)
	r.With(mw.JWTAuth(s.cfg)).Get("/orders/me/{id}", orderHandler.GetOrder)

	// Seller endpoints
	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleSeller))).
		Get("/orders/seller", orderHandler.ListOrdersBySeller)
	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleSeller))).
		Patch("/orders/{id}/status", orderHandler.UpdateOrderStatus)
	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleSeller))).
		Patch("/orders/{id}/tracking", orderHandler.UpdateOrderTracking)

	// Admin endpoints
	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleAdmin))).
		Get("/admin/orders", adminOrderHandler.ListAllOrders)
	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleAdmin))).
		Get("/admin/orders/status", adminOrderHandler.ListOrdersByStatus)
	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleAdmin))).
		Get("/admin/orders/user", adminOrderHandler.ListOrdersByUser)
	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleAdmin))).
		Get("/admin/orders/{id}", adminOrderHandler.GetOrderDetail)
	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleAdmin))).
		Get("/admin/stats", adminStatsHandler.GetStats)
}

func (s *Server) registerPaymentRoutes(r chi.Router) {
	queries := db.New(s.db)

	// Reuse the same notification service pattern
	emailSvc := notifications.NewEmailService(s.cfg.SMTP, s.log)
	receiptSvc := notifications.NewReceiptService()
	notifSvc := notifications.NewNotificationService(emailSvc, receiptSvc, s.log)

	paymentRepo := payments.NewPaymentRepository(queries, s.db)
	orderRepo := orders.NewOrderRepository(queries)
	userRepo := users.NewUserRepository(queries, s.db)
	paymentSvc := payments.NewPaymentService(paymentRepo, orderRepo, userRepo, s.db, notifSvc, s.log)
	paymentHandler := payments.NewPaymentHandler(paymentSvc, s.cfg.Payments.WebhookSecret)

	r.With(mw.JWTAuth(s.cfg)).Post("/payments", paymentHandler.CreatePayment)
	r.With(mw.JWTAuth(s.cfg)).Get("/payments/me", paymentHandler.ListPaymentsByUser)
	r.With(mw.JWTAuth(s.cfg)).Get("/payments/{id}", paymentHandler.GetPayment)
	r.Post("/payments/webhook", paymentHandler.Webhook)
	r.With(mw.JWTAuth(s.cfg)).Post("/payments/{id}/refund", paymentHandler.RefundPayment)
}

func (s *Server) registerSellerApplicationRoutes(r chi.Router) {
	queries := db.New(s.db)

	appRepo := sellerapplications.NewSellerApplicationRepository(queries)
	userRepo := users.NewUserRepository(queries, s.db)

	// Create notification service for seller applications
	emailSvc := notifications.NewEmailService(s.cfg.SMTP, s.log)
	receiptSvc := notifications.NewReceiptService()
	notifSvc := notifications.NewNotificationService(emailSvc, receiptSvc, s.log)

	appSvc := sellerapplications.NewSellerApplicationService(appRepo, userRepo, notifSvc)
	appHandler := sellerapplications.NewSellerApplicationHandler(appSvc)

	// User-facing
	r.With(mw.JWTAuth(s.cfg)).Post("/seller-applications", appHandler.Create)
	r.With(mw.JWTAuth(s.cfg)).Get("/seller-applications/me", appHandler.ListMine)

	// Admin
	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleAdmin))).
		Get("/admin/seller-applications", appHandler.ListPending)
	r.With(mw.JWTAuth(s.cfg), mw.RoleGuard(string(db.UserRoleAdmin))).
		Patch("/admin/seller-applications/{id}", appHandler.UpdateStatus)
}
