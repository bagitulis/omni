---
alwaysApply: true
description: Agent delegation, escalation, failure handling, and session continuity
---

| Situation             | Delegate To                   |
| --------------------- | ----------------------------- |
| Large task/feature    | Prometheus → Sisyphus         |
| Search in codebase    | `@explore`                    |
| Search external docs  | `@librarian`                  |
| Architecture question | `@oracle`                     |
| UI/Frontend           | category="visual-engineering" |
| Quick fix             | category="quick"              |

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

- oracle → librarian → manual
- librarian → explore + websearch → manual
- explore → grep/glob → manual

Every `task()` output includes a `session_id`. **ALWAYS use it.**

> **Note:** `task(...)` is the delegation primitive. Older examples may reference `delegate_task()` — treat them as equivalent.

| Scenario               | Action                                             |
| ---------------------- | -------------------------------------------------- |
| Task failed/incomplete | `session_id="...", prompt="Fix: [specific error]"` |
| Follow-up question     | `session_id="...", prompt="Also: [question]"`      |
| Verification failed    | `session_id="...", prompt="Failed: [error]. Fix."` |
| Multi-turn with agent  | `session_id="..."` — NEVER start fresh             |

**Why session_id is CRITICAL:**
- Agent has FULL conversation context preserved
- No repeated file reads, exploration, or setup
- Saves 70%+ tokens on follow-ups
- Agent knows what it already tried/learned

```
❌ WRONG: Task failed → new delegation from scratch (loses all context)
✅ CORRECT: Task failed → session_id="ses_xxx", prompt="Fix: [error]"
```

**Failure counter tracks SAME error/issue.** If a DIFFERENT error occurs, reset counter to 1.

| Count | Action                                                                                |
| ----- | ------------------------------------------------------------------------------------- |
| 1     | Fix directly, record error. Document what was tried.                                  |
| 2     | **STOP fixing.** FULL RESEARCH: trace flow + docs + SDK + references. Fix with evidence. |
| 3-4   | Continue fixing, but MUST use research from step 2. No guessing.                      |
| 5+    | **ASK USER.** Confirm: continue / skip / try different approach. Full failure log.    |

**Reset rule:** Different error = new counter starting at 1. Same error repeating = increment counter.

Every delegation prompt MUST include 6 sections:

```
1. TASK: Atomic, specific goal (one action per delegation)
2. EXPECTED OUTCOME: Concrete deliverables with success criteria
3. REQUIRED TOOLS: Explicit tool whitelist (prevents tool sprawl)
4. MUST DO: Exhaustive requirements — leave NOTHING implicit
5. MUST NOT DO: Forbidden actions — anticipate and block rogue behavior
6. CONTEXT: File paths, existing patterns, constraints
```

**Vague prompts = poor results. Be exhaustive.**
