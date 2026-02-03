package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/database"
	"github.com/yourusername/oview/internal/indexer"
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

	// Create indexer
	fmt.Println("🔍 Starting indexing process...")
	fmt.Println()

	idx := indexer.New(projectPath, projectConfig.ProjectID, db, ragConfig)

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
	fmt.Println("Indexed data is now available for RAG queries!")
	fmt.Println()
	fmt.Println("Note: Using stub embeddings (hash-based, no semantic meaning)")
	fmt.Println("To use real embeddings, implement a proper embeddings generator")

	return nil
}
