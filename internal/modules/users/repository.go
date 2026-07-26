package users

import (
	"context"
	"fmt"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id int64) (db.User, error)
	UpdateUser(ctx context.Context, id int64, username, avatarURL, phone *string) (db.User, error)
	UpdateUserRole(ctx context.Context, id int64, role db.UserRole) (db.User, error)
	GetStoreByUserID(ctx context.Context, userID int64) (db.Store, error)
	CreateStoreWithRole(ctx context.Context, userID int64, name string, description, logoURL *string) (db.Store, error)
	GetStoreOwnerByStoreID(ctx context.Context, storeID int64) (int64, error)
}

type userRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewUserRepository(queries *db.Queries, pool *pgxpool.Pool) UserRepository {
	return &userRepository{queries: queries, pool: pool}
}

func (r *userRepository) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	return r.queries.GetUserByID(ctx, id)
}

func (r *userRepository) UpdateUser(ctx context.Context, id int64, username, avatarURL, phone *string) (db.User, error) {
	return r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:        id,
		Username:  pgutil.NullText(username),
		AvatarUrl: pgutil.NullText(avatarURL),
		Phone:     pgutil.NullText(phone),
	})
}

func (r *userRepository) UpdateUserRole(ctx context.Context, id int64, role db.UserRole) (db.User, error) {
	return r.queries.UpdateUserRole(ctx, db.UpdateUserRoleParams{
		ID:   id,
		Role: role,
	})
}

func (r *userRepository) GetStoreByUserID(ctx context.Context, userID int64) (db.Store, error) {
	return r.queries.GetStoreByUserID(ctx, userID)
}

func (r *userRepository) CreateStoreWithRole(ctx context.Context, userID int64, name string, description, logoURL *string) (db.Store, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.Store{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := r.queries.WithTx(tx)

	store, err := q.CreateStore(ctx, db.CreateStoreParams{
		UserID:      userID,
		Name:        name,
		Description: pgutil.NullText(description),
		LogoUrl:     pgutil.NullText(logoURL),
	})
	if err != nil {
		return db.Store{}, fmt.Errorf("create store: %w", err)
	}

	if _, err := q.UpdateUserRole(ctx, db.UpdateUserRoleParams{
		ID:   userID,
		Role: db.UserRoleSeller,
	}); err != nil {
		return db.Store{}, fmt.Errorf("update user role: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Store{}, fmt.Errorf("commit tx: %w", err)
	}

	return store, nil
}

func (r *userRepository) GetStoreOwnerByStoreID(ctx context.Context, storeID int64) (int64, error) {
	return r.queries.GetStoreOwnerByStoreID(ctx, storeID)
}