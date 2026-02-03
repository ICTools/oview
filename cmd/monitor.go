package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor MCP activity in real-time with visual feedback",
	Long: `Monitor MCP server activity to see when Claude Code uses oview tools.

This command shows a live dashboard of:
- MCP requests from Claude Code
- Tool calls (search, get_context, project_info)
- Response times and result counts
- Total statistics

Perfect for verifying that Claude is actually using oview!`,
	RunE: runMonitor,
}

func init() {
	rootCmd.AddCommand(monitorCmd)
}

type MCPMessage struct {
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Method  string                 `json:"method,omitempty"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Time    time.Time              `json:"time"`
}

type MonitorStats struct {
	TotalRequests   int
	SearchCalls     int
	ContextCalls    int
	ProjectInfoCalls int
	StartTime       time.Time
	LastActivity    time.Time
}

func runMonitor(cmd *cobra.Command, args []string) error {
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║         🔍 oview MCP Activity Monitor                         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("📡 Monitoring MCP activity from Claude Code...")
	fmt.Println("   Waiting for Claude to use oview tools...")
	fmt.Println()
	fmt.Println("💡 TIP: In Claude Code, try:")
	fmt.Println("   > Use search to find authentication code")
	fmt.Println("   > Use project_info")
	fmt.Println("   > Use get_context for cmd/init.go")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop monitoring")
	fmt.Println("─────────────────────────────────────────────────────────────────")
	fmt.Println()

	stats := &MonitorStats{
		StartTime: time.Now(),
	}

	// Read from stdin (expecting MCP server logs piped to this command)
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()

		// Try to parse as JSON
		var msg MCPMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Not JSON, might be regular log
			if strings.Contains(line, "MCP") || strings.Contains(line, "search") {
				fmt.Printf("📝 %s\n", line)
			}
			continue
		}

		msg.Time = time.Now()
		stats.LastActivity = msg.Time
		stats.TotalRequests++

		// Handle different message types
		if strings.Contains(msg.Message, "tools/call") {
			handleToolCall(&msg, stats)
		} else if strings.Contains(msg.Message, "initialize") {
			handleInitialize(&msg)
		} else if strings.Contains(msg.Message, "tools/list") {
			handleToolsList(&msg)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	printFinalStats(stats)
	return nil
}

func handleToolCall(msg *MCPMessage, stats *MonitorStats) {
	timestamp := msg.Time.Format("15:04:05")

	// Try to extract tool name and arguments from params
	toolName := "unknown"
	var args map[string]interface{}

	if msg.Params != nil {
		if name, ok := msg.Params["name"].(string); ok {
			toolName = name
		}
		if arguments, ok := msg.Params["arguments"].(map[string]interface{}); ok {
			args = arguments
		}
	}

	// Update stats
	switch toolName {
	case "search":
		stats.SearchCalls++
	case "get_context":
		stats.ContextCalls++
	case "project_info":
		stats.ProjectInfoCalls++
	}

	// Print colored output
	fmt.Println()
	fmt.Printf("┌─ 🎯 TOOL CALL @ %s ─────────────────────────────────────\n", timestamp)
	fmt.Printf("│\n")

	switch toolName {
	case "search":
		fmt.Printf("│  🔍 SEARCH\n")
		if query, ok := args["query"].(string); ok {
			fmt.Printf("│     Query:  \"%s\"\n", query)
		}
		if limit, ok := args["limit"].(float64); ok {
			fmt.Printf("│     Limit:  %d results\n", int(limit))
		}

	case "get_context":
		fmt.Printf("│  📖 GET CONTEXT\n")
		if path, ok := args["path"].(string); ok {
			fmt.Printf("│     File:   %s\n", path)
		}
		if symbol, ok := args["symbol"].(string); ok && symbol != "" {
			fmt.Printf("│     Symbol: %s\n", symbol)
		}

	case "project_info":
		fmt.Printf("│  ℹ️  PROJECT INFO\n")
		fmt.Printf("│     (No arguments)\n")

	default:
		fmt.Printf("│  ⚙️  %s\n", strings.ToUpper(toolName))
		if len(args) > 0 {
			argsJSON, _ := json.MarshalIndent(args, "│     ", "  ")
			fmt.Printf("│     Args: %s\n", argsJSON)
		}
	}

	fmt.Printf("│\n")
	fmt.Printf("└────────────────────────────────────────────────────────────────\n")

	// Print current stats
	duration := time.Since(stats.StartTime)
	fmt.Printf("\n📊 Stats: %d total | %d search | %d context | %d info | ⏱️  %s\n\n",
		stats.TotalRequests,
		stats.SearchCalls,
		stats.ContextCalls,
		stats.ProjectInfoCalls,
		formatDuration(duration))
}

func handleInitialize(msg *MCPMessage) {
	timestamp := msg.Time.Format("15:04:05")
	fmt.Printf("🔌 [%s] Claude Code connected to MCP server\n", timestamp)
	fmt.Println()
}

func handleToolsList(msg *MCPMessage) {
	timestamp := msg.Time.Format("15:04:05")
	fmt.Printf("📋 [%s] Claude Code requested available tools\n", timestamp)
	fmt.Println()
}

func printFinalStats(stats *MonitorStats) {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("📊 FINAL STATISTICS")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	duration := time.Since(stats.StartTime)

	fmt.Printf("⏱️  Session Duration:     %s\n", formatDuration(duration))
	fmt.Printf("📡 Total MCP Requests:   %d\n", stats.TotalRequests)
	fmt.Println()
	fmt.Printf("🔍 Search calls:         %d\n", stats.SearchCalls)
	fmt.Printf("📖 Get context calls:    %d\n", stats.ContextCalls)
	fmt.Printf("ℹ️  Project info calls:  %d\n", stats.ProjectInfoCalls)
	fmt.Println()

	if stats.TotalRequests > 0 {
		fmt.Println("✅ Claude Code is using oview MCP server!")
	} else {
		fmt.Println("⚠️  No activity detected. Make sure:")
		fmt.Println("   1. Claude Code is running")
		fmt.Println("   2. MCP is configured in ~/.claude/mcp_servers.json")
		fmt.Println("   3. You're asking Claude to use the tools")
	}
	fmt.Println()
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}
