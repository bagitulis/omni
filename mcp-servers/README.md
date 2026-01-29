# MCP Servers

Koleksi MCP (Model Context Protocol) servers untuk integrasi AI dengan berbagai layanan.

## ⚠️ UNTUK AI/BOT: INSTRUKSI WAJIB

1. **JANGAN** buat script Go/Python baru untuk operasi yang sudah ada MCP-nya
2. **LANGSUNG** panggil MCP tool via JSON-RPC ke binary yang sesuai
3. **BACA** README di folder masing-masing MCP sebelum melakukan operasi

---

## Build Semua MCP Servers

```powershell
cd mcp-servers
go mod download
go build -o bin/mcp-github.exe ./cmd/github
go build -o bin/mcp-antigravity.exe ./cmd/antigravity
go build -o bin/mcp-shopee-ads.exe ./cmd/shopee-ads
go build -o bin/mcp-tiktok-ads.exe ./cmd/tiktok-ads
```

Atau build satu per satu sesuai kebutuhan.

---

## Daftar MCP Servers

| Server               | Binary                    | Trigger        | Deskripsi                                              |
| -------------------- | ------------------------- | -------------- | ------------------------------------------------------ |
| GitHub Accounts      | `bin/mcp-github.exe`      | `MCP GH`, `GH` | Kelola akun GitHub (add, delete, status, backup codes) |
| Antigravity Accounts | `bin/mcp-antigravity.exe` | `MCP AG`, `AG` | Kelola akun Antigravity                                |
| Shopee Ads           | `bin/mcp-shopee-ads.exe`  | `MCP Shopee`   | Analisis performa iklan Shopee                         |
| TikTok Ads           | `bin/mcp-tiktok-ads.exe`  | `MCP TikTok`   | Analisis performa iklan TikTok                         |

---

## Cara Panggil MCP Tool (PowerShell)

Format umum:

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"TOOL_NAME","arguments":{...}}}
'@ | mcp-servers/bin/MCP-BINARY.exe 2>$null
```

Contoh - tambah akun GitHub:

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"add_account","arguments":{"email":"user@gmail.com","password":"xxx","codes":["xxxxx-xxxxx"]}}}
'@ | mcp-servers/bin/mcp-github.exe
```

---

## Struktur Folder

```
mcp-servers/
├── bin/                    # Binary hasil build
│   ├── mcp-github.exe
│   ├── mcp-antigravity.exe
│   ├── mcp-shopee-ads.exe
│   └── mcp-tiktok-ads.exe
├── cmd/                    # Source code tiap MCP
│   ├── github/             # GitHub accounts management
│   ├── antigravity/        # Antigravity accounts management
│   ├── shopee-ads/         # Shopee ads analyzer
│   ├── tiktok-ads/         # TikTok ads analyzer
│   ├── add-accounts/       # Utility script (bukan MCP)
│   └── check-accounts/     # Utility script (bukan MCP)
├── pkg/                    # Shared packages
│   ├── mcp/                # MCP protocol implementation
│   ├── sheets/             # Google Sheets integration
│   ├── totp/               # TOTP/2FA generator
│   └── python/             # Python runner for ads analysis
├── go.mod
├── go.sum
└── mcp-config.json         # VS Code MCP configuration
```

---

## Troubleshooting

### "google config directory not found"

Binary harus dijalankan dari root project atau pastikan ada folder `backend/config/static/google/` dengan credential JSON files.

### Binary tidak ada

Build dulu dengan perintah di atas.

### Python error (untuk Shopee/TikTok Ads)

Install dependencies:

```powershell
pip install pandas numpy scikit-learn
```

---

## Dokumentasi per MCP

- [GitHub Accounts](cmd/github/README.md)
- [Antigravity Accounts](cmd/antigravity/README.md)
- [Shopee Ads](cmd/shopee-ads/README.md)
- [TikTok Ads](cmd/tiktok-ads/README.md)
