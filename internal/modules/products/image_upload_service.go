package products

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/dtt4h/go-marketplace/internal/storage"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
)

var ErrStorageNotConfigured = errors.New("image storage is not configured")

// ImageUploadService handles image upload and deletion operations.
type ImageUploadService interface {
	UploadImage(ctx context.Context, productID int64, userID int64, reader io.Reader, size int64, contentType string) (dtos.ProductImageResponse, error)
	DeleteImage(ctx context.Context, imageID int64, userID int64) error
	GetPresignedUploadURL(ctx context.Context, productID int64, userID int64, contentType string, objectKey string) (string, error)
	GetPresignedDownloadURL(ctx context.Context, imageID int64, userID int64, expirySeconds int64) (string, error)
}

type imageUploadService struct {
	repo  ProductRepository
	store storage.ObjectStorage
}

// NewImageUploadService creates a new ImageUploadService.
func NewImageUploadService(repo ProductRepository, store storage.ObjectStorage) ImageUploadService {
	return &imageUploadService{repo: repo, store: store}
}

func (s *imageUploadService) checkProductOwnership(ctx context.Context, productID, userID int64) error {
	row, err := s.repo.GetProductStoreOwner(ctx, productID)
	if err != nil {
		if pgutil.IsNoRows(err) {
			return ErrProductNotFound
		}
		return fmt.Errorf("check ownership: %w", err)
	}
	if row.StoreOwnerID != userID {
		return ErrForbidden
	}
	return nil
}

func (s *imageUploadService) UploadImage(ctx context.Context, productID int64, userID int64, reader io.Reader, size int64, contentType string) (dtos.ProductImageResponse, error) {
	if s.store == nil {
		return dtos.ProductImageResponse{}, ErrStorageNotConfigured
	}

	if err := s.checkProductOwnership(ctx, productID, userID); err != nil {
		return dtos.ProductImageResponse{}, err
	}

	key, err := s.store.Upload(ctx, reader, size, storage.UploadOptions{
		ContentType: contentType,
	})
	if err != nil {
		return dtos.ProductImageResponse{}, fmt.Errorf("upload to storage: %w", err)
	}

	images, err := s.repo.ListProductImages(ctx, productID)
	if err != nil {
		return dtos.ProductImageResponse{}, fmt.Errorf("list images: %w", err)
	}

	position := int32(len(images))

	img, err := s.repo.CreateProductImage(ctx, productID, key, position)
	if err != nil {
		return dtos.ProductImageResponse{}, fmt.Errorf("create image record: %w", err)
	}

	return dtos.ProductImageResponse{
		ID:       img.ID,
		URL:      img.Url,
		Position: img.Position,
	}, nil
}

func (s *imageUploadService) DeleteImage(ctx context.Context, imageID int64, userID int64) error {
	if s.store == nil {
		return ErrStorageNotConfigured
	}

	img, err := s.repo.GetImageByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("get image: %w", err)
	}

	if err := s.checkProductOwnership(ctx, img.ProductID, userID); err != nil {
		return err
	}

	if img.ObjectKey.Valid {
		if err := s.store.Delete(ctx, img.ObjectKey.String); err != nil {
			if !storage.IsNotFoundError(err) {
				return fmt.Errorf("delete from storage: %w", err)
			}
		}
	}

	if err := s.repo.DeleteImageByID(ctx, imageID); err != nil {
		return fmt.Errorf("delete image record: %w", err)
	}

	return nil
}

func (s *imageUploadService) GetPresignedUploadURL(ctx context.Context, productID int64, userID int64, contentType string, objectKey string) (string, error) {
	if s.store == nil {
		return "", ErrStorageNotConfigured
	}

	if err := s.checkProductOwnership(ctx, productID, userID); err != nil {
		return "", err
	}

	url, err := s.store.PresignedPutObject(ctx, objectKey, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("generate presigned URL: %w", err)
	}

	return url, nil
}

func (s *imageUploadService) GetPresignedDownloadURL(ctx context.Context, imageID int64, userID int64, expirySeconds int64) (string, error) {
	if s.store == nil {
		return "", ErrStorageNotConfigured
	}

	img, err := s.repo.GetImageByID(ctx, imageID)
	if err != nil {
		return "", fmt.Errorf("get image: %w", err)
	}

	if err := s.checkProductOwnership(ctx, img.ProductID, userID); err != nil {
		return "", err
	}

	url, err := s.store.PresignedGetObject(ctx, img.ObjectKey.String, time.Duration(expirySeconds)*time.Second)
	if err != nil {
		return "", fmt.Errorf("generate presigned URL: %w", err)
	}

	return url, nil
}
