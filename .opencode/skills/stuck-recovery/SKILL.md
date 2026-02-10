---
description: Stuck recovery protocol for handling subagent timeouts and failures
---

# Stuck Recovery Protocol

> **Purpose:** Prevent fix→test→fail loops by enforcing flow understanding BEFORE fixing.
> **For:** Sisyphus, Sisyphus-Junior, Atlas, Hephaestus — ALL agents that implement or fix code.

---

## ⚠️ THE #1 RULE: UNDERSTAND THE FLOW FIRST

**The most common reason AI gets stuck is: it doesn't understand the relevant flow.**

The flow depends on what you're working on — it's NOT always the same path:

| Topic                     | Relevant Flow to Research                                      |
| ------------------------- | -------------------------------------------------------------- |
| **Platform integration**  | Platform API → SDK → Handler → Service → Repository → DB       |
| **Database/Schema issue** | Migration → Schema → Repository → Service → Handler            |
| **Backend API bug**       | Handler → Service → Repository → DB query → Response           |
| **Frontend bug**          | Component → API call → Response → State → Render               |
| **Full-stack feature**    | DB schema → Repository → Service → Handler → API → Frontend UI |

Before writing ANY fix, you MUST:

1. **Identify which layers are relevant** to THIS specific issue
2. **Research those layers thoroughly** (read files, @explore for patterns, @librarian for external docs)
3. **Trace the data flow** through each relevant layer
4. **Identify the root cause layer** — fix ONLY there

```
❌ WRONG: See error → guess fix → test → fail → guess again → test → fail (STUCK)
✅ RIGHT: See error → research flow → trace root cause → fix precisely → test → done
```

> **If you fix the same thing twice, STOP. You don't understand the flow.**

---

## Research Protocol (MANDATORY before fix attempt)

### For Bug Fixes

```
Step 1: @explore → Search local SDK for the relevant API/function
Step 2: Read the full chain: Handler → Service → Repository
Step 3: @librarian → Search official API docs (if SDK behavior is unclear)
Step 4: Map the data flow end-to-end
Step 5: IDENTIFY which layer has the root cause
Step 6: FIX only that layer
```

### For New Features

```
Step 1: @explore → Search existing similar patterns in codebase
Step 2: Read reference implementations (same module or similar feature)
Step 3: @librarian → Search API docs for any external integrations
Step 4: Plan implementation following existing patterns
Step 5: Implement
```

---

## Failure Counter & Escalation

Track each attempt:

| Count | Action                                                                              |
| ----- | ----------------------------------------------------------------------------------- |
| 1     | Fix directly, record error. Document what you tried.                                |
| 2     | **STOP.** TRACE FLOW activated. Research full chain before next fix.                |
| 3+    | **TOTAL STOP.** RESEARCH activated. Delegate `@explore` + `@librarian` in parallel. |

### Mandatory Format

```markdown
## Fix Attempt #N

**Failure Count:** N
**Previous Error:** [error message]
**Hypothesis:** [why this fix should work]
**Flow Understanding:**

- Platform API expects: [what]
- SDK sends: [what]
- Handler receives: [what]
- Service processes: [what]
- Root cause: [which layer, what's wrong]

**Action:** [specific fix in specific layer]
```

---

## Subagent Timeout & Failure Recovery

### Detection Signals

| Signal                     | Meaning                 | Severity |
| -------------------------- | ----------------------- | -------- |
| `Tool execution aborted`   | Agent hit timeout/limit | Medium   |
| Empty response             | Agent stuck in thinking | Medium   |
| `session_id` but no answer | Partial execution       | Low      |
| 3x same error              | Stuck in loop           | High     |
| No response after 60s      | Complete timeout        | High     |

### Recovery Actions

#### TIMEOUT/ABORT

```
Step 1: Check session_id from response
Step 2: If exists → delegate_task(session_id="...", prompt="continue from where you stopped")
Step 3: If not exists → retry with more specific prompt
Step 4: Max 2 retries, then fallback
```

#### EMPTY RESPONSE

```
Step 1: Wait 5 seconds
Step 2: Try background_output(task_id=...)
Step 3: If still empty → cancel task
Step 4: Retry with simpler prompt
Step 5: If fails again → fallback to alternative agent
```

#### REPEATED ERROR (3x same error)

```
Step 1: STOP delegating to that agent immediately
Step 2: Log the error pattern
Step 3: Switch to fallback agent
Step 4: If all fallbacks fail → handle manually
```

---

## Fallback Chain

| Primary Agent | Fallback 1                  | Fallback 2       | Last Resort       |
| ------------- | --------------------------- | ---------------- | ----------------- |
| oracle        | librarian + manual analysis | handle yourself  | ask user          |
| librarian     | explore + websearch         | grep/glob direct | ask user          |
| explore       | grep/glob direct            | AST search       | ask user          |
| momus         | manual review checklist     | skip review      | proceed carefully |
| metis         | manual pre-analysis         | skip analysis    | proceed carefully |

---

## Session Continuity Pattern

```typescript
// ALWAYS save session_id from delegation
const result = await delegate_task({...});
const sessionId = result.session_id;

// On failure, RESUME instead of retry fresh
if (failed) {
  await delegate_task({
    session_id: sessionId,
    prompt: "Fix: [specific error]. Continue."
  });
}
```

---

## Manual Takeover Triggers

Switch to manual handling when:

- 2+ retries failed for same task
- Critical path blocked
- User explicitly requests
- All fallback agents exhausted

---

## Anti-Patterns (FORBIDDEN)

| Don't                          | Do Instead                                     |
| ------------------------------ | ---------------------------------------------- |
| Fix without understanding flow | Research the RELEVANT flow for the topic FIRST |
| Retry infinitely               | Max 2 retries, then fallback                   |
| Guess the fix                  | Trace data flow, find root cause               |
| Fix→Test→Fail loop             | STOP at failure 2, trace flow                  |
| Ignore session_id              | Always use for continuity                      |
| Fresh start after fail         | Resume with context                            |
| Silent failure                 | Log and escalate                               |
| Same prompt on retry           | Adjust prompt specificity                      |
| Skip SDK/API reference         | ALWAYS check local SDK first                   |
