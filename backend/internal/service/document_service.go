package service

import (
	"github.com/PuvaanRaaj/personal-rag-agent/internal/repository"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/storage"
)

// DocumentService handles document operations
type DocumentService struct {
	documentRepo     *repository.DocumentRepository
	vectorRepo       *repository.VectorRepository
	s3Client         *storage.S3Client
	embeddingService *EmbeddingService
}

// NewDocumentService creates a new document service
func NewDocumentService(
	documentRepo *repository.DocumentRepository,
	vectorRepo *repository.VectorRepository,
	s3Client *storage.S3Client,
	embeddingService *EmbeddingService,
) *DocumentService {
	return &DocumentService{
		documentRepo:     documentRepo,
		vectorRepo:       vectorRepo,
		s3Client:         s3Client,
		embeddingService: embeddingService,
	}
}

// TODO: Implement document processing methods
