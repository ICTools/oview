package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/agents"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/detector"
)

var (
	forceInit bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize oview for the current project",
	Long: `Detects the project stack and creates the .oview/ directory structure:
- Detects Symfony, Docker, Makefile, frontend stack
- Creates .oview/project.yaml with project configuration
- Creates .oview/rag.yaml with chunking rules
- Generates Claude agent instruction files in .oview/agents/`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&forceInit, "force", false, "Overwrite existing .oview configuration")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Println("🔍 Initializing oview for this project...")
	fmt.Println()

	// Check if .oview already exists
	oviewDir := filepath.Join(projectPath, ".oview")
	if _, err := os.Stat(oviewDir); err == nil && !forceInit {
		return fmt.Errorf(".oview directory already exists. Use --force to overwrite")
	}

	// Create .oview directory structure
	fmt.Println("📁 Creating .oview directory structure...")
	dirs := []string{
		".oview",
		".oview/agents",
		".oview/index",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectPath, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	fmt.Println("   ✓ Directory structure created")

	// Detect stack
	fmt.Println("🔎 Detecting project stack...")
	detect := detector.New(projectPath)
	stack, err := detect.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect stack: %w", err)
	}

	// Print detected stack
	fmt.Println("   ✓ Stack detected:")
	if stack.Symfony {
		fmt.Println("     - Symfony")
	}
	if stack.Docker {
		fmt.Println("     - Docker")
	}
	if stack.Makefile {
		fmt.Println("     - Makefile")
	}
	if stack.Frontend.Detected {
		fmt.Printf("     - Frontend: %v\n", stack.Frontend.Frameworks)
	}
	if len(stack.Languages) > 0 {
		fmt.Printf("     - Languages: %v\n", stack.Languages)
	}
	if stack.Infrastructure.Redis || stack.Infrastructure.RabbitMQ || stack.Infrastructure.Elasticsearch {
		infra := []string{}
		if stack.Infrastructure.Redis {
			infra = append(infra, "Redis")
		}
		if stack.Infrastructure.RabbitMQ {
			infra = append(infra, "RabbitMQ")
		}
		if stack.Infrastructure.Elasticsearch {
			infra = append(infra, "Elasticsearch")
		}
		fmt.Printf("     - Infrastructure: %v\n", infra)
	}

	// Generate project slug and ID
	slug := detector.GenerateProjectSlug(projectPath)
	projectID := generateProjectID()

	// Detect commands
	commands := detect.DetectCommands(stack)

	// Create project config
	fmt.Println("📝 Creating project configuration...")
	projectConfig := &config.ProjectConfig{
		ProjectID:   projectID,
		ProjectSlug: slug,
		Stack:       *stack,
		Commands:    commands,
		Trello: config.TrelloConfig{
			BoardID: "",
			ListIDs: map[string]string{
				"backlog":        "",
				"todo":           "",
				"in_progress":    "",
				"review":         "",
				"done":           "",
			},
		},
	}

	if err := projectConfig.Save(projectPath); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}
	fmt.Printf("   ✓ Project config saved (slug: %s)\n", slug)

	// Create RAG config
	fmt.Println("📋 Creating RAG configuration...")
	ragConfig := config.DefaultRAGConfig()
	if err := config.SaveRAGConfig(projectPath, ragConfig); err != nil {
		return fmt.Errorf("failed to save RAG config: %w", err)
	}
	fmt.Println("   ✓ RAG config saved")

	// Create empty manifest and stats
	fmt.Println("📊 Creating index manifests...")
	emptyJSON := []byte("{}\n")
	manifestPath := filepath.Join(projectPath, ".oview", "index", "manifest.json")
	if err := os.WriteFile(manifestPath, emptyJSON, 0644); err != nil {
		return fmt.Errorf("failed to create manifest: %w", err)
	}
	statsPath := filepath.Join(projectPath, ".oview", "index", "stats.json")
	if err := os.WriteFile(statsPath, emptyJSON, 0644); err != nil {
		return fmt.Errorf("failed to create stats: %w", err)
	}
	fmt.Println("   ✓ Index manifests created")

	// Generate agent files
	fmt.Println("🤖 Generating Claude agent instruction files...")
	agentGen := agents.New(projectPath, stack)
	if err := agentGen.GenerateAll(); err != nil {
		return fmt.Errorf("failed to generate agent files: %w", err)
	}
	fmt.Println("   ✓ Agent files generated")

	// Summary
	fmt.Println()
	fmt.Println("✅ Initialization complete!")
	fmt.Println()
	fmt.Println("Created:")
	fmt.Println("  .oview/project.yaml     - Project configuration")
	fmt.Println("  .oview/rag.yaml         - RAG indexing rules")
	fmt.Println("  .oview/agents/          - Claude agent instructions")
	fmt.Println("  .oview/index/           - Index metadata (empty)")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review and customize .oview/project.yaml if needed")
	fmt.Println("  2. Add Trello credentials if using Trello integration")
	fmt.Println("  3. Run: oview up")
	fmt.Println("  4. Run: oview index")

	return nil
}

// generateProjectID generates a random project ID
func generateProjectID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
