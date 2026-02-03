package embeddings

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
)

// StubGenerator is a placeholder embedding generator for MVP
// It generates deterministic vectors based on content hash
// TODO: Replace with real embedding model (OpenAI, local model, etc.)
type StubGenerator struct {
	dimension int
}

// NewStubGenerator creates a new stub embedding generator
func NewStubGenerator(dimension int) *StubGenerator {
	if dimension == 0 {
		dimension = 1536 // Default to OpenAI ada-002 dimension
	}
	return &StubGenerator{
		dimension: dimension,
	}
}

// Embed generates a deterministic embedding vector from text
// This is a PLACEHOLDER implementation that:
// 1. Hashes the input text
// 2. Uses the hash bytes to seed a deterministic vector
// 3. Normalizes the vector
//
// WARNING: This does NOT capture semantic meaning!
// It only provides consistent vectors for identical text.
func (g *StubGenerator) Embed(text string) ([]float32, error) {
	// Hash the text to get deterministic bytes
	hash := sha256.Sum256([]byte(text))

	// Generate vector from hash
	vector := make([]float32, g.dimension)

	// Use hash bytes to seed the vector generation
	seed := binary.BigEndian.Uint64(hash[:8])

	// Simple deterministic pseudo-random generation
	for i := 0; i < g.dimension; i++ {
		// Generate deterministic value using hash and position
		seed = seed*1103515245 + 12345 // Linear congruential generator
		value := float32(seed % 10000) / 10000.0
		// Center around 0
		vector[i] = (value - 0.5) * 2.0
	}

	// Normalize the vector (important for cosine similarity)
	magnitude := float32(0.0)
	for _, v := range vector {
		magnitude += v * v
	}
	magnitude = float32(math.Sqrt(float64(magnitude)))

	if magnitude > 0 {
		for i := range vector {
			vector[i] /= magnitude
		}
	}

	return vector, nil
}

// Dimension returns the dimension of the embedding vectors
func (g *StubGenerator) Dimension() int {
	return g.dimension
}

// Name returns the name of the embedding model
func (g *StubGenerator) Name() string {
	return "stub-hash-based (PLACEHOLDER - NO SEMANTIC MEANING)"
}
