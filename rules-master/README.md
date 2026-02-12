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

### File Structure

```
rules-master/
  rules.json      ← Master: blocks + prompt_blocks + prompt_compose + targets
  sync_rules.py   ← Sync tool (3 modes + --check + --dry-run)
  README.md       ← This file
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
