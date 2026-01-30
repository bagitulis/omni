# 🚀 OpenCode Setup Guide - Complete Installation Tutorial

> **For AI Agents:** Baca dan jalankan tutorial ini step-by-step untuk setup OpenCode di komputer baru.

**Last Updated:** January 30, 2026  
**Tested on:** Windows 11

---

## 📑 Table of Contents

1. [Prerequisites](#prerequisites)
2. [Install OpenCode CLI](#install-opencode-cli)
3. [Create Configuration Files](#create-configuration-files)
4. [Setup Google Authentication](#setup-google-authentication)
5. [Verify Installation](#verify-installation)
6. [Troubleshooting](#troubleshooting)
7. [References](#references)

---

## Prerequisites

Pastikan sudah terinstall:

- **Node.js** (v18+) - https://nodejs.org/
- **npm** (included with Node.js)
- **Git** (optional) - https://git-scm.com/

Cek dengan command:

```powershell
node --version
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

### Alternative Methods

| Method      | Command                      |
| ----------- | ---------------------------- |
| NPM         | `npm install -g opencode-ai` |
| Bun         | `bun add -g opencode-ai`     |
| Chocolatey  | `choco install opencode`     |
| Scoop       | `scoop install opencode`     |
| Desktop App | https://opencode.ai/download |

**Reference:** https://opencode.ai/docs

---

## Create Configuration Files

### Step 1: Create Config Directory

```powershell
# Create config directory (if not exists)
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.config\opencode"
```

### Step 2: Create opencode.json (Main Config)

Buat file `~/.config/opencode/opencode.json`:

```powershell
$opencodeConfig = @'
{
  "$schema": "https://opencode.ai/config.json",
  "theme": "system",
  "plugin": [
    "opencode-antigravity-auth@beta",
    "oh-my-opencode@latest"
  ],
  "provider": {
    "google": {
      "npm": "@ai-sdk/google",
      "models": {
        "antigravity-gemini-3-pro": {
          "name": "Gemini 3 Pro (Antigravity)",
          "limit": { "context": 1048576, "output": 65535 },
          "modalities": { "input": ["text", "image", "pdf"], "output": ["text"] },
          "variants": {
            "low": { "thinkingLevel": "low" },
            "high": { "thinkingLevel": "high" }
          }
        },
        "antigravity-gemini-3-flash": {
          "name": "Gemini 3 Flash (Antigravity)",
          "limit": { "context": 1048576, "output": 65536 },
          "modalities": { "input": ["text", "image", "pdf"], "output": ["text"] },
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
          "modalities": { "input": ["text", "image", "pdf"], "output": ["text"] }
        },
        "antigravity-claude-sonnet-4-5-thinking": {
          "name": "Claude Sonnet 4.5 Thinking (Antigravity)",
          "limit": { "context": 200000, "output": 64000 },
          "modalities": { "input": ["text", "image", "pdf"], "output": ["text"] },
          "variants": {
            "low": { "thinkingConfig": { "thinkingBudget": 8192 } },
            "max": { "thinkingConfig": { "thinkingBudget": 32768 } }
          }
        },
        "antigravity-claude-opus-4-5-thinking": {
          "name": "Claude Opus 4.5 Thinking (Antigravity)",
          "limit": { "context": 200000, "output": 64000 },
          "modalities": { "input": ["text", "image", "pdf"], "output": ["text"] },
          "variants": {
            "low": { "thinkingConfig": { "thinkingBudget": 8192 } },
            "max": { "thinkingConfig": { "thinkingBudget": 32768 } }
          }
        },
        "gemini-2.5-flash": {
          "name": "Gemini 2.5 Flash (Gemini CLI)",
          "limit": { "context": 1048576, "output": 65536 },
          "modalities": { "input": ["text", "image", "pdf"], "output": ["text"] }
        },
        "gemini-2.5-pro": {
          "name": "Gemini 2.5 Pro (Gemini CLI)",
          "limit": { "context": 1048576, "output": 65536 },
          "modalities": { "input": ["text", "image", "pdf"], "output": ["text"] }
        },
        "gemini-3-flash-preview": {
          "name": "Gemini 3 Flash Preview (Gemini CLI)",
          "limit": { "context": 1048576, "output": 65536 },
          "modalities": { "input": ["text", "image", "pdf"], "output": ["text"] }
        },
        "gemini-3-pro-preview": {
          "name": "Gemini 3 Pro Preview (Gemini CLI)",
          "limit": { "context": 1048576, "output": 65535 },
          "modalities": { "input": ["text", "image", "pdf"], "output": ["text"] }
        }
      }
    }
  }
}
'@
$opencodeConfig | Set-Content -Path "$env:USERPROFILE\.config\opencode\opencode.json" -Encoding UTF8
Write-Host "✅ opencode.json created!"
```

### Step 3: Create antigravity.json (Plugin Settings)

Buat file `~/.config/opencode/antigravity.json`:

```powershell
$antigravityConfig = @'
{
  "$schema": "https://raw.githubusercontent.com/NoeFabris/opencode-antigravity-auth/main/assets/antigravity.schema.json",
  "quiet_mode": true,
  "account_selection_strategy": "round-robin",
  "switch_on_first_rate_limit": true,
  "pid_offset_enabled": true,
  "quota_fallback": true
}
'@
$antigravityConfig | Set-Content -Path "$env:USERPROFILE\.config\opencode\antigravity.json" -Encoding UTF8
Write-Host "✅ antigravity.json created!"
```

#### Antigravity Config Options

| Setting                      | Value           | Description                                                |
| ---------------------------- | --------------- | ---------------------------------------------------------- |
| `quiet_mode`                 | `true`          | Suppress verbose plugin logs                               |
| `account_selection_strategy` | `"round-robin"` | **Required for 4+ accounts** - distributes requests evenly |
| `switch_on_first_rate_limit` | `true`          | Auto-switch account on rate limit                          |
| `pid_offset_enabled`         | `true`          | Support multiple OpenCode instances                        |
| `quota_fallback`             | `true`          | Use Gemini's secondary quota pool                          |

### Step 4: Create oh-my-opencode.json (Agent Settings)

Buat file `~/.config/opencode/oh-my-opencode.json`:

```powershell
$ohmyopencodeConfig = @'
{
  "google_auth": false,
  "default_model": "google/antigravity-claude-sonnet-4-5",
  "agents": {
    "sisyphus": { "model": "google/antigravity-claude-opus-4-5-thinking" },
    "atlas": { "model": "google/antigravity-claude-opus-4-5-thinking" },
    "prometheus": { "model": "google/antigravity-claude-sonnet-4-5-thinking" },
    "oracle": { "model": "google/antigravity-claude-sonnet-4-5" },
    "librarian": { "model": "google/antigravity-claude-sonnet-4-5" },
    "frontend-ui-ux-engineer": { "model": "google/antigravity-gemini-3-pro" },
    "document-writer": { "model": "google/antigravity-gemini-3-flash" },
    "multimodal-looker": { "model": "google/antigravity-gemini-3-flash" },
    "explore": { "model": "google/antigravity-gemini-3-flash" },
    "default": { "model": "google/antigravity-claude-sonnet-4-5" }
  }
}
'@
$ohmyopencodeConfig | Set-Content -Path "$env:USERPROFILE\.config\opencode\oh-my-opencode.json" -Encoding UTF8
Write-Host "✅ oh-my-opencode.json created!"
```

---

## Setup Google Authentication

### Option A: Fresh Login (New Computer)

```powershell
# Run OpenCode first to install plugins
opencode

# Then login with Google account
opencode auth login
```

Pilih:

1. **Google** sebagai provider
2. **OAuth with Google (Antigravity)**
3. **Add new account**
4. Ikuti instruksi browser untuk authenticate

### Option B: Migrate from Old Computer

Copy file `antigravity-accounts.json` dari komputer lama:

**Lokasi file:**

- **Windows:** `C:\Users\<USERNAME>\AppData\Roaming\opencode\antigravity-accounts.json`
- **Linux/Mac:** `~/.config/opencode/antigravity-accounts.json`

```powershell
# Copy from backup/old computer to:
Copy-Item "path\to\backup\antigravity-accounts.json" -Destination "$env:APPDATA\opencode\antigravity-accounts.json"
```

> ⚠️ **Note:** Jika refresh token expired, jalankan `opencode auth login` untuk re-authenticate.

---

## Verify Installation

### Check All Files Exist

```powershell
Write-Host "=== Checking OpenCode Installation ===" -ForegroundColor Green

# Check OpenCode CLI
Write-Host "`n1. OpenCode CLI:"
opencode --version

# Check config files
Write-Host "`n2. Config Files:"
$configPath = "$env:USERPROFILE\.config\opencode"
@("opencode.json", "antigravity.json", "oh-my-opencode.json") | ForEach-Object {
    if (Test-Path "$configPath\$_") {
        Write-Host "   ✅ $_" -ForegroundColor Green
    } else {
        Write-Host "   ❌ $_ NOT FOUND" -ForegroundColor Red
    }
}

# Check accounts file
Write-Host "`n3. Account Data:"
$accountsPath = "$env:APPDATA\opencode\antigravity-accounts.json"
if (Test-Path $accountsPath) {
    Write-Host "   ✅ antigravity-accounts.json" -ForegroundColor Green
} else {
    Write-Host "   ⚠️ Not logged in yet - run 'opencode auth login'" -ForegroundColor Yellow
}

Write-Host "`n=== Verification Complete ===" -ForegroundColor Green
```

### Test OpenCode

```powershell
# Start OpenCode
opencode

# Or run a quick test
opencode run "Hello, what model are you?" --model=google/antigravity-claude-sonnet-4-5
```

---

## Configuration Summary

### File Locations

| File                        | Location              | Purpose                            |
| --------------------------- | --------------------- | ---------------------------------- |
| `opencode.json`             | `~/.config/opencode/` | Main config (plugins, models)      |
| `antigravity.json`          | `~/.config/opencode/` | Antigravity plugin settings        |
| `oh-my-opencode.json`       | `~/.config/opencode/` | Agent orchestration settings       |
| `antigravity-accounts.json` | `%APPDATA%/opencode/` | Google account tokens (sensitive!) |

### Installed Plugins

| Plugin                      | Version | Purpose                                                             |
| --------------------------- | ------- | ------------------------------------------------------------------- |
| `opencode-antigravity-auth` | @beta   | Google OAuth for Claude/Gemini models (use @beta for latest fixes!) |
| `oh-my-opencode`            | @latest | Multi-agent orchestration (Sisyphus, Oracle, etc.)                  |

### Available Models

| Model ID                                        | Description                  |
| ----------------------------------------------- | ---------------------------- |
| `google/antigravity-claude-opus-4-5-thinking`   | Claude Opus 4.5 (thinking)   |
| `google/antigravity-claude-sonnet-4-5-thinking` | Claude Sonnet 4.5 (thinking) |
| `google/antigravity-claude-sonnet-4-5`          | Claude Sonnet 4.5            |
| `google/antigravity-gemini-3-pro`               | Gemini 3 Pro                 |
| `google/antigravity-gemini-3-flash`             | Gemini 3 Flash               |
| `google/gemini-2.5-pro`                         | Gemini 2.5 Pro (CLI quota)   |
| `google/gemini-2.5-flash`                       | Gemini 2.5 Flash (CLI quota) |

### oh-my-opencode Agents

| Agent                 | Model             | Specialization          |
| --------------------- | ----------------- | ----------------------- |
| **default_model**     | Claude Sonnet 4.5 | Fallback for unknown    |
| **Sisyphus**          | Claude Opus 4.5   | Main orchestrator       |
| **Atlas**             | Claude Opus 4.5   | Heavy lifting tasks     |
| **Prometheus**        | Claude Opus 4.5   | Complex reasoning       |
| **Oracle**            | Claude Sonnet 4.5 | Architecture, debugging |
| **Librarian**         | Claude Sonnet 4.5 | Docs, code search       |
| **Frontend Engineer** | Gemini 3 Pro      | UI/UX development       |
| **Document Writer**   | Gemini 3 Flash    | Documentation           |
| **Multimodal Looker** | Gemini 3 Flash    | Image analysis          |
| **Explore**           | Gemini 3 Flash    | Fast codebase grep      |
| **default**           | Claude Sonnet 4.5 | Default agent           |

---

## Troubleshooting

### Error: "This version of Antigravity is no longer supported"

**Cause:** Plugin outdated or account token expired.

**Solution:**

```powershell
# Clear plugin cache
Remove-Item "$env:USERPROFILE\.config\opencode\node_modules" -Recurse -Force
Remove-Item "$env:USERPROFILE\.config\opencode\bun.lock" -Force
Remove-Item "$env:USERPROFILE\.config\opencode\package.json" -Force
Remove-Item "$env:USERPROFILE\.config\opencode\package-lock.json" -Force

# Restart opencode (will reinstall plugins)
opencode

# Re-authenticate if needed
opencode auth login
```

### Error: "ProviderModelNotFoundError: claude-haiku-4-5"

**Cause:** oh-my-opencode using default model not available in Antigravity.

**Solution:** Create/update `oh-my-opencode.json` with Antigravity models (see Step 4 above). Make sure to include `default_model` as fallback.

### Error: "ProviderModelNotFoundError: gemini-3-pro" atau model lain

**Cause:** Plugin tidak bisa intercept model karena `npm` provider tidak dikonfigurasi.

**Solution:** Pastikan `opencode.json` menggunakan:

1. Plugin versi `@beta` (bukan `@latest`)
2. Provider `google` punya entry `"npm": "@ai-sdk/google"`

```json
{
  "plugin": ["opencode-antigravity-auth@beta", "oh-my-opencode@latest"],
  "provider": {
    "google": {
      "npm": "@ai-sdk/google",
      "models": { ... }
    }
  }
}
```

> 📌 **Reference:** [GitHub Issue #303](https://github.com/NoeFabris/opencode-antigravity-auth/issues/303), [GitHub Issue #310](https://github.com/NoeFabris/opencode-antigravity-auth/issues/310)

### Error: "opencode is not recognized as a command"

**Cause:** OpenCode not installed globally.

**Solution:**

```powershell
npm install -g opencode-ai
```

### Error: "This version of Antigravity is no longer supported" (After Fresh Install)

**Cause:** The plugin generates random version identifiers that may be blocked.

**Solution:** Apply the version 1.15.8 patch:

```powershell
# 1. Find the fingerprint.js file
$fingerprintPath = Get-ChildItem -Path "$env:USERPROFILE\.config\opencode\node_modules" -Recurse -Filter "fingerprint.js" |
    Where-Object { $_.FullName -match "opencode-antigravity-auth" } |
    Select-Object -First 1 -ExpandProperty FullName

Write-Host "Found: $fingerprintPath" -ForegroundColor Yellow

# 2. Backup original
Copy-Item $fingerprintPath "$fingerprintPath.backup"

# 3. Read and patch the file
$content = Get-Content $fingerprintPath -Raw

# Replace version generation with fixed version 1.15.8
$pattern = 'const versions = \[`1\.\$\{.*?\}`\];'
$replacement = 'const versions = ["1.15.8"];'
$newContent = $content -replace $pattern, $replacement

# Save patched file
$newContent | Set-Content $fingerprintPath -Encoding UTF8
Write-Host "✅ Patched to version 1.15.8" -ForegroundColor Green

# 4. Also fix accounts file if needed
$accountsPath = "$env:APPDATA\opencode\antigravity-accounts.json"
if (Test-Path $accountsPath) {
    $accounts = Get-Content $accountsPath -Raw | ConvertFrom-Json
    $accounts.version = "1.15.8"
    $accounts | ConvertTo-Json -Depth 10 | Set-Content $accountsPath -Encoding UTF8
    Write-Host "✅ Accounts file updated to version 1.15.8" -ForegroundColor Green
}
```

**Atau gunakan script otomatis:**

Script `scripts/fix-antigravity.ps1` sudah tersedia di project ini:

```powershell
# Jalankan dari root project
.\scripts\fix-antigravity.ps1
```

> 📌 **Reference:** [GitHub Issue #324](https://github.com/NoeFabris/opencode-antigravity-auth/issues/324)

### Error: "All Accounts Rate-Limited"

**Cause:** Google account quota exceeded.

**Solution:**

1. Add more Google accounts: `opencode auth login`
2. Wait for rate limit to expire
3. Switch strategy in `antigravity.json`:

```json
{
  "account_selection_strategy": "sticky"
}
```

---

## Quick Setup Script (All-in-One)

Copy dan jalankan script ini untuk setup lengkap:

```powershell
# ============================================
# OpenCode Complete Setup Script
# ============================================

Write-Host "🚀 Starting OpenCode Setup..." -ForegroundColor Cyan

# Step 1: Install OpenCode
Write-Host "`n📦 Installing OpenCode CLI..." -ForegroundColor Yellow
npm install -g opencode-ai

# Step 2: Create config directory
Write-Host "`n📁 Creating config directory..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.config\opencode" | Out-Null

# Step 3: Create opencode.json
Write-Host "`n📝 Creating opencode.json..." -ForegroundColor Yellow
@'
{
  "$schema": "https://opencode.ai/config.json",
  "theme": "system",
  "plugin": ["opencode-antigravity-auth@beta", "oh-my-opencode@latest"],
  "provider": {
    "google": {
      "npm": "@ai-sdk/google",
      "models": {
        "antigravity-gemini-3-pro": {"name": "Gemini 3 Pro (Antigravity)", "limit": {"context": 1048576, "output": 65535}, "modalities": {"input": ["text", "image", "pdf"], "output": ["text"]}, "variants": {"low": {"thinkingLevel": "low"}, "high": {"thinkingLevel": "high"}}},
        "antigravity-gemini-3-flash": {"name": "Gemini 3 Flash (Antigravity)", "limit": {"context": 1048576, "output": 65536}, "modalities": {"input": ["text", "image", "pdf"], "output": ["text"]}, "variants": {"minimal": {"thinkingLevel": "minimal"}, "low": {"thinkingLevel": "low"}, "medium": {"thinkingLevel": "medium"}, "high": {"thinkingLevel": "high"}}},
        "antigravity-claude-sonnet-4-5": {"name": "Claude Sonnet 4.5 (Antigravity)", "limit": {"context": 200000, "output": 64000}, "modalities": {"input": ["text", "image", "pdf"], "output": ["text"]}},
        "antigravity-claude-sonnet-4-5-thinking": {"name": "Claude Sonnet 4.5 Thinking (Antigravity)", "limit": {"context": 200000, "output": 64000}, "modalities": {"input": ["text", "image", "pdf"], "output": ["text"]}, "variants": {"low": {"thinkingConfig": {"thinkingBudget": 8192}}, "max": {"thinkingConfig": {"thinkingBudget": 32768}}}},
        "antigravity-claude-opus-4-5-thinking": {"name": "Claude Opus 4.5 Thinking (Antigravity)", "limit": {"context": 200000, "output": 64000}, "modalities": {"input": ["text", "image", "pdf"], "output": ["text"]}, "variants": {"low": {"thinkingConfig": {"thinkingBudget": 8192}}, "max": {"thinkingConfig": {"thinkingBudget": 32768}}}},
        "gemini-2.5-flash": {"name": "Gemini 2.5 Flash (Gemini CLI)", "limit": {"context": 1048576, "output": 65536}, "modalities": {"input": ["text", "image", "pdf"], "output": ["text"]}},
        "gemini-2.5-pro": {"name": "Gemini 2.5 Pro (Gemini CLI)", "limit": {"context": 1048576, "output": 65536}, "modalities": {"input": ["text", "image", "pdf"], "output": ["text"]}},
        "gemini-3-flash-preview": {"name": "Gemini 3 Flash Preview (Gemini CLI)", "limit": {"context": 1048576, "output": 65536}, "modalities": {"input": ["text", "image", "pdf"], "output": ["text"]}},
        "gemini-3-pro-preview": {"name": "Gemini 3 Pro Preview (Gemini CLI)", "limit": {"context": 1048576, "output": 65535}, "modalities": {"input": ["text", "image", "pdf"], "output": ["text"]}}
      }
    }
  }
}
'@ | Set-Content -Path "$env:USERPROFILE\.config\opencode\opencode.json" -Encoding UTF8

# Step 4: Create antigravity.json (with round-robin for multi-account)
Write-Host "`n📝 Creating antigravity.json..." -ForegroundColor Yellow
@'
{"$schema": "https://raw.githubusercontent.com/NoeFabris/opencode-antigravity-auth/main/assets/antigravity.schema.json", "quiet_mode": true, "account_selection_strategy": "round-robin", "switch_on_first_rate_limit": true, "pid_offset_enabled": true, "quota_fallback": true}
'@ | Set-Content -Path "$env:USERPROFILE\.config\opencode\antigravity.json" -Encoding UTF8

# Step 5: Create oh-my-opencode.json
Write-Host "`n📝 Creating oh-my-opencode.json..." -ForegroundColor Yellow
@'
{"google_auth": false, "default_model": "google/antigravity-claude-sonnet-4-5", "agents": {"sisyphus": {"model": "google/antigravity-claude-opus-4-5-thinking"}, "atlas": {"model": "google/antigravity-claude-opus-4-5-thinking"}, "prometheus": {"model": "google/antigravity-claude-opus-4-5-thinking"}, "oracle": {"model": "google/antigravity-claude-sonnet-4-5"}, "librarian": {"model": "google/antigravity-claude-sonnet-4-5"}, "frontend-ui-ux-engineer": {"model": "google/antigravity-gemini-3-pro"}, "document-writer": {"model": "google/antigravity-gemini-3-flash"}, "multimodal-looker": {"model": "google/antigravity-gemini-3-flash"}, "explore": {"model": "google/antigravity-gemini-3-flash"}, "default": {"model": "google/antigravity-claude-sonnet-4-5"}}}
'@ | Set-Content -Path "$env:USERPROFILE\.config\opencode\oh-my-opencode.json" -Encoding UTF8

# Done
Write-Host "`n✅ Setup Complete!" -ForegroundColor Green
Write-Host "`nNext steps:" -ForegroundColor Cyan
Write-Host "1. Run 'opencode' to start (plugins will install automatically)"
Write-Host "2. Run 'opencode auth login' to authenticate with Google"
Write-Host "3. Enjoy coding with AI! 🎉"
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

### Related Resources

- **NPM opencode-antigravity-auth:** https://www.npmjs.com/package/opencode-antigravity-auth
- **NPM oh-my-opencode:** https://www.npmjs.com/package/oh-my-opencode

---

**Created:** January 30, 2026  
**Author:** AI Assistant  
**Project:** OMNI
