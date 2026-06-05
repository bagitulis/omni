# rules-master — DRY Rules Sync System

Single source of truth for all AI agent rules. Edit once in `rules.json`, run `sync_rules.py` to propagate everywhere.

## Quick Start

```bash
# Sync all targets
python rules-master/sync_rules.py

# Check for drift (CI mode — exits 1 if stale)
python rules-master/sync_rules.py --check

# Preview changes without writing
python rules-master/sync_rules.py --dry-run

# Sync specific block or file only
python rules-master/sync_rules.py --section failure-escalation
python rules-master/sync_rules.py --target AGENTS.md
```

## How It Works

### Three Sync Modes

| Mode                | Purpose                                               | Target Files             |
| ------------------- | ----------------------------------------------------- | ------------------------ |
| `inject_block`      | Replace content between `<!-- MASTER:key -->` markers | `.md` rule files         |
| `generate_full_doc` | Generate entire file from template blocks             | `DELEGATION_RULES.md`    |
| `compose_json`      | Compose `prompt_append` strings from DRY blocks       | `opencode-profiles.json` |
| Mode                | Purpose                                               | Target Files             |
| ------------------- | ----------------------------------------------------- | ------------------------ |
| `inject_block`      | Replace content between `<!-- MASTER:key -->` markers | `.md` rule files         |
| `generate_full_doc` | Generate entire file from template blocks             | `DELEGATION_RULES.md`    |
| `compose_json`      | Compose `prompt_append` strings from DRY blocks       | `opencode-profiles.json` |

### File Structure

```
rules-master/
  rules.json                   ← Master: blocks + prompt_blocks + prompt_compose + targets
  sync_rules.py                ← Sync tool (3 modes + --check + --dry-run)
  validate_drift.py            ← Independent structural validator (targets, markers, compose)
  validate_skill_references.py ← Cross-repo skill reference integrity validator
  README.md                    ← This file
```

## rules.json Structure

### `blocks` — Markdown rule blocks for .md injection

Used by `inject_block` and `generate_full_doc` modes. Each key maps to markdown content that gets injected between markers.

```json
{
  "blocks": {
    "failure-escalation": "| Failure Type | Action |\n| --- | --- |\n...",
    "concurrency": "| Setting | Value | ...",
    "fallback-chain": "- oracle → librarian → manual\n..."
  }
}
```

### `prompt_blocks` — DRY prompt_append building blocks

Atomic prompt fragments that get composed into full `prompt_append` strings for agents/categories.

```json
{
  "prompt_blocks": {
    "code-quality": "📏 CODE QUALITY: After editing code files...",
    "understand-flow": "🔍 UNDERSTAND FLOW FIRST: Before fixing ANY bug...",
    "no-premature-done": "✅ NO PREMATURE DONE: Before saying 'done'..."
  }
}
```

### `prompt_compose` — Composition mapping

Defines which `prompt_blocks` each agent/category gets (and in what order).

```json
{
  "prompt_compose": {
    "agents": {
      "sisyphus": [
        "read-file-sisyphus",
        "code-quality",
        "understand-flow",
        "no-premature-done"
      ]
    },
    "categories": {
      "visual-engineering": [
        "constitution",
        "screenshot-rule",
        "code-quality",
        "ui-bug-reporting",
        "understand-flow",
        "no-premature-done"
      ]
    }
  }
}
```

### `targets` — File mapping and sync mode

```json
{
  "targets": {
    "DELEGATION_RULES.md": {
      "mode": "generate_full_doc",
      "sections": ["concurrency", "..."]
    },
    "AGENTS.md": {
      "mode": "inject_block",
      "markers": ["delegation-routing-compact", "..."]
    },
    "opencode-configs/opencode-profiles.json": { "mode": "compose_json" }
  }
}
```

## Adding New Rules

### Add a markdown block (for .md files)

1. Add to `rules.json` → `blocks`:

   ```json
   "my-new-rule": "| Column | Value |\n| --- | --- |\n| Rule 1 | Do X |"
   ```

2. Add marker in target `.md` file:

   ```markdown
   <!-- MASTER:my-new-rule -->
   <!-- /MASTER:my-new-rule -->
   ```

3. Register in `targets` → file → `markers` array

4. Run `python rules-master/sync_rules.py`

### Add a prompt block (for JSON prompt_append)

1. Add to `rules.json` → `prompt_blocks`:

   ```json
   "my-prompt-rule": "🆕 MY RULE: Description of what agents must do."
   ```

2. Add to `prompt_compose` → agent/category arrays where needed

3. Run `python rules-master/sync_rules.py`

## Marker Format

In `.md` files, content between markers is auto-replaced on sync:

```markdown
Some manual content above...

<!-- MASTER:failure-escalation -->

...this content is auto-injected by sync_rules.py...

<!-- /MASTER:failure-escalation -->

More manual content below...
```

**Do NOT edit content between markers** — it will be overwritten on next sync.
## Portability

To use in another project:

1. Copy the `rules-master/` folder
2. Edit `rules.json` blocks/targets for your project
3. Add `<!-- MASTER:key -->` markers in your `.md` files
4. Run `python rules-master/sync_rules.py`

## CI Integration

```bash
# Fails with exit code 1 if any target is out of sync
python rules-master/sync_rules.py --check
```

## Claude Code Integration

Claude Code rules live di `.claude/rules/*.md` dan **auto-generated** dari `rules.json` pakai mode `generate_full_doc`.

### Current Targets

| File | Scope | Blocks Included |
| --- | --- | --- |
| `.claude/rules/delegation-rules.md` | `alwaysApply: true` | delegation-routing-compact, failure-escalation, fallback-chain, session-continuity, failure-counter, prompt-structure |
| `.claude/rules/code-quality.md` | `alwaysApply: true` | file-size-quality, git-commit-gate, stuck-recovery-understand-flow, stuck-recovery-research-protocol |
| `.claude/rules/react-rules.md` | `globs: frontend/**` | skill-react-role, stack, critical-principles, design-system, naming-conventions, anti-patterns |

### Sync Command

```bash
# Sync all (termasuk Claude Code rules)
python rules-master/sync_rules.py

# Sync cuma Claude Code rules
python rules-master/sync_claude_rules.py

# CI check (drift detection)
python rules-master/sync_claude_rules.py --check
```

### Adding New Claude Code Rules

1. **Pilih blocks** dari `rules.json` → `blocks` yang mau di-include
2. **Tambah target** di `rules.json` → `targets.generate_full_doc`:
   ```json
   ".claude/rules/nama-rules.md": {
     "uses": ["block-key-1", "block-key-2"],
     "frontmatter": "---\nalwaysApply: true\ndescription: Deskripsi singkat\n---\n\n<!-- AUTO-GENERATED -->"
   }
   ```
3. **Update `sync_claude_rules.py`** → tambah ke list `CLAUDE_TARGETS`
4. **Run** `python rules-master/sync_rules.py`

### Frontmatter Options

| Field | Type | Purpose |
| --- | --- | --- |
| `alwaysApply` | boolean | `true` = load di semua session |
| `description` | string | Deskripsi 1-line buat Claude |
| `globs` | array | File patterns yang trigger rule (ex: `["frontend/**/*.ts"]`) |

### Auto-Update Strategy

Rules auto-update tiap kali lu jalanin `python rules-master/sync_rules.py`. Cara paling simple:
- **Manual**: Jalain command itu setiap kali edit `rules.json`
- **Pre-commit**: Tambahin ke pre-commit hook (lihat `.git/hooks/`)
- **CI**: Jalain `--check` di CI pipeline, fail kalau ada drift
- **Claude session**: Claude auto-run sync tiap kali ada perubahan di `rules.json`

## Validation

```bash
# Independent structural validation (targets, markers, compose drift)
python rules-master/validate_drift.py

# Cross-repo skill reference integrity
python rules-master/validate_skill_references.py
```

`validate_drift.py` checks:
1. All declared target files exist on disk
2. No orphan markers (markers in files not registered in targets)
3. All markers have proper open/close pairs
4. Composed prompt_append output matches source prompt_blocks

`validate_skill_references.py` checks:
1. All skills referenced in opencode-profiles.json exist in `.opencode/skills/`
2. Skills on disk that are unregistered are documented
3. Uses `--seed-failure REPO:SKILL` to test failure detection

## Policy Notes

### Schema Variant
This project uses the **STANDARD GROUPED** target schema (`inject_block`, `compose_json`
subsections under `targets`). No go_rules targets (n/a — Go backend rules embedded in
per-service sub-routers). Version: 3.0.

### go_rules
Not applicable — Go backend coding rules are managed through per-service sub-routers
(`backend/internal/handlers/AGENTS.md`, `backend/internal/services/AGENTS.md`), not as
a separate go_rules target.

### Architecture
omni has the most mature chunk/router architecture:
- **Thin router**: `AGENTS.md` (55 lines) — delegates to full constitution and sub-routers
- **Canonical constitution**: `AGENTS.MD` (435 lines) — full immutable ruleset
- **Domain sub-routers**: `backend/AGENTS.md`, `frontend/AGENTS.md`, `mcp-servers/AGENTS.md`
- **10 skill inject targets**: `.opencode/skills/*/SKILL.md`
