package model

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Document represents an uploaded document
type Document struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Filename    string    `json:"filename" db:"filename"`
	FileType    string    `json:"file_type" db:"file_type"`
	FileSize    int64     `json:"file_size" db:"file_size"`
	FileHash    string    `json:"file_hash" db:"file_hash"`
	StoragePath string    `json:"storage_path" db:"storage_path"`
	TotalChunks int       `json:"total_chunks" db:"total_chunks"`
	UploadDate  time.Time `json:"upload_date" db:"upload_date"`
}

// QueryHistory represents a query made by a user
type QueryHistory struct {
	ID        string                 `json:"id" db:"id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	Question  string                 `json:"question" db:"question"`
	Answer    string                 `json:"answer" db:"answer"`
	Sources   map[string]interface{} `json:"sources" db:"sources"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// DocumentChunk represents a chunk of text from a document
type DocumentChunk struct {
	ID         string
	DocumentID string
	Content    string
	Page       int
	ChunkIndex int
	Metadata   map[string]interface{}
}

// VectorPoint represents a point in the vector database
type VectorPoint struct {
	ID      string
	Vector  []float32
	Payload map[string]interface{}
}
