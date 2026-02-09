# RAG System Improvements - Advanced Search Guide

## Overview

The oview RAG system has been enhanced with powerful filtering capabilities, query strategies, and improved relevance scoring. These improvements maximize the precision and relevance of search results for different tasks.

## Table of Contents

1. [Advanced Filters](#advanced-filters)
2. [Query Strategies](#query-strategies)
3. [Examples](#examples)
4. [Best Practices](#best-practices)

---

## Advanced Filters

### Available Filters

| Filter | Type | Description | Example |
|--------|------|-------------|---------|
| `language` / `languages` | string / array | Filter by programming language | `"PHP"`, `["PHP", "JavaScript"]` |
| `type` / `types` | string / array | Filter by chunk type | `"code"`, `["code", "doc"]` |
| `path_pattern` | string | Filter by path pattern (wildcards supported) | `"src/Service/*"` |
| `component` / `components` | string / array | Filter by component/module | `"Auth"`, `["Auth", "User"]` |
| `symbol_pattern` | string | Filter by symbol name (wildcards supported) | `"Auth*"`, `"*Controller"` |
| `min_similarity` | number (0-1) | Minimum similarity threshold | `0.4` |
| `limit` | integer (1-20) | Maximum results to return | `10` |

### Chunk Types

- **`code`**: Source code files (PHP, JS, TS, etc.)
- **`doc`**: Documentation files (Markdown, comments)
- **`config`**: Configuration files (YAML, JSON, .env)
- **`test`**: Test files (PHPUnit, Jest, etc.)

---

## Query Strategies

Strategies optimize search behavior for different tasks by adjusting limits, similarity thresholds, and result diversity.

### Available Strategies

#### 1. **default** (Standard Search)
```json
{
  "query": "authentication logic",
  "strategy": "default"
}
```

- **Use case**: General-purpose semantic search
- **Limit**: 5 results
- **Similarity threshold**: None
- **Best for**: Quick searches, finding specific functionality

#### 2. **analysis** (Broad Context)
```json
{
  "query": "user management system",
  "strategy": "analysis"
}
```

- **Use case**: Understanding architecture and relationships
- **Limit**: 10 results
- **Similarity threshold**: 0.3
- **Features**: Includes related chunks, diverse results
- **Best for**: Code analysis, architecture understanding, impact analysis

#### 3. **debug** (Focused Search)
```json
{
  "query": "null pointer exception in checkout",
  "strategy": "debug"
}
```

- **Use case**: Finding bugs and error sources
- **Limit**: 5 results
- **Similarity threshold**: 0.4
- **Filters**: Prioritizes code, excludes documentation
- **Best for**: Bug hunting, error investigation, debugging

#### 4. **exploration** (Discovery Mode)
```json
{
  "query": "payment processing",
  "strategy": "exploration"
}
```

- **Use case**: Discovering new parts of the codebase
- **Limit**: 15 results
- **Similarity threshold**: 0.2
- **Features**: Maximum diversity, all chunk types included
- **Best for**: Codebase discovery, learning unfamiliar code, onboarding

#### 5. **documentation** (Docs-Focused)
```json
{
  "query": "how to configure webhooks",
  "strategy": "documentation"
}
```

- **Use case**: Learning from documentation
- **Limit**: 8 results
- **Similarity threshold**: 0.3
- **Filters**: Prioritizes documentation chunks
- **Best for**: Learning, understanding concepts, finding guides

---

## Examples

### Example 1: Find all authentication code in PHP

```json
{
  "query": "user authentication login",
  "language": "PHP",
  "type": "code",
  "path_pattern": "src/*",
  "limit": 10
}
```

**Result**: Returns up to 10 PHP code chunks from `src/` directory related to authentication.

### Example 2: Debug a specific error with focused search

```json
{
  "query": "InvalidArgumentException in PaymentService",
  "strategy": "debug",
  "symbol_pattern": "*Payment*",
  "min_similarity": 0.5
}
```

**Result**: Returns code chunks with high similarity (>50%) containing payment-related symbols.

### Example 3: Explore API endpoints across the codebase

```json
{
  "query": "REST API endpoints",
  "strategy": "exploration",
  "types": ["code", "doc", "config"],
  "languages": ["PHP", "JavaScript"]
}
```

**Result**: Returns diverse results (up to 15) from both code and documentation about API endpoints.

### Example 4: Analyze security implementation

```json
{
  "query": "security authorization access control",
  "strategy": "analysis",
  "components": ["Security", "Auth"],
  "limit": 15
}
```

**Result**: Broad context (15 results) from Security and Auth components with related chunks included.

### Example 5: Find configuration for a specific service

```json
{
  "query": "redis cache configuration",
  "type": "config",
  "path_pattern": "config/*"
}
```

**Result**: Configuration files in `config/` directory related to Redis cache.

---

## Best Practices

### 1. **Choose the Right Strategy**

| Task | Recommended Strategy | Why |
|------|---------------------|-----|
| "Where is X implemented?" | `default` or `debug` | Focused search for specific functionality |
| "How does Y work?" | `analysis` | Broad context to understand architecture |
| "What payment-related code exists?" | `exploration` | Discover all related code |
| "How do I configure Z?" | `documentation` | Find guides and docs |
| "Fix bug in X" | `debug` | Focused on code, high similarity threshold |

### 2. **Use Filters to Narrow Results**

**Too many results?** Add filters:
```json
{
  "query": "database query",
  "language": "PHP",              // Narrows to PHP only
  "path_pattern": "src/Repository/*",  // Only repositories
  "min_similarity": 0.4           // Higher quality threshold
}
```

**Too few results?** Relax filters:
```json
{
  "query": "authentication",
  "strategy": "exploration",      // Lower similarity threshold
  "limit": 15                     // More results
}
```

### 3. **Combine Filters Effectively**

**Finding test failures:**
```json
{
  "query": "test failure assertion error",
  "type": "test",
  "language": "PHP",
  "strategy": "debug"
}
```

**Understanding a feature:**
```json
{
  "query": "shopping cart checkout flow",
  "types": ["code", "doc"],
  "strategy": "analysis",
  "components": ["Cart", "Checkout", "Payment"]
}
```

### 4. **Use Path Patterns for Specific Modules**

```json
{
  "path_pattern": "src/Service/*"          // All services
  "path_pattern": "src/Controller/Api/*"   // API controllers
  "path_pattern": "*/Test/*"               // All test directories
  "path_pattern": "config/packages/*"      // Package configs
}
```

### 5. **Symbol Patterns for Specific Functions/Classes**

```json
{
  "symbol_pattern": "*Controller"     // All controllers
  "symbol_pattern": "Auth*"           // Classes/functions starting with Auth
  "symbol_pattern": "*Service"        // All services
  "symbol_pattern": "test*"           // Test functions
}
```

---

## Result Fields

Each search result includes:

```json
{
  "path": "src/Service/AuthService.php",
  "type": "code",
  "language": "PHP",
  "symbol": "authenticate",
  "component": "Auth",
  "content": "public function authenticate($credentials) { ... }",
  "similarity": "87.50%",
  "score": "95.25%"
}
```

- **similarity**: Raw cosine similarity to query
- **score**: Re-ranked score after strategy adjustments
- **component**: Module/directory extracted from path

---

## Strategy Comparison

| Aspect | default | analysis | debug | exploration | documentation |
|--------|---------|----------|-------|-------------|---------------|
| Results | 5 | 10 | 5 | 15 | 8 |
| Similarity threshold | None | 0.3 | 0.4 | 0.2 | 0.3 |
| Diversity | Low | Medium | Low | High | Medium |
| Include related | No | Yes | No | No | No |
| Type preference | Balanced | Code+Doc | Code only | All types | Doc+Code |

---

## Advanced Usage in Claude Code

When using these features through Claude Code MCP:

```
> Search for authentication logic in PHP services only
  Uses: language="PHP", path_pattern="src/Service/*", query="authentication"

> Debug the payment error with high precision
  Uses: strategy="debug", query="payment error", min_similarity=0.5

> Explore all API-related code
  Uses: strategy="exploration", query="API endpoints"

> Find documentation about configuring webhooks
  Uses: strategy="documentation", query="webhook configuration"
```

---

## Performance Tips

1. **Use specific queries**: "user login authentication" > "auth"
2. **Add language filters** for multi-language projects
3. **Use path patterns** to limit scope to relevant directories
4. **Adjust similarity threshold** based on query specificity
5. **Choose appropriate strategy** for the task at hand

---

## Migration from Old API

Old syntax (still supported):
```json
{
  "query": "authentication",
  "limit": 5
}
```

New syntax with filters:
```json
{
  "query": "authentication",
  "language": "PHP",
  "strategy": "analysis",
  "limit": 10,
  "min_similarity": 0.3
}
```

All old queries work without modification. New filters are optional enhancements.

---

## Troubleshooting

### No results found
- Try `strategy="exploration"` for broader search
- Lower `min_similarity` threshold
- Remove restrictive filters
- Check if files are indexed: `oview index`

### Too many irrelevant results
- Increase `min_similarity` (try 0.4-0.6)
- Use `strategy="debug"` for focused search
- Add `language`, `type`, or `path_pattern` filters
- Use more specific query terms

### Results from wrong file types
- Add `type="code"` or `types=["code", "config"]`
- Use `path_pattern` to limit directories
- Use appropriate strategy (debug for code, documentation for docs)

---

## Next Steps

- See [MCP Integration](MCP_INTEGRATION.md) for usage with Claude Code
- See [Chunking Guide](CHUNKING_GUIDE.md) for how code is indexed
- Run `oview benchmark` to test search performance

**Questions?** Check the [FAQ](FAQ.md) or open an issue.
