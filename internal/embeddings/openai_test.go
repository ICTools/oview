package embeddings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenAIGenerator_Dimension(t *testing.T) {
	tests := []struct {
		model       string
		expectedDim int
	}{
		{"text-embedding-3-small", 1536},
		{"text-embedding-ada-002", 1536},
		{"text-embedding-3-large", 3072},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			gen := NewOpenAIGenerator("fake-key", tt.model)
			assert.Equal(t, tt.expectedDim, gen.Dimension())
		})
	}
}

func TestOpenAIGenerator_MaxContextLength(t *testing.T) {
	gen := NewOpenAIGenerator("fake-key", "text-embedding-3-small")
	assert.Equal(t, 8191, gen.MaxContextLength())
}

func TestOpenAIGenerator_Name(t *testing.T) {
	tests := []struct {
		model        string
		expectedName string
	}{
		{"text-embedding-3-small", "OpenAI text-embedding-3-small"},
		{"text-embedding-ada-002", "OpenAI text-embedding-ada-002"},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			gen := NewOpenAIGenerator("fake-key", tt.model)
			assert.Equal(t, tt.expectedName, gen.Name())
		})
	}
}

func TestOpenAIGenerator_CountTokens(t *testing.T) {
	gen := NewOpenAIGenerator("fake-key", "text-embedding-3-small")

	tests := []struct {
		name string
		text string
		// Note: exact token counts depend on tiktoken, these are approximate
		minTokens int
		maxTokens int
	}{
		{"empty", "", 0, 0},
		{"simple", "Hello", 1, 2},
		{"sentence", "Hello world", 2, 3},
		{"longer", "This is a longer sentence with more tokens", 7, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := gen.CountTokens(context.Background(), tt.text)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, tokens, tt.minTokens)
			assert.LessOrEqual(t, tokens, tt.maxTokens)
		})
	}
}

func TestNewOpenAIGenerator_DefaultModel(t *testing.T) {
	// Test with empty model string - should use default
	gen := NewOpenAIGenerator("fake-key", "")
	assert.NotNil(t, gen)
	// Default model is text-embedding-3-small
	assert.Equal(t, 1536, gen.Dimension())
}
