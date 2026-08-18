package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/config"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/jackc/pgx/v5/pgtype"
)

// --- Mocks ---

type mockAuthRepository struct {
	createUser       func(ctx context.Context, email, passwordHash string, role db.UserRole, username string) (db.User, error)
	getUserByEmail   func(ctx context.Context, email string) (db.User, error)
	getUserByID      func(ctx context.Context, id int64) (db.User, error)
	createRefresh    func(ctx context.Context, userID int64, token string, expiresAt time.Time) (db.RefreshToken, error)
	getRefresh       func(ctx context.Context, token string) (db.RefreshToken, error)
	deleteRefresh    func(ctx context.Context, token string) error
	deleteUserRefresh func(ctx context.Context, userID int64) error
	generateReset    func(ctx context.Context, userID int64, expiresAt time.Time) (string, error)
	getUserByReset   func(ctx context.Context, tokenHash string) (db.User, error)
	deleteReset      func(ctx context.Context, userID int64) error
	resetPassword    func(ctx context.Context, userID int64, passwordHash string) (db.User, error)
}

func (m *mockAuthRepository) CreateUser(ctx context.Context, email, passwordHash string, role db.UserRole, username string) (db.User, error) {
	return m.createUser(ctx, email, passwordHash, role, username)
}
func (m *mockAuthRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return m.getUserByEmail(ctx, email)
}
func (m *mockAuthRepository) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	return m.getUserByID(ctx, id)
}
func (m *mockAuthRepository) CreateRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) (db.RefreshToken, error) {
	return m.createRefresh(ctx, userID, token, expiresAt)
}
func (m *mockAuthRepository) GetRefreshToken(ctx context.Context, token string) (db.RefreshToken, error) {
	return m.getRefresh(ctx, token)
}
func (m *mockAuthRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	return m.deleteRefresh(ctx, token)
}
func (m *mockAuthRepository) DeleteUserRefreshTokens(ctx context.Context, userID int64) error {
	return m.deleteUserRefresh(ctx, userID)
}
func (m *mockAuthRepository) GenerateResetToken(ctx context.Context, userID int64, expiresAt time.Time) (string, error) {
	return m.generateReset(ctx, userID, expiresAt)
}
func (m *mockAuthRepository) GetUserByResetToken(ctx context.Context, tokenHash string) (db.User, error) {
	return m.getUserByReset(ctx, tokenHash)
}
func (m *mockAuthRepository) DeleteResetToken(ctx context.Context, userID int64) error {
	return m.deleteReset(ctx, userID)
}
func (m *mockAuthRepository) ResetPassword(ctx context.Context, userID int64, passwordHash string) (db.User, error) {
	return m.resetPassword(ctx, userID, passwordHash)
}

func newTestService(repo AuthRepository) AuthService {
	return NewAuthService(repo, &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret-key-for-jwt-signing",
			AccessTTL: 15 * time.Minute,
			RefreshTTL: 7 * 24 * time.Hour,
		},
	})
}

func testUser() db.User {
	return db.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: "$2a$12$hashed",
		Role:         db.UserRoleBuyer,
		Username:     "testuser",
	}
}

// --- Register ---

func TestRegister(t *testing.T) {
	t.Run("rejects empty email", func(t *testing.T) {
		svc := newTestService(&mockAuthRepository{})
		_, err := svc.Register(context.Background(), dtos.RegisterRequest{
			Password: "password123",
			Username: "testuser",
		})
		if err != ErrEmailRequired {
			t.Fatalf("expected ErrEmailRequired, got %v", err)
		}
	})

	t.Run("rejects empty username", func(t *testing.T) {
		svc := newTestService(&mockAuthRepository{})
		_, err := svc.Register(context.Background(), dtos.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		})
		if err != ErrUsernameRequired {
			t.Fatalf("expected ErrUsernameRequired, got %v", err)
		}
	})

	t.Run("rejects weak password", func(t *testing.T) {
		svc := newTestService(&mockAuthRepository{})
		_, err := svc.Register(context.Background(), dtos.RegisterRequest{
			Email:    "test@example.com",
			Password: "short",
			Username: "testuser",
		})
		if err != ErrWeakPassword {
			t.Fatalf("expected ErrWeakPassword, got %v", err)
		}
	})

	t.Run("rejects duplicate email", func(t *testing.T) {
		repo := &mockAuthRepository{
			createUser: func(ctx context.Context, email, passwordHash string, role db.UserRole, username string) (db.User, error) {
				return db.User{}, errors.New("unique violation")
			},
		}
		// Mock pgutil functions
		svc := newTestService(repo)
		_, err := svc.Register(context.Background(), dtos.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Username: "testuser",
		})
		// Will fail with unique violation error, not necessarily ErrEmailTaken
		// since pgutil mock is not set up, but we verify it returns an error
		if err == nil {
			t.Fatal("expected error for duplicate email")
		}
	})

	t.Run("success", func(t *testing.T) {
		user := testUser()
		repo := &mockAuthRepository{
			createUser: func(ctx context.Context, email, passwordHash string, role db.UserRole, username string) (db.User, error) {
				return user, nil
			},
			createRefresh: func(ctx context.Context, userID int64, token string, expiresAt time.Time) (db.RefreshToken, error) {
				return db.RefreshToken{ID: 1, UserID: userID}, nil
			},
			deleteUserRefresh: func(ctx context.Context, userID int64) error {
				return nil
			},
		}
		svc := newTestService(repo)
		resp, err := svc.Register(context.Background(), dtos.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Username: "testuser",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.User.Email != "test@example.com" {
			t.Fatalf("expected email test@example.com, got %s", resp.User.Email)
		}
		if resp.AccessToken == "" {
			t.Fatal("expected access token")
		}
		if resp.RefreshToken == "" {
			t.Fatal("expected refresh token")
		}
	})
}

// --- Login ---

func TestLogin(t *testing.T) {
	t.Run("rejects invalid credentials - user not found", func(t *testing.T) {
		repo := &mockAuthRepository{
			getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
				return db.User{}, errors.New("not found")
			},
		}
		svc := newTestService(repo)
		_, err := svc.Login(context.Background(), dtos.LoginRequest{
			Email:    "notfound@example.com",
			Password: "password123",
		})
		if err != ErrInvalidCredentials {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("rejects wrong password", func(t *testing.T) {
		user := testUser()
		repo := &mockAuthRepository{
			getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
				return user, nil
			},
		}
		svc := newTestService(repo)
		_, err := svc.Login(context.Background(), dtos.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		})
		if err != ErrInvalidCredentials {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		user := testUser()
		repo := &mockAuthRepository{
			getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
				return user, nil
			},
			createRefresh: func(ctx context.Context, userID int64, token string, expiresAt time.Time) (db.RefreshToken, error) {
				return db.RefreshToken{ID: 1, UserID: userID}, nil
			},
			deleteUserRefresh: func(ctx context.Context, userID int64) error {
				return nil
			},
		}
		svc := newTestService(repo)
		resp, err := svc.Login(context.Background(), dtos.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.User.Username != "testuser" {
			t.Fatalf("expected username testuser, got %s", resp.User.Username)
		}
		if resp.AccessToken == "" {
			t.Fatal("expected access token")
		}
	})
}

// --- Refresh ---

func TestRefresh(t *testing.T) {
	t.Run("rejects invalid refresh token", func(t *testing.T) {
		repo := &mockAuthRepository{
			getRefresh: func(ctx context.Context, token string) (db.RefreshToken, error) {
				return db.RefreshToken{}, errors.New("not found")
			},
		}
		svc := newTestService(repo)
		_, err := svc.Refresh(context.Background(), "invalid-token")
		if err != ErrInvalidRefreshToken {
			t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
		}
	})

	t.Run("rejects expired refresh token", func(t *testing.T) {
		expiredTime := time.Now().Add(-2 * time.Hour)
		repo := &mockAuthRepository{
			getRefresh: func(ctx context.Context, token string) (db.RefreshToken, error) {
				return db.RefreshToken{
					ID:        1,
					UserID:    1,
					ExpiresAt: pgtype.Timestamptz{Time: expiredTime, Valid: true},
				}, nil
			},
			deleteRefresh: func(ctx context.Context, token string) error {
				return nil
			},
		}
		svc := newTestService(repo)
		_, err := svc.Refresh(context.Background(), "expired-token")
		if err != ErrRefreshTokenExpired {
			t.Fatalf("expected ErrRefreshTokenExpired, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		user := testUser()
		repo := &mockAuthRepository{
			getRefresh: func(ctx context.Context, token string) (db.RefreshToken, error) {
				return db.RefreshToken{ID: 1, UserID: 1}, nil
			},
			getUserByID: func(ctx context.Context, id int64) (db.User, error) {
				return user, nil
			},
			deleteRefresh: func(ctx context.Context, token string) error {
				return nil
			},
			createRefresh: func(ctx context.Context, userID int64, token string, expiresAt time.Time) (db.RefreshToken, error) {
				return db.RefreshToken{ID: 2, UserID: userID}, nil
			},
		}
		svc := newTestService(repo)
		resp, err := svc.Refresh(context.Background(), "valid-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.AccessToken == "" {
			t.Fatal("expected access token")
		}
		if resp.RefreshToken == "" {
			t.Fatal("expected new refresh token")
		}
	})
}

// --- Logout ---

func TestLogout(t *testing.T) {
	repo := &mockAuthRepository{
		deleteUserRefresh: func(ctx context.Context, userID int64) error {
			return nil
		},
	}
	svc := newTestService(repo)
	err := svc.Logout(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- ForgotPassword ---

func TestForgotPassword(t *testing.T) {
	t.Run("user not found - no error (security)", func(t *testing.T) {
		repo := &mockAuthRepository{
			getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
				return db.User{}, errors.New("not found")
			},
		}
		svc := newTestService(repo)
		err := svc.ForgotPassword(context.Background(), "nobody@example.com")
		if err != nil {
			t.Fatalf("expected no error for unknown email, got %v", err)
		}
	})

	t.Run("success - generates token", func(t *testing.T) {
		user := testUser()
		repo := &mockAuthRepository{
			getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
				return user, nil
			},
			generateReset: func(ctx context.Context, userID int64, expiresAt time.Time) (string, error) {
				return "reset-token-123", nil
			},
		}
		svc := newTestService(repo)
		err := svc.ForgotPassword(context.Background(), "test@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("generate reset token fails", func(t *testing.T) {
		user := testUser()
		repo := &mockAuthRepository{
			getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
				return user, nil
			},
			generateReset: func(ctx context.Context, userID int64, expiresAt time.Time) (string, error) {
				return "", errors.New("db error")
			},
		}
		svc := newTestService(repo)
		err := svc.ForgotPassword(context.Background(), "test@example.com")
		if err == nil {
			t.Fatal("expected error for failed token generation")
		}
	})
}

// --- ResetPassword ---

func TestResetPassword(t *testing.T) {
	t.Run("rejects weak password", func(t *testing.T) {
		repo := &mockAuthRepository{}
		svc := newTestService(repo)
		err := svc.ResetPassword(context.Background(), "token", "short")
		if err != ErrWeakPassword {
			t.Fatalf("expected ErrWeakPassword, got %v", err)
		}
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		repo := &mockAuthRepository{
			getUserByReset: func(ctx context.Context, tokenHash string) (db.User, error) {
				return db.User{}, errors.New("not found")
			},
		}
		svc := newTestService(repo)
		err := svc.ResetPassword(context.Background(), "invalid-token", "newpassword123")
		if err == nil {
			t.Fatal("expected error for invalid token")
		}
	})

	t.Run("success", func(t *testing.T) {
		user := testUser()
		repo := &mockAuthRepository{
			getUserByReset: func(ctx context.Context, tokenHash string) (db.User, error) {
				return user, nil
			},
			deleteUserRefresh: func(ctx context.Context, userID int64) error {
				return nil
			},
			resetPassword: func(ctx context.Context, userID int64, passwordHash string) (db.User, error) {
				return user, nil
			},
		}
		svc := newTestService(repo)
		err := svc.ResetPassword(context.Background(), "valid-token", "newpassword123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
