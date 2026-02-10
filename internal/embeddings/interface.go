package embeddings

import "context"

// Generator generates embeddings for text
type Generator interface {
	// Embed generates an embedding vector for the given text
	// Context can be used to cancel the operation or set timeouts
	Embed(ctx context.Context, text string) ([]float32, error)

	// Dimension returns the dimension of the embedding vectors
	Dimension() int

	// MaxContextLength returns the maximum context length in tokens
	MaxContextLength() int

	// CountTokens counts the actual number of tokens in the text using the model's tokenizer
	// Returns -1 if token counting is not supported (fallback to estimation)
	// Context can be used to cancel the operation or set timeouts
	CountTokens(ctx context.Context, text string) (int, error)

	// Name returns the name of the embedding model
	Name() string
}
