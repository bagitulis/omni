---
alwaysApply: true
description: Project agents, identity, and runtime rules (synced from project root)
synced_from: AGENTS.md
---

<!-- AUTO-SYNCED from AGENTS.md by .claude/sync_project_rules.py -->
<!-- Source: AGENTS.md | DO NOT EDIT — edit the source file and re-run sync -->

# OMNI Agent Rules

> **STATUS: MANDATORY** | **Version: 6.0** | **Updated: 2026-05-24**
>
> This is a **thin router** pointing to the full AGENTS.MD file.
> Edit `AGENTS.MD` (uppercase) — the canonical constitution.
> This file is auto-synced by `sync_rules.py` for target compatibility.

---

## Delegation Quick Reference

<!-- MASTER:delegation-routing-compact -->
| Situation             | Delegate To                   |
| --------------------- | ----------------------------- |
| Large task/feature    | Prometheus → Sisyphus         |
| Search in codebase    | `@explore`                    |
| Search external docs  | `@librarian`                  |
| Architecture question | `@oracle`                     |
| UI/Frontend           | category="visual-engineering" |
| Quick fix             | category="quick"              |
<!-- /MASTER:delegation-routing-compact -->

### Delegation Failure Escalation (MANDATORY)

<!-- MASTER:failure-escalation -->
When a delegated task **fails or produces incorrect results** (NOT due to connection loss or timeout):

| Failure Type             | Action                                                     |
| ------------------------ | ---------------------------------------------------------- |
| **Timeout / connection** | Retry with `session_id` in same category                   |
| **Wrong output / error** | **MUST retry using `category="deep"`** on the same task    |
| **Deep also fails**      | Escalate to `@oracle` for analysis, then retry or ask user |

> **Why `deep`?** The `deep` category uses a stronger reasoning model with autonomous problem-solving.
> It performs thorough research before acting — ideal for tasks that lighter categories failed on.
> This prevents wasting retries on the same weak model that already failed.

> **Fallback note:** If `@oracle` is unavailable (rate limit, timeout), use `category="deep"` with detailed analysis prompt instead. See AGENTS.md § Advisory Agent Fallback Matrix.
<!-- /MASTER:failure-escalation -->

## Subagent Stuck Recovery

Fallback chain:

<!-- MASTER:fallback-chain -->
- oracle → librarian → manual
- librarian → explore + websearch → manual
- explore → grep/glob → manual
<!-- /MASTER:fallback-chain -->

---

**See [`AGENTS.MD`](./AGENTS.MD) for the full constitution, critical invariants,**
**architecture pattern, build/test/deploy instructions, and all other rules.**
