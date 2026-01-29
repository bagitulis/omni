# MCP TikTok Ads Analyzer

## ⚠️ UNTUK AI/BOT: BACA INI DULU!

**JANGAN** buat script Python baru untuk analisis TikTok Ads.  
**LANGSUNG** panggil MCP tool via JSON-RPC ke `mcp-servers/bin/mcp-tiktok-ads.exe`.

### Cara Panggil MCP Tool (PowerShell)

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"TOOL_NAME","arguments":{...}}}
'@ | mcp-servers/bin/mcp-tiktok-ads.exe 2>$null
```

---

## Build & Install

```powershell
cd mcp-servers
go mod download
go build -o bin/mcp-tiktok-ads.exe ./cmd/tiktok-ads
```

**Dependency Python** (diperlukan untuk analisis):

```powershell
pip install pandas numpy scikit-learn
```

---

## Tools yang Tersedia

### 1. `get_menu` - Menu

**Trigger**: `MCP Menu TikTok` atau `MCP TikTok`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_menu","arguments":{}}}
'@ | mcp-servers/bin/mcp-tiktok-ads.exe
```

### 2. `analyze_ads` - Analisis Ringkas

**Trigger**: `MCP Analisis TikTok`

Menampilkan statistik umum: total cost, revenue, ROI, breakdown per bidding mode.

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"analyze_ads","arguments":{}}}
'@ | mcp-servers/bin/mcp-tiktok-ads.exe
```

### 3. `get_top_products` - Produk Terbaik

**Trigger**: `MCP Produk Terbaik TikTok`

Menampilkan produk dengan performa terbaik berdasarkan ML score.

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_top_products","arguments":{"limit":10}}}
'@ | mcp-servers/bin/mcp-tiktok-ads.exe
```

### 4. `get_stop_products` - Produk Harus Stop

**Trigger**: `MCP Produk Stop TikTok`

Menampilkan produk yang harus dihentikan iklannya (ROI negatif).

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_stop_products","arguments":{}}}
'@ | mcp-servers/bin/mcp-tiktok-ads.exe
```

### 5. `generate_insights` - Insight Lengkap

**Trigger**: `MCP Insight TikTok`

Analisis mendalam dengan ML, rekomendasi, dan distribusi produk.

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"generate_insights","arguments":{}}}
'@ | mcp-servers/bin/mcp-tiktok-ads.exe
```

### 6. `generate_report` - Laporan HTML

**Trigger**: `MCP Laporan TikTok`

Generate laporan lengkap dalam format HTML/Markdown.

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"generate_report","arguments":{}}}
'@ | mcp-servers/bin/mcp-tiktok-ads.exe
```

---

## Data Source

Python module `tiktok_ads` di `mcp-servers/pkg/python/tiktok_ads/` yang mengakses data dari database/API TikTok.
