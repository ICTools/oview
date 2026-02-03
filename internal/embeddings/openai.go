package embeddings

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIGenerator generates embeddings using OpenAI API
type OpenAIGenerator struct {
	client *openai.Client
	model  openai.EmbeddingModel
}

// NewOpenAIGenerator creates a new OpenAI embeddings generator
// apiKey: Your OpenAI API key
// model: "text-embedding-3-small" (recommended, $0.02/1M tokens) or "text-embedding-ada-002"
func NewOpenAIGenerator(apiKey string, model string) *OpenAIGenerator {
	var embeddingModel openai.EmbeddingModel
	if model == "" {
		embeddingModel = openai.SmallEmbedding3 // text-embedding-3-small (default)
	} else {
		embeddingModel = openai.EmbeddingModel(model)
	}

	return &OpenAIGenerator{
		client: openai.NewClient(apiKey),
		model:  embeddingModel,
	}
}

// Embed generates an embedding vector for the given text
func (g *OpenAIGenerator) Embed(text string) ([]float32, error) {
	// Truncate text if too long (OpenAI limit: 8191 tokens ≈ 30k chars)
	if len(text) > 30000 {
		text = text[:30000]
	}

	// Create embedding request
	resp, err := g.client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequestStrings{
			Input: []string{text},
			Model: g.model,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned from OpenAI")
	}

	// Convert []float64 to []float32
	embedding := make([]float32, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// Dimension returns the dimension of the embedding vectors
func (g *OpenAIGenerator) Dimension() int {
	// text-embedding-3-small: 1536 dimensions
	// text-embedding-ada-002: 1536 dimensions
	// text-embedding-3-large: 3072 dimensions
	if g.model == "text-embedding-3-large" {
		return 3072
	}
	return 1536
}

// Name returns the name of the embedding model
func (g *OpenAIGenerator) Name() string {
	return fmt.Sprintf("OpenAI %s", g.model)
}
