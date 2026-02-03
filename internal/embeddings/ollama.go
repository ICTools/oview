package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// Embed generates an embedding vector for the given text
func (g *OllamaGenerator) Embed(text string) ([]float32, error) {
	// Truncate text if too long
	if len(text) > 30000 {
		text = text[:30000]
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

	// Make HTTP request
	url := fmt.Sprintf("%s/api/embeddings", g.baseURL)
	resp, err := g.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
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

// Name returns the name of the embedding model
func (g *OllamaGenerator) Name() string {
	return fmt.Sprintf("Ollama %s", g.model)
}
