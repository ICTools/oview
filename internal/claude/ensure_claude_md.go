package claude

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ClaudeMDStatus represents the status of CLAUDE.md creation
type ClaudeMDStatus string

const (
	StatusAlreadyExists  ClaudeMDStatus = "already_exists"
	StatusCreatedViaCLI  ClaudeMDStatus = "created_via_cli"
	StatusCreatedFallback ClaudeMDStatus = "created_fallback"
)

// EnsureClaudeMd ensures that CLAUDE.md exists in the project root
// It attempts to create it via Claude CLI first, then falls back to a template
func EnsureClaudeMd(projectRoot string) (ClaudeMDStatus, error) {
	claudeMdPath := filepath.Join(projectRoot, "CLAUDE.md")

	// Check if CLAUDE.md already exists
	if _, err := os.Stat(claudeMdPath); err == nil {
		return StatusAlreadyExists, nil
	}

	// Suggest using claude /init if available
	if IsClaudeCLIAvailable() {
		fmt.Println()
		fmt.Println("   💡 To create a rich CLAUDE.md, run: claude /init")
		fmt.Println("   ℹ️  Using minimal fallback template for now...")
		fmt.Println()
	}

	// Fallback: create a minimal template
	if err := createFallbackClaudeMd(claudeMdPath); err != nil {
		return "", fmt.Errorf("failed to create CLAUDE.md: %w", err)
	}

	return StatusCreatedFallback, nil
}


// createFallbackClaudeMd creates a minimal CLAUDE.md template
func createFallbackClaudeMd(path string) error {
	template := generateFallbackTemplate()
	return os.WriteFile(path, []byte(template), 0644)
}

// generateFallbackTemplate generates a minimal CLAUDE.md template
func generateFallbackTemplate() string {
	return `# CLAUDE.md

**CRITICAL: Use oview MCP search tool for ALL code queries. Don't rely on this file for implementation details.**

## Project Overview

This project uses oview for RAG-powered code search. Use the MCP search tool to explore the codebase semantically.

## Essential Commands

` + "```bash" + `
# Build
go build -o oview .

# Run
./oview <command>

# Test
go test ./...
` + "```" + `

## How to Explore This Codebase

**DO NOT look for implementation details in this file.** Instead:

1. **Search semantically**: Ask questions like "how does the indexer work?" or "where is database schema defined?"
2. **Use MCP tools**: The oview MCP server provides semantic search across all code
3. **Get context**: Use get_context for specific files you want to modify

## Architecture Overview

Use MCP search to explore:
- "indexing pipeline" - How code is chunked and embedded
- "MCP server" - How the RAG interface works
- "database schema" - PostgreSQL + pgvector structure
- "embeddings" - Ollama/OpenAI integration

For detailed documentation, check the docs/ directory or use MCP search.
`
}

// IsClaudeCLIAvailable checks if the Claude CLI is installed
func IsClaudeCLIAvailable() bool {
	var cmd *exec.Cmd

	// The check differs by platform
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("where", "claude")
	default:
		cmd = exec.Command("which", "claude")
	}

	return cmd.Run() == nil
}

