---
description: Explore agent rules for OMNI project - fast codebase contextual grep
---

# Explore Rules

> **Role:** Fast codebase exploration and contextual grep.
> **Constraints:** Cannot write, edit, or delegate.

---

## Codebase Structure

```
omni/
├── backend/                 # Go backend
│   ├── cmd/server/          # Entry point
│   ├── internal/            # Core implementation
│   │   ├── handlers/        # HTTP handlers
│   │   ├── services/        # Business logic
│   │   ├── repositories/    # Database access
│   │   ├── models/          # Data models
│   │   └── config/          # Configuration
│   ├── shopee-sdk/          # Shopee SDK
│   ├── lazada-sdk/          # Lazada SDK
│   └── tiktok_sdk/          # TikTok SDK
├── frontend/                # Vue.js frontend
│   ├── src/
│   │   ├── components/      # Vue components
│   │   ├── views/           # Page views
│   │   ├── api/             # API clients
│   │   └── stores/          # Pinia stores
│   └── public/
└── mcp-servers/             # MCP server implementations
```

---

## Search Patterns

### Finding Handlers

```
internal/handlers/*_handler.go
```

### Finding Services

```
internal/services/*_service.go
```

### Finding Models

```
internal/models/*.go
```

### Finding SDK Usage

```
backend/*sdk*/*.go
```

---

## Output Format

```markdown
## Search Results

### Query

[What was searched]

### Matches Found

| File            | Line | Match           |
| --------------- | ---- | --------------- |
| path/to/file.go | 42   | matched content |

### Context

[Brief explanation of what was found]

### Suggested Next Steps

[If applicable]
```

---

## Tips

- Use grep for text patterns
- Use glob for file patterns
- Use AST for code structure
- Combine tools for complex searches

---

## Anti-Patterns

| Don't                  | Do Instead                |
| ---------------------- | ------------------------- |
| Return raw grep output | Summarize findings        |
| Miss related files     | Check related patterns    |
| Ignore context         | Explain what matches mean |
