package indexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/oview/internal/config"
)

func TestChunker_detectLanguage(t *testing.T) {
	c := NewChunker(&config.RAGConfig{})

	tests := []struct {
		ext      string
		basename string
		expected string
	}{
		{".py", "main.py", "python"},
		{".js", "app.js", "javascript"},
		{".jsx", "App.jsx", "javascript"},
		{".ts", "main.ts", "typescript"},
		{".tsx", "App.tsx", "typescript"},
		{".go", "main.go", "go"},
		{".php", "Controller.php", "php"},
		{".java", "Main.java", "java"},
		{".rs", "main.rs", "rust"},
		{".c", "main.c", "c"},
		{".h", "header.h", "c"},
		{".cpp", "main.cpp", "cpp"},
		{".hpp", "header.hpp", "cpp"},
		{".cc", "main.cc", "cpp"},
		{".cxx", "main.cxx", "cpp"},
		{".hxx", "header.hxx", "cpp"},
		{".rb", "app.rb", "ruby"},
		{".cs", "Program.cs", "csharp"},
		{".txt", "README.txt", ""},
		{".md", "README.md", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := c.detectLanguage(tt.ext, tt.basename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChunker_chunkBySize(t *testing.T) {
	ragConfig := &config.RAGConfig{
		Chunking: config.ChunkingRules{
			Generic: config.ChunkRule{
				MaxSize: 100,
			},
		},
	}
	c := NewChunker(ragConfig)

	tests := []struct {
		name          string
		content       string
		maxSize       int
		expectedCount int
	}{
		{
			name:          "small content",
			content:       "Hello world",
			maxSize:       100,
			expectedCount: 1,
		},
		{
			name:          "exact size",
			content:       "x",
			maxSize:       1,
			expectedCount: 1,
		},
		{
			name:          "needs splitting",
			content:       "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10\nLine 11\nLine 12\nLine 13\nLine 14\nLine 15",
			maxSize:       50,
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := c.chunkBySize("test.txt", tt.content, tt.maxSize, "Text", "doc")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(chunks))

			// Verify all chunks have correct fields
			for i, chunk := range chunks {
				assert.Equal(t, "test.txt", chunk.Path)
				assert.Equal(t, "Text", chunk.Language)
				assert.Equal(t, "doc", chunk.Type)
				assert.NotEmpty(t, chunk.Content)
				if tt.expectedCount > 1 {
					assert.Contains(t, chunk.Symbol, "chunk-")
				}

				// Verify chunk size doesn't exceed maxSize (allow for newlines)
				if i < len(chunks)-1 {
					assert.LessOrEqual(t, len(chunk.Content), tt.maxSize+10) // +10 for newlines
				}
			}
		})
	}
}

func TestChunker_chunkPHP(t *testing.T) {
	ragConfig := &config.RAGConfig{
		Chunking: config.ChunkingRules{
			PHP: config.ChunkRule{
				MaxSize: 1000,
			},
		},
	}
	c := NewChunker(ragConfig)

	tests := []struct {
		name          string
		content       string
		expectedCount int
		expectedSymbols []string
	}{
		{
			name: "single class",
			content: `<?php
class UserController {
    public function index() {
        return "list";
    }
}`,
			expectedCount: 1,
			expectedSymbols: []string{"UserController"},
		},
		{
			name: "class with multiple methods (small class)",
			content: `<?php
class UserService {
    public function create() {
        return "create";
    }

    public function update() {
        return "update";
    }
}`,
			expectedCount: 1, // Small class is kept as one chunk
			expectedSymbols: []string{"UserService"},
		},
		{
			name: "multiple classes",
			content: `<?php
class ClassA {
    public function methodA() {}
}

class ClassB {
    public function methodB() {}
}`,
			expectedCount: 2,
			expectedSymbols: []string{"ClassA", "ClassB"}, // Each class becomes one chunk when small
		},
		{
			name: "no classes",
			content: `<?php
function standalone() {
    return "test";
}`,
			expectedCount: 1,
			expectedSymbols: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := c.chunkPHP("test.php", tt.content)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(chunks))

			if len(tt.expectedSymbols) > 0 {
				for i, symbol := range tt.expectedSymbols {
					if i < len(chunks) {
						assert.Contains(t, chunks[i].Symbol, symbol)
					}
				}
			}

			// Verify chunk properties
			for _, chunk := range chunks {
				assert.Equal(t, "test.php", chunk.Path)
				assert.Equal(t, "PHP", chunk.Language)
				assert.NotEmpty(t, chunk.Content)
			}
		})
	}
}

func TestChunker_chunkYAML(t *testing.T) {
	ragConfig := &config.RAGConfig{
		Chunking: config.ChunkingRules{
			YAML: config.ChunkRule{
				MaxSize: 500,
			},
		},
	}
	c := NewChunker(ragConfig)

	tests := []struct {
		name          string
		content       string
		expectedCount int
		expectedKeys  []string
	}{
		{
			name: "small yaml",
			content: `name: test
version: 1.0`,
			expectedCount: 1,
			expectedKeys: []string{},
		},
		{
			name: "multiple sections (small)",
			content: `database:
  host: localhost
  port: 5432

cache:
  enabled: true
  ttl: 3600`,
			expectedCount: 1, // Small YAML is kept as single chunk
			expectedKeys: []string{},
		},
		{
			name: "nested sections (small)",
			content: `services:
  web:
    image: nginx
    ports:
      - 80:80
  db:
    image: postgres
    environment:
      POSTGRES_DB: mydb`,
			expectedCount: 1, // Small YAML is kept as single chunk
			expectedKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := c.chunkYAML("config.yaml", tt.content)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(chunks), tt.expectedCount)

			// Verify chunk properties
			for _, chunk := range chunks {
				assert.Equal(t, "config.yaml", chunk.Path)
				assert.Equal(t, "YAML", chunk.Language)
				assert.NotEmpty(t, chunk.Content)
			}

			// Check for expected keys in symbols
			if len(tt.expectedKeys) > 0 {
				symbols := make([]string, len(chunks))
				for i, chunk := range chunks {
					symbols[i] = chunk.Symbol
				}
				for _, key := range tt.expectedKeys {
					found := false
					for _, symbol := range symbols {
						if symbol == key {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected key %s not found in symbols", key)
				}
			}
		})
	}
}

func TestChunker_chunkMarkdown(t *testing.T) {
	c := NewChunker(&config.RAGConfig{})

	tests := []struct {
		name          string
		content       string
		expectedCount int
		expectedHeadings []string
	}{
		{
			name: "single heading",
			content: `# Main Title
Some content here`,
			expectedCount: 1,
			expectedHeadings: []string{"Main Title"},
		},
		{
			name: "multiple headings",
			content: `# Introduction
Intro text

## Section 1
Section 1 content

## Section 2
Section 2 content`,
			expectedCount: 3,
			expectedHeadings: []string{"Introduction", "Section 1", "Section 2"},
		},
		{
			name: "nested headings",
			content: `# Title
Content

## Subsection
More content

### Sub-subsection
Even more content`,
			expectedCount: 3,
			expectedHeadings: []string{"Title", "Subsection", "Sub-subsection"},
		},
		{
			name: "no headings",
			content: `Just plain text
No headings here`,
			expectedCount: 1,
			expectedHeadings: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := c.chunkMarkdown("README.md", tt.content)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(chunks))

			// Verify chunk properties
			for _, chunk := range chunks {
				assert.Equal(t, "README.md", chunk.Path)
				assert.Equal(t, "Markdown", chunk.Language)
				assert.Equal(t, "doc", chunk.Type)
				assert.NotEmpty(t, chunk.Content)
			}

			// Check headings
			if len(tt.expectedHeadings) > 0 {
				for i, heading := range tt.expectedHeadings {
					if i < len(chunks) {
						assert.Contains(t, chunks[i].Symbol, heading)
					}
				}
			}
		})
	}
}

func TestChunker_chunkMakefile(t *testing.T) {
	c := NewChunker(&config.RAGConfig{})

	content := `build:
	go build -o app

test:
	go test ./...

clean:
	rm -rf dist/`

	chunks, err := c.chunkMakefile("Makefile", content)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(chunks))

	expectedTargets := []string{"build", "test", "clean"}
	for i, chunk := range chunks {
		assert.Equal(t, "Makefile", chunk.Path)
		assert.Equal(t, "Makefile", chunk.Language)
		assert.Equal(t, "config", chunk.Type)
		assert.Equal(t, "build", chunk.Component)
		assert.Equal(t, expectedTargets[i], chunk.Symbol)
		assert.NotEmpty(t, chunk.Content)
	}
}

func TestChunker_chunkDockerCompose(t *testing.T) {
	c := NewChunker(&config.RAGConfig{})

	content := `version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
  db:
    image: postgres:13
    environment:
      POSTGRES_PASSWORD: secret`

	chunks, err := c.chunkDockerCompose("docker-compose.yml", content)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(chunks))

	expectedServices := []string{"web", "db"}
	for i, chunk := range chunks {
		assert.Equal(t, "docker-compose.yml", chunk.Path)
		assert.Equal(t, "YAML", chunk.Language)
		assert.Equal(t, "config", chunk.Type)
		assert.Equal(t, "docker", chunk.Component)
		assert.Equal(t, expectedServices[i], chunk.Symbol)
		assert.NotEmpty(t, chunk.Content)
	}
}

func TestChunker_ChunkFile(t *testing.T) {
	ragConfig := &config.RAGConfig{
		Chunking: config.ChunkingRules{
			PHP: config.ChunkRule{MaxSize: 1000},
			JavaScript: config.ChunkRule{MaxSize: 2000},
			YAML: config.ChunkRule{MaxSize: 500},
			Generic: config.ChunkRule{MaxSize: 1500},
		},
	}
	c := NewChunker(ragConfig)

	tests := []struct {
		name     string
		path     string
		content  []byte
		wantErr  bool
	}{
		{
			name:    "PHP file",
			path:    "src/Controller.php",
			content: []byte("<?php\nclass Test {}"),
			wantErr: false,
		},
		{
			name:    "JavaScript file",
			path:    "app.js",
			content: []byte("function test() {}"),
			wantErr: false,
		},
		{
			name:    "YAML file",
			path:    "config.yaml",
			content: []byte("key: value"),
			wantErr: false,
		},
		{
			name:    "Makefile",
			path:    "Makefile",
			content: []byte("build:\n\tgo build"),
			wantErr: false,
		},
		{
			name:    "Markdown file",
			path:    "README.md",
			content: []byte("# Title\nContent"),
			wantErr: false,
		},
		{
			name:    "Unknown extension",
			path:    "file.unknown",
			content: []byte("some content"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := c.ChunkFile(tt.path, tt.content)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, chunks)

				// Verify all chunks have valid fields
				for _, chunk := range chunks {
					assert.Equal(t, tt.path, chunk.Path)
					assert.NotEmpty(t, chunk.Content)
					assert.NotEmpty(t, chunk.Language)
					assert.NotEmpty(t, chunk.Type)
				}
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("getComponent", func(t *testing.T) {
		tests := []struct {
			path     string
			expected string
		}{
			{"src/Controller/UserController.php", "Controller"},
			{"internal/database/client.go", "database"},
			{"./file.txt", ""},
			{"file.txt", ""},
		}

		for _, tt := range tests {
			result := getComponent(tt.path)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("getFileType", func(t *testing.T) {
		tests := []struct {
			path     string
			expected string
		}{
			{"src/tests/UserTest.php", "test"},
			{"internal/test/helper.go", "test"},
			{"src/config/services.yaml", "config"}, // Must have /config/ in path
			{"docs/README.md", "doc"},
			{"src/Controller.php", "code"},
		}

		for _, tt := range tests {
			result := getFileType(tt.path)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("detectLanguage", func(t *testing.T) {
		tests := []struct {
			path     string
			expected string
		}{
			{"main.php", "PHP"},
			{"app.js", "JavaScript"},
			{"main.ts", "TypeScript"},
			{"App.jsx", "JSX"},
			{"Component.tsx", "TSX"},
			{"main.py", "Python"},
			{"main.go", "Go"},
			{"Main.java", "Java"},
			{"app.rb", "Ruby"},
			{"config.yaml", "YAML"},
			{"data.json", "JSON"},
			{"page.html", "HTML"},
			{"styles.css", "CSS"},
			{"unknown.xyz", "Unknown"},
		}

		for _, tt := range tests {
			result := detectLanguage(tt.path)
			assert.Equal(t, tt.expected, result)
		}
	})
}
