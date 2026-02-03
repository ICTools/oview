#!/bin/bash
# Real-time MCP activity monitor
# Shows visual feedback when Claude Code uses oview

echo "🚀 Starting MCP Monitor..."
echo ""
echo "This will show real-time activity when Claude uses oview tools."
echo ""
echo "In another terminal, run: claude"
echo "Then ask Claude to search or get context."
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Start MCP server with monitoring
oview mcp 2>&1 | oview monitor
