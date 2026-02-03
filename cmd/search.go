package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/embeddings"
)

var (
	searchLimit int
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search the codebase using semantic similarity",
	Long: `Search your indexed codebase using semantic similarity.
The query is embedded using the same model configured in project.yaml,
then similar code chunks are retrieved using cosine similarity.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "n", 5, "Number of results to return")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	query := strings.Join(args, " ")

	fmt.Println("🔍 Searching codebase...")
	fmt.Println()

	// Load project config
	projectConfig, err := config.LoadProjectConfig(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load project config: %w\nHint: Run 'oview init' first", err)
	}

	// Load global config
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Print embeddings info
	fmt.Printf("📊 Query: \"%s\"\n", query)
	fmt.Printf("🤖 Using embeddings: %s / %s (%d dimensions)\n",
		projectConfig.Embeddings.Provider,
		projectConfig.Embeddings.Model,
		projectConfig.Embeddings.Dim)
	fmt.Println()

	// Initialize embeddings generator
	var generator embeddings.Generator
	switch projectConfig.Embeddings.Provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			apiKey = projectConfig.Embeddings.APIKey
		}
		if apiKey == "" {
			return fmt.Errorf("OPENAI_API_KEY not set")
		}
		generator = embeddings.NewOpenAIGenerator(apiKey, projectConfig.Embeddings.Model)

	case "ollama":
		baseURL := projectConfig.Embeddings.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		generator = embeddings.NewOllamaGenerator(baseURL, projectConfig.Embeddings.Model)

	default:
		return fmt.Errorf("unsupported embeddings provider: %s", projectConfig.Embeddings.Provider)
	}

	// Generate embedding for query
	fmt.Println("🧮 Generating query embedding...")
	queryEmbedding, err := generator.Embed(query)
	if err != nil {
		return fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Connect to database
	dbName := fmt.Sprintf("oview_%s", projectConfig.ProjectSlug)
	dbUser := fmt.Sprintf("oview_%s", projectConfig.ProjectSlug)
	dbPassword := projectConfig.Database.Password

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbUser, dbPassword, globalConfig.PostgresHost, globalConfig.PostgresPort, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Search for similar chunks
	fmt.Println("🔎 Searching for similar chunks...")
	fmt.Println()

	results, err := searchSimilarChunks(db, projectConfig.ProjectID, queryEmbedding, searchLimit)
	if err != nil {
		return fmt.Errorf("failed to search: %w", err)
	}

	// Display results
	if len(results) == 0 {
		fmt.Println("❌ No results found")
		return nil
	}

	fmt.Printf("✅ Found %d results:\n", len(results))
	fmt.Println()

	for i, result := range results {
		fmt.Printf("═══════════════════════════════════════════════════════════════\n")
		fmt.Printf("Result #%d - Similarity: %.2f%%\n", i+1, result.Similarity*100)
		fmt.Printf("───────────────────────────────────────────────────────────────\n")
		fmt.Printf("📁 File:     %s\n", result.Path)
		if result.Symbol != "" {
			fmt.Printf("🔤 Symbol:   %s\n", result.Symbol)
		}
		fmt.Printf("📂 Type:     %s\n", result.Type)
		if result.Language != "" {
			fmt.Printf("💻 Language: %s\n", result.Language)
		}
		fmt.Println()
		fmt.Printf("📝 Content preview:\n")
		fmt.Println("───────────────────────────────────────────────────────────────")

		// Show first 300 characters
		content := result.Content
		if len(content) > 300 {
			content = content[:300] + "..."
		}
		fmt.Println(content)
		fmt.Println()
	}

	return nil
}

// SearchResult represents a search result
type SearchResult struct {
	ID         int
	Path       string
	Type       string
	Language   string
	Symbol     string
	Content    string
	Similarity float64
}

// searchSimilarChunks searches for chunks similar to the query embedding
func searchSimilarChunks(db *sql.DB, projectID string, queryEmbedding []float32, limit int) ([]SearchResult, error) {
	// Convert embedding to PostgreSQL array format
	embeddingStr := embeddingToString(queryEmbedding)

	query := `
		SELECT
			id, path, type, COALESCE(language, ''), COALESCE(symbol, ''), content,
			1 - (embedding <=> $1::vector) as similarity
		FROM chunks
		WHERE project_id = $2
		ORDER BY embedding <=> $1::vector
		LIMIT $3
	`

	rows, err := db.Query(query, embeddingStr, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		err := rows.Scan(&r.ID, &r.Path, &r.Type, &r.Language, &r.Symbol, &r.Content, &r.Similarity)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, r)
	}

	return results, nil
}

// embeddingToString converts a float32 slice to PostgreSQL vector string format
func embeddingToString(embedding []float32) string {
	parts := make([]string, len(embedding))
	for i, v := range embedding {
		parts[i] = fmt.Sprintf("%f", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
