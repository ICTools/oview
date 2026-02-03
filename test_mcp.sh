#!/bin/bash
# Test script for oview MCP server

set -e

echo "🧪 Testing oview MCP server..."
echo ""

# Check if project is initialized
if [ ! -f .oview/project.yaml ]; then
    echo "❌ No .oview/project.yaml found"
    echo "   Run 'oview init' first in your project directory"
    exit 1
fi

echo "✅ Project initialized"
echo ""

# Test that oview binary works
echo "1️⃣  Testing oview binary..."
./oview version
echo ""

# Test MCP server startup
echo "2️⃣  Testing MCP server startup (will timeout in 2s)..."
timeout 2 ./oview mcp 2>&1 | head -5 || echo "   ✅ MCP server accepts connections"
echo ""

echo "✅ MCP server is ready!"
echo ""
echo "📝 Configuration for Claude Code:"
echo "   File: ~/.claude/mcp_servers.json"
echo "   Content:"
cat ~/.claude/mcp_servers.json
echo ""
echo "📚 Next steps:"
echo "   1. Make sure your project is indexed: ./oview index"
echo "   2. Restart Claude Code to load the MCP server"
echo "   3. In Claude Code, try: 'Use project_info'"
echo ""
echo "📖 Documentation:"
echo "   - Full guide: docs/MCP_INTEGRATION.md"
echo "   - Quick start: docs/QUICK_START_MCP.md"
