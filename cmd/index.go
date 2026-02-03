package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/database"
	"github.com/yourusername/oview/internal/embeddings"
	"github.com/yourusername/oview/internal/indexer"
)

var (
	forceReindex bool
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index the project codebase into pgvector",
	Long: `Indexes the project repository for RAG:
- Scans source code, config files, documentation
- Applies chunking rules from .oview/rag.yaml
- Generates embeddings (using stub for MVP)
- Stores chunks in the project database
- Updates .oview/index/stats.json and manifest.json`,
	RunE: runIndex,
}

func init() {
	indexCmd.Flags().BoolVar(&forceReindex, "force", false, "Force full reindex (clears existing embeddings)")
	rootCmd.AddCommand(indexCmd)
}

func runIndex(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Println("📚 Indexing project codebase...")
	fmt.Println()

	// Load project config
	fmt.Println("📋 Loading project configuration...")
	projectConfig, err := config.LoadProjectConfig(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load project config: %w\nHint: Run 'oview init' first", err)
	}

	if projectConfig.Database.Name == "" {
		return fmt.Errorf("project database not configured\nHint: Run 'oview up' first")
	}

	fmt.Printf("   ✓ Project: %s\n", projectConfig.ProjectSlug)

	// Load RAG config
	ragConfig, err := config.LoadRAGConfig(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load RAG config: %w", err)
	}
	fmt.Println("   ✓ RAG config loaded")

	// Load global config
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Connect to project database
	fmt.Println("🔗 Connecting to project database...")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		projectConfig.Database.User,
		projectConfig.Database.Password,
		globalConfig.PostgresHost,
		globalConfig.PostgresPort,
		projectConfig.Database.Name,
	)

	dbClient, err := database.NewClient(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbClient.Close()

	db, err := dbClient.GetConnection(projectConfig.Database.Name)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	defer db.Close()

	fmt.Println("   ✓ Connected")

	// Create embeddings generator based on project config
	embConfig := projectConfig.Embeddings
	var embedder embeddings.Generator

	fmt.Printf("📊 Embeddings config: provider=%s, model=%s, dim=%d\n",
		embConfig.Provider, embConfig.Model, embConfig.Dim)

	switch embConfig.Provider {
	case "openai":
		// Get API key from config or environment
		apiKey := embConfig.APIKey
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		if apiKey == "" {
			return fmt.Errorf("OpenAI API key required. Set in .oview/project.yaml or OPENAI_API_KEY environment variable")
		}
		embedder = embeddings.NewOpenAIGenerator(apiKey, embConfig.Model)
		fmt.Printf("🤖 Using OpenAI embeddings: %s\n", embedder.Name())

	case "ollama":
		baseURL := embConfig.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		embedder = embeddings.NewOllamaGenerator(baseURL, embConfig.Model)
		fmt.Printf("🤖 Using Ollama embeddings: %s\n", embedder.Name())
		fmt.Printf("   ⚠️  Make sure: ollama serve && ollama pull %s\n", embConfig.Model)

	case "stub":
		embedder = embeddings.NewStubGenerator(embConfig.Dim)
		fmt.Println("⚠️  Using stub embeddings (no semantic meaning)")

	default:
		return fmt.Errorf("unknown embeddings provider: %s (edit .oview/project.yaml)", embConfig.Provider)
	}

	// Verify dimensions match
	if embedder.Dimension() != embConfig.Dim {
		fmt.Printf("⚠️  Warning: Model dimension (%d) doesn't match config (%d)\n",
			embedder.Dimension(), embConfig.Dim)
		fmt.Println("   Update .oview/project.yaml or run: oview up")
	}

	// Create indexer
	fmt.Println("🔍 Starting indexing process...")
	fmt.Println()

	idx := indexer.New(projectPath, projectConfig.ProjectID, db, ragConfig, embedder, embConfig.Model)

	// Run indexing
	stats, err := idx.Index()
	if err != nil {
		return fmt.Errorf("indexing failed: %w", err)
	}

	// Print summary
	fmt.Println()
	fmt.Println("✅ Indexing complete!")
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Files indexed:  %d\n", stats.FilesIndexed)
	fmt.Printf("  Chunks stored:  %d\n", stats.ChunksStored)
	fmt.Printf("  Total size:     %d bytes\n", stats.TotalBytes)
	fmt.Printf("  Duration:       %s\n", stats.Duration)
	if stats.CommitSHA != "" {
		fmt.Printf("  Git commit:     %s\n", stats.CommitSHA)
	}
	fmt.Println()
	fmt.Println("✅ Indexed data is now available for RAG queries!")
	fmt.Println()
	fmt.Printf("Embedding model: %s\n", embConfig.Model)
	fmt.Println()

	// Show tips based on provider
	if embConfig.Provider == "stub" {
		fmt.Println("💡 To use real embeddings, edit .oview/project.yaml:")
		fmt.Println("   embeddings:")
		fmt.Println("     provider: openai    # or ollama")
		fmt.Println("     model: text-embedding-3-small")
		fmt.Println("     dim: 1536")
		fmt.Println()
		fmt.Println("   Then run: oview index")
	}

	return nil
}
