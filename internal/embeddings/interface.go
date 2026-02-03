package embeddings

// Generator generates embeddings for text
type Generator interface {
	// Embed generates an embedding vector for the given text
	Embed(text string) ([]float32, error)

	// Dimension returns the dimension of the embedding vectors
	Dimension() int

	// Name returns the name of the embedding model
	Name() string
}
