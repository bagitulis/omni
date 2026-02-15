# OpenCode Configs - AI Rules

> **STATUS: MANDATORY** | **For: All AI Agents**

---

## Architecture Overview (v5.0)

```
AI.py (root, CLI entry point) imports from opencode-configs/:
    ├── ai_profiles.py  ← Profile loading, merging, LSP detection, plugin transform
    └── ai_sync.py      ← Account/config file sync across filesystem locations

opencode-configs/
    ├── opencode-profiles.json  ← Single source of truth (shared + profiles)
    ├── ai_profiles.py          ← Helper module (imported by AI.py)
    ├── ai_sync.py              ← Helper module (imported by AI.py)
    ├── test-accounts.js        ← Account tester (imports test-accounts-helpers.js)
    └── test-accounts-helpers.js← Shared helpers for test-accounts.js

AI.py reads profiles.json → merges shared + profile → generates:
    ~/.config/opencode/oh-my-opencode.json   (valid schema, consumed by opencode)
    ~/.config/opencode/opencode.json         (plugin provider config)
```

**Delivery method**: Plugin only — Via Auth Plugin, transformed names (google/antigravity-gemini-3-pro)

---

## File Ownership & Purpose

| File                             | Purpose                                          | Who Updates            |
| -------------------------------- | ------------------------------------------------ | ---------------------- |
| `opencode-profiles.json`         | **PRIMARY** - Unified profiles (shared + models) | User / AI              |
| `antigravity-accounts copy.json` | **MASTER BACKUP** - Complete accounts            | **USER ONLY (manual)** |
| `antigravity-accounts.json`      | Working file for daily use                       | AI.py smart_sync       |
| `opencode-plugin.json`           | Plugin mode provider config                      | User                   |
| `antigravity.json`               | Plugin settings                                  | AI.py sync             |
| `test-accounts.js`               | Account tester script                            | AI (with approval)     |
| `test-accounts-helpers.js`       | Shared helpers for test-accounts.js              | AI (with approval)     |
| `ai_profiles.py`                 | Profile loading, merging, LSP, plugin transform  | AI (with approval)     |
| `ai_sync.py`                     | Account/config file sync utilities               | AI (with approval)     |
| `../AI.py` (root)                | Provider switcher v5.0 (CLI entry point)         | AI (with approval)     |

### Legacy Files (Deleted)

| File                              | Status  | Notes                              |
| --------------------------------- | ------- | ---------------------------------- |
| `oh-my-opencode-antigravity.json` | Deleted | Replaced by opencode-profiles.json |
| `oh-my-opencode-mix.json`         | Deleted | Replaced by opencode-profiles.json |
| `oh-my-opencode-copilot.json`     | Deleted | Replaced by opencode-profiles.json |
| `switch-provider.ps1`             | Deleted | Replaced by AI.py CLI args         |
| `test-provider.ps1`               | Deleted | Replaced by AI.py CLI args         |
| `opencode-proxy.json`             | Deleted | Proxy mode removed in v5.0         |

---

## CRITICAL RESTRICTIONS

### DO NOT TOUCH

| File                             | Reason                                                 |
| -------------------------------- | ------------------------------------------------------ |
| `antigravity-accounts copy.json` | Master backup - NEVER modify, read, or include in sync |
| `_backup_*.json`                 | Legacy files - should not exist anymore                |

### Sync Flow (4 Locations)

```
AI.py smart_sync_accounts() syncs ONLY these 4:
1. opencode-configs/antigravity-accounts.json
2. ~/.config/opencode/antigravity-accounts.json
3. AppData/Roaming/opencode/antigravity-accounts.json
4. AppData/Local/opencode/antigravity-accounts.json

NEVER add "copy.json" to sync!
```

---

## Transform Rules

Model names are always transformed for plugin mode:

```
google/claude-*  →  google/antigravity-claude-*
google/gemini-*  →  google/antigravity-gemini-*
```

**Note:** Transform happens AFTER merge (on serialized JSON string), so only google/_ models are affected. github-copilot/_ and openai/\* models pass through unchanged.

---

## AI.py Flow (v5.0)

```
MENU (single-step selection):
  [1] Mix Copilot      (Copilot + Google + OpenAI)
  [2] Mix Antigravity   (Google + OpenAI, no Copilot)
  [S] Sync accounts across locations
  [C] Show current provider
  [Q] Quit

PROCESS:
  1. Load opencode-profiles.json           (ai_profiles.load_profiles)
  2. Deep-merge shared + selected profile   (ai_profiles.merge_profile)
  3. Inject LSP servers (gopls, biome)      (ai_profiles.inject_lsp_config)
  4. Serialize to JSON
  5. Transform for plugin                   (ai_profiles.transform_for_plugin)
  6. Write oh-my-opencode.json to ~/.config/opencode/
  7. Copy opencode-plugin.json → opencode.json
  8. Copy antigravity.json
  9. Sync accounts                          (ai_sync.smart_sync_accounts)
  10. Start opencode

CLI ARGS:
  python AI.py mix-copilot        → option 1
  python AI.py mix-antigravity    → option 2
  python AI.py sync               → sync accounts only
  python AI.py current            → show current provider
```

---

## opencode-profiles.json Structure

```json
{
  "shared": {
    "$schema": "...",
    "google_auth": false,
    "browser_automation_engine": { "provider": "agent-browser" },
    "agents": {
      "sisyphus": { "prompt_append": "...", "skills": ["..."] }
    },
    "categories": {
      "visual-engineering": { "skills": ["..."], "prompt_append": "..." }
    }
  },
  "profiles": {
    "mix-copilot": {
      "default_model": "github-copilot/claude-opus-4.5",
      "agents": {
        "sisyphus": { "model": "github-copilot/claude-opus-4.5" }
      },
      "categories": {}
    },
    "mix-antigravity": {
      "default_model": "google/gemini-3-pro",
      "variant": "high",
      "agents": {},
      "categories": {}
    }
  }
}
```

**Merge logic:** For each agent/category, start with deep copy of shared entry (prompt_append, skills), then `.update()` with profile entry (model, variant, temperature, reasoningEffort).

---

## test-accounts.js Flow

```
SETUP:
  opencode-plugin.json            → ~/.config/opencode/opencode.json
  oh-my-opencode (via profiles)   → ~/.config/opencode/oh-my-opencode.json (TRANSFORMED)

TEST LOOP:
  For each account in "copy.json":
    Write single account → ~/.config/opencode/antigravity-accounts.json
    Run opencode test
    Capture result

RESTORE (always):
  "copy.json" → ~/.config/opencode/antigravity-accounts.json
```

---

## When Modifying These Scripts

### Allowed Changes

- Bug fixes in test-accounts.js, AI.py, ai_profiles.py, or ai_sync.py
- Adding new profiles to opencode-profiles.json
- Improving error handling
- Updating model names in profiles

### Forbidden Changes

- Adding `copy.json` to any sync mechanism
- Removing transform logic for plugin mode
- Changing restore source from `copy.json` to anything else
- Creating backup mechanisms that read from `~/.config/opencode/`
- Adding invalid keys to oh-my-opencode.json output (must pass schema validation)

---

## Troubleshooting

| Symptom                        | Cause                                | Fix                                                             |
| ------------------------------ | ------------------------------------ | --------------------------------------------------------------- |
| "Model not found" error        | Plugin auth failed or model mismatch | Check oh-my-opencode.json has antigravity-\* prefixes           |
| Accounts missing after test    | Restore failed                       | Manually copy from `copy.json`                                  |
| Sync overwrites with 1 account | Race condition (legacy)              | Should not happen after fix - restore always from `copy.json`   |
| Profile merge missing fields   | Shared or profile entry incomplete   | Check opencode-profiles.json has entry in both shared + profile |

---

**Version:** 3.0 | **Updated:** 2026-02-15
