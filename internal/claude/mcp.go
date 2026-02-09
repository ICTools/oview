package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MCPServerConfig represents the MCP server configuration snippet
type MCPServerConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// MCPServer represents a single MCP server configuration
type MCPServer struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	CWD     string   `json:"cwd"`
}

// WriteClaudeMcpSnippet creates .oview/claude_mcp.json with the MCP configuration snippet
func WriteClaudeMcpSnippet(projectRoot string) error {
	config := MCPServerConfig{
		MCPServers: map[string]MCPServer{
			"oview": {
				Command: "oview",
				Args:    []string{"mcp"},
				CWD:     ".",
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal MCP config: %w", err)
	}

	snippetPath := filepath.Join(projectRoot, ".oview", "claude_mcp.json")
	if err := os.WriteFile(snippetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write MCP snippet: %w", err)
	}

	return nil
}

// AddToClaudeMcpConfig adds or updates the oview MCP server configuration in ~/.claude/mcp_servers.json
func AddToClaudeMcpConfig(projectRoot string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	mcpConfigPath := filepath.Join(claudeDir, "mcp_servers.json")

	// Ensure .claude directory exists
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Read existing config or create new one
	var existingConfig MCPServerConfig
	if data, err := os.ReadFile(mcpConfigPath); err == nil {
		if err := json.Unmarshal(data, &existingConfig); err != nil {
			// Corrupted config file, backup and start fresh
			backupPath := mcpConfigPath + ".backup"
			os.WriteFile(backupPath, data, 0644)
			existingConfig = MCPServerConfig{MCPServers: make(map[string]MCPServer)}
		}
	} else {
		// File doesn't exist, create new config
		existingConfig = MCPServerConfig{MCPServers: make(map[string]MCPServer)}
	}

	// Add or update oview server config
	existingConfig.MCPServers["oview"] = MCPServer{
		Command: "oview",
		Args:    []string{"mcp"},
		CWD:     projectRoot,
	}

	// Write updated config
	data, err := json.MarshalIndent(existingConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal MCP config: %w", err)
	}

	if err := os.WriteFile(mcpConfigPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write MCP config: %w", err)
	}

	return nil
}
