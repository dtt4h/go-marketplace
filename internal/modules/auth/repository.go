package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

// AuthRepository defines data access methods for authentication.
type AuthRepository interface {
	CreateUser(ctx context.Context, email, passwordHash string, role db.UserRole, username string) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByID(ctx context.Context, id int64) (db.User, error)
	CreateRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) (db.RefreshToken, error)
	GetRefreshToken(ctx context.Context, token string) (db.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteUserRefreshTokens(ctx context.Context, userID int64) error
	GenerateResetToken(ctx context.Context, userID int64, expiresAt time.Time) (string, error)
	GetUserByResetToken(ctx context.Context, tokenHash string) (db.User, error)
	DeleteResetToken(ctx context.Context, userID int64) error
	ResetPassword(ctx context.Context, userID int64, passwordHash string) (db.User, error)
}

type authRepository struct {
	queries *db.Queries
}

// NewAuthRepository creates a new AuthRepository.
func NewAuthRepository(queries *db.Queries) AuthRepository {
	return &authRepository{queries: queries}
}

func (r *authRepository) CreateUser(ctx context.Context, email, passwordHash string, role db.UserRole, username string) (db.User, error) {
	row, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		Username:     username,
	})
	if err != nil {
		return db.User{}, err
	}
	return createUserRowToUser(row), nil
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return db.User{}, err
	}
	return getUserByEmailRowToUser(row), nil
}

func (r *authRepository) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return db.User{}, err
	}
	return getUserByIDRowToUser(row), nil
}

func (r *authRepository) CreateRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) (db.RefreshToken, error) {
	return r.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		Token:     token,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
}

func (r *authRepository) GetRefreshToken(ctx context.Context, token string) (db.RefreshToken, error) {
	return r.queries.GetRefreshToken(ctx, token)
}

func (r *authRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := r.queries.DeleteRefreshToken(ctx, token)
	return err
}

func (r *authRepository) DeleteUserRefreshTokens(ctx context.Context, userID int64) error {
	return r.queries.DeleteUserRefreshTokens(ctx, userID)
}

func (r *authRepository) GenerateResetToken(ctx context.Context, userID int64, expiresAt time.Time) (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	token := hex.EncodeToString(randomBytes)

	tokenHash := hashToken(token)

	_, err := r.queries.GenerateResetToken(ctx, db.GenerateResetTokenParams{
		ID:                  userID,
		ResetToken:          pgtype.Text{String: tokenHash, Valid: true},
		ResetTokenExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return "", fmt.Errorf("generate reset token: %w", err)
	}

	return token, nil
}

func (r *authRepository) GetUserByResetToken(ctx context.Context, tokenHash string) (db.User, error) {
	row, err := r.queries.GetUserByResetToken(ctx, pgtype.Text{String: tokenHash, Valid: true})
	if err != nil {
		return db.User{}, err
	}
	return db.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Role:         row.Role,
		Username:     row.Username,
		AvatarUrl:    row.AvatarUrl,
		Phone:        row.Phone,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

func (r *authRepository) DeleteResetToken(ctx context.Context, userID int64) error {
	return r.queries.DeleteResetToken(ctx, userID)
}

func (r *authRepository) ResetPassword(ctx context.Context, userID int64, passwordHash string) (db.User, error) {
	row, err := r.queries.ResetPassword(ctx, db.ResetPasswordParams{
		ID:           userID,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return db.User{}, err
	}
	return resetPasswordRowToUser(row), nil
}

func createUserRowToUser(row db.CreateUserRow) db.User {
	return db.User{
		ID:       row.ID,
		Email:    row.Email,
		Role:     row.Role,
		Username: row.Username,
	}
}

func getUserByEmailRowToUser(row db.GetUserByEmailRow) db.User {
	return db.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Role:         row.Role,
		Username:     row.Username,
		AvatarUrl:    row.AvatarUrl,
		Phone:        row.Phone,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func getUserByIDRowToUser(row db.GetUserByIDRow) db.User {
	return db.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Role:         row.Role,
		Username:     row.Username,
		AvatarUrl:    row.AvatarUrl,
		Phone:        row.Phone,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func resetPasswordRowToUser(row db.ResetPasswordRow) db.User {
	return db.User{
		ID:       row.ID,
		Email:    row.Email,
		Role:     row.Role,
		Username: row.Username,
	}
}
