package service

import (
	"github.com/PuvaanRaaj/personal-rag-agent/internal/repository"
)

// RAGService handles RAG query operations
type RAGService struct {
	vectorRepo       *repository.VectorRepository
	embeddingService *EmbeddingService
	llmAPIKey        string
}

// NewRAGService creates a new RAG service
func NewRAGService(
	vectorRepo *repository.VectorRepository,
	embeddingService *EmbeddingService,
	llmAPIKey string,
) *RAGService {
	return &RAGService{
		vectorRepo:       vectorRepo,
		embeddingService: embeddingService,
		llmAPIKey:        llmAPIKey,
	}
}

// TODO: Implement RAG query methods
