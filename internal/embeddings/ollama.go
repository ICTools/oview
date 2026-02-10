package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaGenerator generates embeddings using local Ollama API
type OllamaGenerator struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaGenerator creates a new Ollama embeddings generator
// baseURL: Usually "http://localhost:11434"
// model: "nomic-embed-text" (recommended) or "mxbai-embed-large"
func NewOllamaGenerator(baseURL string, model string) *OllamaGenerator {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "nomic-embed-text" // Default, 768 dimensions
	}

	return &OllamaGenerator{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

// OllamaEmbeddingRequest is the request structure for Ollama API
type OllamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// OllamaEmbeddingResponse is the response structure from Ollama API
type OllamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// OllamaTokenizeRequest is the request structure for Ollama tokenize API
type OllamaTokenizeRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// OllamaTokenizeResponse is the response structure from Ollama tokenize API
type OllamaTokenizeResponse struct {
	Tokens []int `json:"tokens"`
}

// Embed generates an embedding vector for the given text
func (g *OllamaGenerator) Embed(ctx context.Context, text string) ([]float32, error) {
	// Set default timeout if context doesn't have one
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	// Count actual tokens using Ollama's tokenizer
	tokenCount, err := g.CountTokens(ctx, text)
	if err != nil {
		// Fallback to character-based estimation if tokenization fails
		maxChars := int(float64(g.MaxContextLength()) * 0.5 * 1.5)
		if len(text) >= maxChars {
			text = text[:maxChars]
		}
	} else {
		// Use real token count with 90% safety factor (more precise)
		maxTokens := int(float64(g.MaxContextLength()) * 0.9)
		if tokenCount > maxTokens {
			// Binary search to find the right length
			text = g.truncateToTokenLimit(ctx, text, maxTokens)
		}
	}

	// Create request
	reqBody := OllamaEmbeddingRequest{
		Model:  g.model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request with context
	url := fmt.Sprintf("%s/api/embeddings", g.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ollama API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert []float64 to []float32
	embedding := make([]float32, len(ollamaResp.Embedding))
	for i, v := range ollamaResp.Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// Dimension returns the dimension of the embedding vectors
func (g *OllamaGenerator) Dimension() int {
	// nomic-embed-text: 768 dimensions
	// mxbai-embed-large: 1024 dimensions
	// all-minilm: 384 dimensions
	switch g.model {
	case "nomic-embed-text":
		return 768
	case "mxbai-embed-large":
		return 1024
	case "all-minilm":
		return 384
	default:
		return 768 // Default
	}
}

// MaxContextLength returns the maximum context length in tokens
func (g *OllamaGenerator) MaxContextLength() int {
	// Different Ollama models have different context limits
	switch g.model {
	case "nomic-embed-text":
		return 8192 // 8192 tokens
	case "mxbai-embed-large":
		return 512 // 512 tokens
	case "all-minilm":
		return 256 // 256 tokens
	default:
		return 2048 // Conservative default
	}
}

// CountTokens counts the actual number of tokens using Ollama's tokenizer
func (g *OllamaGenerator) CountTokens(ctx context.Context, text string) (int, error) {
	// Set default timeout if context doesn't have one
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	// Create request
	reqBody := OllamaTokenizeRequest{
		Model:  g.model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return -1, fmt.Errorf("failed to marshal tokenize request: %w", err)
	}

	// Make HTTP request to /api/tokenize with context
	url := fmt.Sprintf("%s/api/tokenize", g.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return -1, fmt.Errorf("failed to create tokenize request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return -1, fmt.Errorf("Ollama tokenize API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return -1, fmt.Errorf("Ollama tokenize API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenizeResp OllamaTokenizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenizeResp); err != nil {
		return -1, fmt.Errorf("failed to decode tokenize response: %w", err)
	}

	return len(tokenizeResp.Tokens), nil
}

// truncateToTokenLimit truncates text to fit within token limit using binary search
func (g *OllamaGenerator) truncateToTokenLimit(ctx context.Context, text string, maxTokens int) string {
	if len(text) == 0 {
		return text
	}

	// Binary search for the right character length
	left, right := 0, len(text)
	result := text

	for left < right {
		mid := (left + right + 1) / 2
		tokens, err := g.CountTokens(ctx, text[:mid])
		if err != nil {
			// Fallback to simple truncation if tokenization fails
			return text[:mid]
		}

		if tokens <= maxTokens {
			result = text[:mid]
			left = mid
		} else {
			right = mid - 1
		}
	}

	return result
}

// Name returns the name of the embedding model
func (g *OllamaGenerator) Name() string {
	return fmt.Sprintf("Ollama %s", g.model)
}
