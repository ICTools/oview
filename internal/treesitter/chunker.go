package treesitter

import (
	"context"
	"fmt"
	"strings"
)

// Chunk represents a semantic code chunk
type Chunk struct {
	Content   string
	Symbol    string // Function/class name
	Type      string // function, class, method, module
	StartLine int
	EndLine   int
	Language  string
}

// ChunkerConfig defines chunking behavior
type ChunkerConfig struct {
	MaxTokens        int     // Maximum tokens per chunk (from embedding model)
	MinTokens        int     // Minimum tokens to avoid tiny chunks
	SafetyFactor     float64 // Use 80% of max to be safe
	GroupSmall       bool    // Group small functions together
	SubdivideLarge   bool    // Subdivide large functions
	PreferFunctions  bool    // Prefer function-level chunks over file-level
}

// Chunker performs semantic chunking using Tree-sitter
type Chunker struct {
	config  *ChunkerConfig
	parsers *ParserManager
}

// NewChunker creates a new Tree-sitter based chunker
func NewChunker(embeddingMaxTokens int) *Chunker {
	config := &ChunkerConfig{
		MaxTokens:        int(float64(embeddingMaxTokens) * 0.8), // 80% safety margin
		MinTokens:        50,
		SafetyFactor:     0.8,
		GroupSmall:       true,
		SubdivideLarge:   true,
		PreferFunctions:  true,
	}

	return &Chunker{
		config:  config,
		parsers: NewParserManager(),
	}
}

// NewChunkerWithConfig creates a chunker with custom config
func NewChunkerWithConfig(config *ChunkerConfig) *Chunker {
	return &Chunker{
		config:  config,
		parsers: NewParserManager(),
	}
}

// ChunkFile chunks a file using Tree-sitter AST parsing
func (c *Chunker) ChunkFile(path string, content []byte, language string) ([]Chunk, error) {
	// Get parser for language
	parser, err := c.parsers.GetParser(language)
	if err != nil {
		return nil, fmt.Errorf("failed to get parser for %s: %w", language, err)
	}

	// Parse the content
	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}
	defer tree.Close()

	// Extract semantic units (functions, classes, etc.)
	extractor := NewExtractor(language)
	units := extractor.Extract(tree.RootNode(), content)

	if len(units) == 0 {
		// Fallback: treat entire file as one chunk if no units found
		return []Chunk{{
			Content:   string(content),
			Symbol:    "",
			Type:      "file",
			StartLine: 1,
			EndLine:   countLines(content),
			Language:  language,
		}}, nil
	}

	// Apply adaptive chunking based on token limits
	chunks := c.adaptiveChunking(units, content)

	return chunks, nil
}

// adaptiveChunking applies intelligent grouping and subdivision
func (c *Chunker) adaptiveChunking(units []SemanticUnit, content []byte) []Chunk {
	chunks := make([]Chunk, 0)
	pendingGroup := make([]SemanticUnit, 0)
	pendingTokens := 0

	for i, unit := range units {
		tokens := estimateTokens(unit.Content)

		if tokens > c.config.MaxTokens && c.config.SubdivideLarge {
			// Unit too large - subdivide it
			if len(pendingGroup) > 0 {
				// Flush pending group first
				chunks = append(chunks, c.createGroupChunk(pendingGroup, content))
				pendingGroup = nil
				pendingTokens = 0
			}

			// Subdivide large unit
			subChunks := c.subdivideUnit(unit, content)
			chunks = append(chunks, subChunks...)

		} else if tokens < c.config.MinTokens && c.config.GroupSmall {
			// Unit too small - try to group with others
			pendingGroup = append(pendingGroup, unit)
			pendingTokens += tokens

			// Check if we should flush the group
			shouldFlush := pendingTokens >= c.config.MinTokens ||
				i == len(units)-1 || // Last unit
				pendingTokens+estimateTokens(units[min(i+1, len(units)-1)].Content) > c.config.MaxTokens

			if shouldFlush {
				chunks = append(chunks, c.createGroupChunk(pendingGroup, content))
				pendingGroup = nil
				pendingTokens = 0
			}

		} else {
			// Unit is good size - use as-is
			if len(pendingGroup) > 0 {
				// Flush pending group first
				chunks = append(chunks, c.createGroupChunk(pendingGroup, content))
				pendingGroup = nil
				pendingTokens = 0
			}

			chunks = append(chunks, Chunk{
				Content:   unit.Content,
				Symbol:    unit.Name,
				Type:      unit.Type,
				StartLine: unit.StartLine,
				EndLine:   unit.EndLine,
				Language:  unit.Language,
			})
		}
	}

	return chunks
}

// subdivideUnit subdivides a large semantic unit into smaller chunks
func (c *Chunker) subdivideUnit(unit SemanticUnit, content []byte) []Chunk {
	// For now, simple line-based subdivision
	// TODO: Use Tree-sitter to find logical blocks (if statements, loops, etc.)
	lines := strings.Split(unit.Content, "\n")
	chunks := make([]Chunk, 0)

	currentChunk := make([]string, 0)
	currentTokens := 0
	chunkStartLine := unit.StartLine

	for i, line := range lines {
		lineTokens := int(float64(len(line)) / 1.5)

		if currentTokens+lineTokens > c.config.MaxTokens && len(currentChunk) > 0 {
			// Flush current chunk
			chunkContent := strings.Join(currentChunk, "\n")
			chunks = append(chunks, Chunk{
				Content:   chunkContent,
				Symbol:    fmt.Sprintf("%s_part%d", unit.Name, len(chunks)+1),
				Type:      unit.Type + "_part",
				StartLine: chunkStartLine,
				EndLine:   chunkStartLine + len(currentChunk) - 1,
				Language:  unit.Language,
			})

			currentChunk = nil
			currentTokens = 0
			chunkStartLine = unit.StartLine + i
		}

		currentChunk = append(currentChunk, line)
		currentTokens += lineTokens
	}

	// Flush remaining
	if len(currentChunk) > 0 {
		chunkContent := strings.Join(currentChunk, "\n")
		chunks = append(chunks, Chunk{
			Content:   chunkContent,
			Symbol:    fmt.Sprintf("%s_part%d", unit.Name, len(chunks)+1),
			Type:      unit.Type + "_part",
			StartLine: chunkStartLine,
			EndLine:   unit.EndLine,
			Language:  unit.Language,
		})
	}

	return chunks
}

// createGroupChunk creates a single chunk from multiple small units
func (c *Chunker) createGroupChunk(units []SemanticUnit, content []byte) Chunk {
	if len(units) == 0 {
		return Chunk{}
	}

	if len(units) == 1 {
		return Chunk{
			Content:   units[0].Content,
			Symbol:    units[0].Name,
			Type:      units[0].Type,
			StartLine: units[0].StartLine,
			EndLine:   units[0].EndLine,
			Language:  units[0].Language,
		}
	}

	// Combine multiple units
	contents := make([]string, len(units))
	symbols := make([]string, len(units))
	for i, unit := range units {
		contents[i] = unit.Content
		symbols[i] = unit.Name
	}

	return Chunk{
		Content:   strings.Join(contents, "\n\n"),
		Symbol:    strings.Join(symbols, ", "),
		Type:      "group",
		StartLine: units[0].StartLine,
		EndLine:   units[len(units)-1].EndLine,
		Language:  units[0].Language,
	}
}

// estimateTokens estimates token count from text (1 token ≈ 1.5 chars for code with whitespace)
func estimateTokens(text string) int {
	return int(float64(len(text)) / 1.5)
}

// countLines counts lines in content
func countLines(content []byte) int {
	return strings.Count(string(content), "\n") + 1
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
