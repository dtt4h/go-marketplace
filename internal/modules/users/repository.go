package users

import (
	"context"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id int64) (db.User, error)
	UpdateUser(ctx context.Context, id int64, username, avatarURL, phone *string) (db.User, error)
	UpdateUserRole(ctx context.Context, id int64, role db.UserRole) (db.User, error)
	GetStoreByUserID(ctx context.Context, userID int64) (db.Store, error)
	CreateStore(ctx context.Context, userID int64, name string, description, logoURL *string) (db.Store, error)
}

type userRepository struct {
	queries *db.Queries
}

func NewUserRepository(queries *db.Queries) UserRepository {
	return &userRepository{queries: queries}
}

func (r *userRepository) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	return r.queries.GetUserByID(ctx, id)
}

func (r *userRepository) UpdateUser(ctx context.Context, id int64, username, avatarURL, phone *string) (db.User, error) {
	return r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:        id,
		Username:  nullText(username),
		AvatarUrl: nullText(avatarURL),
		Phone:     nullText(phone),
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

func (r *userRepository) CreateStore(ctx context.Context, userID int64, name string, description, logoURL *string) (db.Store, error) {
	return r.queries.CreateStore(ctx, db.CreateStoreParams{
		UserID:      userID,
		Name:        name,
		Description: nullText(description),
		LogoUrl:     nullText(logoURL),
	})
}

func nullText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}