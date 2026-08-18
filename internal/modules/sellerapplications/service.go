package sellerapplications

import (
	"context"
	"errors"
	"fmt"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrAlreadyPending    = errors.New("you already have a pending application")
	ErrNotFound          = errors.New("application not found")
	ErrInvalidStatus     = errors.New("invalid status, must be 'approved' or 'rejected'")
	ErrAlreadyProcessed  = errors.New("application has already been processed")
)

// UserRepository — subset of users.UserRepository needed here.
type UserRepository interface {
	UpdateUserRole(ctx context.Context, id int64, role db.UserRole) (db.User, error)
	CreateStoreWithRole(ctx context.Context, userID int64, name string, description, logoURL *string) (db.Store, error)
	GetUserByID(ctx context.Context, id int64) (db.User, error)
}

// EmailSender sends email notifications.
type EmailSender interface {
	SendSellerApplicationApproved(to, username, storeName string) error
	SendSellerApplicationRejected(to, username, reason string) error
	SendNewApplicationNotification(to, username, storeName, description string) error
}

type SellerApplicationRepository interface {
	Create(ctx context.Context, arg db.CreateSellerApplicationParams) (db.SellerApplication, error)
	GetByID(ctx context.Context, id int64) (db.SellerApplication, error)
	GetPendingByUserID(ctx context.Context, userID int64) (db.SellerApplication, error)
	ListByUser(ctx context.Context, userID int64) ([]db.SellerApplication, error)
	ListPending(ctx context.Context, limit, offset int32) ([]db.SellerApplication, error)
	CountPending(ctx context.Context) (int64, error)
	UpdateStatus(ctx context.Context, id int64, status string) (db.SellerApplication, error)
}

type SellerApplicationService interface {
	Create(ctx context.Context, userID int64, storeName string, description *string) (db.SellerApplication, error)
	ListMine(ctx context.Context, userID int64) ([]db.SellerApplication, error)
	ListPending(ctx context.Context, page, limit int) ([]db.SellerApplication, int64, error)
	Approve(ctx context.Context, applicationID int64) (db.SellerApplication, error)
	Reject(ctx context.Context, applicationID int64) (db.SellerApplication, error)
}

type sellerApplicationService struct {
	repo       SellerApplicationRepository
	userRepo   UserRepository
	email      EmailSender
}

func NewSellerApplicationService(repo SellerApplicationRepository, userRepo UserRepository, email EmailSender) SellerApplicationService {
	return &sellerApplicationService{repo: repo, userRepo: userRepo, email: email}
}

func (s *sellerApplicationService) Create(ctx context.Context, userID int64, storeName string, description *string) (db.SellerApplication, error) {
	if storeName == "" {
		return db.SellerApplication{}, errors.New("store name is required")
	}

	// Check for existing pending application
	if _, err := s.repo.GetPendingByUserID(ctx, userID); err == nil {
		return db.SellerApplication{}, ErrAlreadyPending
	}

	desc := pgtype.Text{Valid: false}
	if description != nil {
		desc = pgtype.Text{String: *description, Valid: true}
	}

	app, err := s.repo.Create(ctx, db.CreateSellerApplicationParams{
		UserID:      userID,
		StoreName:   storeName,
		Description: desc,
	})
	if err != nil {
		return db.SellerApplication{}, fmt.Errorf("create seller application: %w", err)
	}

	// Send notification to admin about new application
	go func() {
		user, userErr := s.userRepo.GetUserByID(ctx, userID)
		if userErr != nil {
			return
		}
		descStr := ""
		if description != nil {
			descStr = *description
		}
		_ = s.email.SendNewApplicationNotification("admin@marketplace.local", user.Username, storeName, descStr)
	}()

	return app, nil
}

func (s *sellerApplicationService) ListMine(ctx context.Context, userID int64) ([]db.SellerApplication, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *sellerApplicationService) ListPending(ctx context.Context, page, limit int) ([]db.SellerApplication, int64, error) {
	offset := int32((page - 1) * limit)
	apps, err := s.repo.ListPending(ctx, int32(limit), offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list pending applications: %w", err)
	}
	total, err := s.repo.CountPending(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count pending applications: %w", err)
	}
	return apps, total, nil
}

func (s *sellerApplicationService) Approve(ctx context.Context, applicationID int64) (db.SellerApplication, error) {
	app, err := s.repo.GetByID(ctx, applicationID)
	if err != nil {
		return db.SellerApplication{}, ErrNotFound
	}
	if app.Status != "pending" {
		return db.SellerApplication{}, ErrAlreadyProcessed
	}

	app, err = s.repo.UpdateStatus(ctx, applicationID, "approved")
	if err != nil {
		return db.SellerApplication{}, fmt.Errorf("approve application: %w", err)
	}

	// Promote user to seller role
	user, userErr := s.userRepo.GetUserByID(ctx, app.UserID)
	if userErr == nil {
		go func() {
			_ = s.email.SendSellerApplicationApproved(user.Email, user.Username, app.StoreName)
		}()
	}

	if _, err := s.userRepo.UpdateUserRole(ctx, app.UserID, db.UserRoleSeller); err != nil {
		return db.SellerApplication{}, fmt.Errorf("promote user to seller: %w", err)
	}

	return app, nil
}

func (s *sellerApplicationService) Reject(ctx context.Context, applicationID int64) (db.SellerApplication, error) {
	app, err := s.repo.GetByID(ctx, applicationID)
	if err != nil {
		return db.SellerApplication{}, ErrNotFound
	}
	if app.Status != "pending" {
		return db.SellerApplication{}, ErrAlreadyProcessed
	}

	app, err = s.repo.UpdateStatus(ctx, applicationID, "rejected")
	if err != nil {
		return db.SellerApplication{}, fmt.Errorf("reject application: %w", err)
	}

	// Send rejection notification to user
	go func() {
		user, userErr := s.userRepo.GetUserByID(ctx, app.UserID)
		if userErr != nil {
			return
		}
		_ = s.email.SendSellerApplicationRejected(user.Email, user.Username, "Ваша заявка была отклонена.")
	}()

	return app, nil
}
