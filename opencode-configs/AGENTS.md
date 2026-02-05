# OpenCode Configs - AI Rules

> **STATUS: MANDATORY** | **For: All AI Agents**

---

## File Ownership & Purpose

| File                              | Purpose                               | Who Updates            |
| --------------------------------- | ------------------------------------- | ---------------------- |
| `antigravity-accounts copy.json`  | **MASTER BACKUP** - Complete accounts | **USER ONLY (manual)** |
| `antigravity-accounts.json`       | Working file for daily use            | AI.py smart_sync       |
| `oh-my-opencode-antigravity.json` | Source config (proxy mode names)      | User                   |
| `oh-my-opencode-mix.json`         | Source config (Copilot + Antigravity) | User                   |
| `oh-my-opencode-copilot.json`     | Source config (Copilot only)          | User                   |
| `opencode-plugin.json`            | Plugin mode config                    | User                   |
| `opencode-proxy.json`             | Proxy mode config                     | User                   |
| `antigravity.json`                | Plugin settings                       | AI.py sync             |
| `test-accounts.js`                | Account tester script                 | AI (with approval)     |
| `AI.py`                           | Provider switcher                     | AI (with approval)     |

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

When switching to **Plugin mode**, model names must be transformed:

```
google/claude-*  →  google/antigravity-claude-*
google/gemini-*  →  google/antigravity-gemini-*
```

When switching to **Proxy mode**, model names stay as-is (no prefix).

---

## test-accounts.js Flow

```
SETUP:
  opencode-plugin.json           → ~/.config/opencode/opencode.json
  oh-my-opencode-antigravity.json → ~/.config/opencode/oh-my-opencode.json (TRANSFORMED)

TEST LOOP:
  For each account in "copy.json":
    Write single account → ~/.config/opencode/antigravity-accounts.json
    Run opencode test
    Capture result

RESTORE (always):
  "copy.json" → ~/.config/opencode/antigravity-accounts.json
```

---

## AI.py Flow

```
MODE SELECTION:
  1. Copilot         - oh-my-opencode-copilot.json (no transform)
  2. Antigravity Proxy  - oh-my-opencode-antigravity.json + opencode-proxy.json
  3. Antigravity Plugin - oh-my-opencode-antigravity.json (TRANSFORM) + opencode-plugin.json
  4. Mix Proxy       - oh-my-opencode-mix.json + opencode-proxy.json
  5. Mix Plugin      - oh-my-opencode-mix.json (TRANSFORM) + opencode-plugin.json

SYNC (Plugin modes only):
  smart_sync_accounts() - newest wins across 4 locations
```

---

## When Modifying These Scripts

### Allowed Changes

- Bug fixes in test-accounts.js or AI.py
- Adding new provider modes
- Improving error handling
- Updating model names in config files

### Forbidden Changes

- Adding `copy.json` to any sync mechanism
- Removing transform logic for plugin mode
- Changing restore source from `copy.json` to anything else
- Creating backup mechanisms that read from `~/.config/opencode/`

---

## Troubleshooting

| Symptom                        | Cause                                 | Fix                                                           |
| ------------------------------ | ------------------------------------- | ------------------------------------------------------------- |
| "Model not found" error        | Wrong mode (plugin vs proxy mismatch) | Check oh-my-opencode.json has correct prefixes                |
| Accounts missing after test    | Restore failed                        | Manually copy from `copy.json`                                |
| Sync overwrites with 1 account | Race condition (legacy)               | Should not happen after fix - restore always from `copy.json` |

---

**Version:** 1.0 | **Updated:** 2026-02-05
