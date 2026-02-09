package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	markerStart = "<!-- OVIEW_MCP_RAG_FIRST_START -->"
	markerEnd   = "<!-- OVIEW_MCP_RAG_FIRST_END -->"
)

// UpsertOviewRagFirstSection adds or updates the oview RAG-first policy section in CLAUDE.md
func UpsertOviewRagFirstSection(projectRoot string) error {
	claudeMdPath := filepath.Join(projectRoot, "CLAUDE.md")

	// Read existing CLAUDE.md
	content, err := os.ReadFile(claudeMdPath)
	if err != nil {
		return fmt.Errorf("failed to read CLAUDE.md: %w", err)
	}

	currentContent := string(content)
	ragSection := generateRagFirstSection()

	// Check if markers exist
	startIdx := strings.Index(currentContent, markerStart)
	endIdx := strings.Index(currentContent, markerEnd)

	var newContent string

	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		// Markers exist - replace content between them
		before := currentContent[:startIdx]
		after := currentContent[endIdx+len(markerEnd):]
		newContent = before + ragSection + after
	} else {
		// Markers don't exist - append section at the end
		// Ensure proper spacing
		trimmed := strings.TrimRight(currentContent, "\n")
		newContent = trimmed + "\n\n" + ragSection
	}

	// Write updated content
	if err := os.WriteFile(claudeMdPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write CLAUDE.md: %w", err)
	}

	return nil
}

// generateRagFirstSection generates the oview RAG-first policy section
func generateRagFirstSection() string {
	// Split the template to avoid backticks inside raw string literals
	part1 := markerStart + `
## oview MCP RAG-First Policy

**CRITICAL INSTRUCTION: When working with this codebase, you MUST use oview MCP tools FIRST for all codebase understanding tasks.**

### Mandatory Tool Usage Order

1. **ALWAYS start with oview MCP tools** when you need to:
   - Understand code architecture or structure
   - Find where specific functionality is implemented
   - Locate files related to a feature or concept
   - Search for patterns, configurations, or implementations

2. **Use semantic search FIRST**:
   - Use the 'search' tool with natural language queries
   - Apply filters when relevant: 'type', 'path', 'language', 'component'
   - Examples of effective queries:
     - "authentication flow"
     - "security.yaml firewall configuration"
     - "messenger rabbitmq transport setup"
     - "redis cache configuration"
     - "elasticsearch mapping definitions"

3. **Use 'get_context' ONLY after identifying relevant files**:
   - After 'search' returns relevant files, use 'get_context' to retrieve full content
   - Specify exact file paths returned by search results
   - Request only the minimal set of files needed

4. **Open/edit files only when necessary**:
   - After using MCP tools to understand the codebase
   - Limit to the specific files identified through semantic search
   - Never scan or grep files manually when MCP tools are available

### Fallback Behavior

**If MCP tools are unavailable or return errors:**
- Explicitly state: "oview MCP is not available, falling back to manual exploration"
- Only then use grep, glob, or file scanning as alternatives
- Inform the user that semantic search would be more effective

### MCP Configuration

To enable oview MCP integration with Claude Code, add this to your ~/.claude/mcp_servers.json:

`

	jsonBlock := "```json\n" + `{
  "mcpServers": {
    "oview": {
      "command": "oview",
      "args": ["mcp"],
      "cwd": "."
    }
  }
}
` + "```\n"

	part2 := `
A copy of this configuration is available in .oview/claude_mcp.json for easy copying.

### Example Workflow

**Correct approach:**
1. Use 'search' tool: "user authentication implementation"
2. Review semantic search results with similarity scores
3. Use 'get_context' for top-ranked files
4. Open/edit only the identified files

**Incorrect approach (DON'T DO THIS):**
1. ~~Grep for "auth" across the codebase~~
2. ~~Manually browse directory structure~~
3. ~~Open multiple files speculatively~~

### Benefits of RAG-First Approach

- **Semantic understanding**: Find code by intent, not just keywords
- **Ranked results**: See most relevant files first by similarity score
- **Context-aware**: Understands relationships between code components
- **Efficient**: Avoid scanning thousands of files manually
- **Accurate**: Vector embeddings capture code meaning beyond text matching

**Remember: oview MCP tools are your PRIMARY interface for understanding this codebase. Use them first, always.**
` + markerEnd

	return part1 + jsonBlock + part2
}
