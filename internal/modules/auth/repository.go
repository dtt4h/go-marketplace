package auth

import (
	"context"
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
}

type authRepository struct {
	queries *db.Queries
}

// NewAuthRepository creates a new AuthRepository.
func NewAuthRepository(queries *db.Queries) AuthRepository {
	return &authRepository{queries: queries}
}

func (r *authRepository) CreateUser(ctx context.Context, email, passwordHash string, role db.UserRole, username string) (db.User, error) {
	return r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		Username:     username,
	})
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *authRepository) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	return r.queries.GetUserByID(ctx, id)
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
