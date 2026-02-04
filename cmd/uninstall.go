package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/docker"
)

var (
	uninstallForce      bool
	uninstallKeepData   bool
	uninstallKeepConfig bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall oview global infrastructure",
	Long: `Removes all oview global infrastructure:
- Stops and removes Docker containers (oview-postgres)
- Removes Docker volumes (data will be lost unless --keep-data is used)
- Removes Docker network
- Removes global configuration file

WARNING: This will delete all project databases!
Use --keep-data to preserve volumes for later reinstall.`,
	RunE: runUninstall,
}

func init() {
	uninstallCmd.Flags().BoolVarP(&uninstallForce, "force", "f", false, "Skip confirmation prompt")
	uninstallCmd.Flags().BoolVar(&uninstallKeepData, "keep-data", false, "Keep Docker volumes (databases)")
	uninstallCmd.Flags().BoolVar(&uninstallKeepConfig, "keep-config", false, "Keep ~/.oview/config.yaml")
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	fmt.Println("🗑️  oview Uninstall")
	fmt.Println()

	// Load config to get container/volume names
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		fmt.Println("⚠️  Could not load config, will use default names")
		globalConfig = config.DefaultGlobalConfig()
	}

	// Show what will be removed
	fmt.Println("The following will be removed:")
	fmt.Printf("  🐳 Container: %s\n", globalConfig.PostgresContainerName)
	fmt.Printf("  🌐 Network:   %s\n", globalConfig.DockerNetworkName)

	if !uninstallKeepData {
		fmt.Printf("  💾 Volume:    %s (⚠️  ALL PROJECT DATABASES)\n", globalConfig.PostgresVolume)
	} else {
		fmt.Println("  ✓ Volumes will be kept (--keep-data)")
	}

	if !uninstallKeepConfig {
		configPath, _ := config.GetConfigPath()
		fmt.Printf("  📄 Config:    %s\n", configPath)
	} else {
		fmt.Println("  ✓ Config will be kept (--keep-config)")
	}

	fmt.Println()

	// Confirmation prompt
	if !uninstallForce {
		fmt.Print("⚠️  Are you sure you want to continue? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("❌ Uninstall cancelled")
			return nil
		}
	}

	fmt.Println()
	fmt.Println("🔧 Starting uninstall process...")
	fmt.Println()

	// Create Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	// Stop and remove Postgres container
	fmt.Printf("🛑 Stopping %s...\n", globalConfig.PostgresContainerName)
	if err := dockerClient.StopContainer(globalConfig.PostgresContainerName); err != nil {
		fmt.Printf("   ⚠️  Warning: %v\n", err)
	} else {
		fmt.Println("   ✓ Stopped")
	}

	fmt.Printf("🗑️  Removing %s...\n", globalConfig.PostgresContainerName)
	if err := dockerClient.RemoveContainer(globalConfig.PostgresContainerName); err != nil {
		fmt.Printf("   ⚠️  Warning: %v\n", err)
	} else {
		fmt.Println("   ✓ Removed")
	}

	// Remove volumes if requested
	if !uninstallKeepData {
		fmt.Println()
		fmt.Println("💾 Removing volumes...")

		// Remove Postgres volume
		fmt.Printf("🗑️  Removing volume %s...\n", globalConfig.PostgresVolume)
		if _, err := dockerClient.RunCommand("volume", "rm", globalConfig.PostgresVolume); err != nil {
			fmt.Printf("   ⚠️  Warning: %v\n", err)
		} else {
			fmt.Println("   ✓ Removed")
		}
	}

	// Remove network
	fmt.Println()
	fmt.Printf("🌐 Removing network %s...\n", globalConfig.DockerNetworkName)
	if _, err := dockerClient.RunCommand("network", "rm", globalConfig.DockerNetworkName); err != nil {
		fmt.Printf("   ⚠️  Warning: %v\n", err)
	} else {
		fmt.Println("   ✓ Removed")
	}

	// Remove config file
	if !uninstallKeepConfig {
		fmt.Println()
		fmt.Println("📄 Removing configuration...")
		configPath, _ := config.GetConfigPath()
		if err := os.Remove(configPath); err != nil {
			if !os.IsNotExist(err) {
				fmt.Printf("   ⚠️  Warning: %v\n", err)
			}
		} else {
			fmt.Println("   ✓ Config file removed")
		}

		// Try to remove .oview directory if empty
		configDir, _ := config.GetConfigDir()
		if err := os.Remove(configDir); err != nil {
			// Ignore error if directory is not empty or doesn't exist
		} else {
			fmt.Println("   ✓ Config directory removed")
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("✅ Uninstall complete!")
	fmt.Println()

	if uninstallKeepData {
		fmt.Println("💡 Your data volumes were preserved.")
		fmt.Println("   To reinstall with existing data: oview install")
		fmt.Println("   To remove data volumes manually:")
		fmt.Printf("     docker volume rm %s\n", globalConfig.PostgresVolume)
	} else {
		fmt.Println("⚠️  All project databases have been deleted.")
		fmt.Println("   To reinstall: oview install")
	}

	if uninstallKeepConfig {
		configPath, _ := config.GetConfigPath()
		fmt.Printf("\n💡 Configuration preserved at: %s\n", configPath)
	}

	fmt.Println()
	fmt.Println("To remove the oview binary itself:")
	fmt.Println("  sudo rm /usr/local/bin/oview")

	return nil
}
