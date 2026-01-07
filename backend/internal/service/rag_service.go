package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PuvaanRaaj/personal-rag-agent/internal/logger"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/repository"
)

// RAGService handles RAG query operations
type RAGService struct {
	vectorRepo       *repository.VectorRepository
	embeddingService *EmbeddingService
	documentRepo     *repository.DocumentRepository
	llmAPIKey        string
	httpClient       *http.Client
}

// NewRAGService creates a new RAG service
func NewRAGService(
	vectorRepo *repository.VectorRepository,
	embeddingService *EmbeddingService,
	llmAPIKey string,
	documentRepo *repository.DocumentRepository,
) *RAGService {
	return &RAGService{
		vectorRepo:       vectorRepo,
		embeddingService: embeddingService,
		documentRepo:     documentRepo,
		llmAPIKey:        llmAPIKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// QueryRequest represents a RAG query request
type QueryRequest struct {
	Question string `json:"question"`
}

// QueryResponse represents a RAG query response
type QueryResponse struct {
	Answer  string                   `json:"answer"`
	Sources []map[string]interface{} `json:"sources"`
}

// ChatCompletionRequest represents an OpenAI chat completion request
type ChatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents an OpenAI chat completion response
type ChatCompletionResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

// Query performs a RAG query
func (s *RAGService) Query(ctx context.Context, userID, question string) (*QueryResponse, error) {
	// 1. Generate embedding for the question
	questionEmbedding, err := s.embeddingService.GenerateEmbedding(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("failed to generate question embedding: %w", err)
	}

	// 2. Search for similar chunks
	results, err := s.vectorRepo.Search(ctx, userID, questionEmbedding, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	// 3. Build context from results
	var contextChunks []string
	var sources []map[string]interface{}

	for _, result := range results {
		if content, ok := result.Payload["content"].(string); ok {
			contextChunks = append(contextChunks, content)
		}

		// Extract source metadata
		source := map[string]interface{}{
			"filename": result.Payload["filename"],
			"page":     result.Payload["page"],
		}
		sources = append(sources, source)
	}

	// 4. Build prompt with context
	systemPrompt := `You are a helpful AI assistant with access to the user's uploaded documents.

Your role is to:
1. Answer questions accurately using information from the provided context
2. Cite specific sources when providing information
3. Be concise and actionable in your responses
4. If the information isn't in the context, clearly state that

CRITICAL: Base your answer ONLY on the provided context. Do not use external knowledge.`

	contextText := ""
	for i, chunk := range contextChunks {
		contextText += fmt.Sprintf("\n[Document %d]: %s\n", i+1, chunk)
	}

	userPrompt := fmt.Sprintf("Context from user's documents:\n%s\n\nQuestion: %s\n\nAnswer based on the above context:", contextText, question)

	// 5. Call LLM
	answer, err := s.callLLM(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	// 6. Save to query history
	if err := s.documentRepo.SaveQueryHistory(ctx, userID, question, answer, map[string]interface{}{
		"sources": sources,
	}); err != nil {
		// Log error but don't fail the request
		logger.Error("Failed to save query history",
			"user_id", userID,
			"error", err,
		)
	}

	return &QueryResponse{
		Answer:  answer,
		Sources: sources,
	}, nil
}

// callLLM calls the OpenAI API for chat completion
func (s *RAGService) callLLM(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	requestBody := ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.llmAPIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var completionResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completionResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(completionResp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return completionResp.Choices[0].Message.Content, nil
}
