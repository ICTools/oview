# MCP Server Logging

This guide explains how to use the MCP server logging feature to monitor and debug Claude Code integration.

## Overview

The MCP server logs all activities to `~/.oview/mcp.log`, including:
- Server startup and configuration
- Incoming MCP requests (initialize, tools/list, tools/call)
- Tool executions with arguments and results
- Performance metrics (response times)
- Errors and warnings

## Log Format

All logs are in JSON format for easy parsing:

```json
{
  "timestamp": "2026-02-04T10:30:45Z",
  "level": "info",
  "message": "Calling tool",
  "context": {
    "tool": "search",
    "arguments": {
      "query": "authentication logic",
      "limit": 5
    }
  }
}
```

## Viewing Logs in Real-Time

### Start Monitoring

```bash
oview mcp logs
```

This will:
1. Check if the log file exists
2. Display the log file path
3. Stream logs in real-time using `tail -f`

### Stop Monitoring

Press `Ctrl+C` to stop streaming logs.

## Log Levels

- **debug**: Detailed information for debugging (MCP protocol details)
- **info**: General informational messages (tool calls, successful operations)
- **warn**: Warning messages (non-critical issues, failed requests)
- **error**: Error messages (critical failures, configuration issues)

## Common Log Scenarios

### Server Startup

```json
{"timestamp":"2026-02-04T10:30:00Z","level":"info","message":"Starting oview MCP server","context":{"project_path":"/path/to/project","log_file":"/home/user/.oview/mcp.log"}}
{"timestamp":"2026-02-04T10:30:01Z","level":"info","message":"MCP server ready to accept requests","context":{"project":"my-project","version":"0.2.0"}}
```

### Tool Call: Search

```json
{"timestamp":"2026-02-04T10:30:45Z","level":"debug","message":"Received MCP request","context":{"method":"tools/call","id":1}}
{"timestamp":"2026-02-04T10:30:45Z","level":"info","message":"Calling tool","context":{"tool":"search","arguments":{"query":"authentication logic","limit":5}}}
{"timestamp":"2026-02-04T10:30:46Z","level":"info","message":"Tool call succeeded","context":{"tool":"search","duration":"523ms","summary":"5 results"}}
{"timestamp":"2026-02-04T10:30:46Z","level":"debug","message":"Request handled successfully","context":{"method":"tools/call","id":1,"duration":"524ms"}}
```

### Tool Call: Get Context

```json
{"timestamp":"2026-02-04T10:31:00Z","level":"info","message":"Calling tool","context":{"tool":"get_context","arguments":{"path":"src/Service/AuthService.php","limit":3}}}
{"timestamp":"2026-02-04T10:31:01Z","level":"info","message":"Tool call succeeded","context":{"tool":"get_context","duration":"201ms","summary":"3 chunks"}}
```

### Error: Project Not Initialized

```json
{"timestamp":"2026-02-04T10:32:00Z","level":"error","message":"Failed to load project config","context":{"error":"failed to load project config: open .oview/project.yaml: no such file or directory","hint":"Run 'oview init' first"}}
{"timestamp":"2026-02-04T10:32:00Z","level":"error","message":"MCP server error","context":{"error":"failed to load project config: open .oview/project.yaml: no such file or directory\nHint: Run 'oview init' first"}}
```

## Debugging Tips

### Check if MCP server is running

If you're using Claude Code and MCP tools aren't working:

1. **Check the log file exists:**
   ```bash
   ls -lh ~/.oview/mcp.log
   ```

2. **View recent logs:**
   ```bash
   tail -20 ~/.oview/mcp.log
   ```

3. **Monitor logs while using Claude Code:**
   ```bash
   # In one terminal:
   oview mcp logs

   # In another terminal, use Claude Code to trigger MCP calls
   ```

### Common Issues

#### No log file found

**Symptom:**
```
Log file not found: /home/user/.oview/mcp.log
The MCP server may not have been started yet.
```

**Solution:**
The MCP server hasn't been started yet. Claude Code will start it automatically when needed, or you can start it manually for testing:
```bash
cd /path/to/your/project
oview mcp
```

#### Permission errors

**Symptom:**
```json
{"level":"error","message":"failed to open log file: permission denied"}
```

**Solution:**
Check permissions on `~/.oview/` directory:
```bash
chmod 755 ~/.oview
```

#### Slow tool responses

**Symptom:**
```json
{"level":"warn","message":"Request failed","context":{"duration":"5.2s"}}
```

**Solution:**
- Check database connectivity
- Verify index exists: `oview status`
- Consider reindexing: `oview index`

## Log Rotation

The log file grows continuously. To manage log size:

### Manual cleanup

```bash
# Backup and truncate
cp ~/.oview/mcp.log ~/.oview/mcp.log.backup
echo > ~/.oview/mcp.log
```

### Automatic rotation (future feature)

Log rotation is planned for a future release. Current workaround:

```bash
# Add to crontab for weekly rotation
0 0 * * 0 mv ~/.oview/mcp.log ~/.oview/mcp.log.$(date +\%Y\%m\%d) && echo > ~/.oview/mcp.log
```

## Filtering Logs

### View only errors

```bash
tail -f ~/.oview/mcp.log | jq 'select(.level=="error")'
```

### View only tool calls

```bash
tail -f ~/.oview/mcp.log | jq 'select(.message=="Calling tool")'
```

### View slow queries (>1s)

```bash
tail -f ~/.oview/mcp.log | jq 'select(.context.duration | tonumber > 1000)'
```

### Extract search queries

```bash
cat ~/.oview/mcp.log | jq -r 'select(.message=="Calling tool" and .context.tool=="search") | .context.arguments.query'
```

## Performance Monitoring

### Average tool response time

```bash
cat ~/.oview/mcp.log | jq -r 'select(.message=="Tool call succeeded") | .context.duration' | sed 's/ms//' | awk '{sum+=$1; count++} END {print "Average:", sum/count, "ms"}'
```

### Tool call frequency

```bash
cat ~/.oview/mcp.log | jq -r 'select(.message=="Calling tool") | .context.tool' | sort | uniq -c | sort -rn
```

## Integration with Claude Code

When Claude Code uses oview MCP tools, you'll see:

1. **Initialize request** (once per session)
   ```json
   {"method":"initialize"}
   ```

2. **Tools list request** (once per session)
   ```json
   {"method":"tools/list"}
   ```

3. **Tool calls** (multiple times)
   ```json
   {"method":"tools/call","tool":"search"}
   ```

### Example Claude Code session

```
User: "Find authentication code in this project"
  ↓
Claude → MCP: search(query="authentication code")
  ↓
MCP → Database: SELECT ... WHERE embedding <=> ...
  ↓
MCP → Claude: [5 results with similarity scores]
  ↓
Claude: "I found authentication code in src/Service/AuthService.php..."
```

All these steps are logged in `~/.oview/mcp.log`.

## Troubleshooting Workflow

1. **Start log monitoring:**
   ```bash
   oview mcp logs
   ```

2. **Reproduce the issue** in Claude Code

3. **Check for errors** in the log stream

4. **Fix the issue** based on error messages

5. **Verify the fix** by checking logs again

## Advanced: Log Analysis

### Find most searched topics

```bash
cat ~/.oview/mcp.log | jq -r 'select(.context.tool=="search") | .context.arguments.query' | sort | uniq -c | sort -rn | head -10
```

### Identify slow operations

```bash
cat ~/.oview/mcp.log | jq -r 'select(.context.duration) | [.context.tool, .context.duration] | @tsv' | sort -k2 -rn | head -10
```

### Count errors by type

```bash
cat ~/.oview/mcp.log | jq -r 'select(.level=="error") | .context.error' | sort | uniq -c | sort -rn
```

## Future Enhancements

Planned logging improvements:

- **Log rotation**: Automatic cleanup of old logs
- **Log levels filter**: Configure verbosity via `~/.oview/config.yaml`
- **Structured metrics**: Export Prometheus metrics
- **Web UI**: Real-time dashboard for log visualization
- **Alerts**: Notify on critical errors

---

For more information, see:
- [MCP Integration Guide](MCP_INTEGRATION.md)
- [Quick Start](QUICK_START_MCP.md)
- [Verification Guide](VERIFICATION_GUIDE.md)
