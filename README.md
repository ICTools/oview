# oview - Local Software Factory Environment

`oview` is a CLI tool that bootstraps a local Software Factory environment for multiple projects with shared infrastructure (Postgres+pgvector) and per-project RAG indexing.

## Features

- **Shared Infrastructure**: One Postgres (with pgvector) instance for all projects
- **Per-Project Databases**: Isolated database for each project
- **Stack Detection**: Automatically detects Symfony, Docker, Makefile, frontend frameworks
- **Claude Agent Templates**: Generates role-specific AI agent instruction files
- **RAG Indexing**: Indexes your codebase into pgvector for semantic search

## Prerequisites

- Go 1.23+
- Docker and Docker Compose
- Git (optional, for commit tracking)

## Installation

### Quick Install (Recommended)

```bash
# From the repository
cd oview
./install.sh
```

The script will:
- ✅ Check prerequisites (Docker, Go)
- ✅ Build or download the binary
- ✅ Install to `/usr/local/bin/`
- ✅ Optionally set up infrastructure
- ✅ Show next steps

### Manual Installation

```bash
# Clone the repository
git clone <repository-url>
cd oview

# Build
go build -o oview .

# Install globally
sudo cp oview /usr/local/bin/oview

# Set up infrastructure
oview install
```

For detailed installation instructions, see [INSTALL.md](INSTALL.md).

## Quick Start

### 1. Install Global Infrastructure

Run once to set up shared Postgres:

```bash
oview install
```

**Output:**
```
🚀 Installing oview global infrastructure...

📡 Creating Docker network 'oview-net'...
   ✓ Network ready
🐘 Creating Postgres container 'oview-postgres'...
   ✓ Postgres running on port 5432
💾 Saving configuration...
   ✓ Configuration saved to /home/user/.oview/config.yaml

✅ Installation complete!

Connection details:
  Postgres: localhost:5432
  User:     oview
  Password: oview_password_change_me
```

### 2. Initialize a Project

Navigate to your project directory and initialize oview:

```bash
cd /path/to/your/project
oview init
```

**Output:**
```
🔍 Initializing oview for this project...

📁 Creating .oview directory structure...
   ✓ Directory structure created
🔎 Detecting project stack...
   ✓ Stack detected:
     - Symfony
     - Docker
     - Makefile
     - Frontend: [Webpack Encore, Stimulus]
     - Languages: [PHP, JavaScript]
     - Infrastructure: [Redis, RabbitMQ]
📝 Creating project configuration...
   ✓ Project config saved (slug: my-project)
📋 Creating RAG configuration...
   ✓ RAG config saved
📊 Creating index manifests...
   ✓ Index manifests created
🤖 Generating Claude agent instruction files...
   ✓ Agent files generated

✅ Initialization complete!

Created:
  .oview/project.yaml     - Project configuration
  .oview/rag.yaml         - RAG indexing rules
  .oview/agents/          - Claude agent instructions
  .oview/index/           - Index metadata (empty)
```

### 3. Start Project Runtime

Create the project database and enable pgvector:

```bash
oview up
```

**Output:**
```
🚀 Starting project runtime...

📋 Loading project configuration...
   ✓ Project: abc123 (slug: my-project)
🔍 Checking global infrastructure...
   ✓ Global infrastructure is running
🔗 Connecting to Postgres...
   ✓ Connected
💾 Creating database 'oview_my-project'...
   ✓ Database created
👤 Creating database user 'oview_my-project'...
   ✓ User created
🔐 Granting access permissions...
   ✓ Permissions granted
🧮 Enabling pgvector extension...
   ✓ pgvector enabled
🏗️  Creating RAG schema...
   ✓ Schema created

✅ Project runtime is ready!

Database connection:
  DSN: postgres://oview_my-project:xxx@localhost:5432/oview_my-project
```

### 4. Index Your Codebase

Index your project for RAG:

```bash
oview index
```

**Output:**
```
📚 Indexing project codebase...

📋 Loading project configuration...
   ✓ Project: my-project
   ✓ RAG config loaded
🔗 Connecting to project database...
   ✓ Connected
🔍 Starting indexing process...

Found 142 files to index
[1/142] Indexing src/Controller/HomeController.php...
  ✓ 3 chunks stored
[2/142] Indexing src/Service/UserService.php...
  ✓ 5 chunks stored
...

✅ Indexing complete!

Summary:
  Files indexed:  142
  Chunks stored:  487
  Total size:     2.4 MB
  Duration:       3.2s
  Git commit:     abc123def

Indexed data is now available for RAG queries!
```

## Project Structure

After initialization, your project will have:

```
your-project/
├── .oview/
│   ├── project.yaml          # Project metadata and stack info
│   ├── rag.yaml              # Chunking and indexing rules
│   ├── agents/               # Claude agent instruction files
│   │   ├── pm.md
│   │   ├── po.md
│   │   ├── techlead.md
│   │   ├── dev_backend.md
│   │   ├── dev_frontend.md
│   │   ├── dba.md
│   │   ├── devops.md
│   │   └── qa.md
│   └── index/
│       ├── manifest.json     # Indexed files manifest
│       └── stats.json        # Indexing statistics
└── [your existing code]
```

## Commands

### `oview install`

Installs global infrastructure (run once):
- Creates `oview-net` Docker network
- Starts `oview-postgres` container (Postgres 16 + pgvector)
- Saves configuration to `~/.oview/config.yaml`
- Handles port conflicts automatically

**Options:**
- None

**Example:**
```bash
oview install
```

### `oview init`

Initializes oview for the current project:
- Detects project stack (Symfony, Docker, Makefile, frontend, infrastructure)
- Creates `.oview/` directory structure
- Generates project configuration
- Creates RAG indexing rules
- Generates Claude agent instruction files

**Options:**
- `--force`: Overwrite existing `.oview` configuration

**Example:**
```bash
oview init
oview init --force  # Overwrite existing config
```

### `oview up`

Sets up project runtime:
- Verifies global infrastructure is running
- Creates project-specific database
- Creates database user
- Enables pgvector extension
- Creates RAG schema (chunks table)
- Saves database credentials

**Options:**
- None

**Example:**
```bash
oview up
```

### `oview index`

Indexes project codebase:
- Scans files based on `.oview/rag.yaml` rules
- Chunks files by type (PHP by class/function, YAML by section, etc.)
- Generates embeddings (stub implementation for MVP)
- Stores chunks in project database with metadata
- Updates manifest and statistics

**Options:**
- None

**Example:**
```bash
oview index
```

### `oview version`

Shows the oview version:

```bash
oview version
```

### `oview uninstall`

Uninstalls oview global infrastructure:
- Stops and removes Docker containers
- Removes Docker volumes (unless `--keep-data`)
- Removes Docker network
- Removes global configuration (unless `--keep-config`)

**Options:**
- `-f, --force`: Skip confirmation prompt
- `--keep-data`: Keep Docker volumes (preserve all databases)
- `--keep-config`: Keep ~/.oview/config.yaml

**Examples:**
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

**See also:** [UNINSTALL.md](UNINSTALL.md) for detailed guide.

## Configuration Files

### `.oview/project.yaml`

Project configuration with detected stack:

```yaml
project_id: abc123def
project_slug: my-project
stack:
  symfony: true
  docker: true
  makefile: true
  frontend:
    detected: true
    package_manager: npm
    frameworks:
      - Webpack Encore
      - Stimulus
    build_tools:
      - Webpack
  infrastructure:
    redis: true
    rabbitmq: true
    elasticsearch: false
  languages:
    - PHP
    - JavaScript
  frameworks:
    - Symfony
commands:
  test:
    - bin/phpunit
    - npm test
  lint:
    - vendor/bin/php-cs-fixer fix --dry-run
    - bin/console lint:yaml config
  static_analysis:
    - vendor/bin/phpstan analyse
  build:
    - npm run build
  start:
    - docker-compose up -d
database:
  name: oview_my-project
  user: oview_my-project
  password: "xxx"
```

### `.oview/rag.yaml`

RAG indexing configuration:

```yaml
chunking:
  php:
    strategy: function
    max_size: 2000
    max_tokens: 500
    overlap: 100
  javascript:
    strategy: function
    max_size: 2000
    max_tokens: 500
    overlap: 100
  twig:
    strategy: file
    max_size: 1500
    max_tokens: 400
    overlap: 50
  yaml:
    strategy: section
    max_size: 1000
    max_tokens: 300
    overlap: 50
  makefile:
    strategy: section
    max_size: 800
    max_tokens: 200
    overlap: 20
  docker:
    strategy: section
    max_size: 1000
    max_tokens: 300
    overlap: 50
  generic:
    strategy: size
    max_size: 1500
    max_tokens: 400
    overlap: 100
indexing:
  include_paths:
    - src/
    - config/
    - templates/
    - assets/
    - tests/
    - Makefile
    - docker-compose.yml
    - compose.yaml
    - README.md
    - docs/
  exclude_paths:
    - vendor/
    - node_modules/
    - var/
    - public/bundles/
    - .git/
  extensions:
    - .php
    - .twig
    - .yaml
    - .yml
    - .js
    - .ts
    - .jsx
    - .tsx
    - .json
    - .md
    - .txt
```

### `~/.oview/config.yaml`

Global configuration (created by `oview install`):

```yaml
postgres_host: localhost
postgres_port: 5432
postgres_user: oview
postgres_password: oview_password_change_me
postgres_container_name: oview-postgres
postgres_volume: oview-postgres-data
docker_network_name: oview-net
```

## Claude Agent Files

Each project gets role-specific Claude agent instruction files in `.oview/agents/`:

- **pm.md**: Project Manager agent (triage, prioritize, break down tasks)
- **po.md**: Product Owner agent (clarify requirements, acceptance criteria)
- **techlead.md**: Tech Lead agent (architecture, design patterns, code review)
- **dev_backend.md**: Backend Developer agent (implement features, write tests)
- **dev_frontend.md**: Frontend Developer agent (UI implementation, responsive design)
- **dba.md**: Database Administrator agent (schema design, migrations, optimization)
- **devops.md**: DevOps agent (infrastructure, Docker, deployment)
- **qa.md**: QA Engineer agent (testing, quality assurance, bug verification)

Each agent file includes:
- Role mission and responsibilities
- Expected inputs (task requirements, RAG context, etc.)
- **Strict JSON output format** for structured responses
- Stack-specific skills and best practices
- Safety rules (no destructive commands, test before commit, etc.)

**Example output format:**
```json
{
  "summary": "Brief summary of what was done",
  "actions": ["List of actions taken"],
  "files_changed": ["paths/to/changed/files"],
  "commands": ["commands that were run"],
  "blocking": false,
  "errors": []
}
```

## Database Schema

The RAG schema includes a `chunks` table optimized for pgvector:

```sql
CREATE TABLE chunks (
    id SERIAL PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,

    -- Source information
    source VARCHAR(50) NOT NULL,  -- 'repo', 'docs', 'external'
    type VARCHAR(50) NOT NULL,    -- 'code', 'doc', 'config', 'test'
    path TEXT NOT NULL,
    language VARCHAR(50),
    symbol VARCHAR(255),          -- function/class name
    component VARCHAR(255),       -- module/component name

    -- Content
    content TEXT NOT NULL,
    content_hash VARCHAR(64) NOT NULL,

    -- Embedding (1536 dimensions for OpenAI ada-002 compatibility)
    embedding vector(1536),

    -- Metadata
    metadata JSONB,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    commit_sha VARCHAR(40),

    CONSTRAINT unique_chunk UNIQUE (project_id, content_hash)
);

-- Indexes for fast queries
CREATE INDEX idx_chunks_project_id ON chunks(project_id);
CREATE INDEX idx_chunks_type ON chunks(type);
CREATE INDEX idx_chunks_path ON chunks(path);
CREATE INDEX idx_chunks_embedding ON chunks USING hnsw (embedding vector_cosine_ops);
```

## Embeddings

**Current Implementation (MVP):**
- Uses stub embeddings based on SHA256 hash
- Generates deterministic, normalized vectors
- **Does NOT capture semantic meaning**
- Suitable for testing infrastructure, not production

**To Implement Real Embeddings:**
1. Implement the `embeddings.Generator` interface in `internal/embeddings/`
2. Options:
   - OpenAI API (text-embedding-ada-002 or text-embedding-3)
   - Local models (sentence-transformers, all-MiniLM-L6-v2)
   - Custom models via Ollama or similar
3. Update the indexer to use your implementation

**Example:**
```go
// internal/embeddings/openai.go
type OpenAIGenerator struct {
    apiKey string
    client *openai.Client
}

func (g *OpenAIGenerator) Embed(text string) ([]float32, error) {
    // Call OpenAI API
    // Return embedding vector
}
```

## Troubleshooting

### Docker containers not starting

```bash
# Check if Docker is running
docker ps

# Check container logs
docker logs oview-postgres

# Restart containers
docker restart oview-postgres
```

### Port conflicts

oview automatically detects port conflicts and uses alternative ports. Check `~/.oview/config.yaml` for actual ports.

### Database connection errors

```bash
# Verify database exists
docker exec oview-postgres psql -U oview -l

# Test connection
docker exec oview-postgres psql -U oview -d oview_myproject -c "SELECT 1"
```

### Indexing fails

```bash
# Check RAG config
cat .oview/rag.yaml

# Verify database permissions
docker exec oview-postgres psql -U oview -d oview_myproject -c "\dt"

# Re-run up to fix permissions
oview up
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    User Machine                         │
│                                                         │
│  ┌────────────┐        ┌───────────────────────────┐    │
│  │   oview    │        │   Docker Infrastructure   │    │
│  │    CLI     │──────▶ │                           │    │
│  └────────────┘        │  ┌─────────────────────┐  │    │
│         │              │  │  oview-postgres     │  │    │
│         │              │  │  (pgvector)         │  │    │
│         │              │  └─────────────────────┘  │    │
│         │              │                           │    │
│         │              │   oview-net (network)     │    │
│         │              └───────────────────────────┘    │
│         │                                               │
│         ▼                                               │
│  ┌──────────────┐                                       │
│  │  Project 1   │                                       │
│  │  .oview/     │ ──▶ DB: oview_project1                │
│  └──────────────┘                                       │
│                                                         │
│  ┌──────────────┐                                       │
│  │  Project 2   │                                       │
│  │  .oview/     │ ──▶ DB: oview_project2                │
│  └──────────────┘                                       │
└─────────────────────────────────────────────────────────┘
```

## 🔌 Claude Code Integration

**oview integrates with Claude Code via MCP (Model Context Protocol)!**

This allows Claude to:
- 🔍 **Search your codebase semantically** - "Where is authentication handled?"
- 📖 **Get context automatically** - Understand your code before suggesting changes
- 🎯 **Work more efficiently** - Access your indexed knowledge base in real-time

### Quick Setup

1. **Index your project:**
   ```bash
   cd /path/to/your/project
   oview init
   oview up
   oview index
   ```

2. **Configure Claude Code:**
   ```bash
   mkdir -p ~/.claude
   cat > ~/.claude/mcp_servers.json << 'JSON'
   {
     "mcpServers": {
       "oview": {
         "command": "oview",
         "args": ["mcp"]
       }
     }
   }
   JSON
   ```

3. **Use it:**
   ```
   > Claude, search for error handling code in this project
   ```

4. **Monitor logs (optional):**
   ```bash
   oview mcp logs
   ```
   View MCP server activity in real-time to debug issues or monitor performance.

📚 **Full guide:** See [docs/MCP_INTEGRATION.md](docs/MCP_INTEGRATION.md)
🚀 **Quick start:** See [docs/QUICK_START_MCP.md](docs/QUICK_START_MCP.md)
🔍 **Logging guide:** See [docs/MCP_LOGGING.md](docs/MCP_LOGGING.md)

## Roadmap

- [x] Real embeddings integration (OpenAI, Ollama)
- [x] MCP integration with Claude Code
- [ ] Incremental indexing (only changed files)
- [ ] Web UI for management
- [ ] Multi-project dashboards
- [ ] Docker Compose export for projects

## Contributing

Contributions welcome! Please open an issue or PR.

## License

MIT License - See LICENSE file for details.

## Version

Current version: **0.2.0**
