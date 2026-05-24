---
description: Explore agent rules for OMNI project - fast codebase contextual grep
---

# Explore Rules

<!-- MASTER:skill-explore-role -->
> **Role:** Fast codebase exploration and contextual grep.
> **Constraints:** Cannot write, edit, or delegate.
<!-- /MASTER:skill-explore-role -->

---

<!-- MASTER:codebase-structure -->
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
├── frontend/                # React frontend
│   ├── src/
│   │   ├── api/             # Axios API clients
│   │   ├── components/      # Reusable components
│   │   │   ├── layout/      # AppLayout, Sidebar, Header
│   │   │   ├── ui/          # Extended Ant Design
│   │   │   ├── tables/      # Data tables
│   │   │   ├── forms/       # Form components
│   │   │   └── modals/      # Modal components
│   │   ├── pages/           # Route pages
│   │   ├── hooks/           # Custom hooks (useOrders, etc.)
│   │   ├── stores/          # Zustand stores
│   │   ├── types/           # TypeScript (snake_case!)
│   │   ├── lib/             # Utilities
│   │   └── styles/          # Theme (Ant Design overrides)
│   └── public/
├── mcp-servers/             # MCP server implementations
└── .sisyphus/plans/         # Migration and feature plans
```
<!-- /MASTER:codebase-structure -->

---

## Search Patterns

<!-- MASTER:skill-explore-search-patterns -->
Before searching the codebase, read ARCHITECTURE.md in the project root.
Use the search patterns documented there — they reflect the actual folder structure of this project.
Do not use generic patterns (internal/handlers/*.go, etc.) without first confirming they exist in ARCHITECTURE.md.
<!-- /MASTER:skill-explore-search-patterns -->

---

## Output Format

<!-- MASTER:skill-explore-output-format -->
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
<!-- /MASTER:skill-explore-output-format -->

---

## Tips

<!-- MASTER:skill-explore-tips -->
- Use grep for text patterns
- Use glob for file patterns
- Use AST for code structure
- Combine tools for complex searches
<!-- /MASTER:skill-explore-tips -->

---

## Anti-Patterns

<!-- MASTER:skill-explore-anti-patterns -->
| Don't                  | Do Instead                |
| ---------------------- | ------------------------- |
| Return raw grep output | Summarize findings        |
| Miss related files     | Check related patterns    |
| Ignore context         | Explain what matches mean |
<!-- /MASTER:skill-explore-anti-patterns -->
