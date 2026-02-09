# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**oview** is a Go CLI tool that creates a local Software Factory environment with shared Postgres+pgvector infrastructure for RAG-powered codebase indexing. It enables semantic search across multiple projects and integrates with Claude Code via the Model Context Protocol (MCP).

## Build & Development Commands

```bash
# Build the binary
go build -o oview .

# Run tests
go test ./...

# Run a specific test
go test ./internal/claude -v -run TestRAGPolicy

# Install locally
sudo cp oview /usr/local/bin/oview

# Quick development workflow
go build -o oview . && ./oview <command>
```

## Architecture Overview

### Three-Layer Design

1. **Global Infrastructure Layer** (once per machine)
   - Docker container: `oview-postgres` (Postgres 16 + pgvector)
   - Docker network: `oview-net`
   - Config: `~/.oview/config.yaml`

2. **Per-Project Configuration Layer** (`.oview/` directory)
   - `project.yaml`: Project metadata, stack detection, database credentials, embeddings config
   - `rag.yaml`: File patterns, chunking strategies per language
   - `agents/`: 8 role-specific Claude instruction files (PM, PO, Tech Lead, Backend/Frontend Dev, DBA, DevOps, QA)
   - `index/`: Manifest and statistics from indexing

3. **RAG Indexing Pipeline**
   - File scanning → Language detection → Chunking → Embedding generation → Storage in pgvector
   - Each project gets isolated database: `oview_<project-slug>`

### Database Naming Convention

- Container: `oview-postgres`
- Network: `oview-net`
- Per-project database: `oview_<project-slug>`
- Per-project user: `oview_<project-slug>` (matches database name)

Database names are slugified from project name (lowercase, hyphens, alphanumeric only).

## Core Components

### Embeddings System (`internal/embeddings/`)

Implements pluggable `Generator` interface:
```go
type Generator interface {
    Embed(text string) ([]float32, error)
    Dimension() int
    MaxContextLength() int
    Name() string
}
```

**Available Implementations:**
- `stub.go`: Deterministic SHA256-based (testing only, no semantic meaning)
- `ollama.go`: Local models via HTTP API (nomic-embed-text, mxbai-embed-large, all-minilm)
- `openai.go`: Cloud models via REST API (text-embedding-3-small, text-embedding-ada-002)

**Context Length Adaptation:**
- Automatically truncates text exceeding model limits
- Warns when chunks exceed 80% of context capacity
- Configured per model in embedding provider implementations

### Chunking Strategies (`internal/indexer/chunker.go`)

Language-specific chunking in `languageChunker` map:
- **PHP**: Regex-based class/function extraction
- **JavaScript/TypeScript**: Size-based (2000 chars default)
- **Twig**: Template blocks
- **YAML**: Top-level sections
- **Makefile**: By targets
- **Docker Compose**: By services
- **Markdown**: By headings
- **Generic**: Size-based with configurable overlap

All chunking strategies respect `max_size` and `overlap` from `.oview/rag.yaml`.

### MCP Server (`internal/mcp/`)

JSON-RPC 2.0 over stdio implementing three tools:
- `search`: Semantic similarity search (cosine distance via pgvector)
- `get_context`: Related chunks for specific file/symbol
- `project_info`: Project metadata, stack info, embedding config

**Critical Detail**: MCP server auto-detects project by working directory. Must be run from project root containing `.oview/`.

### Stack Detection (`internal/detector/detector.go`)

Scans for indicator files to detect:
- **Symfony**: symfony.lock, bin/console, config/bundles.php, src/Kernel.php
- **Docker**: docker-compose.yml, Dockerfile
- **Frontend**: package.json, webpack.config.js, vite.config.js
- **Infrastructure**: redis.conf, rabbitmq.conf, elasticsearch.yml
- **Build tools**: Makefile, tsconfig.json

Results stored in `.oview/project.yaml` under `stack:` section.

## Important Non-Obvious Details

### Idempotent Operations
All commands can run multiple times safely:
- `oview install`: Docker operations check for existing resources
- `oview up`: Uses "IF NOT EXISTS" for database/user creation
- `oview index`: Clears existing chunks before reindexing (full reindex by default)

### Chunk Deduplication
- Content hashing via SHA256 prevents duplicate storage
- Unique constraint: `UNIQUE (project_id, content_hash)`
- Same content in different files = single stored chunk

### Git Integration
- Stores current commit SHA with each indexing operation
- Tracks which version of code was indexed
- Useful for understanding index staleness

### Vector Dimensions
- Defaults to 1536 (OpenAI text-embedding-ada-002 compatible)
- Configurable via `.oview/project.yaml` `embeddings.dim`
- Must match embedding model's output dimension
- Changing dimension requires reindexing

### Port Conflict Handling
`internal/docker/ports.go` automatically detects busy ports and finds alternatives. Actual port stored in `~/.oview/config.yaml`.

### Agent JSON Output Format
All agent instruction files specify strict JSON output for structured responses:
```json
{
  "summary": "Brief summary of actions",
  "actions": ["List of actions taken"],
  "files_changed": ["paths/to/files"],
  "commands": ["commands executed"],
  "blocking": false,
  "errors": []
}
```

### Claude Code MCP Integration

**Setup Requirements:**
1. Project must be indexed: `oview init && oview up && oview index`
2. Add to `~/.claude/mcp_servers.json`:
```json
{
  "mcpServers": {
    "oview": {
      "command": "oview",
      "args": ["mcp"]
    }
  }
}
```
3. Run Claude Code from project root (where `.oview/` exists)

**How MCP Tools Work:**
- `search`: Takes query string, returns ranked chunks with similarity scores
- `get_context`: Takes file path + optional symbol, returns related chunks
- `project_info`: Returns stack detection results, embeddings config, database status

## Configuration Files

### `.oview/project.yaml`
- `project_id`: UUID-like identifier
- `project_slug`: Used for database naming
- `stack`: Detection results (symfony, docker, frontend, infrastructure)
- `embeddings`: Provider (ollama/openai/stub), model name, dimension, API key
- `database`: Connection details (auto-generated by `oview up`)
- `llm`: LLM provider config (claude-code, ollama, openai)

### `.oview/rag.yaml`
- `chunking`: Per-language settings (strategy, max_size, overlap, max_tokens)
- `indexing`: File patterns (include_paths, exclude_paths, extensions)

Language strategies:
- `function`: Extract functions/classes (PHP, JavaScript)
- `section`: Split by sections (YAML, Makefile, Docker Compose)
- `file`: Whole file (small Twig templates)
- `size`: Fixed-size chunks with overlap (generic fallback)

### `~/.oview/config.yaml`
Global infrastructure credentials:
- Postgres host/port/user/password
- Container and volume names
- Docker network name

## Package Structure

```
cmd/                    # Cobra commands (install, init, up, index, search, etc.)
internal/
  ├─ config/           # Config file loading/saving (ProjectConfig, GlobalConfig, RAGConfig)
  ├─ docker/           # Docker client wrapper, port detection
  ├─ detector/         # Stack detection logic
  ├─ database/         # Postgres client, schema creation, chunk storage
  ├─ indexer/          # File scanning, chunking, indexing orchestration
  ├─ embeddings/       # Embedding generators (stub, ollama, openai)
  ├─ agents/           # Agent instruction template generation
  ├─ claude/           # Claude Desktop config generation, RAG policy
  └─ mcp/              # MCP server implementation (JSON-RPC, tool handlers)
```

## Common Workflows

### Initial Setup (New Machine)
```bash
oview install              # Global infrastructure
cd /path/to/project
oview init                 # Detect stack, create config
oview up                   # Create database
oview index                # Index codebase
```

### Re-indexing After Code Changes
```bash
oview index                # Full reindex (clears existing chunks)
```

### Debugging Indexing Issues
```bash
# Check what files will be indexed
cat .oview/rag.yaml

# Verify database connection
docker exec oview-postgres psql -U oview -d oview_<project-slug> -c "\dt"

# Check indexing statistics
cat .oview/index/stats.json

# View indexed files
cat .oview/index/manifest.json
```

### Testing MCP Integration
```bash
# Start MCP server manually (for debugging)
cd /path/to/project
oview mcp

# View MCP server logs in real-time (for monitoring/debugging)
oview mcp logs

# Send test request (in another terminal)
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | oview mcp
```

### MCP Server Logging

The MCP server logs all activities to `~/.oview/mcp.log`:
- Server startup/shutdown events
- Incoming MCP requests with method and ID
- Tool calls with arguments and results
- Performance metrics (response times)
- Errors and warnings

**View logs in real-time:**
```bash
oview mcp logs
```

**Log format (JSON):**
```json
{
  "timestamp": "2026-02-04T10:30:45Z",
  "level": "info",
  "message": "Calling tool",
  "context": {
    "tool": "search",
    "arguments": {"query": "authentication logic", "limit": 5},
    "duration": "523ms",
    "summary": "5 results"
  }
}
```

For detailed logging documentation, see [docs/MCP_LOGGING.md](docs/MCP_LOGGING.md).

## Key Code Locations

- Chunking logic: `internal/indexer/chunker.go` (`languageChunker` map)
- Embedding generation: `internal/embeddings/*.go` (interface implementations)
- Database schema: `internal/database/database.go:93` (`SetupSchema` function)
- MCP tool handlers: `internal/mcp/handler.go` (`HandleToolCall` function)
- Stack detection: `internal/detector/detector.go` (`detectStack` function)
- Index orchestration: `internal/indexer/indexer.go` (`IndexProject` function)

## Version Information

- **Current Version**: 0.2.0
- **Go Version**: 1.23+
- **Key Dependencies**: Cobra (CLI), Viper (config), lib/pq (Postgres), go-openai (OpenAI SDK)

**Recent Breaking Changes (v0.2.0):**
- Removed N8N integration
- Removed financial cost calculations from compare command
- Added dynamic context length adaptation per embedding model
- Improved chunking error handling for large chunks

## Common Pitfalls

1. **Forgetting to run `oview up`**: Indexing requires database to exist first
2. **Wrong working directory**: MCP server must run from project root with `.oview/`
3. **Mismatched vector dimensions**: Changing embedding model requires updating `embeddings.dim` in `project.yaml` and reindexing
4. **Ollama not running**: If using Ollama embeddings, ensure `ollama serve` is running
5. **Port conflicts**: If Postgres port 5432 is busy, check `~/.oview/config.yaml` for actual port used
6. **Stale index**: After major code changes, run `oview index` to refresh embeddings

## Future Enhancement Areas

- Incremental indexing (only changed files)
- Web UI for index management
- Multi-language support (Python, Java, Rust chunking)
- File watching for auto-reindexing
- Windows support (currently Linux/macOS only)

<!-- OVIEW_MCP_RAG_FIRST_START -->
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

```json
{
  "mcpServers": {
    "oview": {
      "command": "oview",
      "args": ["mcp"],
      "cwd": "."
    }
  }
}
```

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
<!-- OVIEW_MCP_RAG_FIRST_END -->