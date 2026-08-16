package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
)

var (
	ErrStoreAlreadyExists = errors.New("store already exists")
	ErrStoreNotFound      = errors.New("store not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameTaken      = errors.New("username is already taken")
	ErrStoreNameRequired  = errors.New("store name is required")
	ErrUsernameRequired   = errors.New("username is required")
)

// UserService defines business logic for user profile operations.
type UserService interface {
	GetProfile(ctx context.Context, userID int64) (dtos.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID int64, req dtos.UpdateProfileRequest) (dtos.ProfileResponse, error)
	CreateStore(ctx context.Context, userID int64, req dtos.CreateStoreRequest) (dtos.StoreResponse, error)
	UpdateStore(ctx context.Context, userID int64, req dtos.UpdateStoreRequest) (dtos.StoreResponse, error)
}

type userService struct {
	repo UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func validateUsername(username *string) error {
	if username == nil {
		return nil
	}
	if strings.TrimSpace(*username) == "" {
		return ErrUsernameRequired
	}
	return nil
}

func validateStoreName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrStoreNameRequired
	}
	return nil
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
	if err := validateUsername(req.Username); err != nil {
		return dtos.ProfileResponse{}, err
	}

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
	if err := validateStoreName(req.Name); err != nil {
		return dtos.StoreResponse{}, err
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

func (s *userService) UpdateStore(ctx context.Context, userID int64, req dtos.UpdateStoreRequest) (dtos.StoreResponse, error) {
	store, err := s.repo.GetStoreByUserID(ctx, userID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return dtos.StoreResponse{}, ErrStoreNotFound
		}
		return dtos.StoreResponse{}, fmt.Errorf("get store by user id: %w", err)
	}

	_, err = s.repo.UpdateStore(ctx, store.ID, &req.Name, req.Description, req.LogoURL)
	if err != nil {
		return dtos.StoreResponse{}, fmt.Errorf("update store: %w", err)
	}

	updated, err := s.repo.GetStoreByUserID(ctx, userID)
	if err != nil {
		return dtos.StoreResponse{}, fmt.Errorf("get updated store: %w", err)
	}

	return dtos.ToStoreResponse(updated), nil
}
