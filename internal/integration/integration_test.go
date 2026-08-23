//go:build integration

package integration

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/dtt4h/go-marketplace/internal/config"
	"github.com/dtt4h/go-marketplace/internal/database"
	"github.com/dtt4h/go-marketplace/internal/modules/auth"
	"github.com/dtt4h/go-marketplace/internal/modules/cart"
	"github.com/dtt4h/go-marketplace/internal/modules/products"
	"github.com/dtt4h/go-marketplace/internal/modules/users"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testEnv holds test containers and dependencies
type testEnv struct {
	ctx context.Context
	db  *postgres.PostgresContainer
	dsn string
	log *slog.Logger
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	ctx := context.Background()

	dbC, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}

	host, err := dbC.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get db host: %v", err)
	}
	port, err := dbC.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get db port: %v", err)
	}
	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/testdb?sslmode=disable", host, port.Port())

	// Run migrations from migrations/ directory
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		t.Fatalf("failed to create migrator: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to run migrations: %v", err)
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	t.Cleanup(func() {
		dbC.Terminate(ctx)
	})

	return &testEnv{
		ctx: ctx,
		db:  dbC,
		dsn: dsn,
		log: log,
	}
}

// --- Auth integration test ---

func TestAuthIntegration(t *testing.T) {
	env := setupTestEnv(t)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key-for-integration-tests",
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 7 * 24 * time.Hour,
		},
	}

	dbPool, err := database.New(env.ctx, env.dsn, env.log)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer dbPool.Close()

	repo := auth.NewAuthRepository(db.New(dbPool))
	svc := auth.NewAuthService(repo, cfg)

	t.Run("register user", func(t *testing.T) {
		_, err := svc.Register(env.ctx, dtos.RegisterRequest{
			Email:    "int_test@example.com",
			Password: "password123",
			Username: "inttestuser",
		})
		if err != nil {
			t.Fatalf("register failed: %v", err)
		}
	})

	t.Run("register duplicate email", func(t *testing.T) {
		_, err := svc.Register(env.ctx, dtos.RegisterRequest{
			Email:    "int_test@example.com",
			Password: "password123",
			Username: "anotheruser",
		})
		if err == nil {
			t.Fatal("expected error for duplicate email")
		}
	})

	t.Run("login", func(t *testing.T) {
		_, err := svc.Login(env.ctx, dtos.LoginRequest{
			Email:    "int_test@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("login failed: %v", err)
		}
	})

	t.Run("login wrong password", func(t *testing.T) {
		_, err := svc.Login(env.ctx, dtos.LoginRequest{
			Email:    "int_test@example.com",
			Password: "wrongpassword",
		})
		if err == nil {
			t.Fatal("expected error for wrong password")
		}
	})
}

// --- Product integration test ---

func TestProductIntegration(t *testing.T) {
	env := setupTestEnv(t)

	dbPool, err := database.New(env.ctx, env.dsn, env.log)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer dbPool.Close()

	// Create a user first via auth service
	authRepo := auth.NewAuthRepository(db.New(dbPool))
	authSvc := auth.NewAuthService(authRepo, &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 7 * 24 * time.Hour,
		},
	})
	_, err = authSvc.Register(env.ctx, dtos.RegisterRequest{
		Email:    "store@example.com",
		Password: "password123",
		Username: "storeowner",
	})
	if err != nil {
		t.Fatalf("register user failed: %v", err)
	}

	// Get user
	userRepo := users.NewUserRepository(db.New(dbPool), dbPool)
	storeOwner, err := userRepo.GetUserByID(env.ctx, 1)
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}

	// Create a store
	_, err = userRepo.CreateStoreWithRole(env.ctx, storeOwner.ID, "Test Store", ptrString("Test description"), nil)
	if err != nil {
		t.Fatalf("create store failed: %v", err)
	}

	productRepo := products.NewProductRepository(db.New(dbPool), dbPool)
	productSvc := products.NewProductService(productRepo, nil, nil)

	t.Run("create product", func(t *testing.T) {
		_, err := productSvc.CreateProduct(env.ctx, storeOwner.ID, dtos.CreateProductRequest{
			Title:    "Integration Test Product",
			Price:    "999",
			Stock:    10,
			
		})
		if err != nil {
			t.Fatalf("create product failed: %v", err)
		}
	})

	t.Run("list products", func(t *testing.T) {
		items, total, err := productSvc.ListProducts(env.ctx, nil, nil, nil, nil, nil, "created_desc", 1, 20)
		if err != nil {
			t.Fatalf("list products failed: %v", err)
		}
		if total == 0 {
			t.Fatal("expected at least 1 product")
		}
		if len(items) == 0 {
			t.Fatal("expected at least 1 item in response")
		}
	})

	t.Run("get product", func(t *testing.T) {
		// Get the product we just created
		products, _, _ := productSvc.ListProducts(env.ctx, nil, nil, nil, nil, nil, "created_desc", 1, 1)
		if len(products) == 0 {
			t.Fatal("no products found")
		}
		_, err := productSvc.GetProduct(env.ctx, products[0].ID)
		if err != nil {
			t.Fatalf("get product failed: %v", err)
		}
	})
}

// --- Cart integration test ---

func TestCartIntegration(t *testing.T) {
	env := setupTestEnv(t)

	dbPool, err := database.New(env.ctx, env.dsn, env.log)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer dbPool.Close()

	// Create user via auth
	authRepo := auth.NewAuthRepository(db.New(dbPool))
	authSvc := auth.NewAuthService(authRepo, &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 7 * 24 * time.Hour,
		},
	})
	_, err = authSvc.Register(env.ctx, dtos.RegisterRequest{
		Email:    "cartuser@example.com",
		Password: "password123",
		Username: "cartuser",
	})
	if err != nil {
		t.Fatalf("register user failed: %v", err)
	}

	userRepo := users.NewUserRepository(db.New(dbPool), dbPool)
	cartUser, err := userRepo.GetUserByID(env.ctx, 1)
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}

	// Create store
	_, err = userRepo.CreateStoreWithRole(env.ctx, cartUser.ID, "Cart Store", ptrString("Test"), nil)
	if err != nil {
		t.Fatalf("create store failed: %v", err)
	}

	productRepo := products.NewProductRepository(db.New(dbPool), dbPool)
	productSvc := products.NewProductService(productRepo, nil, nil)
	prod, err := productSvc.CreateProduct(env.ctx, cartUser.ID, dtos.CreateProductRequest{
		Title:    "Cart Test Product",
		Price:    "500",
		Stock:    100,
		
	})
	if err != nil {
		t.Fatalf("create product failed: %v", err)
	}

	cartRepo := cart.NewCartRepository(db.New(dbPool))
	cartSvc := cart.NewCartService(cartRepo)

	t.Run("add item to cart", func(t *testing.T) {
		_, err := cartSvc.AddItem(env.ctx, cartUser.ID, dtos.AddCartItemRequest{
			ProductID: prod.ID,
			Quantity:  2,
		})
		if err != nil {
			t.Fatalf("add item failed: %v", err)
		}
	})

	t.Run("get cart", func(t *testing.T) {
		cartResp, err := cartSvc.GetCart(env.ctx, cartUser.ID)
		if err != nil {
			t.Fatalf("get cart failed: %v", err)
		}
		if len(cartResp.Items) == 0 {
			t.Fatal("expected cart items")
		}
	})

	t.Run("update cart item", func(t *testing.T) {
		cartResp, _ := cartSvc.GetCart(env.ctx, cartUser.ID)
		if len(cartResp.Items) == 0 {
			t.Fatal("no cart items to update")
		}
		_, err := cartSvc.UpdateItem(env.ctx, cartUser.ID, cartResp.Items[0].ID, dtos.UpdateCartItemRequest{
			Quantity: 5,
		})
		if err != nil {
			t.Fatalf("update cart item failed: %v", err)
		}
	})

	t.Run("remove cart item", func(t *testing.T) {
		cartResp, _ := cartSvc.GetCart(env.ctx, cartUser.ID)
		if len(cartResp.Items) == 0 {
			t.Fatal("no cart items to remove")
		}
		_, err := cartSvc.RemoveItem(env.ctx, cartUser.ID, cartResp.Items[0].ID)
		if err != nil {
			t.Fatalf("remove cart item failed: %v", err)
		}
	})

	t.Run("clear cart", func(t *testing.T) {
		// Add item first
		_, _ = cartSvc.AddItem(env.ctx, cartUser.ID, dtos.AddCartItemRequest{
			ProductID: prod.ID,
			Quantity:  1,
		})
		_, err := cartSvc.Clear(env.ctx, cartUser.ID)
		if err != nil {
			t.Fatalf("clear cart failed: %v", err)
		}
	})
}

// --- Helpers ---

func ptrString(s string) *string {
	return &s
}
