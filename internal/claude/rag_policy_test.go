package claude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpsertOviewRagFirstSection(t *testing.T) {
	tests := []struct {
		name           string
		initialContent string
		wantContains   []string
		wantNotContain []string
		runTwice       bool // Test idempotence
	}{
		{
			name: "append to file without markers",
			initialContent: `# CLAUDE.md

This is a test project.

## Existing Section

Some existing content.
`,
			wantContains: []string{
				"# CLAUDE.md",
				"This is a test project",
				"## Existing Section",
				markerStart,
				markerEnd,
				"oview MCP RAG-First Policy",
				"authentication flow",
			},
			wantNotContain: []string{},
			runTwice:       true,
		},
		{
			name: "replace existing section between markers",
			initialContent: `# CLAUDE.md

Initial content.

` + markerStart + `
OLD CONTENT TO BE REPLACED
This should not appear in the result.
` + markerEnd + `

Final content.
`,
			wantContains: []string{
				"# CLAUDE.md",
				"Initial content",
				"Final content",
				markerStart,
				markerEnd,
				"oview MCP RAG-First Policy",
			},
			wantNotContain: []string{
				"OLD CONTENT TO BE REPLACED",
				"This should not appear in the result",
			},
			runTwice: true,
		},
		{
			name:           "work with empty file",
			initialContent: "",
			wantContains: []string{
				markerStart,
				markerEnd,
				"oview MCP RAG-First Policy",
			},
			wantNotContain: []string{},
			runTwice:       false,
		},
		{
			name: "preserve content before and after markers",
			initialContent: `# Project

Before markers.

` + markerStart + `
Old section
` + markerEnd + `

After markers.

## More content

Final paragraph.
`,
			wantContains: []string{
				"# Project",
				"Before markers",
				"After markers",
				"## More content",
				"Final paragraph",
				markerStart,
				markerEnd,
				"oview MCP RAG-First Policy",
			},
			wantNotContain: []string{
				"Old section",
			},
			runTwice: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()
			claudeMdPath := filepath.Join(tmpDir, "CLAUDE.md")

			// Write initial content
			if err := os.WriteFile(claudeMdPath, []byte(tt.initialContent), 0644); err != nil {
				t.Fatalf("Failed to write initial CLAUDE.md: %v", err)
			}

			// Run upsert
			if err := UpsertOviewRagFirstSection(tmpDir); err != nil {
				t.Fatalf("UpsertOviewRagFirstSection() error = %v", err)
			}

			// Read result
			result, err := os.ReadFile(claudeMdPath)
			if err != nil {
				t.Fatalf("Failed to read result CLAUDE.md: %v", err)
			}
			resultStr := string(result)

			// Check wanted content
			for _, want := range tt.wantContains {
				if !strings.Contains(resultStr, want) {
					t.Errorf("Result does not contain expected string: %q", want)
				}
			}

			// Check unwanted content
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(resultStr, notWant) {
					t.Errorf("Result contains unwanted string: %q", notWant)
				}
			}

			// Test idempotence if requested
			if tt.runTwice {
				firstResult := resultStr

				// Run upsert again
				if err := UpsertOviewRagFirstSection(tmpDir); err != nil {
					t.Fatalf("Second UpsertOviewRagFirstSection() error = %v", err)
				}

				// Read second result
				secondResult, err := os.ReadFile(claudeMdPath)
				if err != nil {
					t.Fatalf("Failed to read second result CLAUDE.md: %v", err)
				}

				// Results should be identical (idempotence)
				if firstResult != string(secondResult) {
					t.Errorf("Upsert is not idempotent:\nFirst:\n%s\n\nSecond:\n%s", firstResult, string(secondResult))
				}
			}
		})
	}
}

func TestGenerateRagFirstSection(t *testing.T) {
	section := generateRagFirstSection()

	// Must include markers
	if !strings.Contains(section, markerStart) {
		t.Error("Section does not contain start marker")
	}
	if !strings.Contains(section, markerEnd) {
		t.Error("Section does not contain end marker")
	}

	// Must include key instructions
	requiredPhrases := []string{
		"oview MCP RAG-First Policy",
		"CRITICAL INSTRUCTION",
		"authentication flow",
		"security.yaml firewall",
		"messenger rabbitmq",
		"redis cache",
		"elasticsearch mapping",
		"MCP Configuration",
		"~/.claude/mcp_servers.json",
		".oview/claude_mcp.json",
	}

	for _, phrase := range requiredPhrases {
		if !strings.Contains(section, phrase) {
			t.Errorf("Section does not contain required phrase: %q", phrase)
		}
	}

	// Must NOT include OS-specific paths
	forbiddenPhrases := []string{
		"/home/",
		"/Users/",
		"C:\\",
	}

	for _, phrase := range forbiddenPhrases {
		if strings.Contains(section, phrase) {
			t.Errorf("Section contains forbidden OS-specific path: %q", phrase)
		}
	}
}

func TestMarkerExtraction(t *testing.T) {
	content := `Before
` + markerStart + `
Middle
` + markerEnd + `
After`

	startIdx := strings.Index(content, markerStart)
	endIdx := strings.Index(content, markerEnd)

	if startIdx == -1 {
		t.Error("Failed to find start marker")
	}
	if endIdx == -1 {
		t.Error("Failed to find end marker")
	}
	if endIdx <= startIdx {
		t.Error("End marker should come after start marker")
	}

	// Extract content before and after
	before := content[:startIdx]
	after := content[endIdx+len(markerEnd):]

	if !strings.Contains(before, "Before") {
		t.Error("Before section not extracted correctly")
	}
	if !strings.Contains(after, "After") {
		t.Error("After section not extracted correctly")
	}
}
