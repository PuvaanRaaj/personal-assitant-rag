package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalStorage implements StorageDriver for local filesystem
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local filesystem storage driver
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

// UploadFile saves a file to the local filesystem
func (l *LocalStorage) UploadFile(ctx context.Context, key string, file io.Reader) error {
	fullPath := filepath.Join(l.basePath, key)

	// Create directory structure
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	// Copy content
	if _, err := io.Copy(f, file); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GetFile retrieves a file from the local filesystem
func (l *LocalStorage) GetFile(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.basePath, key)

	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return f, nil
}

// DeleteFile removes a file from the local filesystem
func (l *LocalStorage) DeleteFile(ctx context.Context, key string) error {
	fullPath := filepath.Join(l.basePath, key)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // File already deleted, not an error
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Try to clean up empty parent directories
	dir := filepath.Dir(fullPath)
	for dir != l.basePath {
		if err := os.Remove(dir); err != nil {
			break // Directory not empty or other error, stop cleanup
		}
		dir = filepath.Dir(dir)
	}

	return nil
}

// GetPresignedURL returns a file:// URL for local files
// Note: This is only useful for local development/debugging
func (l *LocalStorage) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	fullPath := filepath.Join(l.basePath, key)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", key)
	}

	// Return absolute path as file:// URL
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return "file://" + absPath, nil
}
