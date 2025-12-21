package utils

import (
	"strings"
	"unicode"
)

// ChunkText splits text into chunks with overlap
func ChunkText(text string, chunkSize, overlap int) []string {
	// Tokenize by words (simple approximation)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r)
	})

	if len(words) == 0 {
		return nil
	}

	var chunks []string
	for i := 0; i < len(words); i += chunkSize - overlap {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}

		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)

		// If this was the last chunk, break
		if end == len(words) {
			break
		}
	}

	return chunks
}

// EstimateTokens estimates the number of tokens in text (rough approximation)
func EstimateTokens(text string) int {
	// Rough approximation: 1 token ≈ 4 characters
	return len(text) / 4
}
