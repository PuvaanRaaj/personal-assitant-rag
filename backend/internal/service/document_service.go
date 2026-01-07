package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/PuvaanRaaj/personal-rag-agent/internal/model"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/repository"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/storage"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/utils"
)

// DocumentService handles document operations
type DocumentService struct {
	documentRepo     *repository.DocumentRepository
	vectorRepo       *repository.VectorRepository
	storageDriver    storage.StorageDriver
	embeddingService *EmbeddingService
}

// NewDocumentService creates a new document service
func NewDocumentService(
	documentRepo *repository.DocumentRepository,
	vectorRepo *repository.VectorRepository,
	storageDriver storage.StorageDriver,
	embeddingService *EmbeddingService,
) *DocumentService {
	return &DocumentService{
		documentRepo:     documentRepo,
		vectorRepo:       vectorRepo,
		storageDriver:    storageDriver,
		embeddingService: embeddingService,
	}
}

// UploadDocument handles document upload and processing
func (s *DocumentService) UploadDocument(ctx context.Context, userID string, file *multipart.FileHeader) (*model.Document, error) {
	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedTypes := map[string]bool{
		".pdf": true, ".txt": true, ".md": true,
		".json": true, ".csv": true,
	}
	if !allowedTypes[ext] {
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	// Validate file size (10MB max)
	const maxSize = 10 * 1024 * 1024
	if file.Size > maxSize {
		return nil, fmt.Errorf("file too large (max 10MB)")
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Read file content
	content, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Calculate hash
	hash := sha256.Sum256(content)
	fileHash := hex.EncodeToString(hash[:])

	// Extract text based on file type
	var text string
	switch ext {
	case ".txt", ".md":
		text = string(content)
	case ".json", ".csv":
		text = string(content)
	case ".pdf":
		// Simple text extraction (for now, treat as text)
		// TODO: Add proper PDF parsing with unipdf
		text = string(content)
	default:
		return nil, fmt.Errorf("unsupported file type")
	}

	// Chunk the text
	chunks := utils.ChunkText(text, 500, 50)
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no text content found in document")
	}

	// Generate embeddings
	embeddings, err := s.embeddingService.GenerateEmbeddings(ctx, chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Upload to storage
	storagePath := fmt.Sprintf("%s/%s/%s", userID, fileHash, file.Filename)
	if err := s.storageDriver.UploadFile(ctx, storagePath, bytes.NewReader(content)); err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create document record
	doc := &model.Document{
		UserID:      userID,
		Filename:    file.Filename,
		FileType:    ext,
		FileSize:    file.Size,
		FileHash:    fileHash,
		StoragePath: storagePath,
		TotalChunks: len(chunks),
	}

	if err := s.documentRepo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to create document record: %w", err)
	}

	// Ensure vector collection exists
	vectorSize := uint64(s.embeddingService.GetDimensions())
	if err := s.vectorRepo.EnsureCollection(ctx, userID, vectorSize); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	// Store vectors
	var points []*model.VectorPoint
	for i, embedding := range embeddings {
		point := &model.VectorPoint{
			ID:     fmt.Sprintf("%s_chunk_%d", doc.ID, i),
			Vector: embedding,
			Payload: map[string]interface{}{
				"document_id": doc.ID,
				"user_id":     userID,
				"filename":    file.Filename,
				"file_type":   ext,
				"chunk_index": i,
				"content":     chunks[i],
			},
		}
		points = append(points, point)
	}

	if err := s.vectorRepo.InsertVectors(ctx, userID, points); err != nil {
		return nil, fmt.Errorf("failed to insert vectors: %w", err)
	}

	return doc, nil
}

// ListDocuments lists all documents for a user
func (s *DocumentService) ListDocuments(ctx context.Context, userID string) ([]*model.Document, error) {
	return s.documentRepo.ListByUserID(ctx, userID)
}

// GetDocument gets a single document
func (s *DocumentService) GetDocument(ctx context.Context, userID, documentID string) (*model.Document, error) {
	doc, err := s.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if doc.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	return doc, nil
}

// DeleteDocument deletes a document and its vectors
func (s *DocumentService) DeleteDocument(ctx context.Context, userID, documentID string) error {
	// Get document
	doc, err := s.GetDocument(ctx, userID, documentID)
	if err != nil {
		return err
	}

	// Delete from storage
	if err := s.storageDriver.DeleteFile(ctx, doc.StoragePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Delete vectors
	if err := s.vectorRepo.DeleteByDocumentID(ctx, userID, documentID); err != nil {
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	// Delete database record
	if err := s.documentRepo.Delete(ctx, documentID); err != nil {
		return fmt.Errorf("failed to delete document record: %w", err)
	}

	return nil
}
