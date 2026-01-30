# 🚀 OpenCode Setup Guide - Complete Installation Tutorial

> **For AI Agents:** Baca dan jalankan tutorial ini step-by-step untuk setup OpenCode di komputer baru.
>
> **⚠️ PENTING:** Config files sudah tersedia di folder `opencode-configs/` - tinggal copy ke lokasi yang benar!

**Last Updated:** January 30, 2026  
**Tested on:** Windows 11  
**OpenCode Version:** v1.1.44  
**Antigravity Plugin:** v1.4.2 (@latest contains fix for version enforcement)

---

## 📑 Table of Contents

1. [Quick Start (Copy Existing Config)](#quick-start-copy-existing-config)
2. [File Locations (PENTING!)](#file-locations-penting)
3. [Prerequisites](#prerequisites)
4. [Install OpenCode CLI](#install-opencode-cli)
5. [Configuration Files](#configuration-files)
6. [Setup Google Authentication](#setup-google-authentication)
7. [Verify Installation](#verify-installation)
8. [Troubleshooting](#troubleshooting)
9. [Current Configuration Summary](#current-configuration-summary)
10. [Backup & Restore](#backup--restore)
11. [References](#references)

---

## File Locations (PENTING!)

> ⚠️ **CRITICAL:** OpenCode menyimpan config di `~/.config/opencode/` (BUKAN `%APPDATA%`)

### Lokasi Config (Single Source of Truth)

| File                        | Location              | Purpose                                |
| --------------------------- | --------------------- | -------------------------------------- |
| `opencode.json`             | `~/.config/opencode/` | Main config (plugins, models)          |
| `antigravity.json`          | `~/.config/opencode/` | Antigravity plugin settings            |
| `oh-my-opencode.json`       | `~/.config/opencode/` | Agent orchestration settings           |
| `antigravity-accounts.json` | `~/.config/opencode/` | **Google account tokens (SENSITIVE!)** |

### Path Windows

```powershell
# Config directory
$HOME\.config\opencode\

# Full path contoh:
C:\Users\yumna\.config\opencode\opencode.json
C:\Users\yumna\.config\opencode\antigravity-accounts.json
```

### Backup di Project

| File                              | Location            | Purpose                          |
| --------------------------------- | ------------------- | -------------------------------- |
| `opencode.json`                   | `opencode-configs/` | Main config backup               |
| `antigravity.json`                | `opencode-configs/` | Plugin settings backup           |
| `oh-my-opencode.json`             | `opencode-configs/` | Current agent config             |
| `oh-my-opencode-full-claude.json` | `opencode-configs/` | Claude-based agents              |
| `oh-my-opencode-full-gemini.json` | `opencode-configs/` | Gemini-based agents              |
| `antigravity-accounts.json`       | `opencode-configs/` | **Account backup (GITIGNORED!)** |

> 🔒 **SECURITY:** `antigravity-accounts.json` sudah di-gitignore karena berisi refresh tokens!

---

## Quick Start (Copy Existing Config)

**Jika pindah PC atau setup baru**, config files sudah tersedia di project ini:

```powershell
# 1. Create config directory (LOKASI YANG BENAR!)
New-Item -ItemType Directory -Force -Path "$HOME\.config\opencode"

# 2. Copy config files dari project
Copy-Item "opencode-configs\opencode.json" "$HOME\.config\opencode\opencode.json" -Force
Copy-Item "opencode-configs\antigravity.json" "$HOME\.config\opencode\antigravity.json" -Force

# 3. Pilih variant oh-my-opencode (Claude ATAU Gemini):
# Option A: Claude-based agents (recommended for coding)
Copy-Item "opencode-configs\oh-my-opencode-full-claude.json" "$HOME\.config\opencode\oh-my-opencode.json" -Force

# Option B: Gemini-based agents (faster, more quota)
# Copy-Item "opencode-configs\oh-my-opencode-full-gemini.json" "$HOME\.config\opencode\oh-my-opencode.json" -Force

# 4. Copy accounts file (jika ada backup di project)
if (Test-Path "opencode-configs\antigravity-accounts.json") {
    Copy-Item "opencode-configs\antigravity-accounts.json" "$HOME\.config\opencode\antigravity-accounts.json" -Force
    Write-Host "✅ Accounts restored!" -ForegroundColor Green
} else {
    Write-Host "⚠️ No accounts backup - run 'opencode auth login' to authenticate" -ForegroundColor Yellow
}

# 5. Install OpenCode CLI (jika belum)
npm install -g opencode-ai

# 6. Run OpenCode
opencode
```

> 📁 **Config files backup:** `opencode-configs/` folder di root project
>
> ⚠️ **JANGAN** copy ke `%APPDATA%/opencode/` - itu lokasi lama yang sudah tidak dipakai!

---

## Prerequisites

Pastikan sudah terinstall:

- **Node.js** (v18+) - https://nodejs.org/
- **npm** (included with Node.js)

```powershell
node --version   # v18+ required
npm --version
```

---

## Install OpenCode CLI

### Windows (PowerShell)

```powershell
# Install OpenCode globally
npm install -g opencode-ai

# Verify installation
opencode --version
```

### Alternative: Direct Download

Download binary dari https://github.com/anomalyco/opencode/releases

```powershell
# Download dan simpan sebagai opencode.exe di project root
# Lalu jalankan langsung:
.\opencode.exe
```

---

## Configuration Files

### Backup Files di Project (`opencode-configs/`)

| File                              | Size                                         | Description |
| --------------------------------- | -------------------------------------------- | ----------- |
| `opencode.json`                   | Main config dengan semua 9 models            |
| `antigravity.json`                | Plugin settings optimal untuk 5+ accounts    |
| `oh-my-opencode.json`             | Current active agent config                  |
| `oh-my-opencode-full-claude.json` | Claude-based agents (recommended for coding) |
| `oh-my-opencode-full-gemini.json` | Gemini-based agents (faster)                 |
| `antigravity-accounts.json`       | **12 Google accounts (GITIGNORED!)**         |

> 🔒 `antigravity-accounts.json` berisi refresh tokens - **JANGAN commit ke git!**

### Config 1: opencode.json (Main Config)

```json
{
  "$schema": "https://opencode.ai/config.json",
  "theme": "system",
  "plugin": ["opencode-antigravity-auth@latest", "oh-my-opencode@latest"],
  "provider": {
    "google": {
      "npm": "@ai-sdk/google",
      "models": {
        "antigravity-gemini-3-pro": {
          "name": "Gemini 3 Pro (Antigravity)",
          "limit": { "context": 1048576, "output": 65535 },
          "modalities": {
            "input": ["text", "image", "pdf"],
            "output": ["text"]
          },
          "variants": {
            "low": { "thinkingLevel": "low" },
            "high": { "thinkingLevel": "high" }
          }
        },
        "antigravity-gemini-3-flash": {
          "name": "Gemini 3 Flash (Antigravity)",
          "limit": { "context": 1048576, "output": 65536 },
          "modalities": {
            "input": ["text", "image", "pdf"],
            "output": ["text"]
          },
          "variants": {
            "minimal": { "thinkingLevel": "minimal" },
            "low": { "thinkingLevel": "low" },
            "medium": { "thinkingLevel": "medium" },
            "high": { "thinkingLevel": "high" }
          }
        },
        "antigravity-claude-sonnet-4-5": {
          "name": "Claude Sonnet 4.5 (Antigravity)",
          "limit": { "context": 200000, "output": 64000 },
          "modalities": {
            "input": ["text", "image", "pdf"],
            "output": ["text"]
          }
        },
        "antigravity-claude-sonnet-4-5-thinking": {
          "name": "Claude Sonnet 4.5 Thinking (Antigravity)",
          "limit": { "context": 200000, "output": 64000 },
          "modalities": {
            "input": ["text", "image", "pdf"],
            "output": ["text"]
          },
          "variants": {
            "low": { "thinkingConfig": { "thinkingBudget": 8192 } },
            "max": { "thinkingConfig": { "thinkingBudget": 32768 } }
          }
        },
        "antigravity-claude-opus-4-5-thinking": {
          "name": "Claude Opus 4.5 Thinking (Antigravity)",
          "limit": { "context": 200000, "output": 64000 },
          "modalities": {
            "input": ["text", "image", "pdf"],
            "output": ["text"]
          },
          "variants": {
            "low": { "thinkingConfig": { "thinkingBudget": 8192 } },
            "max": { "thinkingConfig": { "thinkingBudget": 32768 } }
          }
        },
        "gemini-2.5-flash": {
          "name": "Gemini 2.5 Flash (Gemini CLI)",
          "limit": { "context": 1048576, "output": 65536 },
          "modalities": {
            "input": ["text", "image", "pdf"],
            "output": ["text"]
          }
        },
        "gemini-2.5-pro": {
          "name": "Gemini 2.5 Pro (Gemini CLI)",
          "limit": { "context": 1048576, "output": 65536 },
          "modalities": {
            "input": ["text", "image", "pdf"],
            "output": ["text"]
          }
        },
        "gemini-3-flash-preview": {
          "name": "Gemini 3 Flash Preview (Gemini CLI)",
          "limit": { "context": 1048576, "output": 65536 },
          "modalities": {
            "input": ["text", "image", "pdf"],
            "output": ["text"]
          }
        },
        "gemini-3-pro-preview": {
          "name": "Gemini 3 Pro Preview (Gemini CLI)",
          "limit": { "context": 1048576, "output": 65535 },
          "modalities": {
            "input": ["text", "image", "pdf"],
            "output": ["text"]
          }
        }
      }
    }
  }
}
```

**Key Points:**

- Plugin versi `@latest` (v1.4.2+) sudah include fix untuk version enforcement
- `"npm": "@ai-sdk/google"` **WAJIB** agar plugin bisa intercept model requests
- 9 models available: 5 Antigravity + 4 Gemini CLI

---

### Config 2: antigravity.json (Plugin Settings)

```json
{
  "$schema": "https://raw.githubusercontent.com/NoeFabris/opencode-antigravity-auth/main/assets/antigravity.schema.json",
  "quiet_mode": true,
  "account_selection_strategy": "round-robin",
  "switch_on_first_rate_limit": true,
  "pid_offset_enabled": true,
  "quota_fallback": true,
  "session_recovery": true,
  "scheduling_mode": "cache_first",
  "max_cache_first_wait_seconds": 60,
  "failure_ttl_seconds": 3600
}
```

**Settings Explanation:**

| Setting                        | Value           | Description                                               |
| ------------------------------ | --------------- | --------------------------------------------------------- |
| `quiet_mode`                   | `true`          | Suppress verbose plugin logs                              |
| `account_selection_strategy`   | `"round-robin"` | **WAJIB untuk 4+ accounts** - distributes requests evenly |
| `switch_on_first_rate_limit`   | `true`          | Auto-switch account on rate limit                         |
| `pid_offset_enabled`           | `true`          | Support multiple OpenCode instances                       |
| `quota_fallback`               | `true`          | Use Gemini's secondary quota pool                         |
| `session_recovery`             | `true`          | Recover sessions on restart                               |
| `scheduling_mode`              | `"cache_first"` | Prioritize cached responses                               |
| `max_cache_first_wait_seconds` | `60`            | Max wait for cache                                        |
| `failure_ttl_seconds`          | `3600`          | Remember failed accounts for 1 hour                       |

---

### Config 3: oh-my-opencode.json (Agent Orchestration)

**Pilih salah satu variant:**

#### Option A: Claude-based (Recommended for coding)

File: `oh-my-opencode-full-claude.json`

```json
{
  "$schema": "https://raw.githubusercontent.com/code-yeongyu/oh-my-opencode/master/assets/oh-my-opencode.schema.json",
  "google_auth": false,
  "default_model": "google/antigravity-claude-sonnet-4-5",
  "agents": {
    "sisyphus": {
      "model": "google/antigravity-claude-opus-4-5-thinking",
      "variant": "max"
    },
    "oracle": {
      "model": "google/antigravity-claude-opus-4-5-thinking",
      "variant": "low"
    },
    "librarian": { "model": "google/antigravity-claude-sonnet-4-5" },
    "explore": { "model": "google/antigravity-claude-sonnet-4-5" },
    "multimodal-looker": { "model": "google/antigravity-claude-sonnet-4-5" },
    "prometheus": {
      "model": "google/antigravity-claude-opus-4-5-thinking",
      "variant": "max"
    },
    "metis": {
      "model": "google/antigravity-claude-sonnet-4-5-thinking",
      "variant": "low"
    },
    "momus": { "model": "google/antigravity-claude-sonnet-4-5" },
    "atlas": {
      "model": "google/antigravity-claude-opus-4-5-thinking",
      "variant": "low"
    },
    "executor": {
      "model": "google/antigravity-claude-sonnet-4-5-thinking",
      "variant": "low"
    },
    "reviewer": {
      "model": "google/antigravity-claude-sonnet-4-5-thinking",
      "variant": "max"
    },
    "tester": {
      "model": "google/antigravity-claude-sonnet-4-5-thinking",
      "variant": "low"
    },
    "security-auditor": {
      "model": "google/antigravity-claude-opus-4-5-thinking",
      "variant": "max"
    },
    "refactorer": {
      "model": "google/antigravity-claude-sonnet-4-5-thinking",
      "variant": "max"
    },
    "doc-writer": { "model": "google/antigravity-claude-sonnet-4-5" },
    "frontend-ui-ux-engineer": {
      "model": "google/antigravity-gemini-3-pro",
      "variant": "high"
    },
    "document-writer": {
      "model": "google/antigravity-gemini-3-flash",
      "variant": "medium"
    },
    "default": { "model": "google/antigravity-claude-sonnet-4-5" }
  },
  "categories": {
    "visual-engineering": { "model": "google/antigravity-claude-sonnet-4-5" },
    "artistry": { "model": "google/antigravity-claude-sonnet-4-5" },
    "writing": { "model": "google/antigravity-claude-sonnet-4-5" },
    "quick": { "model": "google/antigravity-claude-sonnet-4-5" },
    "ultrabrain": {
      "model": "google/antigravity-claude-opus-4-5-thinking",
      "variant": "max"
    },
    "implementation": {
      "model": "google/antigravity-claude-sonnet-4-5-thinking",
      "variant": "max"
    },
    "review": {
      "model": "google/antigravity-claude-sonnet-4-5-thinking",
      "variant": "max"
    },
    "testing": {
      "model": "google/antigravity-claude-sonnet-4-5-thinking",
      "variant": "low"
    },
    "security": {
      "model": "google/antigravity-claude-opus-4-5-thinking",
      "variant": "max"
    },
    "unspecified-low": { "model": "google/antigravity-claude-sonnet-4-5" },
    "unspecified-high": {
      "model": "google/antigravity-claude-opus-4-5-thinking",
      "variant": "max"
    }
  }
}
```

#### Option B: Gemini-based (Faster, more quota)

File: `oh-my-opencode-full-gemini.json`

```json
{
  "$schema": "https://raw.githubusercontent.com/code-yeongyu/oh-my-opencode/master/assets/oh-my-opencode.schema.json",
  "google_auth": false,
  "default_model": "google/antigravity-gemini-3-pro",
  "agents": {
    "sisyphus": {
      "model": "google/antigravity-gemini-3-pro",
      "variant": "high"
    },
    "oracle": { "model": "google/gemini-3-pro-preview" },
    "librarian": {
      "model": "google/antigravity-gemini-3-flash",
      "variant": "medium"
    },
    "explore": { "model": "google/gemini-3-flash-preview" },
    "multimodal-looker": {
      "model": "google/antigravity-gemini-3-flash",
      "variant": "high"
    },
    "prometheus": {
      "model": "google/antigravity-gemini-3-pro",
      "variant": "high"
    },
    "metis": { "model": "google/gemini-3-pro-preview" },
    "momus": { "model": "google/gemini-2.5-flash" },
    "atlas": { "model": "google/antigravity-gemini-3-pro", "variant": "high" },
    "executor": {
      "model": "google/antigravity-gemini-3-pro",
      "variant": "low"
    },
    "reviewer": { "model": "google/gemini-3-pro-preview" },
    "tester": { "model": "google/gemini-3-flash-preview" },
    "security-auditor": {
      "model": "google/antigravity-gemini-3-pro",
      "variant": "high"
    },
    "refactorer": {
      "model": "google/antigravity-gemini-3-pro",
      "variant": "high"
    },
    "doc-writer": { "model": "google/gemini-2.5-flash" },
    "frontend-ui-ux-engineer": {
      "model": "google/antigravity-gemini-3-pro",
      "variant": "high"
    },
    "document-writer": {
      "model": "google/antigravity-gemini-3-flash",
      "variant": "medium"
    },
    "default": { "model": "google/antigravity-gemini-3-pro", "variant": "low" }
  },
  "categories": {
    "visual-engineering": {
      "model": "google/antigravity-gemini-3-pro",
      "variant": "high"
    },
    "artistry": { "model": "google/gemini-3-pro-preview" },
    "writing": { "model": "google/gemini-2.5-pro" },
    "quick": { "model": "google/gemini-3-flash-preview" },
    "ultrabrain": {
      "model": "google/antigravity-gemini-3-pro",
      "variant": "high"
    },
    "implementation": {
      "model": "google/antigravity-gemini-3-pro",
      "variant": "high"
    },
    "review": { "model": "google/gemini-3-pro-preview" },
    "testing": {
      "model": "google/antigravity-gemini-3-flash",
      "variant": "high"
    },
    "security": {
      "model": "google/antigravity-gemini-3-pro",
      "variant": "high"
    },
    "unspecified-low": { "model": "google/gemini-2.5-flash" },
    "unspecified-high": { "model": "google/gemini-2.5-pro" }
  }
}
```

---

## Setup Google Authentication

### Option A: Fresh Login (New Computer)

```powershell
# Run OpenCode first (will install plugins)
opencode

# Then login with Google account
opencode auth login
```

Pilih:

1. **Google** sebagai provider
2. **OAuth with Google (Antigravity)**
3. **Add new account**
4. Ikuti instruksi browser untuk authenticate
5. **Ulangi untuk setiap account** (recommended: 5+ accounts)

### Option B: Migrate from Old Computer

Copy file `antigravity-accounts.json` dari komputer lama:

```powershell
# Copy from backup (dari project folder)
Copy-Item "opencode-configs\antigravity-accounts.json" "$HOME\.config\opencode\antigravity-accounts.json" -Force

# PENTING: Verify userAgent is 1.15.8
$accounts = Get-Content "$HOME\.config\opencode\antigravity-accounts.json" | ConvertFrom-Json
$accounts.accounts | ForEach-Object {
    $ua = $_.fingerprint.userAgent
    Write-Host "$($_.email): $ua"
}
```

**Jika userAgent bukan 1.15.8, fix dengan:**

```powershell
$accountsPath = "$HOME\.config\opencode\antigravity-accounts.json"
$data = Get-Content $accountsPath | ConvertFrom-Json
foreach ($acc in $data.accounts) {
    if ($acc.fingerprint.userAgent -notmatch "1\.15\.8") {
        $acc.fingerprint.userAgent = $acc.fingerprint.userAgent -replace "antigravity/[\d.]+", "antigravity/1.15.8"
    }
}
$data | ConvertTo-Json -Depth 10 | Set-Content $accountsPath -Encoding UTF8
Write-Host "✅ All accounts updated to userAgent 1.15.8"
```

---

## Verify Installation

```powershell
Write-Host "=== Checking OpenCode Installation ===" -ForegroundColor Green

# Check OpenCode CLI
Write-Host "`n1. OpenCode CLI:"
opencode --version

# Check config files (LOKASI YANG BENAR)
Write-Host "`n2. Config Files (~/.config/opencode/):"
$configPath = "$HOME\.config\opencode"
@("opencode.json", "antigravity.json", "oh-my-opencode.json", "antigravity-accounts.json") | ForEach-Object {
    if (Test-Path "$configPath\$_") {
        $size = (Get-Item "$configPath\$_").Length
        Write-Host "   ✅ $_ ($size bytes)" -ForegroundColor Green
    } else {
        Write-Host "   ❌ $_ NOT FOUND" -ForegroundColor Red
    }
}

# Check accounts detail
Write-Host "`n3. Google Accounts:"
$accountsPath = "$HOME\.config\opencode\antigravity-accounts.json"
if (Test-Path $accountsPath) {
    $accounts = Get-Content $accountsPath | ConvertFrom-Json
    Write-Host "   Total: $($accounts.accounts.Count) accounts" -ForegroundColor Cyan
    $accounts.accounts | ForEach-Object {
        $status = if ($_.enabled -eq $false) { "❌ DISABLED" } else { "✅" }
        Write-Host "   $status $($_.email)"
    }
} else {
    Write-Host "   ⚠️ Not logged in yet - run 'opencode auth login'" -ForegroundColor Yellow
}

# Check for OLD location (should NOT exist)
Write-Host "`n4. Check OLD location (should NOT exist):"
if (Test-Path "$env:APPDATA\opencode") {
    Write-Host "   ⚠️ OLD folder exists at %APPDATA%\opencode - consider deleting" -ForegroundColor Yellow
} else {
    Write-Host "   ✅ No old config folder" -ForegroundColor Green
}

Write-Host "`n=== Verification Complete ===" -ForegroundColor Green
```

---

## Troubleshooting

### Error: "This version of Antigravity is no longer supported"

**Cause:** userAgent version di accounts file tidak valid (bukan 1.15.8)

**Solution 1:** Update plugin ke v1.4.2+ (auto-fix)

```powershell
# Plugin @latest sudah v1.4.2 yang include fix
# Pastikan opencode.json pakai @latest:
# "plugin": ["opencode-antigravity-auth@latest", ...]
```

**Solution 2:** Manual fix accounts file

```powershell
# LOKASI YANG BENAR: ~/.config/opencode/
$accountsPath = "$HOME\.config\opencode\antigravity-accounts.json"
$data = Get-Content $accountsPath | ConvertFrom-Json
foreach ($acc in $data.accounts) {
    if ($acc.fingerprint.userAgent -notmatch "1\.15\.8") {
        $acc.fingerprint.userAgent = $acc.fingerprint.userAgent -replace "antigravity/[\d.]+", "antigravity/1.15.8"
    }
}
$data | ConvertTo-Json -Depth 10 | Set-Content $accountsPath -Encoding UTF8
Write-Host "✅ Fixed!"
```

> 📌 **Reference:** [GitHub Issue #324](https://github.com/NoeFabris/opencode-antigravity-auth/issues/324)

---

### Error: "ProviderModelNotFoundError: model-name"

**Cause:** Model tidak terdaftar di `opencode.json` atau oh-my-opencode pakai model yang tidak ada.

**Solution:**

1. Pastikan model ada di `opencode.json` provider.google.models
2. Pastikan oh-my-opencode.json pakai model dengan prefix `google/`
3. Contoh benar: `"google/antigravity-claude-sonnet-4-5"`

---

### Error: "All Accounts Rate-Limited"

**Cause:** Semua Google accounts kena rate limit.

**Solution:**

1. Tambah lebih banyak accounts: `opencode auth login`
2. Tunggu rate limit expire (biasanya 1 jam)
3. Gunakan `"quota_fallback": true` di antigravity.json

---

### Sub-agent Errors dengan Antigravity

**Cause:** oh-my-opencode pakai model yang tidak ada di Antigravity.

**Solution:** Gunakan config `oh-my-opencode-full-claude.json` atau `oh-my-opencode-full-gemini.json` yang sudah dikonfigurasi dengan model yang valid.

---

## Current Configuration Summary

### Active Config (January 30, 2026)

| Component             | Version/Value         |
| --------------------- | --------------------- |
| OpenCode CLI          | v1.1.44               |
| Antigravity Plugin    | @latest (v1.4.2+)     |
| oh-my-opencode Plugin | @latest               |
| Config Variant        | Claude-based agents   |
| Google Accounts       | **12 accounts**       |
| Account Strategy      | round-robin           |
| Config Location       | `~/.config/opencode/` |

### Available Models

| Model ID                                 | Type              | Best For                        |
| ---------------------------------------- | ----------------- | ------------------------------- |
| `antigravity-claude-opus-4-5-thinking`   | Claude Opus 4.5   | Complex reasoning, architecture |
| `antigravity-claude-sonnet-4-5-thinking` | Claude Sonnet 4.5 | Coding, implementation          |
| `antigravity-claude-sonnet-4-5`          | Claude Sonnet 4.5 | General tasks                   |
| `antigravity-gemini-3-pro`               | Gemini 3 Pro      | UI/UX, visual                   |
| `antigravity-gemini-3-flash`             | Gemini 3 Flash    | Fast tasks, docs                |
| `gemini-2.5-pro`                         | Gemini 2.5 Pro    | General (CLI quota)             |
| `gemini-2.5-flash`                       | Gemini 2.5 Flash  | Fast tasks (CLI quota)          |
| `gemini-3-pro-preview`                   | Gemini 3 Pro      | Preview (CLI quota)             |
| `gemini-3-flash-preview`                 | Gemini 3 Flash    | Preview (CLI quota)             |

### Agent Assignments (Claude Variant)

| Agent                   | Model                            | Purpose              |
| ----------------------- | -------------------------------- | -------------------- |
| sisyphus                | Claude Opus 4.5 (max)            | Main orchestrator    |
| prometheus              | Claude Opus 4.5 (max)            | Complex reasoning    |
| security-auditor        | Claude Opus 4.5 (max)            | Security review      |
| atlas                   | Claude Opus 4.5 (low)            | Heavy lifting        |
| oracle                  | Claude Opus 4.5 (low)            | Architecture         |
| reviewer                | Claude Sonnet 4.5-thinking (max) | Code review          |
| refactorer              | Claude Sonnet 4.5-thinking (max) | Refactoring          |
| implementation          | Claude Sonnet 4.5-thinking (max) | Implementation       |
| executor                | Claude Sonnet 4.5-thinking (low) | Task execution       |
| tester                  | Claude Sonnet 4.5-thinking (low) | Testing              |
| librarian               | Claude Sonnet 4.5                | Documentation search |
| explore                 | Claude Sonnet 4.5                | Code exploration     |
| frontend-ui-ux-engineer | Gemini 3 Pro (high)              | UI/UX development    |
| document-writer         | Gemini 3 Flash (medium)          | Writing docs         |

---

## Quick Setup Script (All-in-One)

```powershell
# ============================================
# OpenCode Complete Setup Script
# Run from project root (where opencode-configs/ exists)
# ============================================

Write-Host "🚀 Starting OpenCode Setup..." -ForegroundColor Cyan

# Step 1: Install OpenCode CLI
Write-Host "`n📦 Installing OpenCode CLI..." -ForegroundColor Yellow
npm install -g opencode-ai

# Step 2: Create config directory (LOKASI YANG BENAR!)
Write-Host "`n📁 Creating config directory..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path "$HOME\.config\opencode" | Out-Null

# Step 3: Copy config files
Write-Host "`n📝 Copying config files..." -ForegroundColor Yellow
Copy-Item "opencode-configs\opencode.json" "$HOME\.config\opencode\opencode.json" -Force
Copy-Item "opencode-configs\antigravity.json" "$HOME\.config\opencode\antigravity.json" -Force
Copy-Item "opencode-configs\oh-my-opencode-full-claude.json" "$HOME\.config\opencode\oh-my-opencode.json" -Force

Write-Host "   ✅ opencode.json" -ForegroundColor Green
Write-Host "   ✅ antigravity.json" -ForegroundColor Green
Write-Host "   ✅ oh-my-opencode.json (Claude variant)" -ForegroundColor Green

# Step 4: Check for accounts backup in project
if (Test-Path "opencode-configs\antigravity-accounts.json") {
    Write-Host "`n🔐 Found accounts backup in opencode-configs/" -ForegroundColor Yellow
    $confirm = Read-Host "Restore accounts? (y/n)"
    if ($confirm -eq "y") {
        Copy-Item "opencode-configs\antigravity-accounts.json" "$HOME\.config\opencode\antigravity-accounts.json" -Force
        $accounts = Get-Content "$HOME\.config\opencode\antigravity-accounts.json" | ConvertFrom-Json
        Write-Host "   ✅ Restored $($accounts.accounts.Count) accounts" -ForegroundColor Green
    }
} else {
    Write-Host "`n⚠️ No accounts backup found in opencode-configs/" -ForegroundColor Yellow
    Write-Host "   Run 'opencode auth login' to authenticate" -ForegroundColor Yellow
}

# Done
Write-Host "`n✅ Setup Complete!" -ForegroundColor Green
Write-Host "`nConfig location: $HOME\.config\opencode\" -ForegroundColor Cyan
Write-Host "`nNext steps:" -ForegroundColor Cyan
Write-Host "1. Run 'opencode' to start"
Write-Host "2. If no accounts, run 'opencode auth login' to authenticate"
Write-Host "3. Add multiple accounts for better rate limit handling (recommended: 5+)"
Write-Host "4. Enjoy coding with AI! 🎉"
```

---

## Backup & Restore

### Backup Config ke Project

```powershell
# Backup semua config dari ~/.config/opencode/ ke opencode-configs/
$source = "$HOME\.config\opencode"
$dest = "opencode-configs"

Copy-Item "$source\opencode.json" "$dest\opencode.json" -Force
Copy-Item "$source\antigravity.json" "$dest\antigravity.json" -Force
Copy-Item "$source\oh-my-opencode.json" "$dest\oh-my-opencode.json" -Force
Copy-Item "$source\antigravity-accounts.json" "$dest\antigravity-accounts.json" -Force

Write-Host "✅ Backup complete to opencode-configs/" -ForegroundColor Green
Get-ChildItem $dest -Filter "*.json" | Format-Table Name, Length, LastWriteTime
```

### Restore Config dari Project

```powershell
# Restore semua config dari opencode-configs/ ke ~/.config/opencode/
$source = "opencode-configs"
$dest = "$HOME\.config\opencode"

New-Item -ItemType Directory -Force -Path $dest | Out-Null

Copy-Item "$source\opencode.json" "$dest\opencode.json" -Force
Copy-Item "$source\antigravity.json" "$dest\antigravity.json" -Force
Copy-Item "$source\oh-my-opencode.json" "$dest\oh-my-opencode.json" -Force

# Restore accounts jika ada
if (Test-Path "$source\antigravity-accounts.json") {
    Copy-Item "$source\antigravity-accounts.json" "$dest\antigravity-accounts.json" -Force
    Write-Host "✅ Accounts restored!" -ForegroundColor Green
}

Write-Host "✅ Restore complete!" -ForegroundColor Green
```

### List Current Accounts

```powershell
$accounts = Get-Content "$HOME\.config\opencode\antigravity-accounts.json" | ConvertFrom-Json
Write-Host "Total: $($accounts.accounts.Count) accounts`n" -ForegroundColor Cyan

$accounts.accounts | ForEach-Object {
    $status = if ($_.enabled -eq $false) { "❌ DISABLED" } else { "✅ ACTIVE" }
    $lastUsed = [DateTimeOffset]::FromUnixTimeMilliseconds($_.lastUsed).LocalDateTime.ToString("yyyy-MM-dd HH:mm")
    Write-Host "$status $($_.email) (last: $lastUsed)"
}
```

### Enable/Disable Account

```powershell
# Enable account
$email = "balerinatung3@gmail.com"
$path = "$HOME\.config\opencode\antigravity-accounts.json"
$data = Get-Content $path | ConvertFrom-Json
$data.accounts | Where-Object { $_.email -eq $email } | ForEach-Object { $_.enabled = $true }
$data | ConvertTo-Json -Depth 10 | Set-Content $path -Encoding UTF8
Write-Host "✅ Enabled $email"
```

---

## References

### Official Documentation

- **OpenCode:** https://opencode.ai/docs
- **OpenCode Download:** https://opencode.ai/download
- **OpenCode GitHub:** https://github.com/anomalyco/opencode

### Plugins

- **opencode-antigravity-auth:** https://github.com/NoeFabris/opencode-antigravity-auth
- **oh-my-opencode:** https://github.com/code-yeongyu/oh-my-opencode
- **Model Variants:** https://github.com/NoeFabris/opencode-antigravity-auth/blob/main/docs/MODEL-VARIANTS.md
- **Configuration:** https://github.com/NoeFabris/opencode-antigravity-auth/blob/main/docs/CONFIGURATION.md

### Reference Configs

- **aikazu/kcmon-opencode-config:** https://github.com/aikazu/kcmon-opencode-config

### Known Issues & Fixes

- **Version Error Fix:** https://github.com/NoeFabris/opencode-antigravity-auth/issues/324
- **Model Not Found:** https://github.com/NoeFabris/opencode-antigravity-auth/issues/303

---

## TL;DR - Quick Reference

```
📁 CONFIG LOCATION (SINGLE SOURCE OF TRUTH):
   ~/.config/opencode/
   └── opencode.json
   └── antigravity.json
   └── oh-my-opencode.json
   └── antigravity-accounts.json  ← 12 Google accounts

📁 BACKUP LOCATION (PROJECT):
   opencode-configs/
   └── *.json (gitignored: antigravity-accounts.json)

🚀 QUICK SETUP:
   1. npm install -g opencode-ai
   2. Copy opencode-configs/* → ~/.config/opencode/
   3. opencode

⚠️ JANGAN PAKAI %APPDATA%\opencode\ (deprecated!)
```

---

**Created:** January 30, 2026  
**Last Updated:** January 30, 2026  
**Author:** Yumna  
**Project:** OMNI
