# MCP Antigravity Accounts Server

## ⚠️ UNTUK AI/BOT: BACA INI DULU!

**JANGAN** buat script Go/Python baru untuk operasi Antigravity accounts.  
**LANGSUNG** panggil MCP tool via JSON-RPC ke `mcp-servers/bin/mcp-antigravity.exe`.

### Cara Panggil MCP Tool (PowerShell)

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"TOOL_NAME","arguments":{...}}}
'@ | mcp-servers/bin/mcp-antigravity.exe 2>$null
```

---

## Build & Install

```powershell
cd mcp-servers
go mod download
go build -o bin/mcp-antigravity.exe ./cmd/antigravity
```

---

## Tools yang Tersedia

### 1. `add_account` - Tambah Akun Baru

**Trigger**: `mcp ag add` atau `AG add`

**Format Input User**:

```
mcp ag add
email = xxx@gmail.com
username = UsernameAG (optional)
pw = PasswordXXX
CODE_RECOVERY (optional, 32 char base32 tanpa dash)
backup-code1
backup-code2
...
```

**JSON-RPC**:

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"add_account","arguments":{"email":"xxx@gmail.com","username":"UsernameAG","password":"PasswordXXX","codeRecovery":"ABCD1234EFGH5678IJKL9012MNOP3456","codes":["0805-0848","0722-7662"]}}}
'@ | mcp-servers/bin/mcp-antigravity.exe
```

**Cara Parsing Input User**:

- Email: baris dengan `email =` atau `email:` atau email format (xxx@xxx.com)
- Username: baris dengan `username =` atau `user =` atau `un =`
- Password: baris dengan `pw =` atau `password =` atau `pass =`
- Code Recovery/2FA Secret: 32 karakter, Base32, **TANPA** dash
- Backup Codes: format `xxxx-xxxx` (9 char dengan dash di tengah)

### 2. `get_accounts_by_status` - Filter Akun by Status

**Trigger**: `mcp ag status [STATUS]`

**Status Valid**: `AVAILABLE`, `CHECK`, `YUMNA`, `KANTOR`, `NO QUOTA`, `SUSPENDED`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_accounts_by_status","arguments":{"status":"AVAILABLE"}}}
'@ | mcp-servers/bin/mcp-antigravity.exe
```

### 3. `get_all_accounts` - Semua Akun

**Trigger**: `mcp ag status all`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_all_accounts","arguments":{}}}
'@ | mcp-servers/bin/mcp-antigravity.exe
```

### 4. `delete_backup_code` - Hapus 1 Backup Code

**Trigger**: `mcp ag [CODE]` atau `AG [CODE]`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"delete_backup_code","arguments":{"code":"0805-0848"}}}
'@ | mcp-servers/bin/mcp-antigravity.exe
```

### 5. `delete_account` - Hapus Akun + Semua Backup Codes

**Trigger**: `mcp ag delete [EMAIL]`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"delete_account","arguments":{"email":"xxx@gmail.com"}}}
'@ | mcp-servers/bin/mcp-antigravity.exe
```

### 6. `get_summary` - Ringkasan

**Trigger**: `mcp ag summary`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_summary","arguments":{}}}
'@ | mcp-servers/bin/mcp-antigravity.exe
```

### 7. `generate_2fa` - Generate TOTP Code

**Trigger**: `ag 2fa [SECRET]`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"generate_2fa","arguments":{"secret":"dbxz2cqmwqv3ntvnhrheq2sbvi6f4yf4"}}}
'@ | mcp-servers/bin/mcp-antigravity.exe
```

---

## Data Storage

- **Spreadsheet ID**: `1ZYonq5Lla0FriY-wfvqW5XEwVyfF6osB1yAtTVABFgA`
- **Sheet Akun**: `AG` (columns: Email, Username, Password, CodeRecovery, Status, UserYumna)
- **Sheet Backup Codes**: `Code_AG` (header row = emails, data rows = backup codes per column)
