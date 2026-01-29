# MCP GitHub Accounts Server

## ⚠️ UNTUK AI/BOT: BACA INI DULU!

**JANGAN** buat script Go/Python baru untuk operasi GitHub accounts.  
**LANGSUNG** panggil MCP tool via JSON-RPC ke `mcp-servers/bin/mcp-github.exe`.

### Cara Panggil MCP Tool (PowerShell)

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"TOOL_NAME","arguments":{...}}}
'@ | mcp-servers/bin/mcp-github.exe 2>$null
```

---

## Build & Install

```powershell
cd mcp-servers
go build -o bin/mcp-github.exe ./cmd/github
```

Jika belum pernah setup:

```powershell
cd mcp-servers
go mod download
go build -o bin/mcp-github.exe ./cmd/github
```

---

## Tools yang Tersedia

### 1. `add_account` - Tambah Akun Baru

**Trigger**: `mcp gh add` atau `GH add`

**Format Input User**:

```
mcp gh add
email = xxx@gmail.com
username = UsernameGH (optional)
pw = PasswordXXX
CODE_RECOVERY (optional, 16 char base32 tanpa dash)
backup-code1
backup-code2
...
```

**JSON-RPC**:

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"add_account","arguments":{"email":"xxx@gmail.com","username":"UsernameGH","password":"PasswordXXX","codeRecovery":"ABCD1234EFGH5678","codes":["xxxxx-xxxxx","yyyyy-yyyyy"]}}}
'@ | mcp-servers/bin/mcp-github.exe
```

**Cara Parsing Input User**:

- Email: baris dengan `email =` atau `email:` atau email format (xxx@xxx.com)
- Username: baris dengan `username =` atau `user =` atau `un =`
- Password: baris dengan `pw =` atau `password =` atau `pass =`
- Code Recovery/2FA Secret: 16 karakter, Base32, **TANPA** dash (contoh: `BUCD6EOGDBBIOAKK`)
- Backup Codes: format `xxxxx-xxxxx` (11 char dengan dash di tengah)

### 2. `get_accounts_by_status` - Filter Akun by Status

**Trigger**: `mcp gh status [STATUS]`

**Status Valid**: `AVAILABLE`, `CHECK`, `YUMNA`, `KANTOR`, `NO QUOTA`, `SUSPENDED`, `BANNED`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_accounts_by_status","arguments":{"status":"AVAILABLE"}}}
'@ | mcp-servers/bin/mcp-github.exe
```

### 3. `get_all_accounts` - Semua Akun

**Trigger**: `mcp gh status all`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_all_accounts","arguments":{}}}
'@ | mcp-servers/bin/mcp-github.exe
```

### 4. `delete_backup_code` - Hapus 1 Backup Code

**Trigger**: `mcp gh [CODE]` atau `GH [CODE]`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"delete_backup_code","arguments":{"code":"fcd9d-3e6d8"}}}
'@ | mcp-servers/bin/mcp-github.exe
```

### 5. `update_status` - Update Status Akun

**Trigger**: `mcp gh update [EMAIL] [STATUS]`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"update_status","arguments":{"email":"xxx@gmail.com","status":"NO QUOTA"}}}
'@ | mcp-servers/bin/mcp-github.exe
```

### 6. `delete_account` - Hapus Akun + Semua Backup Codes

**Trigger**: `mcp gh delete [EMAIL]`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"delete_account","arguments":{"email":"xxx@gmail.com"}}}
'@ | mcp-servers/bin/mcp-github.exe
```

### 7. `get_summary` - Ringkasan

**Trigger**: `mcp gh summary`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_summary","arguments":{}}}
'@ | mcp-servers/bin/mcp-github.exe
```

### 8. `generate_2fa` - Generate TOTP Code

**Trigger**: `mcp 2fa [SECRET]` atau `2fa [SECRET]`

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"generate_2fa","arguments":{"secret":"WKX2OSMHXAWA2VRR"}}}
'@ | mcp-servers/bin/mcp-github.exe
```

---

## Troubleshooting

### "google config directory not found"

Pastikan jalankan dari root project (`c:\Users\PC\Desktop\Project\omni`) atau ada folder `backend/config/static/google/` dengan credential JSON files.

### Binary tidak ada

```powershell
cd mcp-servers; go build -o bin/mcp-github.exe ./cmd/github
```

---

## Data Storage

- **Spreadsheet ID**: `1ZYonq5Lla0FriY-wfvqW5XEwVyfF6osB1yAtTVABFgA`
- **Sheet Akun**: `GH` (columns: Email, Username, Password, CodeRecovery, Status, UserYumna)
- **Sheet Backup Codes**: `Code_GH` (header row = emails, data rows = backup codes per column)
