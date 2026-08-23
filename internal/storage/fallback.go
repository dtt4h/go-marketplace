package storage

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"time"
)

var ErrStorageNotConfigured = errors.New("image storage is not configured")

// FallbackObjectStorage wraps ObjectStorage with graceful degradation.
// When the underlying storage fails, it logs the error and returns a
// sentinel error that callers can handle without breaking the user flow.
type FallbackObjectStorage struct {
	wrapped ObjectStorage
	log     *slog.Logger
}

// NewFallbackObjectStorage creates a storage wrapper with graceful degradation.
// If wrapped is nil, all operations return ErrStorageNotConfigured.
func NewFallbackObjectStorage(wrapped ObjectStorage, log *slog.Logger) ObjectStorage {
	if wrapped == nil {
		return &nilStorage{log: log}
	}
	return &FallbackObjectStorage{
		wrapped: wrapped,
		log:     log,
	}
}

func (f *FallbackObjectStorage) Upload(ctx context.Context, reader io.Reader, size int64, opts UploadOptions) (string, error) {
	if f.wrapped == nil {
		return "", ErrStorageNotConfigured
	}
	key, err := f.wrapped.Upload(ctx, reader, size, opts)
	if err != nil {
		f.log.Error("S3 upload failed, image not saved",
			"error", err.Error(),
			"content_type", opts.ContentType,
		)
		return "", err
	}
	return key, nil
}

func (f *FallbackObjectStorage) Delete(ctx context.Context, key string) error {
	if f.wrapped == nil {
		return ErrStorageNotConfigured
	}
	err := f.wrapped.Delete(ctx, key)
	if err != nil {
		f.log.Error("S3 delete failed, image record not cleaned up",
			"error", err.Error(),
			"key", key,
		)
		return err
	}
	return nil
}

func (f *FallbackObjectStorage) PresignedGetObject(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if f.wrapped == nil {
		return "", ErrStorageNotConfigured
	}
	url, err := f.wrapped.PresignedGetObject(ctx, key, expiry)
	if err != nil {
		f.log.Error("S3 presigned get failed",
			"error", err.Error(),
			"key", key,
		)
		return "", err
	}
	return url, nil
}

func (f *FallbackObjectStorage) PresignedPutObject(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if f.wrapped == nil {
		return "", ErrStorageNotConfigured
	}
	url, err := f.wrapped.PresignedPutObject(ctx, key, expiry)
	if err != nil {
		f.log.Error("S3 presigned put failed",
			"error", err.Error(),
			"key", key,
		)
		return "", err
	}
	return url, nil
}

// nilStorage is a no-op storage when S3 is not configured at all.
type nilStorage struct {
	log *slog.Logger
}

func (n *nilStorage) Upload(ctx context.Context, reader io.Reader, size int64, opts UploadOptions) (string, error) {
	return "", ErrStorageNotConfigured
}

func (n *nilStorage) Delete(ctx context.Context, key string) error {
	return ErrStorageNotConfigured
}

func (n *nilStorage) PresignedGetObject(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "", ErrStorageNotConfigured
}

func (n *nilStorage) PresignedPutObject(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "", ErrStorageNotConfigured
}
