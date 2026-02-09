package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProjectConfig stores the per-project configuration
type ProjectConfig struct {
	ProjectID   string            `yaml:"project_id"`
	ProjectSlug string            `yaml:"project_slug"`
	Stack       StackInfo         `yaml:"stack"`
	Commands    CommandConfig     `yaml:"commands"`
	Database    DatabaseConfig    `yaml:"database,omitempty"`
	Embeddings  EmbeddingsConfig  `yaml:"embeddings"`
	LLM         LLMConfig         `yaml:"llm"`
}

// EmbeddingsConfig contains embeddings configuration
type EmbeddingsConfig struct {
	Provider string `yaml:"provider"` // stub, openai, ollama
	Model    string `yaml:"model"`    // Model name (e.g., "text-embedding-3-small", "nomic-embed-text", "bge-code")
	Dim      int    `yaml:"dim"`      // Vector dimension (768, 1536, etc.)
	APIKey   string `yaml:"api_key,omitempty"` // Optional: API key for OpenAI (prefer env var)
	BaseURL  string `yaml:"base_url,omitempty"` // Optional: Base URL for Ollama or custom endpoint
}

// LLMConfig contains LLM/AI model configuration
type LLMConfig struct {
	Provider string `yaml:"provider"` // claude-code, claude-api, openai, ollama
	Model    string `yaml:"model"`    // Model name (e.g., "claude-sonnet-4.5", "gpt-4", "llama3")
	APIKey   string `yaml:"api_key,omitempty"` // Optional: API key (prefer env var)
	BaseURL  string `yaml:"base_url,omitempty"` // Optional: Custom endpoint
}

// StackInfo contains detected stack information
type StackInfo struct {
	Symfony      bool              `yaml:"symfony"`
	Docker       bool              `yaml:"docker"`
	Makefile     bool              `yaml:"makefile"`
	Frontend     FrontendInfo      `yaml:"frontend"`
	Infrastructure InfraInfo       `yaml:"infrastructure"`
	Languages    []string          `yaml:"languages"`
	Frameworks   []string          `yaml:"frameworks"`
}

// FrontendInfo contains frontend stack details
type FrontendInfo struct {
	Detected       bool     `yaml:"detected"`
	PackageManager string   `yaml:"package_manager,omitempty"`
	Frameworks     []string `yaml:"frameworks,omitempty"`
	BuildTools     []string `yaml:"build_tools,omitempty"`
}

// InfraInfo contains infrastructure component details
type InfraInfo struct {
	Redis        bool `yaml:"redis"`
	RabbitMQ     bool `yaml:"rabbitmq"`
	Elasticsearch bool `yaml:"elasticsearch"`
}

// CommandConfig contains command mappings for the project
type CommandConfig struct {
	Test          []string `yaml:"test,omitempty"`
	Lint          []string `yaml:"lint,omitempty"`
	StaticAnalysis []string `yaml:"static_analysis,omitempty"`
	Build         []string `yaml:"build,omitempty"`
	Start         []string `yaml:"start,omitempty"`
}

// DatabaseConfig contains database connection info
type DatabaseConfig struct {
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password,omitempty"`
}

// LoadProjectConfig loads the project config from .oview/project.yaml
func LoadProjectConfig(projectPath string) (*ProjectConfig, error) {
	configPath := filepath.Join(projectPath, ".oview", "project.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}

	return &config, nil
}

// Save saves the project config to .oview/project.yaml
func (c *ProjectConfig) Save(projectPath string) error {
	configPath := filepath.Join(projectPath, ".oview", "project.yaml")

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create .oview directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}

	return nil
}

// RAGConfig contains RAG indexing configuration
type RAGConfig struct {
	Chunking ChunkingRules `yaml:"chunking"`
	Indexing IndexingRules `yaml:"indexing"`
}

// ChunkingRules defines how different file types should be chunked
type ChunkingRules struct {
	PHP         ChunkRule `yaml:"php"`
	JavaScript  ChunkRule `yaml:"javascript"`
	Twig        ChunkRule `yaml:"twig"`
	YAML        ChunkRule `yaml:"yaml"`
	Makefile    ChunkRule `yaml:"makefile"`
	Docker      ChunkRule `yaml:"docker"`
	Generic     ChunkRule `yaml:"generic"`
}

// ChunkRule defines chunking strategy for a file type
type ChunkRule struct {
	Strategy   string `yaml:"strategy"`     // function, file, size, section
	MaxSize    int    `yaml:"max_size"`     // max characters per chunk
	MaxTokens  int    `yaml:"max_tokens"`   // max tokens per chunk (approximate)
	Overlap    int    `yaml:"overlap"`      // overlap between chunks
}

// IndexingRules defines what to index
type IndexingRules struct {
	IncludePaths []string `yaml:"include_paths"`
	ExcludePaths []string `yaml:"exclude_paths"`
	Extensions   []string `yaml:"extensions"`
}

// DefaultRAGConfig returns a RAG config with defaults based on detected stack
func DefaultRAGConfig(stack *StackInfo) *RAGConfig {
	config := &RAGConfig{
		Chunking: ChunkingRules{
			PHP: ChunkRule{
				Strategy:  "function",
				MaxSize:   2000,
				MaxTokens: 500,
				Overlap:   100,
			},
			JavaScript: ChunkRule{
				Strategy:  "function",
				MaxSize:   2000,
				MaxTokens: 500,
				Overlap:   100,
			},
			Twig: ChunkRule{
				Strategy:  "file",
				MaxSize:   1500,
				MaxTokens: 400,
				Overlap:   50,
			},
			YAML: ChunkRule{
				Strategy:  "section",
				MaxSize:   1000,
				MaxTokens: 300,
				Overlap:   50,
			},
			Makefile: ChunkRule{
				Strategy:  "section",
				MaxSize:   800,
				MaxTokens: 200,
				Overlap:   20,
			},
			Docker: ChunkRule{
				Strategy:  "section",
				MaxSize:   1000,
				MaxTokens: 300,
				Overlap:   50,
			},
			Generic: ChunkRule{
				Strategy:  "size",
				MaxSize:   1500,
				MaxTokens: 400,
				Overlap:   100,
			},
		},
		Indexing: IndexingRules{
			IncludePaths: generateIncludePaths(stack),
			ExcludePaths: generateExcludePaths(stack),
			Extensions:   generateExtensions(stack),
		},
	}

	return config
}

// generateIncludePaths generates include paths based on detected stack
func generateIncludePaths(stack *StackInfo) []string {
	paths := []string{}
	pathsMap := make(map[string]bool) // To avoid duplicates

	// Add paths based on detected languages
	for _, lang := range stack.Languages {
		switch lang {
		case "Go":
			paths = append(paths, "cmd/", "internal/", "pkg/", "main.go")
		case "Python":
			paths = append(paths, "src/", "tests/", "*.py")
		case "Ruby":
			paths = append(paths, "lib/", "app/", "spec/", "test/")
		case "Rust":
			paths = append(paths, "src/", "tests/", "benches/")
		case "Java":
			paths = append(paths, "src/", "test/")
		case "C/C++":
			paths = append(paths, "src/", "include/", "tests/")
		case "C#":
			paths = append(paths, "src/", "tests/")
		case "PHP":
			paths = append(paths, "src/", "tests/")
			if stack.Symfony {
				paths = append(paths, "config/", "templates/")
			}
		case "JavaScript", "TypeScript":
			if stack.Frontend.Detected {
				paths = append(paths, "src/", "assets/", "public/")
			}
		}
	}

	// Add common paths for all projects
	commonPaths := []string{
		"docs/",
		"README.md",
		"Makefile",
	}
	paths = append(paths, commonPaths...)

	// Add Docker files if detected
	if stack.Docker {
		paths = append(paths, "docker-compose.yml", "compose.yaml", "Dockerfile")
	}

	// Deduplicate using map
	for _, p := range paths {
		pathsMap[p] = true
	}

	// Convert back to slice
	result := []string{}
	for p := range pathsMap {
		result = append(result, p)
	}

	return result
}

// generateExcludePaths generates exclude paths based on detected stack
func generateExcludePaths(stack *StackInfo) []string {
	paths := []string{".git/", ".oview/"}

	// Add language-specific excludes
	for _, lang := range stack.Languages {
		switch lang {
		case "Go":
			paths = append(paths, "vendor/")
		case "Python":
			paths = append(paths, "venv/", ".venv/", "__pycache__/", "*.pyc", ".pytest_cache/", ".tox/")
		case "Ruby":
			paths = append(paths, "vendor/", ".bundle/")
		case "Rust":
			paths = append(paths, "target/")
		case "Java":
			paths = append(paths, "target/", "build/", ".gradle/")
		case "C#":
			paths = append(paths, "bin/", "obj/")
		case "PHP":
			paths = append(paths, "vendor/", "var/")
		case "JavaScript", "TypeScript":
			paths = append(paths, "node_modules/", "dist/", "build/")
		}
	}

	// Symfony-specific excludes
	if stack.Symfony {
		paths = append(paths, "public/bundles/", "var/cache/", "var/log/")
	}

	// Deduplicate
	seen := make(map[string]bool)
	result := []string{}
	for _, p := range paths {
		if !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}

	return result
}

// generateExtensions generates file extensions based on detected stack
func generateExtensions(stack *StackInfo) []string {
	exts := []string{}
	extsMap := make(map[string]bool)

	// Add extensions based on detected languages
	for _, lang := range stack.Languages {
		switch lang {
		case "Go":
			exts = append(exts, ".go", ".mod", ".sum")
		case "Python":
			exts = append(exts, ".py", ".pyi", ".pyx")
		case "Ruby":
			exts = append(exts, ".rb", ".rake", ".gemspec")
		case "Rust":
			exts = append(exts, ".rs", ".toml")
		case "Java":
			exts = append(exts, ".java", ".kt", ".gradle")
		case "C/C++":
			exts = append(exts, ".c", ".cpp", ".cc", ".cxx", ".h", ".hpp", ".hxx")
		case "C#":
			exts = append(exts, ".cs", ".csproj", ".sln")
		case "PHP":
			exts = append(exts, ".php", ".twig")
		case "JavaScript":
			exts = append(exts, ".js", ".jsx", ".mjs", ".cjs")
		case "TypeScript":
			exts = append(exts, ".ts", ".tsx")
		case "Shell":
			exts = append(exts, ".sh", ".bash")
		}
	}

	// Add common extensions for all projects
	commonExts := []string{
		".md", ".txt", ".yaml", ".yml", ".json", ".toml",
		".xml", ".ini", ".conf", ".env",
	}
	exts = append(exts, commonExts...)

	// Add CSS/SCSS if frontend detected
	if stack.Frontend.Detected {
		exts = append(exts, ".css", ".scss", ".sass", ".less")
	}

	// Deduplicate
	for _, ext := range exts {
		extsMap[ext] = true
	}

	// Convert back to slice
	result := []string{}
	for ext := range extsMap {
		result = append(result, ext)
	}

	return result
}

// SaveRAGConfig saves the RAG config to .oview/rag.yaml
func SaveRAGConfig(projectPath string, config *RAGConfig) error {
	configPath := filepath.Join(projectPath, ".oview", "rag.yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal RAG config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write RAG config: %w", err)
	}

	return nil
}

// LoadRAGConfig loads the RAG config from .oview/rag.yaml
func LoadRAGConfig(projectPath string) (*RAGConfig, error) {
	configPath := filepath.Join(projectPath, ".oview", "rag.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read RAG config: %w", err)
	}

	var config RAGConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse RAG config: %w", err)
	}

	return &config, nil
}
