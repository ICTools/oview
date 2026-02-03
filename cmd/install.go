package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/docker"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install global oview infrastructure (Postgres, n8n)",
	Long: `Installs the shared infrastructure needed for all oview projects:
- Creates a shared Docker network
- Starts a Postgres container with pgvector extension
- Starts an n8n container
- Stores configuration in ~/.oview/config.yaml`,
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Installing oview global infrastructure...")
	fmt.Println()

	// Create Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	// Load or create config
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check and adjust ports if needed
	postgresPort, postgresPortChanged, err := docker.EnsurePortAvailable(cfg.PostgresPort)
	if err != nil {
		return fmt.Errorf("failed to find available port for Postgres: %w", err)
	}
	if postgresPortChanged {
		fmt.Printf("⚠️  Port %d is in use, using %d for Postgres instead\n", cfg.PostgresPort, postgresPort)
		cfg.PostgresPort = postgresPort
	}

	n8nPort, n8nPortChanged, err := docker.EnsurePortAvailable(cfg.N8nPort)
	if err != nil {
		return fmt.Errorf("failed to find available port for n8n: %w", err)
	}
	if n8nPortChanged {
		fmt.Printf("⚠️  Port %d is in use, using %d for n8n instead\n", cfg.N8nPort, n8nPort)
		cfg.N8nPort = n8nPort
		cfg.N8nURL = fmt.Sprintf("http://localhost:%d", n8nPort)
	}

	// Step 1: Create Docker network
	fmt.Printf("📡 Creating Docker network '%s'...\n", cfg.DockerNetworkName)
	if err := dockerClient.CreateNetwork(cfg.DockerNetworkName); err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}
	fmt.Println("   ✓ Network ready")

	// Step 2: Create Postgres container
	fmt.Printf("🐘 Creating Postgres container '%s'...\n", cfg.PostgresContainerName)
	if err := dockerClient.CreatePostgresContainer(
		cfg.PostgresContainerName,
		cfg.DockerNetworkName,
		cfg.PostgresVolume,
		cfg.PostgresPassword,
		cfg.PostgresPort,
	); err != nil {
		return fmt.Errorf("failed to create Postgres container: %w", err)
	}
	fmt.Printf("   ✓ Postgres running on port %d\n", cfg.PostgresPort)

	// Step 3: Create n8n container
	fmt.Printf("🤖 Creating n8n container '%s'...\n", cfg.N8nContainerName)
	if err := dockerClient.CreateN8nContainer(
		cfg.N8nContainerName,
		cfg.DockerNetworkName,
		cfg.N8nVolume,
		cfg.N8nPort,
	); err != nil {
		return fmt.Errorf("failed to create n8n container: %w", err)
	}
	fmt.Printf("   ✓ n8n running on %s\n", cfg.N8nURL)

	// Step 4: Save configuration
	fmt.Println("💾 Saving configuration...")
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("   ✓ Configuration saved to %s\n", configPath)

	// Print summary
	fmt.Println()
	fmt.Println("✅ Installation complete!")
	fmt.Println()
	fmt.Println("Connection details:")
	fmt.Printf("  Postgres: localhost:%d\n", cfg.PostgresPort)
	fmt.Printf("  User:     %s\n", cfg.PostgresUser)
	fmt.Printf("  Password: %s\n", cfg.PostgresPassword)
	fmt.Println()
	fmt.Printf("  n8n:      %s\n", cfg.N8nURL)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Navigate to your project directory")
	fmt.Println("  2. Run: oview init")
	fmt.Println("  3. Run: oview up")
	fmt.Println("  4. Run: oview index")

	return nil
}
