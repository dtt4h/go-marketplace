package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/dtt4h/go-marketplace/internal/config"
	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
)

var (
	ErrEmailTaken          = errors.New("email already taken")
	ErrUsernameTaken       = errors.New("username already taken")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrWeakPassword        = errors.New("password must be at least 8 characters")
	ErrEmailRequired       = errors.New("email is required")
	ErrUsernameRequired    = errors.New("username is required")
)

// AuthService defines business logic for authentication.
type AuthService interface {
	Register(ctx context.Context, req dtos.RegisterRequest) (dtos.AuthResponse, error)
	Login(ctx context.Context, req dtos.LoginRequest) (dtos.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (dtos.RefreshResponse, error)
	Logout(ctx context.Context, userID int64) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type authService struct {
	repo AuthRepository
	cfg  *config.Config
}

// NewAuthService creates a new AuthService.
func NewAuthService(repo AuthRepository, cfg *config.Config) AuthService {
	return &authService{repo: repo, cfg: cfg}
}

func (s *authService) Register(ctx context.Context, req dtos.RegisterRequest) (dtos.AuthResponse, error) {
	if strings.TrimSpace(req.Email) == "" {
		return dtos.AuthResponse{}, ErrEmailRequired
	}
	if strings.TrimSpace(req.Username) == "" {
		return dtos.AuthResponse{}, ErrUsernameRequired
	}
	if len(req.Password) < 8 {
		return dtos.AuthResponse{}, ErrWeakPassword
	}

	role := db.UserRoleBuyer

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return dtos.AuthResponse{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, req.Email, string(hashedPassword), role, req.Username)
	if err != nil {
		if pgutil.IsUniqueViolation(err) {
			switch pgutil.UniqueViolationConstraint(err) {
			case "users_username_key":
				return dtos.AuthResponse{}, ErrUsernameTaken
			default:
				return dtos.AuthResponse{}, ErrEmailTaken
			}
		}
		return dtos.AuthResponse{}, fmt.Errorf("create user: %w", err)
	}

	accessToken, refreshToken, err := s.generateTokens(user.ID, string(user.Role))
	if err != nil {
		return dtos.AuthResponse{}, fmt.Errorf("generate tokens: %w", err)
	}

	if err := s.repo.DeleteUserRefreshTokens(ctx, user.ID); err != nil {
		return dtos.AuthResponse{}, fmt.Errorf("delete old refresh tokens: %w", err)
	}

	_, err = s.repo.CreateRefreshToken(ctx, user.ID, hashToken(refreshToken), time.Now().Add(s.cfg.JWT.RefreshTTL))
	if err != nil {
		return dtos.AuthResponse{}, fmt.Errorf("save refresh token: %w", err)
	}

	return dtos.AuthResponse{
		User:         dtos.ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Login(ctx context.Context, req dtos.LoginRequest) (dtos.AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.AuthResponse{}, ErrInvalidCredentials
		}
		return dtos.AuthResponse{}, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return dtos.AuthResponse{}, ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.generateTokens(user.ID, string(user.Role))
	if err != nil {
		return dtos.AuthResponse{}, fmt.Errorf("generate tokens: %w", err)
	}

	if err := s.repo.DeleteUserRefreshTokens(ctx, user.ID); err != nil {
		return dtos.AuthResponse{}, fmt.Errorf("delete old refresh tokens: %w", err)
	}

	_, err = s.repo.CreateRefreshToken(ctx, user.ID, hashToken(refreshToken), time.Now().Add(s.cfg.JWT.RefreshTTL))
	if err != nil {
		return dtos.AuthResponse{}, fmt.Errorf("save refresh token: %w", err)
	}

	return dtos.AuthResponse{
		User:         dtos.ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (dtos.RefreshResponse, error) {
	tokenHash := hashToken(refreshToken)

	token, err := s.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.RefreshResponse{}, ErrInvalidRefreshToken
		}
		return dtos.RefreshResponse{}, fmt.Errorf("get refresh token: %w", err)
	}

	if time.Now().After(token.ExpiresAt.Time) {
		_ = s.repo.DeleteRefreshToken(ctx, tokenHash)
		return dtos.RefreshResponse{}, ErrRefreshTokenExpired
	}

	user, err := s.repo.GetUserByID(ctx, token.UserID)
	if err != nil {
		return dtos.RefreshResponse{}, fmt.Errorf("find user: %w", err)
	}

	if err := s.repo.DeleteRefreshToken(ctx, tokenHash); err != nil {
		return dtos.RefreshResponse{}, fmt.Errorf("delete old refresh token: %w", err)
	}

	accessToken, newRefreshToken, err := s.generateTokens(user.ID, string(user.Role))
	if err != nil {
		return dtos.RefreshResponse{}, fmt.Errorf("generate tokens: %w", err)
	}

	_, err = s.repo.CreateRefreshToken(ctx, user.ID, hashToken(newRefreshToken), time.Now().Add(s.cfg.JWT.RefreshTTL))
	if err != nil {
		return dtos.RefreshResponse{}, fmt.Errorf("save new refresh token: %w", err)
	}

	return dtos.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) Logout(ctx context.Context, userID int64) error {
	return s.repo.DeleteUserRefreshTokens(ctx, userID)
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = s.repo.GenerateResetToken(ctx, user.ID, expiresAt)
	if err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrWeakPassword
	}

	tokenHash := hashToken(token)

	user, err := s.repo.GetUserByResetToken(ctx, tokenHash)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return errors.New("invalid or expired reset token")
		}
		return fmt.Errorf("get user by reset token: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	_, err = s.repo.ResetPassword(ctx, user.ID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}

	_ = s.repo.DeleteUserRefreshTokens(ctx, user.ID)

	return nil
}

func (s *authService) generateTokens(userID int64, role string) (accessToken, refreshToken string, err error) {
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(s.cfg.JWT.AccessTTL).Unix(),
		"iat":     time.Now().Unix(),
	}
	accessTok := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTok.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", "", fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(s.cfg.JWT.RefreshTTL).Unix(),
		"iat":     time.Now().Unix(),
	}
	refreshTok := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTok.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", "", fmt.Errorf("sign refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

