package indexer

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/embeddings"
)

// Indexer indexes a project into the database
type Indexer struct {
	projectPath    string
	projectID      string
	db             *sql.DB
	chunker        *Chunker
	embedder       embeddings.Generator
	embeddingModel string // Model name to store in DB
	ragConfig      *config.RAGConfig
}

// Stats tracks indexing statistics
type Stats struct {
	CommitSHA    string    `json:"commit_sha"`
	FilesIndexed int       `json:"files_indexed"`
	ChunksStored int       `json:"chunks_stored"`
	TotalBytes   int64     `json:"total_bytes"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Duration     string    `json:"duration"`
}

// Manifest tracks indexed files
type Manifest struct {
	Files      map[string]FileInfo `json:"files"`
	LastUpdate time.Time           `json:"last_update"`
}

// FileInfo tracks metadata about an indexed file
type FileInfo struct {
	Path       string    `json:"path"`
	Hash       string    `json:"hash"`
	Chunks     int       `json:"chunks"`
	IndexedAt  time.Time `json:"indexed_at"`
}

// New creates a new indexer
func New(projectPath, projectID string, db *sql.DB, ragConfig *config.RAGConfig, embedder embeddings.Generator, embeddingModel string) *Indexer {
	// Use provided embedder, or default to stub if nil
	if embedder == nil {
		embedder = embeddings.NewStubGenerator(1536)
	}

	// Use model name from embedder if not provided
	if embeddingModel == "" {
		embeddingModel = embedder.Name()
	}

	return &Indexer{
		projectPath:    projectPath,
		projectID:      projectID,
		db:             db,
		chunker:        NewChunker(ragConfig),
		embedder:       embedder,
		embeddingModel: embeddingModel,
		ragConfig:      ragConfig,
	}
}

// Index indexes the entire project
func (idx *Indexer) Index() (*Stats, error) {
	stats := &Stats{
		StartTime: time.Now(),
	}

	// Get git commit SHA if available
	stats.CommitSHA = idx.getGitCommitSHA()

	// Scan files
	files, err := idx.scanFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}

	fmt.Printf("Found %d files to index\n", len(files))

	// Clear existing chunks for this project
	if err := idx.clearExistingChunks(); err != nil {
		return nil, fmt.Errorf("failed to clear existing chunks: %w", err)
	}

	manifest := &Manifest{
		Files:      make(map[string]FileInfo),
		LastUpdate: time.Now(),
	}

	// Process each file
	for i, file := range files {
		fmt.Printf("[%d/%d] Indexing %s...\n", i+1, len(files), file)

		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("  ⚠️  Failed to read: %v\n", err)
			continue
		}

		// Chunk the file
		chunks, err := idx.chunker.ChunkFile(file, content)
		if err != nil {
			fmt.Printf("  ⚠️  Failed to chunk: %v\n", err)
			continue
		}

		// Store chunks
		storedCount := 0
		for _, chunk := range chunks {
			if err := idx.storeChunk(chunk, stats.CommitSHA); err != nil {
				fmt.Printf("  ⚠️  Failed to store chunk: %v\n", err)
				continue
			}
			storedCount++
		}

		// Update stats and manifest
		stats.FilesIndexed++
		stats.ChunksStored += storedCount
		stats.TotalBytes += int64(len(content))

		// Update manifest
		hash := sha256.Sum256(content)
		manifest.Files[file] = FileInfo{
			Path:      file,
			Hash:      hex.EncodeToString(hash[:]),
			Chunks:    storedCount,
			IndexedAt: time.Now(),
		}

		fmt.Printf("  ✓ %d chunks stored\n", storedCount)
	}

	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime).String()

	// Save stats and manifest
	if err := idx.saveStats(stats); err != nil {
		return nil, fmt.Errorf("failed to save stats: %w", err)
	}

	if err := idx.saveManifest(manifest); err != nil {
		return nil, fmt.Errorf("failed to save manifest: %w", err)
	}

	return stats, nil
}

// scanFiles scans for files to index based on RAG config
func (idx *Indexer) scanFiles() ([]string, error) {
	var files []string

	includePaths := idx.ragConfig.Indexing.IncludePaths
	excludePaths := idx.ragConfig.Indexing.ExcludePaths
	extensions := idx.ragConfig.Indexing.Extensions

	// Build extension map for quick lookup
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap[ext] = true
	}

	for _, includePath := range includePaths {
		fullPath := filepath.Join(idx.projectPath, includePath)

		// Check if path exists
		info, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}

		// If it's a file, add it directly
		if !info.IsDir() {
			files = append(files, includePath)
			continue
		}

		// Walk directory
		err = filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip files we can't read
			}

			// Skip directories
			if info.IsDir() {
				// Check if we should exclude this directory
				relPath, _ := filepath.Rel(idx.projectPath, path)
				for _, excludePath := range excludePaths {
					if strings.HasPrefix(relPath, strings.TrimSuffix(excludePath, "/")) {
						return filepath.SkipDir
					}
				}
				return nil
			}

			// Check extension
			ext := filepath.Ext(path)
			if len(extMap) > 0 && !extMap[ext] {
				return nil
			}

			// Get relative path
			relPath, err := filepath.Rel(idx.projectPath, path)
			if err != nil {
				return nil
			}

			// Check if excluded
			for _, excludePath := range excludePaths {
				if strings.HasPrefix(relPath, strings.TrimSuffix(excludePath, "/")) {
					return nil
				}
			}

			files = append(files, relPath)
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

// storeChunk stores a chunk in the database
func (idx *Indexer) storeChunk(chunk Chunk, commitSHA string) error {
	// Generate embedding
	embedding, err := idx.embedder.Embed(chunk.Content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Generate content hash
	hash := sha256.Sum256([]byte(chunk.Content))
	contentHash := hex.EncodeToString(hash[:])

	// Prepare metadata
	metadata := map[string]interface{}{
		"file_type": chunk.Type,
	}
	if chunk.Symbol != "" {
		metadata["symbol"] = chunk.Symbol
	}
	if chunk.Component != "" {
		metadata["component"] = chunk.Component
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Convert embedding to postgres array format
	embeddingStr := vectorToPostgresArray(embedding)

	// Insert chunk
	query := `
		INSERT INTO chunks (project_id, source, type, path, language, symbol, component, content, content_hash, embedding, embedding_model, metadata, commit_sha)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (project_id, content_hash) DO UPDATE
		SET updated_at = CURRENT_TIMESTAMP,
		    embedding = EXCLUDED.embedding,
		    embedding_model = EXCLUDED.embedding_model
	`

	_, err = idx.db.Exec(query,
		idx.projectID,
		"repo",
		chunk.Type,
		chunk.Path,
		chunk.Language,
		nullString(chunk.Symbol),
		nullString(chunk.Component),
		chunk.Content,
		contentHash,
		embeddingStr,
		idx.embeddingModel,
		metadataJSON,
		nullString(commitSHA),
	)

	return err
}

// clearExistingChunks clears existing chunks for this project
func (idx *Indexer) clearExistingChunks() error {
	_, err := idx.db.Exec("DELETE FROM chunks WHERE project_id = $1", idx.projectID)
	return err
}

// getGitCommitSHA gets the current git commit SHA
func (idx *Indexer) getGitCommitSHA() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = idx.projectPath
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// saveStats saves indexing stats to stats.json
func (idx *Indexer) saveStats(stats *Stats) error {
	statsPath := filepath.Join(idx.projectPath, ".oview", "index", "stats.json")

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statsPath, data, 0644)
}

// saveManifest saves the file manifest
func (idx *Indexer) saveManifest(manifest *Manifest) error {
	manifestPath := filepath.Join(idx.projectPath, ".oview", "index", "manifest.json")

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(manifestPath, data, 0644)
}

// Helper functions

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func vectorToPostgresArray(vec []float32) string {
	// Convert []float32 to postgres array format: [0.1,0.2,0.3,...]
	parts := make([]string, len(vec))
	for i, v := range vec {
		parts[i] = fmt.Sprintf("%f", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
