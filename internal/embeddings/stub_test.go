package embeddings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStubGenerator_Embed(t *testing.T) {
	gen := NewStubGenerator(1536)

	tests := []struct {
		name string
		text string
	}{
		{"simple text", "Hello world"},
		{"empty text", ""},
		{"long text", "This is a longer text that should still produce a deterministic embedding vector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embedding1, err1 := gen.Embed(tt.text)
			embedding2, err2 := gen.Embed(tt.text)

			// Should not error
			assert.NoError(t, err1)
			assert.NoError(t, err2)

			// Should return same dimension
			assert.Equal(t, 1536, len(embedding1))
			assert.Equal(t, 1536, len(embedding2))

			// Should be deterministic (same text = same embedding)
			assert.Equal(t, embedding1, embedding2)

			// Should be normalized (magnitude ≈ 1.0)
			magnitude := float32(0.0)
			for _, v := range embedding1 {
				magnitude += v * v
			}
			assert.InDelta(t, 1.0, magnitude, 0.001, "embedding should be normalized")
		})
	}
}

func TestStubGenerator_Embed_DifferentTexts(t *testing.T) {
	gen := NewStubGenerator(1536)

	emb1, _ := gen.Embed("Hello")
	emb2, _ := gen.Embed("World")

	// Different texts should produce different embeddings
	assert.NotEqual(t, emb1, emb2)
}

func TestStubGenerator_Dimension(t *testing.T) {
	tests := []int{384, 768, 1536, 3072}

	for _, dim := range tests {
		t.Run("dimension_"+string(rune(dim)), func(t *testing.T) {
			gen := NewStubGenerator(dim)
			assert.Equal(t, dim, gen.Dimension())

			embedding, _ := gen.Embed("test")
			assert.Equal(t, dim, len(embedding))
		})
	}
}

func TestStubGenerator_MaxContextLength(t *testing.T) {
	gen := NewStubGenerator(1536)
	assert.Equal(t, 8192, gen.MaxContextLength())
}

func TestStubGenerator_CountTokens(t *testing.T) {
	gen := NewStubGenerator(1536)

	tests := []struct {
		text           string
		expectedTokens int
	}{
		{"", 0},
		{"Hello", 1},              // 5 chars / 4 = 1
		{"Hello world", 2},        // 11 chars / 4 = 2
		{"This is a test", 3},     // 14 chars / 4 = 3
		{"A longer sentence here", 5}, // 23 chars / 4 = 5
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			tokens, err := gen.CountTokens(tt.text)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTokens, tokens)
		})
	}
}

func TestStubGenerator_Name(t *testing.T) {
	gen := NewStubGenerator(1536)
	assert.Contains(t, gen.Name(), "stub")
	assert.Contains(t, gen.Name(), "PLACEHOLDER")
}
