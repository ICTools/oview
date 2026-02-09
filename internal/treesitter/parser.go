package treesitter

import (
	"fmt"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// ParserManager manages Tree-sitter parsers for different languages
type ParserManager struct {
	parsers map[string]*sitter.Parser
	mu      sync.RWMutex
}

// NewParserManager creates a new parser manager
func NewParserManager() *ParserManager {
	return &ParserManager{
		parsers: make(map[string]*sitter.Parser),
	}
}

// GetParser returns a parser for the given language
func (pm *ParserManager) GetParser(language string) (*sitter.Parser, error) {
	pm.mu.RLock()
	parser, exists := pm.parsers[language]
	pm.mu.RUnlock()

	if exists {
		return parser, nil
	}

	// Create new parser
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Double-check after acquiring write lock
	if parser, exists := pm.parsers[language]; exists {
		return parser, nil
	}

	parser = sitter.NewParser()
	lang, err := pm.getLanguage(language)
	if err != nil {
		return nil, err
	}

	parser.SetLanguage(lang)
	pm.parsers[language] = parser

	return parser, nil
}

// getLanguage returns the Tree-sitter language for a given language name
func (pm *ParserManager) getLanguage(language string) (*sitter.Language, error) {
	switch language {
	case "python", "Python", "py":
		return python.GetLanguage(), nil
	case "javascript", "JavaScript", "js":
		return javascript.GetLanguage(), nil
	case "typescript", "TypeScript", "ts":
		return typescript.GetLanguage(), nil
	case "go", "Go", "golang":
		return golang.GetLanguage(), nil
	case "php", "PHP":
		return php.GetLanguage(), nil
	case "java", "Java":
		return java.GetLanguage(), nil
	case "rust", "Rust", "rs":
		return rust.GetLanguage(), nil
	case "c", "C":
		return c.GetLanguage(), nil
	case "cpp", "C++", "cc", "cxx":
		return cpp.GetLanguage(), nil
	case "ruby", "Ruby", "rb":
		return ruby.GetLanguage(), nil
	case "csharp", "C#", "cs":
		return csharp.GetLanguage(), nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}

// IsSupported checks if a language is supported by Tree-sitter
func (pm *ParserManager) IsSupported(language string) bool {
	_, err := pm.getLanguage(language)
	return err == nil
}

// SupportedLanguages returns a list of supported languages
func (pm *ParserManager) SupportedLanguages() []string {
	return []string{
		"python",
		"javascript",
		"typescript",
		"go",
		"php",
		"java",
		"rust",
		"c",
		"cpp",
		"ruby",
		"csharp",
	}
}

// Close closes all parsers
func (pm *ParserManager) Close() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, parser := range pm.parsers {
		parser.Close()
	}
	pm.parsers = make(map[string]*sitter.Parser)
}
