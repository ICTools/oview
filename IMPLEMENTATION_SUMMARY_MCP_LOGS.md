# MCP Logging Feature - Implementation Summary

## Overview

Added comprehensive logging capabilities to the oview MCP server for monitoring and debugging Claude Code integration.

## What Was Added

### 1. Logger Package (`internal/logger/logger.go`)

A structured logging system that:
- Writes logs to both stderr and file (`~/.oview/mcp.log`)
- Supports 4 log levels: debug, info, warn, error
- Uses JSON format for easy parsing
- Thread-safe with mutex protection
- Includes timestamps and structured context

**API:**
```go
logger.Init(logPath)                          // Initialize logger
logger.Info("message", context...)            // Log info
logger.Debug("message", context...)           // Log debug
logger.Warn("message", context...)            // Log warning
logger.Error("message", context...)           // Log error
logger.Close()                                // Close log file
```

### 2. Enhanced MCP Server (`internal/mcp/server.go`)

Added logging throughout the MCP server lifecycle:
- **Startup**: Project config loading, server ready status
- **Request handling**: Incoming requests with method and ID
- **Tool calls**: Tool name, arguments, execution time, results summary
- **Performance**: Response times for all operations
- **Errors**: Detailed error context with hints

**Log points:**
1. Server initialization (config loading)
2. Server ready state
3. Incoming MCP requests (debug level)
4. Tool call start (with arguments)
5. Tool call completion (with duration and summary)
6. Request completion (with total duration)
7. All errors with context

### 3. New Command: `oview mcp logs` (`cmd/mcp.go`)

Real-time log monitoring command:
```bash
oview mcp logs
```

**Features:**
- Checks if log file exists
- Displays log file path
- Streams logs in real-time using `tail -f`
- User-friendly messages if server hasn't started yet
- Ctrl+C to stop

**Refactored `oview mcp` command:**
- Main command: `oview mcp` (starts server)
- Subcommands:
  - `oview mcp start` (explicit start)
  - `oview mcp logs` (view logs)

### 4. Documentation (`docs/MCP_LOGGING.md`)

Comprehensive guide covering:
- Log format and structure
- How to view logs in real-time
- Common log scenarios (startup, tool calls, errors)
- Debugging tips
- Log filtering with jq
- Performance monitoring
- Troubleshooting workflow
- Advanced log analysis examples

### 5. Updated CLAUDE.md

Added logging section with:
- How to view logs
- Log format example
- Reference to detailed documentation

## Log Examples

### Server Startup
```json
{"timestamp":"2026-02-04T10:30:00Z","level":"info","message":"Starting oview MCP server","context":{"project_path":"/path/to/project","log_file":"/home/user/.oview/mcp.log"}}
{"timestamp":"2026-02-04T10:30:01Z","level":"info","message":"MCP server ready to accept requests","context":{"project":"my-project","version":"0.2.0"}}
```

### Search Tool Call
```json
{"timestamp":"2026-02-04T10:30:45Z","level":"info","message":"Calling tool","context":{"tool":"search","arguments":{"query":"authentication logic","limit":5}}}
{"timestamp":"2026-02-04T10:30:46Z","level":"info","message":"Tool call succeeded","context":{"tool":"search","duration":"523ms","summary":"5 results"}}
```

### Error Handling
```json
{"timestamp":"2026-02-04T10:32:00Z","level":"error","message":"Failed to load project config","context":{"error":"failed to load project config: open .oview/project.yaml: no such file or directory","hint":"Run 'oview init' first"}}
```

## Usage Examples

### Basic Monitoring
```bash
# Start log viewer in one terminal
oview mcp logs

# Use Claude Code in another terminal
# Logs will appear in real-time
```

### Filter for Errors Only
```bash
tail -f ~/.oview/mcp.log | jq 'select(.level=="error")'
```

### Extract Search Queries
```bash
cat ~/.oview/mcp.log | jq -r 'select(.context.tool=="search") | .context.arguments.query'
```

### Performance Analysis
```bash
cat ~/.oview/mcp.log | jq -r 'select(.message=="Tool call succeeded") | .context.duration' | sed 's/ms//' | awk '{sum+=$1; count++} END {print "Average:", sum/count, "ms"}'
```

## Benefits

1. **Debugging**: Easily identify MCP integration issues
2. **Monitoring**: Track tool usage and performance in real-time
3. **Analytics**: Understand which tools and queries are most used
4. **Troubleshooting**: Quick diagnosis of configuration or connectivity issues
5. **Performance**: Identify slow operations and bottlenecks

## Files Changed

1. **Created:**
   - `internal/logger/logger.go` (new package)
   - `docs/MCP_LOGGING.md` (documentation)
   - `IMPLEMENTATION_SUMMARY_MCP_LOGS.md` (this file)

2. **Modified:**
   - `cmd/mcp.go` (added logs command, restructured)
   - `internal/mcp/server.go` (added logging throughout)
   - `CLAUDE.md` (added logging section)

## Testing

1. **Build:**
   ```bash
   go build -o oview .
   ```

2. **Check commands:**
   ```bash
   ./oview mcp --help
   ./oview mcp logs --help
   ```

3. **Test logging:**
   ```bash
   # Start MCP server
   cd /path/to/project
   ./oview mcp

   # In another terminal, view logs
   ./oview mcp logs
   ```

## Future Enhancements

Potential improvements:
- Log rotation (automatic cleanup)
- Configurable log levels via config file
- Log filtering flags (`--level=error`, `--tool=search`)
- Export to Prometheus metrics
- Web UI dashboard
- Real-time alerts on errors

## Migration Notes

- **No breaking changes**: Existing MCP functionality unchanged
- **Backward compatible**: Old `oview mcp` still works
- **New dependencies**: None (uses stdlib only)
- **Log location**: `~/.oview/mcp.log` (created automatically)

## Commands Summary

```bash
# Start MCP server (logs to ~/.oview/mcp.log)
oview mcp
oview mcp start

# View logs in real-time
oview mcp logs

# Check log file directly
tail -f ~/.oview/mcp.log
cat ~/.oview/mcp.log | jq .
```

---

**Version:** 0.2.0+logs
**Date:** 2026-02-04
**Status:** ✅ Complete and tested
