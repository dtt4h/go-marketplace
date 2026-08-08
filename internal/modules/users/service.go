package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
)

var (
	ErrStoreAlreadyExists = errors.New("store already exists")
	ErrStoreNotFound      = errors.New("store not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameTaken      = errors.New("username is already taken")
	ErrStoreNameRequired  = errors.New("store name is required")
)

// UserService defines business logic for user profile operations.
type UserService interface {
	GetProfile(ctx context.Context, userID int64) (dtos.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID int64, req dtos.UpdateProfileRequest) (dtos.ProfileResponse, error)
	CreateStore(ctx context.Context, userID int64, req dtos.CreateStoreRequest) (dtos.StoreResponse, error)
}

type userService struct {
	repo UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetProfile(ctx context.Context, userID int64) (dtos.ProfileResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.ProfileResponse{}, ErrUserNotFound
		}
		return dtos.ProfileResponse{}, fmt.Errorf("get user by id: %w", err)
	}

	return dtos.ToProfileResponse(user), nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID int64, req dtos.UpdateProfileRequest) (dtos.ProfileResponse, error) {
	user, err := s.repo.UpdateUser(ctx, userID, req.Username, req.AvatarURL, req.Phone)
	if err != nil {
		if pgutil.IsUniqueViolation(err) {
			return dtos.ProfileResponse{}, ErrUsernameTaken
		}
		if pgutil.IsNoRows(err) {
			return dtos.ProfileResponse{}, ErrUserNotFound
		}
		return dtos.ProfileResponse{}, fmt.Errorf("update user: %w", err)
	}

	return dtos.ToProfileResponse(user), nil
}

func (s *userService) CreateStore(ctx context.Context, userID int64, req dtos.CreateStoreRequest) (dtos.StoreResponse, error) {
	if req.Name == "" {
		return dtos.StoreResponse{}, ErrStoreNameRequired
	}

	_, err := s.repo.GetStoreByUserID(ctx, userID)
	if err == nil {
		return dtos.StoreResponse{}, ErrStoreAlreadyExists
	}
	if !pgutil.IsNoRows(err) {
		return dtos.StoreResponse{}, fmt.Errorf("get store by user id: %w", err)
	}

	store, err := s.repo.CreateStoreWithRole(ctx, userID, req.Name, req.Description, req.LogoURL)
	if err != nil {
		return dtos.StoreResponse{}, fmt.Errorf("create store with role: %w", err)
	}

	return dtos.ToStoreResponse(store), nil
}