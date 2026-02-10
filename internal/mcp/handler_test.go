package mcp

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/embeddings"
	"github.com/yourusername/oview/internal/query"
)

func TestNewToolHandler(t *testing.T) {
	projectConfig := &config.ProjectConfig{
		ProjectID:   "test-project",
		ProjectSlug: "test-project",
	}
	globalConfig := &config.GlobalConfig{}

	handler := NewToolHandler(projectConfig, globalConfig)

	assert.NotNil(t, handler)
	assert.Equal(t, projectConfig, handler.projectConfig)
	assert.Equal(t, globalConfig, handler.globalConfig)
	assert.Nil(t, handler.db) // Not connected yet
	assert.Nil(t, handler.generator) // Not initialized yet
}

func TestToolHandler_CallTool_UnknownTool(t *testing.T) {
	projectConfig := &config.ProjectConfig{}
	globalConfig := &config.GlobalConfig{}
	handler := NewToolHandler(projectConfig, globalConfig)

	result, err := handler.CallTool(context.Background(), "unknown_tool", map[string]interface{}{})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unknown tool")
}

func TestToolHandler_handleProjectInfo(t *testing.T) {
	projectConfig := &config.ProjectConfig{
		ProjectID:   "test-project-id",
		ProjectSlug: "test-project-slug",
		Embeddings: config.EmbeddingsConfig{
			Provider: "ollama",
			Model:    "nomic-embed-text",
			Dim:      768,
		},
		LLM: config.LLMConfig{
			Provider: "claude-code",
			Model:    "sonnet",
		},
		Database: config.DatabaseConfig{
			Name: "oview_test_project",
		},
		Stack: config.StackInfo{
			Symfony: true,
			Docker:  true,
		},
	}
	globalConfig := &config.GlobalConfig{}
	handler := NewToolHandler(projectConfig, globalConfig)

	result, err := handler.handleProjectInfo(context.Background(), map[string]interface{}{})

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test-project-id", resultMap["project_id"])
	assert.Equal(t, "test-project-slug", resultMap["project_slug"])

	embeddings, ok := resultMap["embeddings"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "ollama", embeddings["provider"])
	assert.Equal(t, "nomic-embed-text", embeddings["model"])
	assert.Equal(t, 768, embeddings["dim"])

	llm, ok := resultMap["llm"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "claude-code", llm["provider"])
	assert.Equal(t, "sonnet", llm["model"])

	db, ok := resultMap["database"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "oview_test_project", db["name"])

	stack, ok := resultMap["stack"].(config.StackInfo)
	assert.True(t, ok)
	assert.True(t, stack.Symfony)
	assert.True(t, stack.Docker)
}

func TestToolHandler_handleGetContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	projectConfig := &config.ProjectConfig{
		ProjectID: "test-project",
	}
	globalConfig := &config.GlobalConfig{}
	handler := NewToolHandler(projectConfig, globalConfig)
	handler.db = db

	// Mock query for get_context
	rows := sqlmock.NewRows([]string{"id", "path", "type", "language", "symbol", "content", "similarity"}).
		AddRow(1, "test.php", "code", "PHP", "TestClass", "<?php class TestClass {}", 0.0).
		AddRow(2, "test.php", "code", "PHP", "TestClass::method", "public function method() {}", 0.0)

	mock.ExpectQuery("SELECT(.+)FROM chunks(.+)WHERE project_id(.+)AND path").
		WithArgs("test-project", "test.php", 3).
		WillReturnRows(rows)

	args := map[string]interface{}{
		"path":  "test.php",
		"limit": float64(3),
	}

	result, err := handler.handleGetContext(context.Background(), args)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test.php", resultMap["path"])
	assert.Equal(t, 2, resultMap["count"])

	context, ok := resultMap["context"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(context))
	assert.Equal(t, "test.php", context[0]["path"])
	assert.Equal(t, "TestClass", context[0]["symbol"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestToolHandler_handleGetContext_WithSymbol(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	projectConfig := &config.ProjectConfig{
		ProjectID: "test-project",
	}
	globalConfig := &config.GlobalConfig{}
	handler := NewToolHandler(projectConfig, globalConfig)
	handler.db = db

	// Mock query for get_context with specific symbol
	rows := sqlmock.NewRows([]string{"id", "path", "type", "language", "symbol", "content", "similarity"}).
		AddRow(1, "test.php", "code", "PHP", "TestClass", "<?php class TestClass {}", 0.0)

	mock.ExpectQuery("SELECT(.+)FROM chunks(.+)UNION ALL(.+)").
		WithArgs("test-project", "test.php", "TestClass", 3).
		WillReturnRows(rows)

	args := map[string]interface{}{
		"path":   "test.php",
		"symbol": "TestClass",
		"limit":  float64(3),
	}

	result, err := handler.handleGetContext(context.Background(), args)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test.php", resultMap["path"])
	assert.Equal(t, "TestClass", resultMap["symbol"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestToolHandler_handleGetContext_MissingPath(t *testing.T) {
	projectConfig := &config.ProjectConfig{}
	globalConfig := &config.GlobalConfig{}
	handler := NewToolHandler(projectConfig, globalConfig)

	args := map[string]interface{}{}

	result, err := handler.handleGetContext(context.Background(), args)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "path is required")
}

func TestToolHandler_handleSearch_MissingQuery(t *testing.T) {
	projectConfig := &config.ProjectConfig{}
	globalConfig := &config.GlobalConfig{}
	handler := NewToolHandler(projectConfig, globalConfig)

	args := map[string]interface{}{}

	result, err := handler.handleSearch(context.Background(), args)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "query is required")
}

func TestToolHandler_handleSearch_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	projectConfig := &config.ProjectConfig{
		ProjectID: "test-project",
		Embeddings: config.EmbeddingsConfig{
			Provider: "stub",
			Model:    "stub",
			Dim:      1536,
		},
	}
	globalConfig := &config.GlobalConfig{}
	handler := NewToolHandler(projectConfig, globalConfig)
	handler.db = db
	handler.generator = embeddings.NewStubGenerator(1536)

	// Mock query for search
	rows := sqlmock.NewRows([]string{"id", "path", "type", "language", "symbol", "component", "content", "similarity"}).
		AddRow(1, "src/Controller.php", "code", "PHP", "UserController", "Controller", "<?php class UserController {}", 0.95)

	mock.ExpectQuery("SELECT(.+)FROM chunks(.+)WHERE(.+)language IN").
		WillReturnRows(rows)

	args := map[string]interface{}{
		"query":    "authentication logic",
		"language": "PHP",
		"limit":    float64(5),
	}

	result, err := handler.handleSearch(context.Background(), args)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "authentication logic", resultMap["query"])
	assert.Equal(t, 1, resultMap["count"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestToolHandler_initGenerator_OpenAI(t *testing.T) {
	projectConfig := &config.ProjectConfig{
		Embeddings: config.EmbeddingsConfig{
			Provider: "openai",
			Model:    "text-embedding-3-small",
			APIKey:   "test-key",
		},
	}
	globalConfig := &config.GlobalConfig{}
	handler := NewToolHandler(projectConfig, globalConfig)

	err := handler.initGenerator()

	assert.NoError(t, err)
	assert.NotNil(t, handler.generator)
	assert.Contains(t, handler.generator.Name(), "OpenAI")
}

func TestToolHandler_initGenerator_Ollama(t *testing.T) {
	projectConfig := &config.ProjectConfig{
		Embeddings: config.EmbeddingsConfig{
			Provider: "ollama",
			Model:    "nomic-embed-text",
			BaseURL:  "http://localhost:11434",
		},
	}
	globalConfig := &config.GlobalConfig{}
	handler := NewToolHandler(projectConfig, globalConfig)

	err := handler.initGenerator()

	assert.NoError(t, err)
	assert.NotNil(t, handler.generator)
	assert.Contains(t, handler.generator.Name(), "Ollama")
}

func TestToolHandler_initGenerator_UnsupportedProvider(t *testing.T) {
	projectConfig := &config.ProjectConfig{
		Embeddings: config.EmbeddingsConfig{
			Provider: "unsupported",
		},
	}
	globalConfig := &config.GlobalConfig{}
	handler := NewToolHandler(projectConfig, globalConfig)

	err := handler.initGenerator()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported embeddings provider")
}

func TestFormatFilters(t *testing.T) {
	filters := &query.SearchFilters{
		Languages:     []string{"PHP", "JavaScript"},
		Types:         []string{"code"},
		PathPattern:   "src/*",
		Components:    []string{"Controller", "Service"},
		SymbolPattern: "User*",
		MinSimilarity: 0.8,
		Limit:         10,
	}

	result := formatFilters(filters)

	assert.NotNil(t, result)
	assert.Equal(t, []string{"PHP", "JavaScript"}, result["languages"])
	assert.Equal(t, []string{"code"}, result["types"])
	assert.Equal(t, "src/*", result["path_pattern"])
	assert.Equal(t, []string{"Controller", "Service"}, result["components"])
	assert.Equal(t, "User*", result["symbol_pattern"])
	assert.Equal(t, 0.8, result["min_similarity"])
}

func TestFormatFilters_Empty(t *testing.T) {
	filters := &query.SearchFilters{
		Limit: 5,
	}

	result := formatFilters(filters)

	assert.NotNil(t, result)
	// Should not contain empty fields
	_, hasLanguages := result["languages"]
	assert.False(t, hasLanguages)
}

func TestEmbeddingToString(t *testing.T) {
	embedding := []float32{0.1, 0.2, 0.3}

	result := embeddingToString(embedding)

	assert.Contains(t, result, "[")
	assert.Contains(t, result, "]")
	assert.Contains(t, result, "0.1")
	assert.Contains(t, result, "0.2")
	assert.Contains(t, result, "0.3")
	assert.Contains(t, result, ",")
}
