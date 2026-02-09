# Using Advanced RAG with Claude Code

This guide shows how to leverage the enhanced RAG system when working with Claude Code.

## Quick Reference

When asking Claude Code to search your codebase, it can now use powerful filters and strategies automatically.

### Natural Language → Smart Search

Claude Code translates your natural language requests into optimized searches:

| You say | Claude uses |
|---------|-------------|
| "Find authentication code in PHP" | `language="PHP"`, `type="code"`, `query="authentication"` |
| "Debug the payment error" | `strategy="debug"`, `query="payment error"` |
| "Show me all API endpoints" | `strategy="exploration"`, `query="API endpoints"` |
| "How do I configure webhooks?" | `strategy="documentation"`, `query="webhook configuration"` |

---

## Common Scenarios

### 1. Understanding New Codebase

**You**: *"I'm new to this project. Give me an overview of the authentication system."*

**Claude searches with**:
```json
{
  "query": "authentication system login security",
  "strategy": "analysis",
  "types": ["code", "doc"],
  "limit": 15
}
```

**Result**: Broad context from code and documentation about auth system.

---

### 2. Fixing a Bug

**You**: *"Fix the bug where users can't login after password reset."*

**Claude searches with**:
```json
{
  "query": "password reset login authentication",
  "strategy": "debug",
  "type": "code",
  "min_similarity": 0.5
}
```

**Result**: Highly relevant code chunks about password reset and login.

---

### 3. Implementing a Feature

**You**: *"Add a new API endpoint for user registration."*

**Claude searches with**:
```json
{
  "query": "API endpoint controller route",
  "strategy": "analysis",
  "path_pattern": "src/Controller/*",
  "limit": 10
}
```

**Result**: Examples of existing API endpoints to follow the pattern.

---

### 4. Writing Tests

**You**: *"Write tests for the PaymentService class."*

**Claude searches with**:
```json
{
  "query": "PaymentService payment",
  "type": "code",
  "symbol_pattern": "PaymentService"
}
```

Then:
```json
{
  "query": "service test PHPUnit",
  "type": "test",
  "path_pattern": "tests/*"
}
```

**Result**: The service code + examples of how services are tested.

---

### 5. Code Review

**You**: *"Review the security of the authentication module."*

**Claude searches with**:
```json
{
  "query": "authentication security validation",
  "component": "Auth",
  "strategy": "analysis",
  "limit": 20
}
```

**Result**: Comprehensive view of auth module for security review.

---

## Advanced Usage Patterns

### Pattern 1: Multi-Stage Search

**Complex task**: Refactor the payment system

**Stage 1 - Discovery**:
```
> Show me all payment-related code
```
Uses: `strategy="exploration"`, `query="payment"`

**Stage 2 - Deep Dive**:
```
> Focus on the PaymentService class
```
Uses: `symbol_pattern="PaymentService"`, `type="code"`

**Stage 3 - Tests**:
```
> Show me payment tests
```
Uses: `type="test"`, `query="payment"`

### Pattern 2: Progressive Filtering

**Start broad**:
```
> Find all database queries
```

**Too many results? Narrow down**:
```
> Find database queries in the Repository layer only
```
Uses: `path_pattern="src/Repository/*"`

**Still too broad? Focus on specific table**:
```
> Find queries for the users table
```
Uses: `path_pattern="src/Repository/*"`, `query="users table"`

### Pattern 3: Cross-Language Search

**Task**: Update both frontend and backend for a feature

```
> Find all authentication code in PHP and JavaScript
```

Uses:
```json
{
  "query": "authentication",
  "languages": ["PHP", "JavaScript"],
  "type": "code"
}
```

---

## Strategy Selection Tips

### When Claude chooses "debug"

Indicators:
- You mention: "bug", "error", "fix", "broken"
- You want: Specific, high-relevance results
- You need: Code only, no docs

**Example**: *"Fix the null pointer error in checkout"*

### When Claude chooses "analysis"

Indicators:
- You mention: "understand", "how does", "architecture"
- You want: Broad context, multiple files
- You need: Code + docs for full picture

**Example**: *"How does the checkout flow work?"*

### When Claude chooses "exploration"

Indicators:
- You mention: "all", "discover", "find everything"
- You want: Diverse results, discovery mode
- You need: Wide coverage, various components

**Example**: *"Show me all API endpoints in the project"*

### When Claude chooses "documentation"

Indicators:
- You mention: "how to", "configure", "setup", "learn"
- You want: Documentation and guides
- You need: Explanations, not just code

**Example**: *"How do I configure the webhook system?"*

---

## Filter Usage Examples

### Language Filters

**Single language**:
```
> Find authentication logic in PHP
```
→ `language="PHP"`

**Multiple languages**:
```
> Find validation in PHP and JavaScript
```
→ `languages=["PHP", "JavaScript"]`

### Type Filters

**Code only**:
```
> Show me the implementation of user registration (code only)
```
→ `type="code"`

**Docs only**:
```
> Find documentation about API usage
```
→ `type="doc"`

**Code + Config**:
```
> Show me the database setup (code and configuration)
```
→ `types=["code", "config"]`

### Path Patterns

**Specific directory**:
```
> Find services in src/Service/
```
→ `path_pattern="src/Service/*"`

**Deep search**:
```
> Find tests in any subdirectory
```
→ `path_pattern="*/Test/*"`

**Root files**:
```
> Show me configuration files in the config directory
```
→ `path_pattern="config/*"`

### Component Filters

**Single component**:
```
> Show me code from the Payment component
```
→ `component="Payment"`

**Multiple components**:
```
> Find security code in Auth and User components
```
→ `components=["Auth", "User"]`

### Symbol Patterns

**Controllers**:
```
> Find all controller classes
```
→ `symbol_pattern="*Controller"`

**Auth-related**:
```
> Find all functions starting with authenticate
```
→ `symbol_pattern="authenticate*"`

**Services**:
```
> Show me service classes
```
→ `symbol_pattern="*Service"`

---

## Combining Filters for Power

### Ultra-Focused Search

```
> Find authentication code in PHP Security directory with high precision
```

Claude uses:
```json
{
  "query": "authentication",
  "language": "PHP",
  "type": "code",
  "path_pattern": "src/Security/*",
  "min_similarity": 0.6,
  "strategy": "debug"
}
```

### Wide Discovery

```
> Show me everything related to payments across the entire codebase
```

Claude uses:
```json
{
  "query": "payment",
  "strategy": "exploration",
  "types": ["code", "doc", "config", "test"],
  "limit": 20
}
```

---

## Pro Tips

### 1. Be Specific About Location

**Vague**: *"Find user code"*
**Better**: *"Find user code in the Service layer"*
**Best**: *"Find user management code in src/Service/User/"*

### 2. Mention File Types

**Vague**: *"Find database stuff"*
**Better**: *"Find database code"*
**Best**: *"Find database code and configuration files"*

### 3. Use Keywords from Strategies

**For broad context**: "understand", "analyze", "overview"
**For precise search**: "debug", "fix", "find specific"
**For learning**: "how to", "documentation", "guide"
**For discovery**: "all", "explore", "discover"

### 4. Combine Locations

```
> Find API endpoints in both controllers and route configs
```

Claude will search:
- `path_pattern="src/Controller/*"`
- `path_pattern="config/routes/*"`

### 5. Progressive Refinement

Start simple, then refine:
1. *"Find authentication code"*
2. *"Too broad - focus on PHP only"*
3. *"Still too much - only in Security directory"*
4. *"Perfect - show me high-similarity results only"*

---

## Common Pitfalls

### ❌ Too Generic

**Problem**: *"Show me code"*
- No filters → thousands of results
- No strategy → default behavior
- No context → low relevance

**Solution**: *"Show me authentication code in PHP services"*

### ❌ Conflicting Filters

**Problem**: *"Find all test code in src/Service"*
- Tests usually in `tests/`, not `src/Service/`
- Filters exclude all results

**Solution**: Check if filters make sense together

### ❌ Wrong Strategy

**Problem**: *"Explore the database error"* (exploration for bug)
- Exploration = low similarity threshold
- Debug would be better for errors

**Solution**: Use "fix", "debug" keywords for bugs

---

## Verifying Search Quality

### Good Signs

✅ Results match your intent
✅ Scores above 60% similarity
✅ Diverse components represented
✅ Right file types included

### Warning Signs

⚠️ All results from one file
⚠️ Very low similarity (<30%)
⚠️ Wrong file types
⚠️ Missing expected results

### How to Fix

1. **Check filters**: Are they too restrictive?
2. **Try different strategy**: Analysis vs Debug vs Exploration
3. **Adjust similarity**: Lower for broad search, raise for precision
4. **Rephrase query**: Use domain terms

---

## Real Example: Full Workflow

**Task**: Add email verification to user registration

### Step 1: Understand Current Flow
```
> How does user registration currently work?
```
Strategy: `analysis`
Result: Broad understanding of registration flow

### Step 2: Find Email Functionality
```
> Find existing email sending code
```
Strategy: `exploration`
Result: Discover email service, mailer config

### Step 3: Find Tests
```
> Show me registration tests
```
Filters: `type="test"`, `query="registration"`
Result: Test examples to follow

### Step 4: Implementation
Claude now has:
- Registration flow understanding (analysis)
- Email examples (exploration)
- Test patterns (filtered search)

Can implement with full context!

---

## Debugging Search Issues

### No Results

**Possible causes**:
1. Filters too restrictive
2. Query too specific
3. Content not indexed

**Solutions**:
```
> Try with exploration strategy
> Remove filters and search again
> Check if files are indexed: oview index
```

### Irrelevant Results

**Possible causes**:
1. Query too vague
2. Similarity threshold too low
3. Wrong strategy

**Solutions**:
```
> Add language or type filters
> Use debug strategy for precision
> Increase min_similarity
> Be more specific in query
```

### Too Many Results

**Possible causes**:
1. Generic query
2. Exploration strategy with high limit
3. No filters applied

**Solutions**:
```
> Add path_pattern filter
> Use debug strategy
> Specify language or type
> Increase min_similarity
```

---

## Summary

With the enhanced RAG system, Claude Code can:

✅ Understand your intent and choose the right strategy
✅ Apply smart filters automatically
✅ Combine semantic search with structural filters
✅ Provide highly relevant, task-appropriate results
✅ Adapt search based on your goals (debug vs learn vs explore)

**Just tell Claude what you want in natural language, and it will use the right combination of strategies and filters!**

---

## Additional Resources

- [RAG Improvements Guide](RAG_IMPROVEMENTS.md)
- [Advanced Search Examples](../examples/advanced_search_examples.md)
- [MCP Integration](MCP_INTEGRATION.md)

**Happy searching! 🔍**
