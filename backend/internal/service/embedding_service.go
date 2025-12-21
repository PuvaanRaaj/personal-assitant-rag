package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EmbeddingService handles embedding generation
type EmbeddingService struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(apiKey string) *EmbeddingService {
	return &EmbeddingService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		model: "text-embedding-3-small",
	}
}

// EmbeddingRequest represents an OpenAI embedding request
type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

// EmbeddingResponse represents an OpenAI embedding response
type EmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// GenerateEmbedding generates an embedding for a single text
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := s.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embeddings[0], nil
}

// GenerateEmbeddings generates embeddings for multiple texts (batch processing)
func (s *EmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// OpenAI supports up to 8191 tokens per request, we'll batch by count
	const batchSize = 100
	var allEmbeddings [][]float32

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := s.generateBatch(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to generate batch %d: %w", i/batchSize, err)
		}

		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	return allEmbeddings, nil
}

// generateBatch generates embeddings for a batch of texts
func (s *EmbeddingService) generateBatch(ctx context.Context, texts []string) ([][]float32, error) {
	requestBody := EmbeddingRequest{
		Input: texts,
		Model: s.model,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Retry logic with exponential backoff
	var resp *http.Response
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = s.httpClient.Do(req)
		if err == nil && resp.StatusCode == 200 {
			break
		}

		if resp != nil && resp.StatusCode == 429 {
			// Rate limited, wait and retry
			waitTime := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(waitTime)
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
	}

	if resp == nil {
		return nil, fmt.Errorf("no response after %d retries", maxRetries)
	}
	defer resp.Body.Close()

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract embeddings in order
	embeddings := make([][]float32, len(embeddingResp.Data))
	for _, data := range embeddingResp.Data {
		if data.Index >= len(embeddings) {
			return nil, fmt.Errorf("invalid embedding index: %d", data.Index)
		}
		embeddings[data.Index] = data.Embedding
	}

	return embeddings, nil
}

// GetDimensions returns the embedding dimensions for the model
func (s *EmbeddingService) GetDimensions() int {
	// text-embedding-3-small has 1536 dimensions
	return 1536
}
