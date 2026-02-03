package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/oview/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
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

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCP(cmd *cobra.Command, args []string) error {
	// Get current working directory (project path)
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Log to stderr (stdout is for MCP protocol)
	logToStderr("Starting oview MCP server...")
	logToStderr(fmt.Sprintf("Project path: %s", projectPath))

	// Create and run MCP server
	server := mcp.NewServer(projectPath)
	if err := server.Run(); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}

	return nil
}

// logToStderr logs messages to stderr (stdout is reserved for MCP protocol)
func logToStderr(msg string) {
	logEntry := map[string]interface{}{
		"level":   "info",
		"message": msg,
	}
	json.NewEncoder(os.Stderr).Encode(logEntry)
}
