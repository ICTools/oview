package mcp

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/embeddings"
)

// ToolHandler handles MCP tool calls
type ToolHandler struct {
	projectConfig *config.ProjectConfig
	globalConfig  *config.GlobalConfig
	db            *sql.DB
	generator     embeddings.Generator
}

// NewToolHandler creates a new tool handler
func NewToolHandler(projectConfig *config.ProjectConfig, globalConfig *config.GlobalConfig) *ToolHandler {
	return &ToolHandler{
		projectConfig: projectConfig,
		globalConfig:  globalConfig,
	}
}

// CallTool executes a tool and returns the result
func (h *ToolHandler) CallTool(name string, args map[string]interface{}) (interface{}, error) {
	switch name {
	case "search":
		return h.handleSearch(args)
	case "get_context":
		return h.handleGetContext(args)
	case "project_info":
		return h.handleProjectInfo(args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

// handleSearch performs semantic search
func (h *ToolHandler) handleSearch(args map[string]interface{}) (interface{}, error) {
	// Parse arguments
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required")
	}

	limit := 5
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
		if limit > 20 {
			limit = 20
		}
	}

	// Initialize embeddings generator if needed
	if h.generator == nil {
		if err := h.initGenerator(); err != nil {
			return nil, fmt.Errorf("failed to initialize embeddings: %w", err)
		}
	}

	// Generate query embedding
	queryEmbedding, err := h.generator.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Connect to database if needed
	if h.db == nil {
		if err := h.connectDB(); err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
	}

	// Search
	results, err := h.searchSimilarChunks(queryEmbedding, limit)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Format results
	formattedResults := make([]map[string]interface{}, len(results))
	for i, r := range results {
		formattedResults[i] = map[string]interface{}{
			"path":       r.Path,
			"type":       r.Type,
			"language":   r.Language,
			"symbol":     r.Symbol,
			"content":    r.Content,
			"similarity": fmt.Sprintf("%.2f%%", r.Similarity*100),
		}
	}

	return map[string]interface{}{
		"query":   query,
		"count":   len(results),
		"results": formattedResults,
	}, nil
}

// handleGetContext gets context for a file/symbol
func (h *ToolHandler) handleGetContext(args map[string]interface{}) (interface{}, error) {
	// Parse arguments
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return nil, fmt.Errorf("path is required")
	}

	symbol := ""
	if s, ok := args["symbol"].(string); ok {
		symbol = s
	}

	limit := 3
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	// Connect to database if needed
	if h.db == nil {
		if err := h.connectDB(); err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
	}

	// Get context
	results, err := h.getFileContext(path, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	// Format results
	formattedResults := make([]map[string]interface{}, len(results))
	for i, r := range results {
		formattedResults[i] = map[string]interface{}{
			"path":     r.Path,
			"type":     r.Type,
			"language": r.Language,
			"symbol":   r.Symbol,
			"content":  r.Content,
		}
	}

	return map[string]interface{}{
		"path":    path,
		"symbol":  symbol,
		"count":   len(results),
		"context": formattedResults,
	}, nil
}

// handleProjectInfo returns project information
func (h *ToolHandler) handleProjectInfo(args map[string]interface{}) (interface{}, error) {
	// Check database status
	dbStatus := "unknown"
	if h.db == nil {
		if err := h.connectDB(); err == nil {
			dbStatus = "connected"
		} else {
			dbStatus = "not connected"
		}
	} else {
		if err := h.db.Ping(); err == nil {
			dbStatus = "connected"
		} else {
			dbStatus = "connection lost"
		}
	}

	// Get chunk count
	chunkCount := 0
	if h.db != nil {
		var count int
		err := h.db.QueryRow("SELECT COUNT(*) FROM chunks WHERE project_id = $1", h.projectConfig.ProjectID).Scan(&count)
		if err == nil {
			chunkCount = count
		}
	}

	return map[string]interface{}{
		"project_id":   h.projectConfig.ProjectID,
		"project_slug": h.projectConfig.ProjectSlug,
		"embeddings": map[string]interface{}{
			"provider": h.projectConfig.Embeddings.Provider,
			"model":    h.projectConfig.Embeddings.Model,
			"dim":      h.projectConfig.Embeddings.Dim,
		},
		"llm": map[string]interface{}{
			"provider": h.projectConfig.LLM.Provider,
			"model":    h.projectConfig.LLM.Model,
		},
		"database": map[string]interface{}{
			"name":        h.projectConfig.Database.Name,
			"status":      dbStatus,
			"chunk_count": chunkCount,
		},
		"stack": h.projectConfig.Stack,
	}, nil
}

// initGenerator initializes the embeddings generator
func (h *ToolHandler) initGenerator() error {
	switch h.projectConfig.Embeddings.Provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			apiKey = h.projectConfig.Embeddings.APIKey
		}
		if apiKey == "" {
			return fmt.Errorf("OPENAI_API_KEY not set")
		}
		h.generator = embeddings.NewOpenAIGenerator(apiKey, h.projectConfig.Embeddings.Model)

	case "ollama":
		baseURL := h.projectConfig.Embeddings.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		h.generator = embeddings.NewOllamaGenerator(baseURL, h.projectConfig.Embeddings.Model)

	default:
		return fmt.Errorf("unsupported embeddings provider: %s", h.projectConfig.Embeddings.Provider)
	}

	return nil
}

// connectDB connects to the project database
func (h *ToolHandler) connectDB() error {
	dbName := fmt.Sprintf("oview_%s", h.projectConfig.ProjectSlug)
	dbUser := fmt.Sprintf("oview_%s", h.projectConfig.ProjectSlug)
	dbPassword := h.projectConfig.Database.Password

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbUser, dbPassword, h.globalConfig.PostgresHost, h.globalConfig.PostgresPort, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	h.db = db
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
func (h *ToolHandler) searchSimilarChunks(queryEmbedding []float32, limit int) ([]SearchResult, error) {
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

	rows, err := h.db.Query(query, embeddingStr, h.projectConfig.ProjectID, limit)
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

// getFileContext gets context for a specific file
func (h *ToolHandler) getFileContext(path string, symbol string, limit int) ([]SearchResult, error) {
	var query string
	var args []interface{}

	if symbol != "" {
		// Get specific symbol + related chunks
		query = `
			SELECT
				id, path, type, COALESCE(language, ''), COALESCE(symbol, ''), content,
				0 as similarity
			FROM chunks
			WHERE project_id = $1 AND path = $2 AND symbol = $3
			UNION ALL
			SELECT
				id, path, type, COALESCE(language, ''), COALESCE(symbol, ''), content,
				0 as similarity
			FROM chunks
			WHERE project_id = $1 AND path = $2 AND symbol != $3
			LIMIT $4
		`
		args = []interface{}{h.projectConfig.ProjectID, path, symbol, limit}
	} else {
		// Get all chunks from the file
		query = `
			SELECT
				id, path, type, COALESCE(language, ''), COALESCE(symbol, ''), content,
				0 as similarity
			FROM chunks
			WHERE project_id = $1 AND path = $2
			LIMIT $3
		`
		args = []interface{}{h.projectConfig.ProjectID, path, limit}
	}

	rows, err := h.db.Query(query, args...)
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
