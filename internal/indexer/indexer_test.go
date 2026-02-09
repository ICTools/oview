package indexer

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/embeddings"
)

func TestNew(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	ragConfig := &config.RAGConfig{}
	embedder := embeddings.NewStubGenerator(1536)

	indexer := New("/test/path", "test-project-id", db, ragConfig, embedder, "test-model")

	assert.NotNil(t, indexer)
	assert.Equal(t, "/test/path", indexer.projectPath)
	assert.Equal(t, "test-project-id", indexer.projectID)
	assert.NotNil(t, indexer.chunker)
	assert.NotNil(t, indexer.embedder)
	assert.Equal(t, "test-model", indexer.embeddingModel)
}

func TestNew_WithNilEmbedder(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	ragConfig := &config.RAGConfig{}

	// Pass nil embedder - should default to stub
	indexer := New("/test/path", "test-project-id", db, ragConfig, nil, "")

	assert.NotNil(t, indexer)
	assert.NotNil(t, indexer.embedder)
	assert.Contains(t, indexer.embeddingModel, "stub")
}

func TestIndexer_storeChunk(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	ragConfig := &config.RAGConfig{}
	embedder := embeddings.NewStubGenerator(1536)
	indexer := New("/test/path", "test-project-id", db, ragConfig, embedder, "test-model")

	chunk := Chunk{
		Path:      "test.php",
		Language:  "PHP",
		Symbol:    "TestClass",
		Component: "Controller",
		Content:   "<?php class TestClass {}",
		Type:      "code",
	}

	// Expect INSERT query
	mock.ExpectExec("INSERT INTO chunks").
		WithArgs(
			"test-project-id",
			"repo",
			"code",
			"test.php",
			"PHP",
			"TestClass",
			"Controller",
			chunk.Content,
			sqlmock.AnyArg(), // content_hash
			sqlmock.AnyArg(), // embedding
			"test-model",
			sqlmock.AnyArg(), // metadata
			"abc123",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = indexer.storeChunk(chunk, "abc123")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndexer_storeChunk_WithNullFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	ragConfig := &config.RAGConfig{}
	embedder := embeddings.NewStubGenerator(1536)
	indexer := New("/test/path", "test-project-id", db, ragConfig, embedder, "test-model")

	chunk := Chunk{
		Path:     "test.txt",
		Language: "Text",
		Content:  "Plain text content",
		Type:     "doc",
		// Symbol and Component are empty
	}

	// Expect INSERT with NULL for symbol and component
	mock.ExpectExec("INSERT INTO chunks").
		WithArgs(
			"test-project-id",
			"repo",
			"doc",
			"test.txt",
			"Text",
			nil, // symbol is NULL
			nil, // component is NULL
			chunk.Content,
			sqlmock.AnyArg(), // content_hash
			sqlmock.AnyArg(), // embedding
			"test-model",
			sqlmock.AnyArg(), // metadata
			nil, // commit_sha is NULL
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = indexer.storeChunk(chunk, "")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndexer_clearExistingChunks(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	ragConfig := &config.RAGConfig{}
	embedder := embeddings.NewStubGenerator(1536)
	indexer := New("/test/path", "test-project-id", db, ragConfig, embedder, "test-model")

	mock.ExpectExec("DELETE FROM chunks WHERE project_id").
		WithArgs("test-project-id").
		WillReturnResult(sqlmock.NewResult(0, 5))

	err = indexer.clearExistingChunks()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndexer_scanFiles(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "vendor"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755)

	os.WriteFile(filepath.Join(tmpDir, "src", "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src", "util.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "vendor", "lib.go"), []byte("package vendor"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "node_modules", "index.js"), []byte("// js"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Readme"), 0644)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	ragConfig := &config.RAGConfig{
		Indexing: config.IndexingRules{
			IncludePaths: []string{"src", "README.md"},
			ExcludePaths: []string{"vendor/", "node_modules/"},
			Extensions:   []string{".go", ".md"},
		},
	}

	embedder := embeddings.NewStubGenerator(1536)
	indexer := New(tmpDir, "test-project", db, ragConfig, embedder, "test-model")

	files, err := indexer.scanFiles()
	assert.NoError(t, err)

	// Should include src/*.go and README.md
	// Should exclude vendor/ and node_modules/
	assert.Contains(t, files, "src/main.go")
	assert.Contains(t, files, "src/util.go")
	assert.Contains(t, files, "README.md")
	assert.NotContains(t, files, "vendor/lib.go")
	assert.NotContains(t, files, "node_modules/index.js")
}

func TestIndexer_scanFiles_WithExtensionFilter(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files with different extensions
	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "src", "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src", "test.js"), []byte("// js"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src", "data.json"), []byte("{}"), 0644)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	ragConfig := &config.RAGConfig{
		Indexing: config.IndexingRules{
			IncludePaths: []string{"src"},
			ExcludePaths: []string{},
			Extensions:   []string{".go"}, // Only .go files
		},
	}

	embedder := embeddings.NewStubGenerator(1536)
	indexer := New(tmpDir, "test-project", db, ragConfig, embedder, "test-model")

	files, err := indexer.scanFiles()
	assert.NoError(t, err)

	// Should only include .go files
	assert.Contains(t, files, "src/main.go")
	assert.NotContains(t, files, "src/test.js")
	assert.NotContains(t, files, "src/data.json")
	assert.Equal(t, 1, len(files))
}

func TestIndexer_saveStats(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(filepath.Join(tmpDir, ".oview", "index"), 0755)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	ragConfig := &config.RAGConfig{}
	embedder := embeddings.NewStubGenerator(1536)
	indexer := New(tmpDir, "test-project", db, ragConfig, embedder, "test-model")

	stats := &Stats{
		CommitSHA:    "abc123",
		FilesIndexed: 10,
		ChunksStored: 50,
		TotalBytes:   1024,
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(1 * time.Minute),
		Duration:     "1m0s",
	}

	err = indexer.saveStats(stats)
	assert.NoError(t, err)

	// Verify file was created
	statsPath := filepath.Join(tmpDir, ".oview", "index", "stats.json")
	_, err = os.Stat(statsPath)
	assert.NoError(t, err)

	// Verify file contains valid JSON
	data, err := os.ReadFile(statsPath)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "abc123")
	assert.Contains(t, string(data), "files_indexed")
}

func TestIndexer_saveManifest(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(filepath.Join(tmpDir, ".oview", "index"), 0755)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	ragConfig := &config.RAGConfig{}
	embedder := embeddings.NewStubGenerator(1536)
	indexer := New(tmpDir, "test-project", db, ragConfig, embedder, "test-model")

	manifest := &Manifest{
		Files: map[string]FileInfo{
			"main.go": {
				Path:      "main.go",
				Hash:      "abc123",
				Chunks:    5,
				IndexedAt: time.Now(),
			},
		},
		LastUpdate: time.Now(),
	}

	err = indexer.saveManifest(manifest)
	assert.NoError(t, err)

	// Verify file was created
	manifestPath := filepath.Join(tmpDir, ".oview", "index", "manifest.json")
	_, err = os.Stat(manifestPath)
	assert.NoError(t, err)

	// Verify file contains valid JSON
	data, err := os.ReadFile(manifestPath)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "main.go")
	assert.Contains(t, string(data), "abc123")
}

func TestHelperFunctions_Indexer(t *testing.T) {
	t.Run("nullString", func(t *testing.T) {
		// Empty string should return nil
		result := nullString("")
		assert.Nil(t, result)

		// Non-empty string should return the string
		result = nullString("test")
		assert.Equal(t, "test", result)
	})

	t.Run("vectorToPostgresArray", func(t *testing.T) {
		vec := []float32{0.1, 0.2, 0.3}
		result := vectorToPostgresArray(vec)
		assert.Contains(t, result, "[")
		assert.Contains(t, result, "]")
		assert.Contains(t, result, ",")
		assert.Contains(t, result, "0.1")
		assert.Contains(t, result, "0.2")
		assert.Contains(t, result, "0.3")
	})
}

func TestIndexer_Index_FullWorkflow(t *testing.T) {
	// Save current directory and restore after test
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(cwd)

	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to temp directory so relative paths work
	os.Chdir(tmpDir)

	// Create test files
	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, ".oview", "index"), 0755)
	testFilePath := filepath.Join(tmpDir, "src", "test.go")
	os.WriteFile(testFilePath, []byte("package main\nfunc main() {}"), 0644)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	ragConfig := &config.RAGConfig{
		Indexing: config.IndexingRules{
			IncludePaths: []string{"src"},
			ExcludePaths: []string{},
			Extensions:   []string{".go"},
		},
		Chunking: config.ChunkingRules{
			Generic: config.ChunkRule{MaxSize: 2000},
		},
	}

	embedder := embeddings.NewStubGenerator(1536)
	indexer := New(tmpDir, "test-project", db, ragConfig, embedder, "test-model")

	// Mock DELETE existing chunks
	mock.ExpectExec("DELETE FROM chunks WHERE project_id").
		WithArgs("test-project").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Mock INSERT for chunk
	mock.ExpectExec("INSERT INTO chunks").
		WithArgs(
			sqlmock.AnyArg(), // project_id
			sqlmock.AnyArg(), // source
			sqlmock.AnyArg(), // type
			sqlmock.AnyArg(), // path
			sqlmock.AnyArg(), // language
			sqlmock.AnyArg(), // symbol
			sqlmock.AnyArg(), // component
			sqlmock.AnyArg(), // content
			sqlmock.AnyArg(), // content_hash
			sqlmock.AnyArg(), // embedding
			sqlmock.AnyArg(), // embedding_model
			sqlmock.AnyArg(), // metadata
			sqlmock.AnyArg(), // commit_sha
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	stats, err := indexer.Index()
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats.FilesIndexed)
	assert.Equal(t, 1, stats.ChunksStored)
	assert.Greater(t, stats.TotalBytes, int64(0))
	assert.NotEmpty(t, stats.Duration)

	// Verify stats file was created
	statsPath := filepath.Join(tmpDir, ".oview", "index", "stats.json")
	_, err = os.Stat(statsPath)
	assert.NoError(t, err)

	// Verify manifest file was created
	manifestPath := filepath.Join(tmpDir, ".oview", "index", "manifest.json")
	_, err = os.Stat(manifestPath)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndexer_getGitCommitSHA(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	ragConfig := &config.RAGConfig{}
	embedder := embeddings.NewStubGenerator(1536)

	// Test in non-git directory
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	indexer := New(tmpDir, "test-project", db, ragConfig, embedder, "test-model")
	sha := indexer.getGitCommitSHA()

	// Should return empty string in non-git directory
	assert.Equal(t, "", sha)
}
