package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/docker"
)

var (
	downStopDocker bool
	downForce      bool
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop running oview processes",
	Long: `Stop all running oview processes (MCP servers, background tasks).

By default, only stops oview processes, not the Docker infrastructure.
Use --docker to also stop the Postgres container.

This is useful when you need to rebuild the binary:
  go build -o oview .
  oview down           # Stop running processes
  sudo cp oview /usr/local/bin/oview`,
	RunE: runDown,
}

func init() {
	downCmd.Flags().BoolVar(&downStopDocker, "docker", false, "Also stop Docker containers")
	downCmd.Flags().BoolVarP(&downForce, "force", "f", false, "Force kill processes (SIGKILL)")
	rootCmd.AddCommand(downCmd)
}

func runDown(cmd *cobra.Command, args []string) error {
	fmt.Println("🛑 Stopping oview processes...")
	fmt.Println()

	// Find and kill oview processes
	killed, err := killOviewProcesses()
	if err != nil {
		fmt.Printf("⚠️  Warning while killing processes: %v\n", err)
	}

	if killed == 0 {
		fmt.Println("✓ No oview processes found running")
	} else {
		fmt.Printf("✓ Stopped %d oview process(es)\n", killed)
	}

	// Stop Docker if requested
	if downStopDocker {
		fmt.Println()
		fmt.Println("🐳 Stopping Docker infrastructure...")
		if err := stopDockerInfra(); err != nil {
			return fmt.Errorf("failed to stop Docker infrastructure: %w", err)
		}
		fmt.Println("   ✓ Docker containers stopped")
	}

	fmt.Println()
	fmt.Println("✅ Done!")
	fmt.Println()
	fmt.Println("You can now rebuild and reinstall the binary:")
	fmt.Println("  go build -o oview .")
	fmt.Println("  sudo cp oview /usr/local/bin/oview")

	return nil
}

// killOviewProcesses finds and kills all oview processes except the current one
func killOviewProcesses() (int, error) {
	currentPID := os.Getpid()

	// Find all oview processes using pgrep
	out, err := exec.Command("pgrep", "-f", "oview").Output()
	if err != nil {
		// pgrep returns exit code 1 if no processes found, which is not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to find oview processes: %w", err)
	}

	pids := strings.Split(strings.TrimSpace(string(out)), "\n")
	killed := 0

	signal := "TERM" // Graceful shutdown
	if downForce {
		signal = "KILL" // Force kill
	}

	for _, pidStr := range pids {
		if pidStr == "" {
			continue
		}

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		// Skip the current process (oview down itself)
		if pid == currentPID {
			continue
		}

		// Get process command line to verify it's actually oview
		cmdlineBytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
		if err != nil {
			continue // Process may have exited
		}

		cmdline := string(cmdlineBytes)
		if !strings.Contains(cmdline, "oview") {
			continue
		}

		// Try to get the command being run (e.g., "oview mcp")
		parts := strings.Split(strings.ReplaceAll(cmdline, "\x00", " "), " ")
		cmdDesc := "oview"
		if len(parts) > 1 && parts[1] != "" {
			cmdDesc = fmt.Sprintf("oview %s", parts[1])
		}

		fmt.Printf("   Stopping %s (PID: %d)\n", cmdDesc, pid)

		// Kill the process
		killCmd := exec.Command("kill", "-"+signal, pidStr)
		if err := killCmd.Run(); err != nil {
			fmt.Printf("   ⚠️  Failed to kill PID %d: %v\n", pid, err)
			continue
		}

		killed++
	}

	return killed, nil
}

// stopDockerInfra stops the Docker infrastructure
func stopDockerInfra() error {
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	// Check if container is running
	running, err := dockerClient.ContainerIsRunning(globalConfig.PostgresContainerName)
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}

	if !running {
		fmt.Println("   ℹ️  Container is not running")
		return nil
	}

	// Stop the container
	fmt.Printf("   Stopping %s...\n", globalConfig.PostgresContainerName)
	if err := dockerClient.StopContainer(globalConfig.PostgresContainerName); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	return nil
}
