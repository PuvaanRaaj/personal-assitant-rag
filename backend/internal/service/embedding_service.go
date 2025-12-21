package service

// EmbeddingService handles embedding generation
type EmbeddingService struct {
	apiKey string
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(apiKey string) *EmbeddingService {
	return &EmbeddingService{apiKey: apiKey}
}

// TODO: Implement OpenAI embedding generation methods
