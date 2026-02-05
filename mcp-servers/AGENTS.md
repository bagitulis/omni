# MCP Servers - Development Context

> Auto-injected when working in `mcp-servers/` directory.
> **Parent rules:** See root `AGENTS.md` for critical rules.

---

## CRITICAL: AI Behavior

| Rule                 | Description                                                              |
| -------------------- | ------------------------------------------------------------------------ |
| **NO new scripts**   | DO NOT create new Go/Python scripts for operations that already have MCP |
| **Use existing MCP** | DIRECTLY call MCP tool via JSON-RPC to the appropriate binary            |
| **Read sub-README**  | READ the README in each MCP folder before operation                      |

---

## Stack

- Go 1.21+
- MCP Protocol (JSON-RPC 2.0)
- Google Sheets API
- Python (for ads analysis)

---

## Structure

```
mcp-servers/
├── bin/                    # Compiled binaries
├── cmd/                    # MCP server entry points
│   ├── github/             # GitHub accounts management
│   ├── antigravity/        # Antigravity accounts management
│   ├── shopee-ads/         # Shopee ads analyzer
│   └── tiktok-ads/         # TikTok ads analyzer
├── pkg/                    # Shared packages
│   ├── mcp/                # MCP protocol implementation
│   ├── sheets/             # Google Sheets integration
│   ├── totp/               # TOTP/2FA generator
│   └── python/             # Python runner for ads
└── mcp-config.json         # VS Code MCP configuration
```

---

## Available MCP Servers

| Server          | Binary                | Trigger      | Purpose                     |
| --------------- | --------------------- | ------------ | --------------------------- |
| GitHub Accounts | `mcp-github.exe`      | `MCP GH`     | Manage GitHub accounts      |
| Antigravity     | `mcp-antigravity.exe` | `MCP AG`     | Manage Antigravity accounts |
| Shopee Ads      | `mcp-shopee-ads.exe`  | `MCP Shopee` | Analyze Shopee ads          |
| TikTok Ads      | `mcp-tiktok-ads.exe`  | `MCP TikTok` | Analyze TikTok ads          |

---

## MCP Call Pattern

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"TOOL_NAME","arguments":{...}}}
'@ | mcp-servers/bin/MCP-BINARY.exe 2>$null
```

---

## Build Commands

```powershell
cd mcp-servers
go mod download
go build -o bin/mcp-github.exe ./cmd/github
go build -o bin/mcp-antigravity.exe ./cmd/antigravity
go build -o bin/mcp-shopee-ads.exe ./cmd/shopee-ads
go build -o bin/mcp-tiktok-ads.exe ./cmd/tiktok-ads
```

---

## Anti-Patterns

| Don't                                        | Do Instead                    |
| -------------------------------------------- | ----------------------------- |
| Create new Python/Go script for existing MCP | Call MCP tool directly        |
| Hardcode credentials                         | Use Google Sheets storage     |
| Skip reading sub-README                      | Always check cmd/\*/README.md |
