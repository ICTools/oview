package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/docker"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install global oview infrastructure (Postgres)",
	Long: `Installs the shared infrastructure needed for all oview projects:
- Creates a shared Docker network
- Starts a Postgres container with pgvector extension
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

	// Step 3: Save configuration
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
	fmt.Println("Next steps:")
	fmt.Println("  1. Navigate to your project directory")
	fmt.Println("  2. Run: oview init")
	fmt.Println("  3. Run: oview up")
	fmt.Println("  4. Run: oview index")

	return nil
}
