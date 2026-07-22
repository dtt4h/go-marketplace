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
	"github.com/jackc/pgx/v5"
)

var (
	ErrEmailTaken          = errors.New("email already taken")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrWeakPassword        = errors.New("password must be at least 8 characters")
)

type AuthService interface {
	Register(ctx context.Context, req dtos.RegisterFormRequest) (dtos.AuthResponse, error)
	Login(ctx context.Context, req dtos.LoginFormRequest) (dtos.AuthResponse, error)
	Refresh(ctx context.Context, req dtos.RefreshFormRequest) (dtos.RefreshResponse, error)
	Logout(ctx context.Context, userID int64) error
}

type authService struct {
	repo AuthRepository
	cfg  *config.Config
}

func NewAuthService(repo AuthRepository, cfg *config.Config) *authService {
	return &authService{repo: repo, cfg: cfg}
}

func (s *authService) Register(ctx context.Context, req dtos.RegisterFormRequest) (dtos.AuthResponse, error) {
	if len(req.Password) < 8 {
		return dtos.AuthResponse{}, ErrWeakPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return dtos.AuthResponse{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, req.Email, string(hashedPassword), db.UserRoleBuyer, req.Username)
	if err != nil {
		if isUniqueViolation(err) {
			return dtos.AuthResponse{}, ErrEmailTaken
		}
		return dtos.AuthResponse{}, fmt.Errorf("create user: %w", err)
	}

	accessToken, refreshToken, err := s.generateTokens(user.ID, string(user.Role))
	if err != nil {
		return dtos.AuthResponse{}, fmt.Errorf("generate tokens: %w", err)
	}

	_, err = s.repo.CreateRefreshToken(ctx, user.ID, refreshToken, time.Now().Add(s.cfg.JWT.RefreshTTL))
	if err != nil {
		return dtos.AuthResponse{}, fmt.Errorf("save refresh token: %w", err)
	}

	return dtos.AuthResponse{
		User:         dtos.ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Login(ctx context.Context, req dtos.LoginFormRequest) (dtos.AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

	_, err = s.repo.CreateRefreshToken(ctx, user.ID, refreshToken, time.Now().Add(s.cfg.JWT.RefreshTTL))
	if err != nil {
		return dtos.AuthResponse{}, fmt.Errorf("save refresh token: %w", err)
	}

	return dtos.AuthResponse{
		User:         dtos.ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Refresh(ctx context.Context, req dtos.RefreshFormRequest) (dtos.RefreshResponse, error) {
	token, err := s.repo.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dtos.RefreshResponse{}, ErrInvalidRefreshToken
		}
		return dtos.RefreshResponse{}, fmt.Errorf("get refresh token: %w", err)
	}

	if time.Now().After(token.ExpiresAt.Time) {
		_ = s.repo.DeleteRefreshToken(ctx, req.RefreshToken)
		return dtos.RefreshResponse{}, ErrRefreshTokenExpired
	}

	user, err := s.repo.GetUserByID(ctx, token.UserID)
	if err != nil {
		return dtos.RefreshResponse{}, fmt.Errorf("find user: %w", err)
	}

	if err := s.repo.DeleteRefreshToken(ctx, req.RefreshToken); err != nil {
		return dtos.RefreshResponse{}, fmt.Errorf("delete old refresh token: %w", err)
	}

	accessToken, newRefreshToken, err := s.generateTokens(user.ID, string(user.Role))
	if err != nil {
		return dtos.RefreshResponse{}, fmt.Errorf("generate tokens: %w", err)
	}

	_, err = s.repo.CreateRefreshToken(ctx, user.ID, newRefreshToken, time.Now().Add(s.cfg.JWT.RefreshTTL))
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

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
