package users

import (
	"context"
	"errors"
	"strings"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/jackc/pgx/v5"
)

var (
	ErrStoreAlreadyExists = errors.New("store already exists")
	ErrNoStore            = errors.New("store not found")
	ErrNotFoundUser       = errors.New("user not found")
	ErrUsernameTaken      = errors.New("username is already taken")
)

type UserService interface {
	GetProfile(ctx context.Context, userID int64) (dtos.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID int64, req dtos.UpdateProfileRequest) (dtos.ProfileResponse, error)
	CreateStore(ctx context.Context, userID int64, req dtos.CreateStoreRequest) (dtos.StoreResponse, error)
}

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *userService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID int64) (dtos.ProfileResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dtos.ProfileResponse{}, ErrNotFoundUser
		}
		return dtos.ProfileResponse{}, err
	}

	return dtos.ToProfileResponse(user), nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID int64, req dtos.UpdateProfileRequest) (dtos.ProfileResponse, error) {
	user, err := s.repo.UpdateUser(ctx, userID, req.Username, req.AvatarURL, req.Phone)
	if err != nil {
		if isUniqueViolation(err) {
			return dtos.ProfileResponse{}, ErrUsernameTaken
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return dtos.ProfileResponse{}, ErrNotFoundUser
		}
		return dtos.ProfileResponse{}, err
	}

	return dtos.ToProfileResponse(user), nil
}

func (s *userService) CreateStore(ctx context.Context, userID int64, req dtos.CreateStoreRequest) (dtos.StoreResponse, error) {

	_, err := s.repo.GetStoreByUserID(ctx, userID)
	if err == nil {
		return dtos.StoreResponse{}, ErrStoreAlreadyExists
	}

	if err != pgx.ErrNoRows {
		return dtos.StoreResponse{}, err
	}

	store, err := s.repo.CreateStore(ctx, userID, req.Name, req.Description, req.LogoURL)
	if err != nil {
		return dtos.StoreResponse{}, err
	}

	_, err = s.repo.UpdateUserRole(ctx, userID, db.UserRoleSeller)
	if err != nil {
		return dtos.StoreResponse{}, err
	}

	return dtos.ToStoreResponse(store), nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
