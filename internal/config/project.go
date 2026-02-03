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
	Trello      TrelloConfig      `yaml:"trello"`
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

// TrelloConfig contains Trello integration settings (placeholders)
type TrelloConfig struct {
	BoardID    string `yaml:"board_id,omitempty"`
	ListIDs    map[string]string `yaml:"list_ids,omitempty"`
	APIKey     string `yaml:"api_key,omitempty"`
	APIToken   string `yaml:"api_token,omitempty"`
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

// DefaultRAGConfig returns a RAG config with sensible defaults
func DefaultRAGConfig() *RAGConfig {
	return &RAGConfig{
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
			IncludePaths: []string{
				"src/",
				"config/",
				"templates/",
				"assets/",
				"tests/",
				"Makefile",
				"docker-compose.yml",
				"compose.yaml",
				"README.md",
				"docs/",
			},
			ExcludePaths: []string{
				"vendor/",
				"node_modules/",
				"var/",
				"public/bundles/",
				".git/",
			},
			Extensions: []string{
				".php", ".twig", ".yaml", ".yml", ".js", ".ts",
				".jsx", ".tsx", ".json", ".md", ".txt",
			},
		},
	}
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
