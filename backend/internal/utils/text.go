package utils

import (
	"strings"
	"unicode"
)

// ChunkText splits text into chunks with overlap, trying to break at natural boundaries
func ChunkText(text string, chunkSize, overlap int) []string {
	if len(text) == 0 {
		return nil
	}

	// If text is smaller than chunk size, return as is
	if len(text) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0
	
	for start < len(text) {
		end := start + chunkSize
		if end >= len(text) {
			chunks = append(chunks, text[start:])
			break
		}

		// Look for a good break point (newline, period, space) within the last 20% of the chunk
		breakPoint := end
		searchRange := chunkSize / 5
		if searchRange < 20 {
			searchRange = 20
		}
		
		found := false
		for _, sep := range []string{"\n\n", "\n", ". ", " "} {
			if idx := strings.LastIndex(text[start+chunkSize-searchRange:end], sep); idx != -1 {
				breakPoint = start + chunkSize - searchRange + idx + len(sep)
				found = true
				break
			}
		}

		if !found {
			// If no natural separator found, just break at chunkSize
			breakPoint = end
		}

		chunks = append(chunks, text[start:breakPoint])
		
		// Move start forward, accounting for overlap
		start = breakPoint - overlap
		if start < 0 {
			start = 0
		}
		
		// Avoid infinite loops if breakPoint doesn't move forward
		if start >= breakPoint {
			start = breakPoint
		}
	}

	return chunks
}

// EstimateTokens estimates the number of tokens in text (rough approximation)
func EstimateTokens(text string) int {
	// Rough approximation: 1 token ≈ 4 characters
	return len(text) / 4
}
