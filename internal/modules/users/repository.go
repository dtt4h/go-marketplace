package users

import (
	"context"
	"fmt"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository defines data access methods for user operations.
type UserRepository interface {
	GetUserByID(ctx context.Context, id int64) (db.User, error)
	UpdateUser(ctx context.Context, id int64, username, avatarURL, phone *string) (db.User, error)
	UpdateUserRole(ctx context.Context, id int64, role db.UserRole) (db.User, error)
	GetStoreByUserID(ctx context.Context, userID int64) (db.Store, error)
	CreateStoreWithRole(ctx context.Context, userID int64, name string, description, logoURL *string) (db.Store, error)
	GetStoreOwnerByStoreID(ctx context.Context, storeID int64) (int64, error)
	UpdateStore(ctx context.Context, storeID int64, name, description, logoURL *string) (db.Store, error)
	GetStoreByID(ctx context.Context, storeID int64) (db.Store, error)
}

type userRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(queries *db.Queries, pool *pgxpool.Pool) UserRepository {
	return &userRepository{queries: queries, pool: pool}
}

func (r *userRepository) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return db.User{}, err
	}
	return getUserByIDRowToUser(row), nil
}

func (r *userRepository) UpdateUser(ctx context.Context, id int64, username, avatarURL, phone *string) (db.User, error) {
	row, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:        id,
		Username:  pgutil.NullText(username),
		AvatarUrl: pgutil.NullText(avatarURL),
		Phone:     pgutil.NullText(phone),
	})
	if err != nil {
		return db.User{}, err
	}
	return updateUserRowToUser(row), nil
}

func (r *userRepository) UpdateUserRole(ctx context.Context, id int64, role db.UserRole) (db.User, error) {
	row, err := r.queries.UpdateUserRole(ctx, db.UpdateUserRoleParams{
		ID:   id,
		Role: role,
	})
	if err != nil {
		return db.User{}, err
	}
	return updateUserRoleRowToUser(row), nil
}

func getUserByIDRowToUser(row db.GetUserByIDRow) db.User {
	return db.User{
		ID:       row.ID,
		Email:    row.Email,
		Role:     row.Role,
		Username: row.Username,
	}
}

func updateUserRowToUser(row db.UpdateUserRow) db.User {
	return db.User{
		ID:       row.ID,
		Email:    row.Email,
		Role:     row.Role,
		Username: row.Username,
	}
}

func updateUserRoleRowToUser(row db.UpdateUserRoleRow) db.User {
	return db.User{
		ID:       row.ID,
		Email:    row.Email,
		Role:     row.Role,
		Username: row.Username,
	}
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

func (r *userRepository) UpdateStore(ctx context.Context, storeID int64, name, description, logoURL *string) (db.Store, error) {
	row, err := r.queries.UpdateStore(ctx, db.UpdateStoreParams{
		ID:          storeID,
		Name:        pgutil.NullText(name),
		Description: pgutil.NullText(description),
		LogoUrl:     pgutil.NullText(logoURL),
	})
	if err != nil {
		return db.Store{}, err
	}
	return row, nil
}

func (r *userRepository) GetStoreByID(ctx context.Context, storeID int64) (db.Store, error) {
	return r.queries.GetStoreByID(ctx, storeID)
}