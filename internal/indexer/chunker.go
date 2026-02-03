package indexer

import (
	"bufio"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yourusername/oview/internal/config"
)

// Chunk represents a chunk of code/documentation
type Chunk struct {
	Path      string
	Language  string
	Symbol    string // function/class name if applicable
	Component string // component/module name
	Content   string
	Type      string // code, doc, config, test
}

// Chunker chunks files based on rules
type Chunker struct {
	rules *config.RAGConfig
}

// NewChunker creates a new chunker
func NewChunker(rules *config.RAGConfig) *Chunker {
	return &Chunker{rules: rules}
}

// ChunkFile chunks a file based on its type
func (c *Chunker) ChunkFile(path string, content []byte) ([]Chunk, error) {
	ext := strings.ToLower(filepath.Ext(path))
	basename := filepath.Base(path)

	// Determine file type
	switch {
	case ext == ".php":
		return c.chunkPHP(path, string(content))
	case ext == ".twig":
		return c.chunkTwig(path, string(content))
	case ext == ".yaml" || ext == ".yml":
		return c.chunkYAML(path, string(content))
	case basename == "Makefile":
		return c.chunkMakefile(path, string(content))
	case basename == "docker-compose.yml" || basename == "docker-compose.yaml" || basename == "compose.yml" || basename == "compose.yaml":
		return c.chunkDockerCompose(path, string(content))
	case ext == ".js" || ext == ".ts" || ext == ".jsx" || ext == ".tsx":
		return c.chunkJavaScript(path, string(content))
	case ext == ".md" || ext == ".txt":
		return c.chunkDocument(path, string(content))
	default:
		return c.chunkGeneric(path, string(content))
	}
}

// chunkPHP chunks PHP files by function/class (simplified approach)
func (c *Chunker) chunkPHP(path string, content string) ([]Chunk, error) {
	rule := c.rules.Chunking.PHP
	chunks := []Chunk{}

	// Simple regex-based approach for MVP
	// Look for class and function definitions
	classRegex := regexp.MustCompile(`(?m)^(?:abstract\s+|final\s+)?class\s+(\w+)`)
	functionRegex := regexp.MustCompile(`(?m)^\s*(?:public|private|protected)?\s*function\s+(\w+)\s*\(`)

	// Split by classes first
	classMatches := classRegex.FindAllStringSubmatchIndex(content, -1)

	if len(classMatches) == 0 {
		// No classes found, chunk by max size
		return c.chunkBySize(path, content, rule.MaxSize, "PHP", "code")
	}

	// Process each class
	for i, match := range classMatches {
		startIdx := match[0]
		endIdx := len(content)
		if i < len(classMatches)-1 {
			endIdx = classMatches[i+1][0]
		}

		className := content[match[2]:match[3]]
		classContent := content[startIdx:endIdx]

		// Try to find functions within this class
		funcMatches := functionRegex.FindAllStringSubmatchIndex(classContent, -1)

		if len(funcMatches) == 0 || len(classContent) < rule.MaxSize {
			// No functions or class is small enough, keep as one chunk
			chunks = append(chunks, Chunk{
				Path:      path,
				Language:  "PHP",
				Symbol:    className,
				Component: getComponent(path),
				Content:   strings.TrimSpace(classContent),
				Type:      getFileType(path),
			})
		} else {
			// Chunk by functions
			for j, funcMatch := range funcMatches {
				funcStartIdx := funcMatch[0]
				funcEndIdx := len(classContent)
				if j < len(funcMatches)-1 {
					funcEndIdx = funcMatches[j+1][0]
				}

				funcName := classContent[funcMatch[2]:funcMatch[3]]
				funcContent := classContent[funcStartIdx:funcEndIdx]

				if len(funcContent) <= rule.MaxSize {
					chunks = append(chunks, Chunk{
						Path:      path,
						Language:  "PHP",
						Symbol:    fmt.Sprintf("%s::%s", className, funcName),
						Component: getComponent(path),
						Content:   strings.TrimSpace(funcContent),
						Type:      getFileType(path),
					})
				} else {
					// Function too large, split by size
					subChunks, err := c.chunkBySize(path, funcContent, rule.MaxSize, "PHP", getFileType(path))
					if err != nil {
						return nil, err
					}
					for k, sc := range subChunks {
						sc.Symbol = fmt.Sprintf("%s::%s#%d", className, funcName, k)
						chunks = append(chunks, sc)
					}
				}
			}
		}
	}

	return chunks, nil
}

// chunkJavaScript chunks JavaScript/TypeScript files (simplified)
func (c *Chunker) chunkJavaScript(path string, content string) ([]Chunk, error) {
	rule := c.rules.Chunking.JavaScript
	// For MVP, use simple size-based chunking
	// TODO: Add proper AST-based chunking for functions/classes
	return c.chunkBySize(path, content, rule.MaxSize, detectLanguage(path), getFileType(path))
}

// chunkTwig chunks Twig template files
func (c *Chunker) chunkTwig(path string, content string) ([]Chunk, error) {
	rule := c.rules.Chunking.Twig
	// Twig files are usually small, chunk by file or blocks
	if len(content) <= rule.MaxSize {
		return []Chunk{{
			Path:      path,
			Language:  "Twig",
			Component: getComponent(path),
			Content:   content,
			Type:      "code",
		}}, nil
	}

	// Try to split by blocks
	blockRegex := regexp.MustCompile(`{%\s*block\s+(\w+)\s*%}`)
	matches := blockRegex.FindAllStringSubmatchIndex(content, -1)

	if len(matches) == 0 {
		return c.chunkBySize(path, content, rule.MaxSize, "Twig", "code")
	}

	chunks := []Chunk{}
	for i, match := range matches {
		startIdx := match[0]
		endIdx := len(content)
		if i < len(matches)-1 {
			endIdx = matches[i+1][0]
		}

		blockName := content[match[2]:match[3]]
		blockContent := content[startIdx:endIdx]

		chunks = append(chunks, Chunk{
			Path:      path,
			Language:  "Twig",
			Symbol:    blockName,
			Component: getComponent(path),
			Content:   strings.TrimSpace(blockContent),
			Type:      "code",
		})
	}

	return chunks, nil
}

// chunkYAML chunks YAML files by top-level sections
func (c *Chunker) chunkYAML(path string, content string) ([]Chunk, error) {
	rule := c.rules.Chunking.YAML

	if len(content) <= rule.MaxSize {
		return []Chunk{{
			Path:      path,
			Language:  "YAML",
			Component: getComponent(path),
			Content:   content,
			Type:      getFileType(path),
		}}, nil
	}

	// Split by top-level keys (lines that start without indentation)
	chunks := []Chunk{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentSection strings.Builder
	var currentKey string

	for scanner.Scan() {
		line := scanner.Text()
		// Top-level key: starts without spaces and contains ':'
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' && strings.Contains(line, ":") {
			// Save previous section
			if currentSection.Len() > 0 {
				chunks = append(chunks, Chunk{
					Path:      path,
					Language:  "YAML",
					Symbol:    currentKey,
					Component: getComponent(path),
					Content:   strings.TrimSpace(currentSection.String()),
					Type:      getFileType(path),
				})
				currentSection.Reset()
			}
			currentKey = strings.TrimSpace(strings.Split(line, ":")[0])
		}
		currentSection.WriteString(line + "\n")
	}

	// Save last section
	if currentSection.Len() > 0 {
		chunks = append(chunks, Chunk{
			Path:      path,
			Language:  "YAML",
			Symbol:    currentKey,
			Component: getComponent(path),
			Content:   strings.TrimSpace(currentSection.String()),
			Type:      getFileType(path),
		})
	}

	return chunks, nil
}

// chunkMakefile chunks Makefile by targets
func (c *Chunker) chunkMakefile(path string, content string) ([]Chunk, error) {
	chunks := []Chunk{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentTarget strings.Builder
	var currentName string

	for scanner.Scan() {
		line := scanner.Text()
		// Target line: starts without space/tab and contains ':'
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' && strings.Contains(line, ":") && !strings.HasPrefix(line, "#") {
			// Save previous target
			if currentTarget.Len() > 0 {
				chunks = append(chunks, Chunk{
					Path:      path,
					Language:  "Makefile",
					Symbol:    currentName,
					Component: "build",
					Content:   strings.TrimSpace(currentTarget.String()),
					Type:      "config",
				})
				currentTarget.Reset()
			}
			currentName = strings.TrimSpace(strings.Split(line, ":")[0])
		}
		currentTarget.WriteString(line + "\n")
	}

	// Save last target
	if currentTarget.Len() > 0 {
		chunks = append(chunks, Chunk{
			Path:      path,
			Language:  "Makefile",
			Symbol:    currentName,
			Component: "build",
			Content:   strings.TrimSpace(currentTarget.String()),
			Type:      "config",
		})
	}

	return chunks, nil
}

// chunkDockerCompose chunks docker-compose by services
func (c *Chunker) chunkDockerCompose(path string, content string) ([]Chunk, error) {
	// Similar to YAML but look for 'services:' section specifically
	chunks := []Chunk{}
	inServices := false
	var currentService strings.Builder
	var currentName string

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "services:") {
			inServices = true
			continue
		}

		if inServices {
			// Service name: 2 spaces indentation with ':'
			if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") && strings.Contains(line, ":") {
				// Save previous service
				if currentService.Len() > 0 {
					chunks = append(chunks, Chunk{
						Path:      path,
						Language:  "YAML",
						Symbol:    currentName,
						Component: "docker",
						Content:   strings.TrimSpace(currentService.String()),
						Type:      "config",
					})
					currentService.Reset()
				}
				currentName = strings.TrimSpace(strings.Split(strings.TrimSpace(line), ":")[0])
			}

			currentService.WriteString(line + "\n")

			// Stop at next top-level key
			if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
				inServices = false
			}
		}
	}

	// Save last service
	if currentService.Len() > 0 {
		chunks = append(chunks, Chunk{
			Path:      path,
			Language:  "YAML",
			Symbol:    currentName,
			Component: "docker",
			Content:   strings.TrimSpace(currentService.String()),
			Type:      "config",
		})
	}

	return chunks, nil
}

// chunkDocument chunks markdown/text documents
func (c *Chunker) chunkDocument(path string, content string) ([]Chunk, error) {
	// Chunk by headings for markdown, or by size for text
	if strings.HasSuffix(path, ".md") {
		return c.chunkMarkdown(path, content)
	}
	return c.chunkBySize(path, content, 1500, "Text", "doc")
}

// chunkMarkdown chunks markdown by headings
func (c *Chunker) chunkMarkdown(path string, content string) ([]Chunk, error) {
	chunks := []Chunk{}
	headingRegex := regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)
	matches := headingRegex.FindAllStringSubmatchIndex(content, -1)

	if len(matches) == 0 {
		return []Chunk{{
			Path:      path,
			Language:  "Markdown",
			Component: "docs",
			Content:   content,
			Type:      "doc",
		}}, nil
	}

	for i, match := range matches {
		startIdx := match[0]
		endIdx := len(content)
		if i < len(matches)-1 {
			endIdx = matches[i+1][0]
		}

		heading := content[match[4]:match[5]]
		sectionContent := content[startIdx:endIdx]

		chunks = append(chunks, Chunk{
			Path:      path,
			Language:  "Markdown",
			Symbol:    heading,
			Component: "docs",
			Content:   strings.TrimSpace(sectionContent),
			Type:      "doc",
		})
	}

	return chunks, nil
}

// chunkGeneric chunks files by size
func (c *Chunker) chunkGeneric(path string, content string) ([]Chunk, error) {
	rule := c.rules.Chunking.Generic
	return c.chunkBySize(path, content, rule.MaxSize, detectLanguage(path), getFileType(path))
}

// chunkBySize chunks content by size with overlap
func (c *Chunker) chunkBySize(path, content string, maxSize int, language, fileType string) ([]Chunk, error) {
	chunks := []Chunk{}

	if len(content) <= maxSize {
		return []Chunk{{
			Path:      path,
			Language:  language,
			Component: getComponent(path),
			Content:   content,
			Type:      fileType,
		}}, nil
	}

	// Split by lines for better readability
	lines := strings.Split(content, "\n")
	var currentChunk strings.Builder
	chunkNum := 0

	for _, line := range lines {
		if currentChunk.Len()+len(line) > maxSize {
			// Save current chunk
			chunks = append(chunks, Chunk{
				Path:      path,
				Language:  language,
				Symbol:    fmt.Sprintf("chunk-%d", chunkNum),
				Component: getComponent(path),
				Content:   strings.TrimSpace(currentChunk.String()),
				Type:      fileType,
			})
			chunkNum++
			currentChunk.Reset()
		}
		currentChunk.WriteString(line + "\n")
	}

	// Save last chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, Chunk{
			Path:      path,
			Language:  language,
			Symbol:    fmt.Sprintf("chunk-%d", chunkNum),
			Component: getComponent(path),
			Content:   strings.TrimSpace(currentChunk.String()),
			Type:      fileType,
		})
	}

	return chunks, nil
}

// Helper functions

func getComponent(path string) string {
	// Extract component from path (e.g., "Controller", "Service", etc.)
	parts := strings.Split(filepath.Dir(path), string(filepath.Separator))
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" && parts[i] != "." {
			return parts[i]
		}
	}
	return ""
}

func getFileType(path string) string {
	if strings.Contains(path, "/tests/") || strings.Contains(path, "/test/") {
		return "test"
	}
	if strings.Contains(path, "/config/") {
		return "config"
	}
	if strings.Contains(path, "/docs/") || strings.HasSuffix(path, ".md") {
		return "doc"
	}
	return "code"
}

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".php":
		return "PHP"
	case ".js":
		return "JavaScript"
	case ".ts":
		return "TypeScript"
	case ".jsx":
		return "JSX"
	case ".tsx":
		return "TSX"
	case ".py":
		return "Python"
	case ".go":
		return "Go"
	case ".java":
		return "Java"
	case ".rb":
		return "Ruby"
	case ".yaml", ".yml":
		return "YAML"
	case ".json":
		return "JSON"
	case ".xml":
		return "XML"
	case ".md":
		return "Markdown"
	case ".html", ".twig":
		return "HTML"
	case ".css", ".scss", ".sass":
		return "CSS"
	default:
		return "Unknown"
	}
}
