package embeddings

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOllamaGenerator_CountTokens(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		mockResponse   string
		expectedTokens int
		expectError    bool
	}{
		{
			name:           "simple text",
			text:           "Hello world",
			mockResponse:   `{"tokens":[1,2,3,4]}`,
			expectedTokens: 4,
			expectError:    false,
		},
		{
			name:           "empty text",
			text:           "",
			mockResponse:   `{"tokens":[]}`,
			expectedTokens: 0,
			expectError:    false,
		},
		{
			name:           "long text",
			text:           "This is a longer text that should be tokenized into many tokens by the model",
			mockResponse:   `{"tokens":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15]}`,
			expectedTokens: 15,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/tokenize", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Create generator with mock URL
			gen := NewOllamaGenerator(server.URL, "nomic-embed-text")

			// Test
			tokens, err := gen.CountTokens(tt.text)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTokens, tokens)
			}
		})
	}
}

func TestOllamaGenerator_Embed(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		mockEmbedding  string
		mockTokenize   string
		expectTruncate bool
		expectError    bool
	}{
		{
			name:          "normal text",
			text:          "Hello world",
			mockEmbedding: `{"embedding":[0.1,0.2,0.3]}`,
			mockTokenize:  `{"tokens":[1,2,3]}`,
			expectError:   false,
		},
		{
			name:           "text exceeding limit",
			text:           "This is a very long text that exceeds the token limit and should be truncated before being sent to the embedding API",
			mockEmbedding:  `{"embedding":[0.1,0.2,0.3]}`,
			mockTokenize:   `{"tokens":[` + generateTokenArray(10000) + `]}`,
			expectTruncate: true,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				if r.URL.Path == "/api/tokenize" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.mockTokenize))
				} else if r.URL.Path == "/api/embeddings" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.mockEmbedding))
				}
			}))
			defer server.Close()

			gen := NewOllamaGenerator(server.URL, "nomic-embed-text")

			embedding, err := gen.Embed(tt.text)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, embedding)
				assert.Equal(t, 3, len(embedding))
			}
		})
	}
}

func TestOllamaGenerator_Dimension(t *testing.T) {
	tests := []struct {
		model         string
		expectedDim   int
	}{
		{"nomic-embed-text", 768},
		{"mxbai-embed-large", 1024},
		{"all-minilm", 384},
		{"unknown-model", 768}, // default
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			gen := NewOllamaGenerator("http://localhost:11434", tt.model)
			assert.Equal(t, tt.expectedDim, gen.Dimension())
		})
	}
}

func TestOllamaGenerator_MaxContextLength(t *testing.T) {
	tests := []struct {
		model       string
		expectedMax int
	}{
		{"nomic-embed-text", 8192},
		{"mxbai-embed-large", 512},
		{"all-minilm", 256},
		{"unknown-model", 2048}, // default
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			gen := NewOllamaGenerator("http://localhost:11434", tt.model)
			assert.Equal(t, tt.expectedMax, gen.MaxContextLength())
		})
	}
}

func TestOllamaGenerator_Name(t *testing.T) {
	gen := NewOllamaGenerator("http://localhost:11434", "nomic-embed-text")
	assert.Equal(t, "Ollama nomic-embed-text", gen.Name())
}

// Helper function to generate token array string
func generateTokenArray(count int) string {
	result := ""
	for i := 0; i < count; i++ {
		if i > 0 {
			result += ","
		}
		result += "1"
	}
	return result
}
