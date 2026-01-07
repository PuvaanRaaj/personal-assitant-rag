package storage

import (
	"fmt"

	"github.com/PuvaanRaaj/personal-rag-agent/internal/config"
)

// DriverType represents the type of storage driver
type DriverType string

const (
	// DriverLocal uses local filesystem storage
	DriverLocal DriverType = "local"
	// DriverLocalStack uses LocalStack S3-compatible storage
	DriverLocalStack DriverType = "localstack"
	// DriverS3 uses AWS S3 storage
	DriverS3 DriverType = "s3"
)

// NewStorageDriver creates a storage driver based on the configuration
func NewStorageDriver(cfg *config.Config) (StorageDriver, error) {
	switch DriverType(cfg.StorageDriver) {
	case DriverLocal:
		return NewLocalStorage(cfg.LocalStoragePath)

	case DriverLocalStack:
		// Use S3Client with LocalStack endpoint
		if cfg.AWSConfig.Endpoint == "" {
			return nil, fmt.Errorf("AWS_ENDPOINT is required for localstack driver")
		}
		return NewS3Client(cfg.AWSConfig)

	case DriverS3:
		// Use S3Client with real AWS S3 (no custom endpoint)
		awsCfg := cfg.AWSConfig
		awsCfg.Endpoint = "" // Ensure we don't use a custom endpoint
		return NewS3Client(awsCfg)

	default:
		return nil, fmt.Errorf("unknown storage driver: %s (valid options: local, localstack, s3)", cfg.StorageDriver)
	}
}
