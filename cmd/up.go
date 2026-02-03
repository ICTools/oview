package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/database"
	"github.com/yourusername/oview/internal/docker"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start the project runtime (create database, setup pgvector)",
	Long: `Ensures the project runtime is ready:
- Checks that global infrastructure is running
- Creates a project-specific database in the shared Postgres
- Creates a database user for the project
- Enables the pgvector extension
- Creates RAG indexing tables`,
	RunE: runUp,
}

func init() {
	rootCmd.AddCommand(upCmd)
}

func runUp(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Println("🚀 Starting project runtime...")
	fmt.Println()

	// Load project config
	fmt.Println("📋 Loading project configuration...")
	projectConfig, err := config.LoadProjectConfig(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load project config: %w\nHint: Run 'oview init' first", err)
	}
	fmt.Printf("   ✓ Project: %s (slug: %s)\n", projectConfig.ProjectID, projectConfig.ProjectSlug)

	// Load global config
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Check if infrastructure is running
	fmt.Println("🔍 Checking global infrastructure...")
	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	running, err := dockerClient.ContainerIsRunning(globalConfig.PostgresContainerName)
	if err != nil {
		return fmt.Errorf("failed to check Postgres status: %w", err)
	}
	if !running {
		return fmt.Errorf("Postgres container is not running. Please run 'oview install' first")
	}
	fmt.Println("   ✓ Global infrastructure is running")

	// Generate database credentials if not already set
	dbName := fmt.Sprintf("oview_%s", projectConfig.ProjectSlug)
	dbUser := fmt.Sprintf("oview_%s", projectConfig.ProjectSlug)
	dbPassword := ""

	if projectConfig.Database.Password != "" {
		dbPassword = projectConfig.Database.Password
	} else {
		// Use a simple fixed password for local development
		// No security concern as it's local-only and not exposed
		dbPassword = "oview_dev"
	}

	// Connect to Postgres
	fmt.Println("🔗 Connecting to Postgres...")
	masterDSN := globalConfig.GetDSN("postgres")
	dbClient, err := database.NewClient(masterDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to Postgres: %w", err)
	}
	defer dbClient.Close()
	fmt.Println("   ✓ Connected")

	// Create database
	fmt.Printf("💾 Creating database '%s'...\n", dbName)
	if err := dbClient.CreateDatabase(dbName); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	fmt.Println("   ✓ Database created")

	// Create user
	fmt.Printf("👤 Creating database user '%s'...\n", dbUser)
	if err := dbClient.CreateUser(dbUser, dbPassword); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	fmt.Println("   ✓ User created")

	// Grant access
	fmt.Println("🔐 Granting access permissions...")
	if err := dbClient.GrantAccess(dbName, dbUser); err != nil {
		return fmt.Errorf("failed to grant access: %w", err)
	}
	fmt.Println("   ✓ Permissions granted")

	// Enable pgvector extension
	fmt.Println("🧮 Enabling pgvector extension...")
	if err := dbClient.EnableExtension(dbName, "vector"); err != nil {
		return fmt.Errorf("failed to enable pgvector: %w", err)
	}
	fmt.Println("   ✓ pgvector enabled")

	// Create schema with the configured embedding dimension
	embeddingDim := projectConfig.Embeddings.Dim
	if embeddingDim == 0 {
		embeddingDim = 1536 // Default if not configured
	}
	fmt.Printf("🏗️  Creating RAG schema (embedding dimension: %d)...\n", embeddingDim)
	if err := dbClient.CreateSchema(dbName, embeddingDim); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	fmt.Println("   ✓ Schema created")

	// Update project config with database info
	projectConfig.Database = config.DatabaseConfig{
		Name:     dbName,
		User:     dbUser,
		Password: dbPassword,
	}
	if err := projectConfig.Save(projectPath); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}

	// Print summary
	projectDSN := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbUser, dbPassword, globalConfig.PostgresHost, globalConfig.PostgresPort, dbName)

	fmt.Println()
	fmt.Println("✅ Project runtime is ready!")
	fmt.Println()
	fmt.Println("Database connection:")
	fmt.Printf("  DSN: %s\n", projectDSN)
	fmt.Println()
	fmt.Println("n8n workflow engine:")
	fmt.Printf("  URL: %s\n", globalConfig.N8nURL)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run: oview index")
	fmt.Println("  2. Start building your automation workflows")

	return nil
}

// generatePassword generates a random password
func generatePassword() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
