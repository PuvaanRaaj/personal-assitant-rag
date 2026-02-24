package config

import (
	"os"
	"strings"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port           string
	AllowedOrigins string

	// Database
	DatabaseURL string

	// Storage
	StorageDriver    string // "local", "localstack", or "s3"
	LocalStoragePath string // Path for local filesystem storage
	KnowledgeBasePath string // Path for local knowledge base folder
	DefaultUserID     string // Default user ID for local indexing

	// AWS S3
	AWSConfig AWSConfig

	// Qdrant
	QdrantURL string

	// OpenAI
	OpenAIKey string

	// JWT
	JWTSecret string
}

// AWSConfig holds AWS S3 configuration
type AWSConfig struct {
	Region          string
	Endpoint        string // For LocalStack
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		AllowedOrigins:   getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		DatabaseURL:      getEnv("DATABASE_URL", buildDatabaseURL()),
		StorageDriver:    getEnv("FILESYSTEM_DRIVER", "localstack"), // Default to localstack for Docker
		LocalStoragePath: getEnv("LOCAL_STORAGE_PATH", "./uploads"),
		KnowledgeBasePath: getEnv("KNOWLEDGE_BASE_PATH", "./knowledgebase"),
		DefaultUserID:     getEnv("DEFAULT_USER_ID", "local-user"),
		AWSConfig: AWSConfig{
			Region:          getEnv("AWS_REGION", "us-east-1"),
			Endpoint:        getEnv("AWS_ENDPOINT", ""), // Empty for real AWS S3
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			Bucket:          getEnv("S3_BUCKET", "rag-assistant-uploads"),
		},
		QdrantURL: getEnv("QDRANT_URL", "http://localhost:6333"),
		OpenAIKey: getEnv("OPENAI_API_KEY", ""),
		JWTSecret: getEnv("JWT_SECRET", "change-this-in-production"),
	}
}

// buildDatabaseURL constructs the PostgreSQL connection string from individual env vars
func buildDatabaseURL() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "rag_user")
	password := getEnv("DB_PASSWORD", "password")
	dbname := getEnv("DB_NAME", "rag_assistant")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname + "?sslmode=" + sslmode
}

// getEnv gets an environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}
