# Claude Code Integration - Manual Test Plan

This document outlines manual testing scenarios for the Claude Code integration feature.

## Prerequisites

- Go 1.21+ installed
- oview built: `go build -o oview .`
- Test project directory (can be any directory, even empty)

## Test Scenarios

### Scenario 1: Project WITHOUT CLAUDE.md + Claude CLI Installed

**Setup:**
1. Ensure Claude CLI is installed: `which claude` should return a path
2. Create a fresh test directory: `mkdir /tmp/test-oview-1 && cd /tmp/test-oview-1`
3. Initialize git: `git init`

**Test Steps:**
1. Run: `oview init`
2. When prompted "Enable Claude Code integration? [Y/n]", press Enter (accept default)
3. Observe output

**Expected Results:**
- ✓ Message: "CLAUDE.md created via claude /init"
- ✓ Message: "oview RAG-first policy added to CLAUDE.md"
- ✓ Message: "MCP configuration snippet created"
- ✓ File exists: `CLAUDE.md` (created by Claude CLI)
- ✓ File exists: `.oview/claude_mcp.json`
- ✓ CLAUDE.md contains markers: `<!-- OVIEW_MCP_RAG_FIRST_START -->` and `<!-- OVIEW_MCP_RAG_FIRST_END -->`
- ✓ CLAUDE.md contains: "oview MCP RAG-First Policy"
- ✓ Summary includes Claude Code integration section

**Verification Commands:**
```bash
# Check CLAUDE.md exists
test -f CLAUDE.md && echo "✓ CLAUDE.md exists" || echo "✗ CLAUDE.md missing"

# Check for markers
grep -q "OVIEW_MCP_RAG_FIRST_START" CLAUDE.md && echo "✓ Markers present" || echo "✗ Markers missing"

# Check MCP snippet
test -f .oview/claude_mcp.json && echo "✓ MCP snippet exists" || echo "✗ MCP snippet missing"
cat .oview/claude_mcp.json | jq .

# Verify JSON is valid
jq empty .oview/claude_mcp.json && echo "✓ Valid JSON" || echo "✗ Invalid JSON"
```

---

### Scenario 2: Project WITHOUT CLAUDE.md + Claude CLI Missing

**Setup:**
1. Temporarily rename/hide Claude CLI: `sudo mv /usr/local/bin/claude /usr/local/bin/claude.bak` (or equivalent)
2. Verify: `which claude` should return nothing
3. Create fresh test directory: `mkdir /tmp/test-oview-2 && cd /tmp/test-oview-2`
4. Initialize git: `git init`

**Test Steps:**
1. Run: `oview init`
2. When prompted "Enable Claude Code integration? [Y/n]", press Enter
3. Observe output

**Expected Results:**
- ✓ Message: "CLAUDE.md created (fallback template)"
- ℹ️ Message: "Claude CLI not found - using fallback template"
- ✓ Message: "oview RAG-first policy added to CLAUDE.md"
- ✓ Message: "MCP configuration snippet created"
- ✓ File exists: `CLAUDE.md` (minimal template)
- ✓ File exists: `.oview/claude_mcp.json`
- ✓ CLAUDE.md contains basic structure (# CLAUDE.md, ## Project Overview, etc.)
- ✓ CLAUDE.md contains oview RAG-first section with markers

**Verification Commands:**
```bash
# Check CLAUDE.md structure
grep -q "# CLAUDE.md" CLAUDE.md && echo "✓ Header present"
grep -q "## Project Overview" CLAUDE.md && echo "✓ Fallback template used"
grep -q "oview MCP RAG-First Policy" CLAUDE.md && echo "✓ RAG policy added"
```

**Cleanup:**
```bash
# Restore Claude CLI
sudo mv /usr/local/bin/claude.bak /usr/local/bin/claude
```

---

### Scenario 3: Project WITH Existing CLAUDE.md

**Setup:**
1. Create test directory: `mkdir /tmp/test-oview-3 && cd /tmp/test-oview-3`
2. Initialize git: `git init`
3. Create existing CLAUDE.md:
```bash
cat > CLAUDE.md << 'EOF'
# My Project

This is my existing CLAUDE.md file.

## Important Instructions

Do not change this content!

## Architecture

Existing architecture docs.
EOF
```

**Test Steps:**
1. Run: `oview init`
2. Accept Claude Code integration
3. Observe output

**Expected Results:**
- ✓ Message: "CLAUDE.md already exists (updated with RAG policy)"
- ✓ CLAUDE.md preserves ALL original content
- ✓ oview RAG-first section APPENDED at end (not replacing anything)
- ✓ Markers added: `<!-- OVIEW_MCP_RAG_FIRST_START -->` and `<!-- OVIEW_MCP_RAG_FIRST_END -->`

**Verification Commands:**
```bash
# Check original content preserved
grep -q "My Project" CLAUDE.md && echo "✓ Original header preserved"
grep -q "Do not change this content!" CLAUDE.md && echo "✓ Original content preserved"
grep -q "Existing architecture docs" CLAUDE.md && echo "✓ All content preserved"

# Check new section added
grep -q "oview MCP RAG-First Policy" CLAUDE.md && echo "✓ RAG policy added"
tail -20 CLAUDE.md  # Should show RAG policy at end
```

---

### Scenario 4: Re-running `oview init` (Idempotence Test)

**Setup:**
1. Use test directory from Scenario 3: `cd /tmp/test-oview-3`
2. CLAUDE.md should already have oview markers

**Test Steps:**
1. Run: `oview init --force` (to allow re-init)
2. Accept Claude Code integration again
3. Compare before/after CLAUDE.md

**Expected Results:**
- ✓ Message: "CLAUDE.md already exists (updated with RAG policy)"
- ✓ NO duplication of oview section
- ✓ Content between markers is replaced (identical result)
- ✓ File size should be roughly the same

**Verification Commands:**
```bash
# Copy CLAUDE.md before re-init
cp CLAUDE.md CLAUDE.md.before

# Run init again
oview init --force

# Compare - should be nearly identical
diff CLAUDE.md.before CLAUDE.md
# Expect minimal or no differences (maybe timestamps)

# Count marker occurrences - should be exactly 2 (start + end)
grep -c "OVIEW_MCP_RAG_FIRST" CLAUDE.md
# Expected: 2
```

---

### Scenario 5: Declining Claude Code Integration

**Setup:**
1. Create test directory: `mkdir /tmp/test-oview-5 && cd /tmp/test-oview-5`
2. Initialize git: `git init`

**Test Steps:**
1. Run: `oview init`
2. When prompted "Enable Claude Code integration? [Y/n]", type `n` and press Enter
3. Observe output

**Expected Results:**
- ✓ NO CLAUDE.md created (unless it existed before)
- ✓ NO .oview/claude_mcp.json created
- ✓ Summary does NOT include Claude Code integration section
- ✓ All other oview functionality works normally (.oview/project.yaml, etc.)

**Verification Commands:**
```bash
test ! -f CLAUDE.md && echo "✓ CLAUDE.md not created" || echo "✗ CLAUDE.md was created"
test ! -f .oview/claude_mcp.json && echo "✓ MCP snippet not created" || echo "✗ MCP snippet was created"
test -f .oview/project.yaml && echo "✓ Normal init completed" || echo "✗ Init failed"
```

---

### Scenario 6: Non-Interactive Mode

**Setup:**
1. Create test directory: `mkdir /tmp/test-oview-6 && cd /tmp/test-oview-6`
2. Initialize git: `git init`

**Test Steps:**
1. Run: `oview init --non-interactive`
2. Observe output

**Expected Results:**
- ✓ Claude Code integration ENABLED by default
- ✓ CLAUDE.md created (via CLI or fallback)
- ✓ .oview/claude_mcp.json created
- ✓ No interactive prompts

**Verification Commands:**
```bash
test -f CLAUDE.md && echo "✓ CLAUDE.md created"
test -f .oview/claude_mcp.json && echo "✓ MCP snippet created"
grep -q "oview MCP RAG-First Policy" CLAUDE.md && echo "✓ RAG policy present"
```

---

## Content Validation Tests

After any successful scenario, validate the generated content:

### Validate CLAUDE.md RAG Section

```bash
# Check for required phrases
grep -q "CRITICAL INSTRUCTION" CLAUDE.md && echo "✓ Has critical instruction"
grep -q "authentication flow" CLAUDE.md && echo "✓ Has example query 1"
grep -q "security.yaml firewall" CLAUDE.md && echo "✓ Has example query 2"
grep -q "messenger rabbitmq" CLAUDE.md && echo "✓ Has example query 3"
grep -q "redis cache" CLAUDE.md && echo "✓ Has example query 4"
grep -q "elasticsearch mapping" CLAUDE.md && echo "✓ Has example query 5"

# Check for MCP configuration reference
grep -q "~/.claude/mcp_servers.json" CLAUDE.md && echo "✓ Has MCP config reference"
grep -q ".oview/claude_mcp.json" CLAUDE.md && echo "✓ References snippet file"

# Ensure NO OS-specific paths
! grep -q "/home/" CLAUDE.md && echo "✓ No /home/ paths"
! grep -q "/Users/" CLAUDE.md && echo "✓ No /Users/ paths"
! grep -q "C:\\\\" CLAUDE.md && echo "✓ No Windows paths"
```

### Validate MCP Snippet JSON

```bash
# Check structure
jq -e '.mcpServers.oview.command == "oview"' .oview/claude_mcp.json && echo "✓ Correct command"
jq -e '.mcpServers.oview.args[0] == "mcp"' .oview/claude_mcp.json && echo "✓ Correct args"
jq -e '.mcpServers.oview.cwd == "."' .oview/claude_mcp.json && echo "✓ Relative cwd"

# Ensure it's valid JSON
jq empty .oview/claude_mcp.json && echo "✓ Valid JSON" || echo "✗ Invalid JSON"
```

---

## Cross-Platform Tests

### Linux
- All scenarios should work
- Claude CLI path: typically `/usr/local/bin/claude`

### macOS
- All scenarios should work
- Claude CLI path: typically `/usr/local/bin/claude` or via Homebrew

### Windows (if applicable)
- Verify `which` alternative works (using `where` command)
- Verify file path handling with backslashes
- Note: Claude CLI may not be available on Windows yet

---

## Success Criteria Summary

✅ All 6 scenarios pass without errors
✅ No content duplication in CLAUDE.md
✅ Original CLAUDE.md content always preserved
✅ Marker-based updates are idempotent
✅ Valid JSON in .oview/claude_mcp.json
✅ No OS-specific paths in generated content
✅ Feature works in both interactive and non-interactive modes
✅ Users can decline integration without breaking init
✅ All existing oview functionality still works

---

## Automated Test Coverage

Current test coverage in `internal/claude/rag_policy_test.go`:
- ✅ Marker insertion on empty file
- ✅ Marker replacement (idempotence)
- ✅ Appending when no markers exist
- ✅ Preserving existing content before/after markers
- ✅ Content validation (required phrases, no OS paths)

Additional tests that could be added:
- Integration test for full `oview init` flow
- Mock test for Claude CLI invocation
- Error handling tests (permissions, disk full, etc.)
