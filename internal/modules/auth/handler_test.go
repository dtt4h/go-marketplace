package auth

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/dtt4h/go-marketplace/internal/config"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// --- Mock AuthService ---

type mockAuthService struct {
	register       func(ctx context.Context, req dtos.RegisterRequest) (dtos.AuthResponse, error)
	login          func(ctx context.Context, req dtos.LoginRequest) (dtos.AuthResponse, error)
	refresh        func(ctx context.Context, refreshToken string) (dtos.RefreshResponse, error)
	logout         func(ctx context.Context, userID int64) error
	forgotPassword func(ctx context.Context, email string) error
	resetPassword  func(ctx context.Context, token, newPassword string) error
}

func (m *mockAuthService) Register(ctx context.Context, req dtos.RegisterRequest) (dtos.AuthResponse, error) {
	return m.register(ctx, req)
}

func (m *mockAuthService) Login(ctx context.Context, req dtos.LoginRequest) (dtos.AuthResponse, error) {
	return m.login(ctx, req)
}

func (m *mockAuthService) Refresh(ctx context.Context, refreshToken string) (dtos.RefreshResponse, error) {
	return m.refresh(ctx, refreshToken)
}

func (m *mockAuthService) Logout(ctx context.Context, userID int64) error {
	return m.logout(ctx, userID)
}

func (m *mockAuthService) ForgotPassword(ctx context.Context, email string) error {
	return m.forgotPassword(ctx, email)
}

func (m *mockAuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	return m.resetPassword(ctx, token, newPassword)
}

// --- Mock AuthRepository for service tests ---

type mockAuthRepo struct {
	createUser        func(ctx context.Context, email, passwordHash string, role db.UserRole, username string) (db.User, error)
	getUserByEmail    func(ctx context.Context, email string) (db.User, error)
	getUserByID       func(ctx context.Context, id int64) (db.User, error)
	createRefreshToken func(ctx context.Context, userID int64, token string, expiresAt time.Time) (db.RefreshToken, error)
	getRefreshToken   func(ctx context.Context, token string) (db.RefreshToken, error)
	deleteRefreshToken func(ctx context.Context, token string) error
	deleteUserRefreshTokens func(ctx context.Context, userID int64) error
	getUserByResetToken func(ctx context.Context, tokenHash string) (db.User, error)
	resetPassword     func(ctx context.Context, userID int64, passwordHash string) (db.User, error)
	generateResetToken func(ctx context.Context, userID int64, expiresAt time.Time) (string, error)
	deleteResetToken  func(ctx context.Context, userID int64) error
}

func (m *mockAuthRepo) CreateUser(ctx context.Context, email, passwordHash string, role db.UserRole, username string) (db.User, error) {
	return m.createUser(ctx, email, passwordHash, role, username)
}

func (m *mockAuthRepo) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return m.getUserByEmail(ctx, email)
}

func (m *mockAuthRepo) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	return m.getUserByID(ctx, id)
}

func (m *mockAuthRepo) CreateRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) (db.RefreshToken, error) {
	return m.createRefreshToken(ctx, userID, token, expiresAt)
}

func (m *mockAuthRepo) GetRefreshToken(ctx context.Context, token string) (db.RefreshToken, error) {
	return m.getRefreshToken(ctx, token)
}

func (m *mockAuthRepo) DeleteRefreshToken(ctx context.Context, token string) error {
	return m.deleteRefreshToken(ctx, token)
}

func (m *mockAuthRepo) DeleteUserRefreshTokens(ctx context.Context, userID int64) error {
	return m.deleteUserRefreshTokens(ctx, userID)
}

func (m *mockAuthRepo) GetUserByResetToken(ctx context.Context, tokenHash string) (db.User, error) {
	return m.getUserByResetToken(ctx, tokenHash)
}

func (m *mockAuthRepo) ResetPassword(ctx context.Context, userID int64, passwordHash string) (db.User, error) {
	return m.resetPassword(ctx, userID, passwordHash)
}

func (m *mockAuthRepo) GenerateResetToken(ctx context.Context, userID int64, expiresAt time.Time) (string, error) {
	return m.generateResetToken(ctx, userID, expiresAt)
}

func (m *mockAuthRepo) DeleteResetToken(ctx context.Context, userID int64) error {
	return m.deleteResetToken(ctx, userID)
}

// --- Helpers ---

func newTestAuthHandler(svc AuthService) *AuthHandler {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret-key-for-testing-only",
			AccessTTL:   15 * time.Minute,
			RefreshTTL:  7 * 24 * time.Hour,
		},
	}
	_ = cfg
	return NewAuthHandler(svc, false)
}

func generateTestJWT(userID int64, role string, secret string, ttl time.Duration) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(ttl).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func TestHandlerRegister(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockResp   dtos.AuthResponse
		mockErr    error
		wantStatus int
	}{
		{
			name: "success",
			body: `{"email":"test@example.com","password":"password123","username":"testuser"}`,
			mockResp: dtos.AuthResponse{
				User:         dtos.UserResponse{ID: 1, Email: "test@example.com", Username: "testuser"},
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
			},
			mockErr:    nil,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing auth body",
			body:       `{invalid`,
			mockResp:   dtos.AuthResponse{},
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email",
			body:       `{"email":"notanemail","password":"password123","username":"testuser"}`,
			mockResp:   dtos.AuthResponse{},
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "weak password",
			body:       `{"email":"test@example.com","password":"short","username":"testuser"}`,
			mockResp:   dtos.AuthResponse{},
			mockErr:    ErrWeakPassword,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "email taken",
			body:       `{"email":"taken@example.com","password":"password123","username":"testuser"}`,
			mockResp:   dtos.AuthResponse{},
			mockErr:    ErrEmailTaken,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "username taken",
			body:       `{"email":"test@example.com","password":"password123","username":"taken"}`,
			mockResp:   dtos.AuthResponse{},
			mockErr:    ErrUsernameTaken,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "empty email",
			body:       `{"email":"","password":"password123","username":"testuser"}`,
			mockResp:   dtos.AuthResponse{},
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty username",
			body:       `{"email":"test@example.com","password":"password123","username":""}`,
			mockResp:   dtos.AuthResponse{},
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{
				register: func(ctx context.Context, req dtos.RegisterRequest) (dtos.AuthResponse, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			h := newTestAuthHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Register(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			// Verify refresh token cookie is set on success
			if tt.wantStatus == http.StatusCreated {
				cookie := w.Header().Get("Set-Cookie")
				if !strings.Contains(cookie, "refresh_token") {
					t.Errorf("expected refresh_token cookie, got: %s", cookie)
				}
			}
		})
	}
}

func TestHandlerLogin(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockResp   dtos.AuthResponse
		mockErr    error
		wantStatus int
	}{
		{
			name: "success",
			body: `{"email":"test@example.com","password":"password123"}`,
			mockResp: dtos.AuthResponse{
				User:         dtos.UserResponse{ID: 1, Email: "test@example.com"},
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid credentials",
			body:       `{"email":"test@example.com","password":"wrong"}`,
			mockResp:   dtos.AuthResponse{},
			mockErr:    ErrInvalidCredentials,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid json",
			body:       `{invalid`,
			mockResp:   dtos.AuthResponse{},
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{
				login: func(ctx context.Context, req dtos.LoginRequest) (dtos.AuthResponse, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			h := newTestAuthHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Login(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestHandlerRefresh(t *testing.T) {
	tests := []struct {
		name       string
		cookie     string
		mockResp   dtos.RefreshResponse
		mockErr    error
		wantStatus int
	}{
		{
			name: "success",
			cookie: "refresh_token=valid_token",
			mockResp: dtos.RefreshResponse{
				AccessToken:  "new_access",
				RefreshToken: "new_refresh",
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing cookie",
			cookie:     "",
			mockResp:   dtos.RefreshResponse{},
			mockErr:    nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			cookie:     "refresh_token=invalid",
			mockResp:   dtos.RefreshResponse{},
			mockErr:    ErrInvalidRefreshToken,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "expired token",
			cookie:     "refresh_token=expired",
			mockResp:   dtos.RefreshResponse{},
			mockErr:    ErrRefreshTokenExpired,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{
				refresh: func(ctx context.Context, refreshToken string) (dtos.RefreshResponse, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			h := newTestAuthHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
			if tt.cookie != "" {
				req.Header.Set("Cookie", tt.cookie)
			}
			w := httptest.NewRecorder()

			h.Refresh(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestHandlerLogout(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		mockErr    error
		wantStatus int
	}{
		{
			name:       "success",
			authHeader: "Bearer " + generateTestJWT(1, "buyer", "test-secret-key-for-testing-only", 15*time.Minute),
			mockErr:    nil,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "missing auth",
			authHeader: "",
			mockErr:    nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "service error",
			authHeader: "Bearer " + generateTestJWT(1, "buyer", "test-secret-key-for-testing-only", 15*time.Minute),
			mockErr:    fmt.Errorf("db error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{
				logout: func(ctx context.Context, userID int64) error {
					return tt.mockErr
				},
			}
			h := newTestAuthHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			// Use chi router with JWT auth middleware to extract userID
			r := chi.NewRouter()
			r.Use(mw.JWTAuth(&config.Config{
				JWT: config.JWTConfig{
					Secret:     "test-secret-key-for-testing-only",
					AccessTTL:  15 * time.Minute,
					RefreshTTL: 7 * 24 * time.Hour,
				},
			}))
			r.Post("/auth/logout", http.HandlerFunc(h.Logout))

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestHandlerForgotPassword(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockErr    error
		wantStatus int
	}{
		{
			name:       "success",
			body:       `{"email":"test@example.com"}`,
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty email",
			body:       `{}`,
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			body:       `{invalid`,
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{
				forgotPassword: func(ctx context.Context, email string) error {
					return tt.mockErr
				},
			}
			h := newTestAuthHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.ForgotPassword(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestHandlerResetPassword(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockErr    error
		wantStatus int
	}{
		{
			name:       "success",
			body:       `{"token":"reset_token","password":"newpassword123"}`,
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing token",
			body:       `{"password":"newpassword123"}`,
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing password",
			body:       `{"token":"reset_token"}`,
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "weak password",
			body:       `{"token":"reset_token","password":"short"}`,
			mockErr:    ErrWeakPassword,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid token",
			body:       `{"token":"invalid","password":"newpassword123"}`,
			mockErr:    fmt.Errorf("invalid token"),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{
				resetPassword: func(ctx context.Context, token, newPassword string) error {
					return tt.mockErr
				},
			}
			h := newTestAuthHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.ResetPassword(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// --- Service tests with mock repo ---

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name       string
		req        dtos.RegisterRequest
		mockUser   db.User
		mockErr    error
		wantStatus int
	}{
		{
			name: "success",
			req: dtos.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Username: "testuser",
			},
			mockUser: db.User{
				ID:       1,
				Email:    "test@example.com",
				Username: "testuser",
				Role:     db.UserRoleBuyer,
			},
			mockErr:    nil,
			wantStatus: http.StatusCreated,
		},
		{
			name: "empty email",
			req: dtos.RegisterRequest{
				Email:    "",
				Password: "password123",
				Username: "testuser",
			},
			mockUser:   db.User{},
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty username",
			req: dtos.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Username: "",
			},
			mockUser:   db.User{},
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "weak password",
			req: dtos.RegisterRequest{
				Email:    "test@example.com",
				Password: "short",
				Username: "testuser",
			},
			mockUser:   db.User{},
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuthRepo{
				createUser: func(ctx context.Context, email, passwordHash string, role db.UserRole, username string) (db.User, error) {
					return tt.mockUser, tt.mockErr
				},
				createRefreshToken: func(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (db.RefreshToken, error) {
					return db.RefreshToken{}, nil
				},
				deleteUserRefreshTokens: func(ctx context.Context, userID int64) error {
					return nil
				},
			}

			cfg := &config.Config{
				JWT: config.JWTConfig{
					Secret:     "test-secret",
					AccessTTL:  15 * time.Minute,
					RefreshTTL: 7 * 24 * time.Hour,
				},
			}

			svc := NewAuthService(repo, cfg)
			_, err := svc.Register(context.Background(), tt.req)

			if tt.wantStatus == http.StatusBadRequest {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name       string
		req        dtos.LoginRequest
		mockUser   db.User
		mockErr    error
		wantStatus int
	}{
		{
			name: "success",
			req: dtos.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockUser: db.User{
				ID:           1,
				Email:        "test@example.com",
				PasswordHash: "$2a$12$fpqiylNA81NMTskU1VvXUu.8YT7svM64X538M03iRGs0uODVRlYlC",
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name: "user not found",
			req: dtos.LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			mockUser:   db.User{},
			mockErr:    fmt.Errorf("not found"),
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuthRepo{
				getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
					return tt.mockUser, tt.mockErr
				},
				createRefreshToken: func(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (db.RefreshToken, error) {
					return db.RefreshToken{}, nil
				},
				deleteUserRefreshTokens: func(ctx context.Context, userID int64) error {
					return nil
				},
			}

			cfg := &config.Config{
				JWT: config.JWTConfig{
					Secret:     "test-secret",
					AccessTTL:  15 * time.Minute,
					RefreshTTL: 7 * 24 * time.Hour,
				},
			}

			svc := NewAuthService(repo, cfg)
			_, err := svc.Login(context.Background(), tt.req)

			if tt.wantStatus == http.StatusUnauthorized {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAuthService_Refresh(t *testing.T) {
	tests := []struct {
		name          string
		refreshToken  string
		mockToken     db.RefreshToken
		mockUser      db.User
		tokenErr      error
		userErr       error
		wantStatus    int
	}{
		{
			name:         "success",
			refreshToken: "valid_token",
			mockToken:    db.RefreshToken{ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(1 * time.Hour), Valid: true}},
			mockUser:     db.User{ID: 1, Email: "test@example.com", Role: db.UserRoleBuyer},
			tokenErr:     nil,
			userErr:      nil,
			wantStatus:   http.StatusOK,
		},
		{
			name:         "token not found",
			refreshToken: "invalid",
			mockToken:    db.RefreshToken{},
			mockUser:     db.User{},
			tokenErr:     fmt.Errorf("not found"),
			userErr:      nil,
			wantStatus:   http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuthRepo{
				getRefreshToken: func(ctx context.Context, tokenHash string) (db.RefreshToken, error) {
					return tt.mockToken, tt.tokenErr
				},
				getUserByID: func(ctx context.Context, id int64) (db.User, error) {
					return tt.mockUser, tt.userErr
				},
				deleteRefreshToken: func(ctx context.Context, tokenHash string) error {
					return nil
				},
				createRefreshToken: func(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (db.RefreshToken, error) {
					return db.RefreshToken{}, nil
				},
			}

			cfg := &config.Config{
				JWT: config.JWTConfig{
					Secret:     "test-secret",
					AccessTTL:  15 * time.Minute,
					RefreshTTL: 7 * 24 * time.Hour,
				},
			}

			svc := NewAuthService(repo, cfg)
			_, err := svc.Refresh(context.Background(), tt.refreshToken)

			if tt.wantStatus == http.StatusUnauthorized {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	repo := &mockAuthRepo{
		deleteUserRefreshTokens: func(ctx context.Context, userID int64) error {
			return nil
		},
	}

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 7 * 24 * time.Hour,
		},
	}

	svc := NewAuthService(repo, cfg)
	err := svc.Logout(context.Background(), 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAuthService_ForgotPassword(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		mockUser   *db.User
		mockErr    error
		wantStatus int
	}{
		{
			name:  "user exists",
			email: "test@example.com",
			mockUser: &db.User{
				ID:    1,
				Email: "test@example.com",
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "user not found - should not error",
			email:      "notfound@example.com",
			mockUser:   nil,
			mockErr:    fmt.Errorf("not found"),
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuthRepo{
				getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
					if tt.mockUser == nil {
						return db.User{}, fmt.Errorf("not found")
					}
					return *tt.mockUser, nil
				},
				generateResetToken: func(ctx context.Context, userID int64, expiresAt time.Time) (string, error) {
					return "", tt.mockErr
				},
			}

			cfg := &config.Config{
				JWT: config.JWTConfig{
					Secret:     "test-secret",
					AccessTTL:  15 * time.Minute,
					RefreshTTL: 7 * 24 * time.Hour,
				},
			}

			svc := NewAuthService(repo, cfg)
			err := svc.ForgotPassword(context.Background(), tt.email)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestAuthService_ResetPassword(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		newPassword string
		mockUser    *db.User
		mockErr     error
		wantStatus  int
	}{
		{
			name:        "success",
			token:       "valid_token",
			newPassword: "newpassword123",
			mockUser: &db.User{
				ID: 1,
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:        "weak password",
			token:       "valid_token",
			newPassword: "short",
			mockUser:    nil,
			mockErr:     nil,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "invalid token",
			token:       "invalid",
			newPassword: "newpassword123",
			mockUser:    nil,
			mockErr:     fmt.Errorf("not found"),
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuthRepo{
				getUserByResetToken: func(ctx context.Context, tokenHash string) (db.User, error) {
					if tt.mockUser == nil {
						return db.User{}, fmt.Errorf("not found")
					}
					return *tt.mockUser, nil
				},
				resetPassword: func(ctx context.Context, userID int64, passwordHash string) (db.User, error) {
					return db.User{}, tt.mockErr
				},
				deleteUserRefreshTokens: func(ctx context.Context, userID int64) error {
					return nil
				},
			}

			cfg := &config.Config{
				JWT: config.JWTConfig{
					Secret:     "test-secret",
					AccessTTL:  15 * time.Minute,
					RefreshTTL: 7 * 24 * time.Hour,
				},
			}

			svc := NewAuthService(repo, cfg)
			err := svc.ResetPassword(context.Background(), tt.token, tt.newPassword)

			if tt.wantStatus == http.StatusBadRequest {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
