package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/PuvaanRaaj/personal-rag-agent/internal/model"
)

// DocumentRepository handles document data operations
type DocumentRepository struct {
	db *sql.DB
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create creates a new document record
func (r *DocumentRepository) Create(ctx context.Context, doc *model.Document) error {
	query := `
		INSERT INTO documents (user_id, filename, file_type, file_size, file_hash, storage_path, total_chunks)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, upload_date
	`

	err := r.db.QueryRowContext(ctx, query,
		doc.UserID, doc.Filename, doc.FileType, doc.FileSize,
		doc.FileHash, doc.StoragePath, doc.TotalChunks).
		Scan(&doc.ID, &doc.UploadDate)

	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

// GetByID retrieves a document by ID
func (r *DocumentRepository) GetByID(ctx context.Context, id string) (*model.Document, error) {
	var doc model.Document
	query := `
		SELECT id, user_id, filename, file_type, file_size, file_hash, storage_path, total_chunks, upload_date
		FROM documents WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&doc.ID, &doc.UserID, &doc.Filename, &doc.FileType, &doc.FileSize,
		&doc.FileHash, &doc.StoragePath, &doc.TotalChunks, &doc.UploadDate,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &doc, nil
}

// ListByUserID lists all documents for a user
func (r *DocumentRepository) ListByUserID(ctx context.Context, userID string) ([]*model.Document, error) {
	query := `
		SELECT id, user_id, filename, file_type, file_size, file_hash, storage_path, total_chunks, upload_date
		FROM documents
		WHERE user_id = $1
		ORDER BY upload_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var documents []*model.Document
	for rows.Next() {
		var doc model.Document
		err := rows.Scan(
			&doc.ID, &doc.UserID, &doc.Filename, &doc.FileType, &doc.FileSize,
			&doc.FileHash, &doc.StoragePath, &doc.TotalChunks, &doc.UploadDate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, &doc)
	}

	return documents, nil
}

// Delete deletes a document
func (r *DocumentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM documents WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

// SaveQueryHistory saves a query to history
func (r *DocumentRepository) SaveQueryHistory(ctx context.Context, userID, question, answer string, sources map[string]interface{}) error {
	sourcesJSON, err := json.Marshal(sources)
	if err != nil {
		return fmt.Errorf("failed to marshal sources: %w", err)
	}

	query := `INSERT INTO query_history (user_id, question, answer, sources) VALUES ($1, $2, $3, $4)`

	_, err = r.db.ExecContext(ctx, query, userID, question, answer, sourcesJSON)
	if err != nil {
		return fmt.Errorf("failed to save query history: %w", err)
	}

	return nil
}
