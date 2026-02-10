package embeddings

import (
	"context"
	"fmt"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/pkoukk/tiktoken-go"
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
func (g *OpenAIGenerator) Embed(ctx context.Context, text string) ([]float32, error) {
	// Set default timeout if context doesn't have one
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	// Count actual tokens using tiktoken
	tokenCount, err := g.CountTokens(ctx, text)
	if err != nil {
		// Fallback to character-based estimation if tokenization fails
		maxChars := g.MaxContextLength() * 4
		if len(text) > maxChars {
			text = text[:maxChars]
		}
	} else {
		// Use real token count with 95% safety factor
		maxTokens := int(float64(g.MaxContextLength()) * 0.95)
		if tokenCount > maxTokens {
			// Truncate to token limit
			text = g.truncateToTokenLimit(ctx, text, maxTokens)
		}
	}

	// Create embedding request with context
	resp, err := g.client.CreateEmbeddings(
		ctx,
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

// MaxContextLength returns the maximum context length in tokens
func (g *OpenAIGenerator) MaxContextLength() int {
	// All OpenAI embedding models have 8191 token limit
	return 8191
}

// CountTokens counts the actual number of tokens using tiktoken
// Note: tiktoken is synchronous and doesn't use context, but we accept it for interface consistency
func (g *OpenAIGenerator) CountTokens(ctx context.Context, text string) (int, error) {
	// Get the appropriate encoding for the model
	// cl100k_base is used by text-embedding-3-* and text-embedding-ada-002
	encoding := "cl100k_base"

	tke, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return -1, fmt.Errorf("failed to get tiktoken encoding: %w", err)
	}

	tokens := tke.Encode(text, nil, nil)
	return len(tokens), nil
}

// truncateToTokenLimit truncates text to fit within token limit
func (g *OpenAIGenerator) truncateToTokenLimit(ctx context.Context, text string, maxTokens int) string {
	encoding := "cl100k_base"
	tke, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		// Fallback to simple truncation
		return text[:min(len(text), maxTokens*4)]
	}

	tokens := tke.Encode(text, nil, nil)
	if len(tokens) <= maxTokens {
		return text
	}

	// Decode only the allowed tokens back to text
	truncatedTokens := tokens[:maxTokens]
	return tke.Decode(truncatedTokens)
}

// Name returns the name of the embedding model
func (g *OpenAIGenerator) Name() string {
	return fmt.Sprintf("OpenAI %s", g.model)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
