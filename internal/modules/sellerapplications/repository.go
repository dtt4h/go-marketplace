package sellerapplications

import (
	"context"
	"fmt"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
)

type sellerApplicationRepository struct {
	queries *db.Queries
}

func NewSellerApplicationRepository(queries *db.Queries) SellerApplicationRepository {
	return &sellerApplicationRepository{queries: queries}
}

func (r *sellerApplicationRepository) Create(ctx context.Context, arg db.CreateSellerApplicationParams) (db.SellerApplication, error) {
	return r.queries.CreateSellerApplication(ctx, arg)
}

func (r *sellerApplicationRepository) GetByID(ctx context.Context, id int64) (db.SellerApplication, error) {
	return r.queries.GetSellerApplicationByID(ctx, id)
}

func (r *sellerApplicationRepository) GetPendingByUserID(ctx context.Context, userID int64) (db.SellerApplication, error) {
	return r.queries.GetPendingSellerApplicationByUserID(ctx, userID)
}

func (r *sellerApplicationRepository) ListByUser(ctx context.Context, userID int64) ([]db.SellerApplication, error) {
	return r.queries.ListSellerApplicationsByUser(ctx, userID)
}

func (r *sellerApplicationRepository) ListPending(ctx context.Context, limit, offset int32) ([]db.SellerApplication, error) {
	apps, err := r.queries.ListPendingSellerApplications(ctx, db.ListPendingSellerApplicationsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list pending: %w", err)
	}
	return apps, nil
}

func (r *sellerApplicationRepository) CountPending(ctx context.Context) (int64, error) {
	return r.queries.CountPendingSellerApplications(ctx)
}

func (r *sellerApplicationRepository) UpdateStatus(ctx context.Context, id int64, status string) (db.SellerApplication, error) {
	return r.queries.UpdateSellerApplicationStatus(ctx, db.UpdateSellerApplicationStatusParams{
		ID:     id,
		Status: status,
	})
}
