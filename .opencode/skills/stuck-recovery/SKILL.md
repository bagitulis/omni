---
description: Stuck recovery protocol for handling subagent timeouts and failures
---

# Stuck Recovery Protocol

> **Purpose:** Handle subagent timeouts, empty responses, and repeated failures.
> **For:** Sisyphus and Sisyphus-Junior orchestrators.

---

## Detection Signals

| Signal                     | Meaning                 | Severity |
| -------------------------- | ----------------------- | -------- |
| `Tool execution aborted`   | Agent hit timeout/limit | Medium   |
| Empty response             | Agent stuck in thinking | Medium   |
| `session_id` but no answer | Partial execution       | Low      |
| 3x same error              | Stuck in loop           | High     |
| No response after 60s      | Complete timeout        | High     |

---

## Recovery Actions

### 1. TIMEOUT/ABORT

```
Step 1: Check session_id from response
Step 2: If exists → delegate_task(session_id="...", prompt="continue from where you stopped")
Step 3: If not exists → retry with more specific prompt
Step 4: Max 2 retries, then fallback
```

### 2. EMPTY RESPONSE

```
Step 1: Wait 5 seconds
Step 2: Try background_output(task_id=...)
Step 3: If still empty → cancel task
Step 4: Retry with simpler prompt
Step 5: If fails again → fallback to alternative agent
```

### 3. REPEATED ERROR (3x same error)

```
Step 1: STOP delegating to that agent immediately
Step 2: Log the error pattern
Step 3: Switch to fallback agent
Step 4: If all fallbacks fail → handle manually
```

### 4. PARTIAL EXECUTION

```
Step 1: Use session_id to continue
Step 2: Provide context: "Previous attempt stopped at: [last output]"
Step 3: Ask to complete remaining work
```

---

## Fallback Chain

| Primary Agent | Fallback 1                  | Fallback 2       | Last Resort       |
| ------------- | --------------------------- | ---------------- | ----------------- |
| oracle        | librarian + manual analysis | handle sendiri   | ask user          |
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

## Anti-Patterns

| Don't                  | Do Instead                   |
| ---------------------- | ---------------------------- |
| Retry infinitely       | Max 2 retries, then fallback |
| Ignore session_id      | Always use for continuity    |
| Fresh start after fail | Resume with context          |
| Silent failure         | Log and escalate             |
| Same prompt on retry   | Adjust prompt specificity    |
