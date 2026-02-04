# oview - Functional Overview

**Version:** 0.2.0
**Description:** Local RAG-powered codebase indexing and semantic search tool for software projects

## What is oview?

oview is a CLI tool that indexes your codebase into a local vector database (Postgres+pgvector) and provides semantic search capabilities through an MCP (Model Context Protocol) server for Claude Code integration.

**Key principle:** One shared Postgres infrastructure, isolated databases per project.

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  User Machine                                           │
│                                                         │
│  ┌──────────────┐      ┌─────────────────────┐         │
│  │   oview CLI  │─────▶│  Postgres+pgvector  │         │
│  │  (Go binary) │      │  (Docker container) │         │
│  └──────────────┘      └─────────────────────┘         │
│         │                        │                      │
│         │                        │                      │
│  ┌──────▼──────────┐    ┌────────▼────────┐            │
│  │  Project 1      │    │  DB: oview_proj1│            │
│  │  .oview/        │    └─────────────────┘            │
│  └─────────────────┘                                   │
│                         ┌─────────────────┐            │
│  ┌─────────────────┐    │  DB: oview_proj2│            │
│  │  Project 2      │    └─────────────────┘            │
│  │  .oview/        │                                   │
│  └─────────────────┘                                   │
│                                                         │
│  ┌─────────────────────────────┐                       │
│  │  Claude Code + MCP Server   │                       │
│  │  (semantic search queries)  │                       │
│  └─────────────────────────────┘                       │
└─────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Global Infrastructure (Shared)

**Installed once with:** `oview install`

- **Postgres+pgvector container** (`oview-postgres`)
  - Image: `pgvector/pgvector:pg16`
  - Port: `5432` (auto-adjusted if busy)
  - Volume: `oview-postgres-data` (persistent)

- **Docker network** (`oview-net`)
  - Connects all oview containers

- **Global configuration** (`~/.oview/config.yaml`)
  - Postgres credentials
  - Container names
  - Network settings

### 2. Per-Project Setup

**Initialized with:** `oview init` (in project directory)

- **Project database** (`oview_<project-slug>`)
  - Isolated from other projects
  - Contains `chunks` table with embeddings
  - pgvector extension enabled

- **Project configuration** (`.oview/config.yaml`)
  - Project ID and slug
  - Embeddings provider/model
  - Database credentials

- **RAG configuration** (`.oview/rag.yaml`)
  - File patterns to include/exclude
  - Chunking rules per file type
  - Max chunk sizes

---

## Commands Reference

### `oview install`

**Purpose:** Install global infrastructure (run once per machine)

**What it does:**
1. Creates Docker network `oview-net`
2. Pulls and starts Postgres+pgvector container
3. Detects port conflicts and auto-adjusts
4. Saves configuration to `~/.oview/config.yaml`

**Options:**
- None

**Output:**
```bash
🚀 Installing oview global infrastructure...
📡 Creating Docker network 'oview-net'...
   ✓ Network ready
🐘 Creating Postgres container 'oview-postgres'...
   ✓ Postgres running on port 5432
💾 Saving configuration...
   ✓ Configuration saved to ~/.oview/config.yaml

✅ Installation complete!

Connection details:
  Postgres: localhost:5432
  User:     oview
  Password: oview_password_change_me
```

**Prerequisites:**
- Docker installed and running
- Port 5432 available (or will auto-adjust)

---

### `oview init`

**Purpose:** Initialize oview for the current project

**What it does:**
1. Detects project stack (PHP/Symfony, JavaScript, Docker, etc.)
2. Creates `.oview/` directory structure
3. Generates `config.yaml` with project settings
4. Generates `rag.yaml` with chunking rules
5. Creates `.oview/agents/` with Claude agent templates
6. Prompts for embeddings provider configuration

**Options:**
- None (interactive prompts)

**Interactive prompts:**
1. **Embeddings provider:** `ollama` or `openai`
2. **Model selection:**
   - Ollama: `nomic-embed-text`, `mxbai-embed-large`, `all-minilm`
   - OpenAI: `text-embedding-3-small`, `text-embedding-ada-002`
3. **API configuration:**
   - Ollama: Base URL (default: `http://localhost:11434`)
   - OpenAI: API Key

**Files created:**
```
.oview/
├── config.yaml          # Project configuration
├── rag.yaml            # RAG indexing rules
└── agents/             # Claude agent templates
    ├── pm.md
    ├── po.md
    ├── techlead.md
    ├── dev_backend.md
    ├── dev_frontend.md
    ├── dba.md
    ├── devops.md
    └── qa.md
```

**Example config.yaml:**
```yaml
project_id: my-project-abc123
project_slug: my-project
database:
  name: oview_my-project
  user: oview_my-project
  password: oview_dev
embeddings:
  provider: ollama
  model: nomic-embed-text
  base_url: http://localhost:11434
  dim: 768
```

---

### `oview up`

**Purpose:** Start project runtime (database setup)

**What it does:**
1. Verifies global infrastructure is running
2. Creates project-specific database
3. Creates database user for the project
4. Enables pgvector extension
5. Creates RAG schema (`chunks` table)
6. Updates project config with database credentials

**Options:**
- None

**Database schema created:**
```sql
CREATE TABLE chunks (
    id SERIAL PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,
    source VARCHAR(50) NOT NULL,      -- 'repo'
    type VARCHAR(50) NOT NULL,        -- 'code', 'doc', 'config', 'test'
    path TEXT NOT NULL,
    language VARCHAR(50),
    symbol VARCHAR(255),              -- function/class name
    component VARCHAR(255),           -- module name
    content TEXT NOT NULL,
    content_hash VARCHAR(64) NOT NULL,
    embedding vector(768),            -- Dimension from config
    embedding_model VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    commit_sha VARCHAR(40)
);

-- Indexes created:
-- - idx_chunks_project_id
-- - idx_chunks_embedding (HNSW for vector similarity)
-- - idx_chunks_type, path, source, symbol, commit, metadata
```

---

### `oview index`

**Purpose:** Index project codebase into vector database

**What it does:**
1. Scans project files based on `.oview/rag.yaml` rules
2. Chunks files according to language-specific strategies
3. Generates embeddings for each chunk
4. Stores chunks in database with metadata
5. Saves indexing manifest to `.oview/manifest.json`

**Options:**
- `-n, --queries <int>`: Number of test queries (for benchmark mode, default: 10)
- `-o, --output <file>`: Output file for results (for benchmark mode)

**Chunking strategies by file type:**

| File Type | Strategy | Max Size | Description |
|-----------|----------|----------|-------------|
| **PHP** | Class/Function | 2000 chars | Splits by class, then by function |
| **JavaScript/TypeScript** | Size-based | 2000 chars | Line-based chunking with overlap |
| **Twig** | Block-based | 1500 chars | Splits by `{% block %}` tags |
| **YAML** | Section-based | 1000 chars | Splits by top-level keys |
| **Makefile** | Target-based | 800 chars | Splits by targets |
| **Markdown** | Heading-based | 2000 chars | Splits by `#` headings |
| **Generic** | Size-based | 1500 chars | Fallback for unknown types |

**File patterns (default):**
```yaml
include:
  - "**/*.php"
  - "**/*.js"
  - "**/*.ts"
  - "**/*.jsx"
  - "**/*.tsx"
  - "**/*.twig"
  - "**/*.yaml"
  - "**/*.yml"
  - "**/*.md"
  - "**/Makefile"
  - "**/docker-compose.yml"

exclude:
  - "**/vendor/**"
  - "**/node_modules/**"
  - "**/.git/**"
  - "**/var/**"
  - "**/public/build/**"
```

**Adaptive truncation:**
- Automatically adapts to embedding model's context limit
- Warns when chunks exceed 80% of model capacity
- Model-specific limits:
  - `nomic-embed-text`: 8192 tokens (~32K chars)
  - `mxbai-embed-large`: 512 tokens (~2K chars)
  - `all-minilm`: 256 tokens (~1K chars)
  - OpenAI models: 8191 tokens (~32K chars)

**Output example:**
```bash
📚 Indexing project codebase...
Found 797 files to index

[1/797] Indexing src/Controller/BookController.php...
  ✓ 5 chunks stored
[2/797] Indexing src/Entity/Book.php...
  ✓ 3 chunks stored
[494/797] Indexing templates/admin/store.html.twig...
  ⚠️  Large chunk (6500 chars, model max: 2048) may be truncated for Ollama mxbai-embed-large
  ✓ 2 chunks stored

✅ Indexing complete!
  Files indexed: 797
  Chunks stored: 2,543
  Total size: 15.2 MB
  Duration: 3m 24s
```

**Manifest file (`.oview/manifest.json`):**
```json
{
  "files": {
    "src/Controller/BookController.php": {
      "path": "src/Controller/BookController.php",
      "hash": "a1b2c3d4...",
      "chunks": 5,
      "indexed_at": "2026-02-04T10:30:00Z"
    }
  },
  "last_update": "2026-02-04T10:35:00Z"
}
```

---

### `oview search`

**Purpose:** Search indexed codebase semantically

**What it does:**
1. Generates embedding for search query
2. Performs vector similarity search in database
3. Returns top N most similar chunks
4. Displays results with context

**Options:**
- `-n, --limit <int>`: Number of results to return (default: 5)

**Usage:**
```bash
oview search "authentication implementation"
oview search "database connection setup" -n 10
```

**Output example:**
```bash
🔍 Searching for: "authentication implementation"

Results:

1. src/Security/LoginAuthenticator.php (similarity: 0.89)
   Symbol: LoginAuthenticator::authenticate
   ─────────────────────────────────────────
   public function authenticate(Request $request): Passport
   {
       $credentials = $this->getCredentials($request);
       // ... authentication logic ...
   }

2. config/packages/security.yaml (similarity: 0.85)
   Symbol: security
   ─────────────────────────────────────────
   security:
       providers:
           app_user_provider:
               entity:
                   class: App\Entity\User
   ...
```

---

### `oview benchmark`

**Purpose:** Run performance benchmarks on RAG system

**What it does:**
1. Tests database connection latency
2. Measures embedding generation speed
3. Tests search performance (end-to-end)
4. Tests concurrent search performance
5. Saves detailed results to JSON

**Options:**
- `-o, --output <file>`: Output file (default: `benchmark_results.json`)
- `-n, --queries <int>`: Number of test queries (default: 10)

**Test queries:**
- "authentication implementation"
- "database connection"
- "error handling"
- "configuration files"
- "user management"
- etc.

**Output:**
```bash
🏁 Starting oview RAG Benchmark
═══════════════════════════════════════════════════════

📊 Project: my-project
📦 Total chunks: 2543
🤖 Embeddings: ollama / nomic-embed-text (768 dim)

🧪 Running Benchmark Tests...

1️⃣  Testing database connection...
2️⃣  Testing embedding generation...
   Embedding 1/10: authentication implementation
   ...
3️⃣  Testing search performance...
   Search 1/10: authentication implementation
   ...
4️⃣  Testing concurrent searches...

📈 Calculating summary statistics...

═══════════════════════════════════════════════════════
📊 BENCHMARK RESULTS
═══════════════════════════════════════════════════════

✅ Success Rate: 41/41 tests (100.0%)

⚡ Performance:
   Avg Embedding Time:  125ms
   Avg Search Time:     28ms
   Min Search Time:     18ms
   Max Search Time:     45ms
   Throughput:          35.71 queries/sec

🎯 Relevance:
   Avg Top Result:      87.3%

💾 Full results saved to: benchmark_results.json

📈 Performance Rating:
   🚀 EXCELLENT - Blazing fast!
   🎯 HIGH RELEVANCE - Results are very pertinent
```

---

### `oview compare`

**Purpose:** Compare Claude Code performance with/without oview

**What it does:**
1. Simulates scenarios with and without oview
2. Measures time, token usage, and accuracy
3. Calculates improvements
4. Saves comparison results to JSON

**Options:**
- `-o, --output <file>`: Output file (default: `comparison_results.json`)

**Scenarios tested:**
1. Find authentication code
2. Understand file context
3. Explore codebase
4. Debug error

**Output:**
```bash
⚖️  oview Impact Comparison
═══════════════════════════════════════════════════════════════

Test 1: Find authentication code
─────────────────────────────────────────────────────────

  ✅ WITH oview (MCP):
     Time:     25ms
     Tokens:   2000
     Results:  5
     Accuracy: High - semantic search finds relevant code

  ❌ WITHOUT oview (Direct):
     Time:     500ms
     Tokens:   8000
     Results:  15
     Accuracy: Medium - keyword search, some irrelevant results

  💎 IMPROVEMENTS:
     Time:   -475ms (95.0% faster)
     Tokens: -6000 (75.0% fewer)

═══════════════════════════════════════════════════════════════
💎 AVERAGE IMPROVEMENTS PER QUERY
═══════════════════════════════════════════════════════════════

  ⚡ Time saved:    947.5ms (96.3% faster)
  🎯 Tokens saved:  6800 (76.5% reduction)

🔑 KEY INSIGHTS:
   • oview is 27.3x FASTER than direct file access
   • Uses 76.5% FEWER tokens (less context, more focused)
   • Better ACCURACY with semantic search
   • 100% LOCAL with Ollama (no API costs for embeddings)
```

---

### `oview verify`

**Purpose:** Verify embedding quality and consistency

**What it does:**
1. Samples random chunks from database
2. Regenerates embeddings
3. Compares with stored embeddings
4. Checks for consistency and quality issues

**Options:**
- `-n, --samples <int>`: Number of samples to verify (default: 100)

**Output:**
```bash
🔍 Verifying embeddings...
Sampling 100 random chunks...

✅ Verification complete!
  Samples: 100
  Matches: 100 (100.0%)
  Avg similarity: 0.9987

All embeddings are consistent!
```

---

### `oview monitor`

**Purpose:** Real-time monitoring of MCP search queries

**What it does:**
1. Monitors incoming MCP search requests
2. Shows queries, response times, and results
3. Displays live statistics

**Options:**
- None (runs until Ctrl+C)

**Output:**
```bash
📊 oview MCP Monitor - Real-time Query Tracking
═══════════════════════════════════════════════════════

[10:30:45] Query: "authentication flow"
           Results: 5 chunks
           Time: 28ms
           Top similarity: 0.89

[10:31:12] Query: "database migrations"
           Results: 5 chunks
           Time: 23ms
           Top similarity: 0.92

─────────────────────────────────────────────────────
Statistics:
  Total queries: 2
  Avg time: 25.5ms
  Avg results: 5.0
  Avg similarity: 0.905
```

---

### `oview uninstall`

**Purpose:** Remove global infrastructure

**What it does:**
1. Stops and removes Postgres container
2. Optionally removes Docker volumes
3. Removes Docker network
4. Optionally removes global configuration

**Options:**
- `-f, --force`: Skip confirmation prompt
- `--keep-data`: Keep Docker volumes (preserve databases)
- `--keep-config`: Keep `~/.oview/config.yaml`

**Usage:**
```bash
# Complete uninstall (with confirmation)
oview uninstall

# Quick uninstall
oview uninstall --force

# Keep data for later reinstall
oview uninstall --keep-data

# Update oview (preserve everything)
oview uninstall --keep-data --keep-config
# ... update binary ...
oview install
```

**Warning message:**
```bash
⚠️  Are you sure you want to continue? [y/N]:

The following will be removed:
  🐳 Container: oview-postgres
  🌐 Network:   oview-net
  💾 Volume:    oview-postgres-data (⚠️  ALL PROJECT DATABASES)
  📄 Config:    ~/.oview/config.yaml
```

---

### `oview version`

**Purpose:** Show oview version

**Output:**
```bash
oview version 0.2.0
```

---

## MCP Server Integration

### Setup

Add to Claude Code's MCP configuration (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "oview": {
      "command": "oview",
      "args": ["mcp"],
      "cwd": "/path/to/your/project"
    }
  }
}
```

### Available MCP Tools

**1. `search`**
- Search codebase semantically
- Parameters:
  - `query` (required): Search query
  - `limit` (optional): Number of results (default: 5)
  - `filter` (optional): Filter by type, language, component

**2. `get_context`**
- Get relevant context for a file
- Parameters:
  - `path` (required): File path
  - `limit` (optional): Number of related chunks (default: 3)

**3. `list_files`**
- List indexed files
- Parameters:
  - `pattern` (optional): Filter by pattern
  - `type` (optional): Filter by type (code, doc, config, test)

---

## Configuration Files

### Global Config (`~/.oview/config.yaml`)

```yaml
postgres_host: localhost
postgres_port: 5432
postgres_user: oview
postgres_password: oview_password_change_me
postgres_container_name: oview-postgres
postgres_volume: oview-postgres-data
docker_network_name: oview-net
```

### Project Config (`.oview/config.yaml`)

```yaml
project_id: my-project-abc123
project_slug: my-project
database:
  name: oview_my-project
  user: oview_my-project
  password: oview_dev
embeddings:
  provider: ollama              # or: openai
  model: nomic-embed-text       # or: text-embedding-3-small
  base_url: http://localhost:11434
  api_key: ""                   # for OpenAI
  dim: 768                      # vector dimension
```

### RAG Config (`.oview/rag.yaml`)

```yaml
sources:
  repo:
    enabled: true
    include:
      - "**/*.php"
      - "**/*.js"
      - "**/*.ts"
      - "**/*.twig"
      - "**/*.yaml"
    exclude:
      - "**/vendor/**"
      - "**/node_modules/**"
      - "**/.git/**"

chunking:
  php:
    strategy: class        # class, function, or file
    max_size: 2000        # characters
    overlap: 100
  javascript:
    strategy: file
    max_size: 2000
    overlap: 100
  twig:
    strategy: file
    max_size: 1500
    overlap: 50
  # ... other file types
```

---

## Typical Workflow

### First-time Setup

```bash
# 1. Install global infrastructure (once per machine)
oview install

# 2. Navigate to your project
cd /path/to/project

# 3. Initialize oview for this project
oview init
# → Select ollama/openai
# → Choose model
# → Configure API

# 4. Start project runtime
oview up

# 5. Index your codebase
oview index
```

### Daily Usage

```bash
# Update index after code changes
oview index

# Search from CLI
oview search "authentication logic"

# Or use via Claude Code MCP integration
# Claude can now semantically search your code!
```

### New Project

```bash
# Navigate to new project
cd /path/to/new-project

# Initialize (infrastructure already installed)
oview init
oview up
oview index
```

---

## Embedding Providers

### Ollama (Recommended for local/free)

**Pros:**
- 100% local, no API costs
- No internet required
- Fast on modern hardware
- Multiple model choices

**Cons:**
- Requires local Ollama installation
- Uses disk space and RAM
- Initial model download required

**Models:**
- `nomic-embed-text` (768 dim, 8192 tokens) - **Recommended**
- `mxbai-embed-large` (1024 dim, 512 tokens) - High quality, small context
- `all-minilm` (384 dim, 256 tokens) - Very small, limited context

**Setup:**
```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull model
ollama pull nomic-embed-text

# Verify
ollama list
```

### OpenAI (Recommended for cloud)

**Pros:**
- State-of-the-art quality
- Large context windows (8191 tokens)
- No local resources needed

**Cons:**
- Requires API key
- Costs money ($0.02-$0.13 per 1M tokens)
- Requires internet connection

**Models:**
- `text-embedding-3-small` (1536 dim, $0.02/1M tokens) - **Recommended**
- `text-embedding-3-large` (3072 dim, $0.13/1M tokens) - Highest quality
- `text-embedding-ada-002` (1536 dim, $0.10/1M tokens) - Legacy

**Setup:**
```bash
# Get API key from https://platform.openai.com/api-keys
export OPENAI_API_KEY="sk-..."

# oview will prompt for it during init
```

---

## Current Limitations

### What oview DOES NOT do:

1. **No incremental indexing**
   - Full re-index required after changes
   - Future: Track changed files only

2. **No code understanding**
   - Simple chunking strategies (not AST-based)
   - May split functions awkwardly
   - Future: Language-specific parsers

3. **No auto-reindexing**
   - Manual `oview index` after changes
   - Future: File watcher for auto-indexing

4. **No multi-language chunking**
   - Each file chunked independently
   - Future: Cross-file context

5. **No semantic caching**
   - Embeddings regenerated on each index
   - Future: Cache based on content hash

6. **Limited file type support**
   - Focused on web development (PHP, JS, Twig, YAML)
   - Future: More languages (Python, Java, Rust, etc.)

---

## Performance Characteristics

### Indexing Speed

**Factors:**
- Project size (files, lines of code)
- Embedding provider (Ollama vs OpenAI)
- Model choice
- Hardware (CPU, RAM for Ollama)
- Network (for OpenAI)

**Typical performance:**
- Small project (100 files): ~30 seconds
- Medium project (500 files): ~2-3 minutes
- Large project (2000+ files): ~10-15 minutes

### Search Speed

**Typical performance:**
- Database query: ~5-15ms
- Embedding generation: ~50-200ms (Ollama) or ~100-300ms (OpenAI)
- **Total end-to-end: ~100-500ms**

### Disk Usage

**Per project:**
- Indexed chunks: ~2-5 MB per 1000 files
- Embeddings: Depends on dimension (768 dim ≈ 3KB per chunk)
- **Total: ~5-20 MB for medium project**

---

## Troubleshooting

### "Input length exceeds context length"

**Cause:** Chunks too large for embedding model

**Solutions:**
1. Model has small context (mxbai: 512 tokens, all-minilm: 256 tokens)
2. Reduce `max_size` in `.oview/rag.yaml`
3. Switch to model with larger context (nomic-embed-text: 8192 tokens)

**Example fix:**
```yaml
# .oview/rag.yaml
chunking:
  twig:
    max_size: 800  # Reduce from 1500
```

### Slow indexing with Ollama

**Causes:**
- Large model
- CPU-only inference
- Many files

**Solutions:**
1. Use smaller model (`all-minilm` instead of `nomic-embed-text`)
2. Reduce files to index (update `.oview/rag.yaml` exclude patterns)
3. Use GPU-accelerated Ollama
4. Switch to OpenAI (faster, but costs money)

### MCP server not connecting

**Checks:**
1. Correct `cwd` in Claude desktop config
2. `oview` binary in PATH
3. Project initialized (`oview init` completed)
4. Database running (`docker ps | grep oview-postgres`)

---

## Version History

### v0.2.0 (Current)
- ✅ Removed N8N integration (simplified architecture)
- ✅ Dynamic context length adaptation per model
- ✅ Improved chunking with size validation
- ✅ Better error handling for large chunks
- ✅ Removed financial cost calculations from compare

### v0.1.0
- ✅ Initial release
- ✅ Basic indexing and search
- ✅ MCP server integration
- ✅ Ollama and OpenAI support
- ✅ Benchmark and compare tools

---

## Summary

**oview** is a local-first RAG system that:
1. Indexes your codebase into vector embeddings
2. Stores in local Postgres+pgvector
3. Provides semantic search via MCP for Claude Code
4. Supports multiple embedding providers (Ollama, OpenAI)
5. Adapts to model capabilities automatically
6. Optimizes for speed and relevance

**Perfect for:**
- Understanding large codebases
- Semantic code search
- Context-aware AI coding with Claude
- Local, privacy-focused development
