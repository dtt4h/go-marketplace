package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ObjectStorage defines methods for object storage operations.
type ObjectStorage interface {
	Upload(ctx context.Context, reader io.Reader, size int64, opts UploadOptions) (string, error)
	Delete(ctx context.Context, key string) error
	PresignedGetObject(ctx context.Context, key string, expiry time.Duration) (string, error)
	PresignedPutObject(ctx context.Context, key string, expiry time.Duration) (string, error)
}

// UploadOptions holds optional parameters for object upload.
type UploadOptions struct {
	ContentType string
	Metadata    map[string]string
}

// ObjectStorageConfig holds connection parameters for the storage backend.
type ObjectStorageConfig struct {
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	Bucket    string
	Secure    bool
}

// NewObjectStorage creates a new S3/MinIO client and ensures the bucket exists.
func NewObjectStorage(cfg ObjectStorageConfig) (*ObjectStorageClient, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3 bucket is required")
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.Secure,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	exists, err := client.BucketExists(context.Background(), cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket exists: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{
			Region: cfg.Region,
		}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
	}

	return &ObjectStorageClient{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// ObjectStorageClient wraps the MinIO SDK client.
type ObjectStorageClient struct {
	client *minio.Client
	bucket string
}

func (o *ObjectStorageClient) Upload(ctx context.Context, reader io.Reader, size int64, opts UploadOptions) (string, error) {
	contentType := opts.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	key := generateObjectKey(contentType)

	if _, err := o.client.PutObject(ctx, o.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType:      contentType,
		UserMetadata:     opts.Metadata,
		DisableMultipart: false,
	}); err != nil {
		return "", fmt.Errorf("upload object %q: %w", key, err)
	}

	return key, nil
}

func (o *ObjectStorageClient) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("delete: key is required")
	}

	if err := o.client.RemoveObject(ctx, o.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	return nil
}

func (o *ObjectStorageClient) PresignedGetObject(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", fmt.Errorf("presigned get: key is required")
	}

	url, err := o.client.PresignedGetObject(ctx, o.bucket, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presigned get object %q: %w", key, err)
	}

	return url.String(), nil
}

func (o *ObjectStorageClient) PresignedPutObject(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", fmt.Errorf("presigned put: key is required")
	}

	url, err := o.client.PresignedPutObject(ctx, o.bucket, key, expiry)
	if err != nil {
		return "", fmt.Errorf("presigned put object %q: %w", key, err)
	}

	return url.String(), nil
}

func generateObjectKey(contentType string) string {
	ext := ".bin"
	if exts, _ := mime.ExtensionsByType(contentType); len(exts) > 0 && exts[0] != "" {
		ext = exts[0]
	}

	filename := uuid.New().String() + ext
	return "products/images/" + filename
}

// IsNotFoundError checks if the error indicates a missing object.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var respErr minio.ErrorResponse
	if errors.As(err, &respErr) {
		return respErr.Code == "NoSuchKey"
	}
	return strings.Contains(err.Error(), "NoSuchKey")
}
