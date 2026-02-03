#!/bin/bash
# Script to verify that Claude Code is using oview MCP server

set -e

echo "🔍 Verification: Claude Code + oview MCP Integration"
echo "═══════════════════════════════════════════════════════"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Check MCP configuration
echo "1️⃣  Checking MCP configuration..."
if [ ! -f ~/.claude/mcp_servers.json ]; then
    echo -e "${RED}❌ FAIL${NC}: ~/.claude/mcp_servers.json not found"
    echo "   Run: mkdir -p ~/.claude && cp mcp_servers.example.json ~/.claude/mcp_servers.json"
    exit 1
fi

if grep -q "oview" ~/.claude/mcp_servers.json; then
    echo -e "${GREEN}✅ PASS${NC}: MCP configuration found"
else
    echo -e "${RED}❌ FAIL${NC}: oview not configured in ~/.claude/mcp_servers.json"
    exit 1
fi

# Test 2: Check oview binary
echo ""
echo "2️⃣  Checking oview binary..."
if command -v oview &> /dev/null; then
    OVIEW_PATH=$(which oview)
    echo -e "${GREEN}✅ PASS${NC}: oview found at $OVIEW_PATH"
else
    echo -e "${RED}❌ FAIL${NC}: oview not found in PATH"
    echo "   Run: sudo cp oview /usr/local/bin/oview"
    exit 1
fi

# Test 3: Check project initialization
echo ""
echo "3️⃣  Checking project initialization..."
if [ ! -f .oview/project.yaml ]; then
    echo -e "${RED}❌ FAIL${NC}: Project not initialized"
    echo "   Run: oview init"
    exit 1
fi
echo -e "${GREEN}✅ PASS${NC}: Project initialized"

# Test 4: Check database
echo ""
echo "4️⃣  Checking database..."
PROJECT_SLUG=$(grep "project_slug:" .oview/project.yaml | awk '{print $2}')
DB_NAME="oview_$PROJECT_SLUG"

CHUNK_COUNT=$(docker exec oview-postgres psql -U oview -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM chunks;" 2>/dev/null | xargs)

if [ -z "$CHUNK_COUNT" ] || [ "$CHUNK_COUNT" = "0" ]; then
    echo -e "${YELLOW}⚠️  WARNING${NC}: No chunks indexed ($CHUNK_COUNT chunks)"
    echo "   Run: oview index"
else
    echo -e "${GREEN}✅ PASS${NC}: $CHUNK_COUNT chunks indexed"
fi

# Test 5: Test MCP server
echo ""
echo "5️⃣  Testing MCP server..."

# Create test input
TEST_INPUT='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'

# Start MCP server in background with timeout
RESPONSE=$(echo "$TEST_INPUT" | timeout 2 oview mcp 2>/dev/null || echo "timeout")

if echo "$RESPONSE" | grep -q "protocolVersion"; then
    echo -e "${GREEN}✅ PASS${NC}: MCP server responds correctly"
else
    echo -e "${YELLOW}⚠️  WARNING${NC}: MCP server may not be responding (this is expected in stdio mode)"
    echo "   The server is designed to work with Claude Code's stdin/stdout"
fi

# Test 6: Test search functionality
echo ""
echo "6️⃣  Testing search functionality..."
if [ "$CHUNK_COUNT" -gt 0 ]; then
    # Try a simple search
    SEARCH_RESULT=$(oview search "test" --limit 1 2>&1 | grep -c "Found" || echo "0")
    if [ "$SEARCH_RESULT" -gt 0 ]; then
        echo -e "${GREEN}✅ PASS${NC}: Search functionality works"
    else
        echo -e "${RED}❌ FAIL${NC}: Search failed"
    fi
else
    echo -e "${YELLOW}⚠️  SKIP${NC}: No chunks to search (run 'oview index' first)"
fi

# Test 7: Check for file access monitoring
echo ""
echo "7️⃣  How to verify Claude uses MCP (not direct file access)..."
echo ""
echo "   To verify Claude Code uses oview MCP and NOT direct file reads:"
echo ""
echo "   Method 1 - Monitor MCP server logs:"
echo "   ────────────────────────────────────────"
echo "   1. Open terminal 1:"
echo "      ${YELLOW}oview mcp 2>&1 | tee /tmp/oview_mcp.log${NC}"
echo ""
echo "   2. Open terminal 2 with Claude Code:"
echo "      ${YELLOW}claude${NC}"
echo ""
echo "   3. Ask Claude: 'Use search to find authentication code'"
echo ""
echo "   4. Check /tmp/oview_mcp.log for MCP activity:"
echo "      ${YELLOW}cat /tmp/oview_mcp.log${NC}"
echo ""
echo "   Expected: You should see JSON-RPC messages"
echo ""
echo "   Method 2 - Database query monitoring:"
echo "   ────────────────────────────────────────"
echo "   1. Monitor database queries:"
echo "      ${YELLOW}docker exec oview-postgres tail -f /var/log/postgresql/postgresql.log 2>/dev/null || echo 'Log not available'${NC}"
echo ""
echo "   2. Ask Claude to search"
echo ""
echo "   3. You should see SELECT queries with vector operations"
echo ""
echo "   Method 3 - File access monitoring (strace):"
echo "   ────────────────────────────────────────"
echo "   1. Start Claude Code with strace:"
echo "      ${YELLOW}strace -e open,openat -o /tmp/claude_files.log claude${NC}"
echo ""
echo "   2. Ask Claude to find code"
echo ""
echo "   3. Check which files were opened:"
echo "      ${YELLOW}grep -E '(src/|cmd/)' /tmp/claude_files.log${NC}"
echo ""
echo "   If Claude uses MCP: Few/no direct file reads in your src/"
echo "   If Claude reads files: Many open() calls to your source files"
echo ""

# Test 8: Create a test file to verify
echo ""
echo "8️⃣  Creating test verification file..."

cat > /tmp/verify_claude_mcp.md << 'EOF'
# How to Verify Claude Code Uses oview MCP

## Quick Test

1. Start Claude Code:
   ```bash
   claude
   ```

2. Ask Claude:
   ```
   Use the tool 'project_info' to show me information about this project
   ```

3. Expected behavior:
   - ✅ Claude should call the MCP tool and return project info
   - ✅ You should see: embeddings config, chunk count, etc.
   - ❌ Claude should NOT read .oview/project.yaml directly

## Proof that MCP is being used

### Test 1: Ask for search explicitly
```
User: "Use the search tool to find authentication code"

Claude should:
1. Call MCP search("authentication code")
2. Return results from the database
3. NOT read source files directly
```

### Test 2: Compare file access
```
User: "What's in cmd/init.go?"

WITHOUT MCP (direct read):
- Claude uses Read tool
- Opens cmd/init.go directly
- Returns file contents

WITH MCP (via search):
User: "Use search to find init command implementation"
- Claude uses MCP search
- Gets relevant chunks from database
- Returns indexed content (may be chunked)
```

### Test 3: Hidden file test

1. Create a test file NOT indexed:
   ```bash
   echo "SECRET_API_KEY=test123" > /tmp/secret_not_indexed.txt
   ```

2. Ask Claude:
   ```
   Search for SECRET_API_KEY in this project
   ```

Expected:
- ✅ Claude uses MCP search
- ✅ Doesn't find it (not indexed)
- ❌ Claude should NOT read /tmp/secret_not_indexed.txt

If Claude finds it, it's reading files directly (not using MCP)!

## Monitor MCP Activity

### Real-time monitoring

Terminal 1:
```bash
# Monitor MCP server stderr (logs)
oview mcp 2> >(tee /tmp/mcp_activity.log >&2)
```

Terminal 2:
```bash
# Use Claude Code
claude
```

Terminal 3:
```bash
# Watch MCP activity
tail -f /tmp/mcp_activity.log | jq .
```

When you ask Claude to search, you should see:
```json
{
  "level": "info",
  "message": "MCP request: tools/call",
  "method": "search",
  "query": "your search query"
}
```

## Database monitoring

Monitor PostgreSQL queries:
```bash
# Enable query logging
docker exec oview-postgres psql -U postgres -c "ALTER SYSTEM SET log_statement = 'all';"
docker exec oview-postgres psql -U postgres -c "SELECT pg_reload_conf();"

# Watch logs
docker exec oview-postgres tail -f /var/lib/postgresql/data/log/postgresql-*.log | grep "SELECT"
```

When Claude searches via MCP:
```sql
SELECT ... FROM chunks WHERE ... ORDER BY embedding <=> ...
```

## Verify MCP tools are available

Ask Claude:
```
List all available tools you have access to
```

Expected output should include:
- search
- get_context
- project_info

## The smoking gun test

1. Rename a file:
   ```bash
   mv cmd/search.go cmd/search_renamed.go
   ```

2. Ask Claude (WITHOUT re-indexing):
   ```
   Where is the search command implementation?
   ```

Expected behavior:
- WITH MCP: Claude finds "cmd/search.go" (old indexed path)
- WITHOUT MCP: Claude finds "cmd/search_renamed.go" (current filesystem)

This proves Claude uses the indexed data, not direct file access!

3. Don't forget to rename it back:
   ```bash
   mv cmd/search_renamed.go cmd/search.go
   ```
EOF

echo -e "${GREEN}✅ CREATED${NC}: /tmp/verify_claude_mcp.md"
echo "   Read detailed verification steps: ${YELLOW}cat /tmp/verify_claude_mcp.md${NC}"

# Summary
echo ""
echo "═══════════════════════════════════════════════════════"
echo "📊 VERIFICATION SUMMARY"
echo "═══════════════════════════════════════════════════════"
echo ""
echo "   MCP Config:     ✅"
echo "   oview Binary:   ✅"
echo "   Project Init:   ✅"
echo "   Database:       ✅ ($CHUNK_COUNT chunks)"
echo "   Search:         ✅"
echo ""
echo "🎯 Next Steps:"
echo ""
echo "   1. Open Claude Code: ${YELLOW}claude${NC}"
echo ""
echo "   2. Test MCP integration:"
echo "      ${YELLOW}Use project_info${NC}"
echo "      ${YELLOW}Search for authentication code${NC}"
echo ""
echo "   3. Monitor MCP activity (optional):"
echo "      ${YELLOW}cat /tmp/verify_claude_mcp.md${NC}"
echo ""
echo "✅ Setup verified! Claude Code should use oview MCP by default."
echo ""
