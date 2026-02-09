package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/logger"
	"github.com/yourusername/oview/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP server and logs management",
	Long: `Manage the Model Context Protocol (MCP) server for Claude Code integration.

Available subcommands:
- oview mcp          Start the MCP server
- oview mcp logs     View MCP server logs in real-time
`,
}

var mcpStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start MCP server for Claude Code integration",
	Long: `Start a Model Context Protocol (MCP) server that exposes oview's RAG capabilities to Claude Code.

The server runs in stdio mode and provides tools for:
- Semantic code search
- Context retrieval
- Project information

Usage with Claude Code:
1. Add to ~/.claude/mcp_servers.json:
   {
     "oview": {
       "command": "oview",
       "args": ["mcp"],
       "cwd": "/path/to/your/project"
     }
   }

2. Claude Code will automatically connect when in the project directory.
`,
	RunE: runMCP,
}

var mcpLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View MCP server logs in real-time",
	Long: `View MCP server logs in real-time using tail -f.

The logs include:
- Server startup/shutdown events
- Incoming MCP requests (method, params)
- Tool calls (search, get_context, project_info)
- Response times and results
- Errors and warnings

Logs are stored in: ~/.oview/mcp.log
`,
	RunE: runMCPLogs,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpStartCmd)
	mcpCmd.AddCommand(mcpLogsCmd)

	// Make 'oview mcp' work without subcommand (defaults to start)
	mcpCmd.RunE = runMCP
}

func runMCP(cmd *cobra.Command, args []string) error {
	// Get current working directory (project path)
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Initialize logger
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	logPath := filepath.Join(homeDir, ".oview", "mcp.log")
	if err := logger.Init(logPath); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Close()

	logger.Info("Starting oview MCP server", map[string]interface{}{
		"project_path": projectPath,
		"log_file":     logPath,
	})

	// Create and run MCP server
	server := mcp.NewServer(projectPath)
	if err := server.Run(); err != nil {
		logger.Error("MCP server error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("MCP server error: %w", err)
	}

	logger.Info("MCP server stopped")
	return nil
}

func runMCPLogs(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	logPath := filepath.Join(homeDir, ".oview", "mcp.log")

	// Check if log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		fmt.Printf("Log file not found: %s\n", logPath)
		fmt.Println("The MCP server may not have been started yet.")
		fmt.Println("\nTo start the MCP server:")
		fmt.Println("  oview mcp")
		return nil
	}

	fmt.Printf("Streaming logs from: %s\n", logPath)
	fmt.Println("Press Ctrl+C to stop\n")
	fmt.Println("---")

	// Use tail -f to follow the log file
	return execTailF(logPath)
}

// execTailF executes tail -f on the log file
func execTailF(logPath string) error {
	cmd := exec.Command("tail", "-f", logPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
