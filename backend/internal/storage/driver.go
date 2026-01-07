package storage

import (
	"context"
	"io"
	"time"
)

// StorageDriver defines the interface for file storage operations
type StorageDriver interface {
	// UploadFile uploads a file to storage
	UploadFile(ctx context.Context, key string, file io.Reader) error

	// GetFile retrieves a file from storage
	GetFile(ctx context.Context, key string) (io.ReadCloser, error)

	// DeleteFile deletes a file from storage
	DeleteFile(ctx context.Context, key string) error

	// GetPresignedURL generates a URL for accessing a file
	// For local storage, this returns a file:// URL
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}
