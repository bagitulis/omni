---
description: ULP loop-mode rules - pipeline management, verification cadence, and Definition of Done for autonomous work loops
---

# ULP Loop Mode Rules

> **Purpose:** Governs Sisyphus behavior during autonomous loop execution (`/ulw-loop`).
> **For:** Sisyphus (main orchestrator) when running in loop mode.
> These rules COMPLEMENT (not replace) SISYPHUS_RULES.md and AGENTS.md.

---

## Your Role in Loop Mode

You are the **Main Controller** and **Quality Lead**. Your job is to:

1. **Manage the delegation pipeline** — keep executors busy, slots filled
2. **Inspect and verify** every result — don't just collect reports
3. **Enforce Definition of Done** — no task is complete without passing all gates

---

## 1. Pipeline Management

### Parallel Execution

<!-- MASTER:concurrency -->
| Setting               | Value | Controlled By                               |
| --------------------- | ----- | ------------------------------------------- |
| **Max running tasks** | 3     | System config (`defaultConcurrency: 3`)     |
| **Max queued tasks**  | ∞     | System auto-queues excess                   |
| **Stale timeout**     | 10min | System config (`staleTimeoutMs: 600000`)    |
| **Manual throttling** | NONE  | System handles concurrency — don't throttle |

**Strategy:**

- **Launch aggressively**: Fire as many delegations as the task warrants (5, 6, 10 — doesn't matter)
- **System queues excess**: Only 3 run simultaneously; rest wait in queue automatically
- **Stay productive**: While agents run, continue on other work items or standby for user
- **Never block**: Don't wait idle for delegation results — system notifies on completion

```
❌ WRONG: "I'll limit myself to 3 delegations"
✅ CORRECT: Fire all needed delegations → system queues → collect results when ready

❌ WRONG: Fire 3 agents → wait → wait → wait → respond
✅ CORRECT: Fire agents → continue working / standby → process results as they arrive
```
<!-- /MASTER:concurrency -->

### Task Flow

```
Prometheus Plan → TODO Queue → Delegate (max 3) → Verify → Next
                                    ↑                    |
                                    └────────────────────┘
                                    (if verification fails)
```

### Dynamic Planning

- When a batch finishes, update the master TODO immediately
- Planning updates happen in the background — don't let them block execution
- If scope changes mid-loop, update todos before continuing

---

## 2. Verification Protocol (Double Evaluation)

Every delegated task MUST pass **two evaluations** before being marked DONE:

### Evaluation 1: Backend/Logic Verification

Run these checks **directly** (don't delegate — faster and more reliable):

- [ ] `go build ./...` passes (backend)
- [ ] `go test ./...` passes (backend)
- [ ] `npm run build` passes (frontend, if changed)
- [ ] `npm run lint` passes (frontend, if changed)
- [ ] `lsp_diagnostics` clean on changed files
- [ ] No `as any`, `@ts-ignore`, empty catches

### Evaluation 2: Integration/UI Verification

**Delegate** UI/integration checks where visual verification adds value:

- [ ] Data shows correctly in the UI (use Playwright/browser tools)
- [ ] API responses match expected format (snake_case JSON)
- [ ] No broken layouts, missing columns, or UI bugs
- [ ] End-to-end flow works (user action → backend → database → UI update)

> **Rule:** Use the most reliable tool for each check.
> Deterministic checks (build/test/lint) → run directly.
> Visual/integration checks → delegate with browser tools.

---

## 3. Definition of Done (MANDATORY GATES)

A task is only **DONE** when ALL gates pass. **Never claim done without running the self-check.**

### Completion Self-Check (Before EVERY "done"):

```
ASK YOURSELF before marking ANY task complete:
1. Did ALL verification gates below actually pass (not "I think they pass")?
2. Are there improvements I'm skipping?
3. Did I re-read the original task — does my work fully address it?
4. Are there next steps I should continue with or mention?
```

**If ANY answer is "no" → NOT done. Keep working.**

| Gate       | Check                               | Method                                                       |
| ---------- | ----------------------------------- | ------------------------------------------------------------ |
| **Build**  | All builds pass                     | Direct: `go build`, `npm run build`                          |
| **Test**   | All tests pass                      | Direct: `go test`, `npm run lint`                            |
| **LSP**    | No diagnostics errors               | Direct: `lsp_diagnostics`                                    |
| **Schema** | Backup executed if schema changed   | Direct: `python build.py backup`                             |
| **Docker** | Deployed if needed                  | Direct: `python build.py smart`                              |
| **Git**    | All files committed and pushed      | Delegate: task(category="quick", load_skills=["git-master"]) |
| **UI**     | Data displays correctly             | Delegate: Playwright verification                            |
| **Flow**   | Executor followed "Understand Flow" | Review: check executor traced the flow, not trial-and-error  |

### Stop Conditions (When to EXIT the Loop)

| Condition                                   | Action                               |
| ------------------------------------------- | ------------------------------------ |
| All TODO items completed + all gates passed | Exit with completion promise         |
| Same error 3+ times across attempts         | STOP — escalate to user              |
| Executor stuck in fix→test→fail loop        | STOP executor, invoke stuck-recovery |
| Blocking question requires user input       | STOP — ask user, resume after answer |
| Max iterations reached                      | Exit with progress summary           |

---

## 4. Session Continuity (CRITICAL in Loops)

Long-running loops MUST use `session_id` for delegation:

- **Every** delegation returns a `session_id` — STORE IT
- If task fails → retry with `session_id`, NOT a fresh delegation
- If follow-up needed → use `session_id` to preserve full context

```
❌ WRONG: Task failed → new delegation from scratch (loses context)
✅ CORRECT: Task failed → session_id="ses_xxx", prompt="Fix: [error]"
```

---

## 5. Failure Handling in Loops

<!-- MASTER:failure-counter -->
| Count | Action                                                                          |
| ----- | ------------------------------------------------------------------------------- |
| 1     | Fix directly, record error. Document what was tried.                            |
| 2     | **STOP.** TRACE FLOW activated. Research full chain before fix.                 |
| 3+    | **TOTAL STOP.** RESEARCH activated. Delegate @explore + @librarian in parallel. |
| 5+    | **STOP the task.** Report to orchestrator/user with full failure log.           |
<!-- /MASTER:failure-counter -->

---

## 6. Loop Iteration Checklist

At the START of each loop iteration:

```
1. Check: any background tasks completed? → Collect + verify results
2. Check: any slots free? → Delegate next task from queue
3. Check: any verification pending? → Run Evaluation 1 + 2
4. Check: all tasks done? → Run final gates → exit if passed
5. Check: stuck anywhere? → Apply stuck-recovery protocol
```

---

## 7. Anti-Patterns in Loop Mode

| Forbidden                                  | Do Instead                         |
| ------------------------------------------ | ---------------------------------- |
| Mark task done without verification        | Run both evaluations first         |
| Claim "done" without self-check            | Ask the 4 self-check questions     |
| Assume task is complete without re-reading | Re-read original task requirement  |
| Wait idle for all 3 delegations            | Process completions as they arrive |
| Skip git commit between batches            | Commit after each completed batch  |
| Ignore UI bugs found during verification   | Report per EXECUTOR_RULES.md §12   |
| Restart delegation from scratch on failure | Use session_id to continue         |
| Let loop spin > 5 failures on same issue   | STOP and escalate to user          |

---

**Version:** 1.0 | **Updated:** 2026-02-11
