---
description: ULW loop-mode rules - pipeline management, verification cadence, and Definition of Done for autonomous work loops
---

# ULW Loop Mode Rules

## Your Role in Loop Mode

<!-- MASTER:skill-ulw-loop-role -->
> **Purpose:** Governs Sisyphus behavior during autonomous loop execution (`/ulw-loop`).
> **For:** Sisyphus (main orchestrator) when running in loop mode.
> These rules COMPLEMENT (not replace) SISYPHUS_RULES.md and AGENTS.md.

You are the **Main Controller** and **Quality Lead**. Your job is to:

1. **Manage the delegation pipeline** — keep executors busy, slots filled
2. **Inspect and verify** every result — don't just collect reports
3. **Enforce Definition of Done** — no task is complete without passing all gates
<!-- /MASTER:skill-ulw-loop-role -->

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

<!-- MASTER:skill-ulw-loop-task-flow -->
### Task Flow

Tasks come from **either** a Prometheus plan or direct user request:

```
Task Source (Plan or User) → TODO Queue → Delegate → Verify → Next
                                              ↑                    |
                                              └────────────────────┘
                                              (if verification fails)
```

### Dynamic Planning

- When a batch finishes, update the master TODO immediately
- Planning updates happen in the background — don't let them block execution
- If scope changes mid-loop, update todos before continuing
<!-- /MASTER:skill-ulw-loop-task-flow -->

---

## 2. Task Classification

<!-- MASTER:skill-ulw-loop-task-classification -->
Tasks in loop mode come from a **Prometheus plan** or **direct user request**. Classify each task to determine delegation target and verification type:

| Task Type | Delegation Target | Verification Focus | Special Actions |
| --------- | ----------------- | ------------------ | --------------- |
| **Backend-only** (.go) | `category="implementation"` | Build + test + API curl | — |
| **Frontend-only** (.ts/.tsx) | `category="visual-engineering"` | Build + lint + multi-viewport screenshots | Load `react-frontend-rules` |
| **Full-stack** (backend + frontend) | Split into 2 delegations (backend first) | Both backend + frontend gates | Sequence: backend → frontend |
| **Config/Rules** (.json, .md) | `category="quick"` or do directly | Sync check + JSON validation | Run `sync_rules.py --check` if rules-master |
| **Schema change** (migrations) | `category="implementation"` | Migration + build + backup | MUST run `python build.py backup` AFTER |
| **Bug fix** | `category="deep"` (default) | Root cause evidence + test | Verify executor traced flow, not trial-and-error |

**Note on bug fixes:** Loop mode defaults to `deep` (not `implementation`) because bug fixes benefit from deeper root-cause analysis. For trivial/obvious bugs, the orchestrator may override to `category="quick"` or `category="implementation"`.

**Classification drives delegation.** Wrong classification → wrong executor model → suboptimal result.
<!-- /MASTER:skill-ulw-loop-task-classification -->

---

## 3. Delegation Strategy

<!-- MASTER:skill-ulw-loop-delegation-strategy -->
In loop mode, delegation is **batch-oriented**, not request-response.

### Decision Matrix

| Situation | Strategy |
| --------- | -------- |
| 3+ independent tasks (no dependencies) | Fire ALL in parallel — system queues excess |
| Task B depends on Task A output | Sequence: delegate A → verify → delegate B |
| Trivial task (typo, config tweak, <5 min) | Do directly — delegation overhead > benefit |
| Frontend depends on backend API | Backend FIRST → verify API works → then frontend |
| Same module, multiple changes | Batch into ONE delegation with clear scope |

### Prompt Discipline (MANDATORY)

Every delegation in loop mode MUST use the **6-section prompt structure**:

```
1. TASK: [atomic goal]
2. EXPECTED OUTCOME: [concrete deliverable]
3. REQUIRED TOOLS: [tool whitelist]
4. MUST DO: [exhaustive requirements]
5. MUST NOT DO: [forbidden actions]
6. CONTEXT: [file paths, patterns, task/todo number]
```

**Include task/todo number** (from plan or user request) in every delegation prompt so executors can reference it.

```
❌ WRONG: "Fix the order handler" (vague, no context)
✅ CORRECT: Full 6-section prompt with task/todo number, file paths, existing patterns
```
<!-- /MASTER:skill-ulw-loop-delegation-strategy -->

---

## 4. Todo Lifecycle

<!-- MASTER:skill-ulw-loop-todo-lifecycle -->
In loop mode, todos follow a source-driven lifecycle:

```
Task Source (Plan or User) → todowrite (all tasks) → in_progress (1 per slot) → completed/failed
```

### Rules

| Rule | Behavior |
| ---- | -------- |
| **Initialize** | Convert ALL tasks (from plan or user request) to todos at loop start via `todowrite` |
| **One active per slot** | Max 1 `in_progress` todo per delegation slot (3 max) |
| **Immediate completion** | Mark `completed` the MOMENT verification passes — never batch |
| **Failure handling** | Mark failed todo back to `pending` with retry note, increment failure counter |
| **Dynamic updates** | If executor discovers new work → add todo BEFORE continuing |
| **Scope changes** | If plan changes mid-loop → update todos BEFORE next delegation |
| **Evidence linking** | Each completed todo MUST reference its evidence (build output, screenshot path) |

**Never leave todos stale.** Stale todos = lost progress visibility.
<!-- /MASTER:skill-ulw-loop-todo-lifecycle -->

---

## 5. Verification Protocol (Double Evaluation)

<!-- MASTER:skill-ulw-loop-verification-protocol -->
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

### Evidence Validation (MANDATORY after both evaluations)

Before marking any task DONE, verify the executor provided real evidence — see **§6 Evidence Collection** for the full validation protocol. **Never accept "done" without evidence.**
<!-- /MASTER:skill-ulw-loop-verification-protocol -->

---

## 6. Evidence Collection

<!-- MASTER:skill-ulw-loop-evidence-collection -->
As orchestrator, you MUST **demand and validate** evidence from executors. Accepting results without evidence is FORBIDDEN.

### Required Evidence by Task Type

| Task Type | Required Evidence from Executor | How Orchestrator Validates |
| --------- | ------------------------------- | -------------------------- |
| **Backend** | `go build` exit 0 + `go test` pass + `lsp_diagnostics` clean | Re-run `go build ./...` + `go test ./...` + `lsp_diagnostics` on changed files |
| **Frontend** | `npm run build` exit 0 + `npm run lint` clean + screenshots | Re-run `npm run build` + `npm run lint` + `lsp_diagnostics` on changed files |
| **Full-stack** | Both backend + frontend evidence | Re-run all backend checks (`go build` + `go test` + `lsp_diagnostics`) + all frontend checks (`npm run build` + `npm run lint` + `lsp_diagnostics`) |
| **Schema** | Migration output + `python build.py backup` confirmation | Verify backup file timestamp |
| **Bug fix** | Root cause trace + before/after evidence + test proving fix | Check executor documented flow trace (not trial-and-error) |

### Validation Protocol

```
When executor reports "done":
1. CHECK: Did executor provide evidence in required format?
   → NO: Reject immediately — "Provide [missing evidence]"
2. CHECK: Is evidence real (not phantom)?
   → Re-run deterministic checks (build/test/lint/lsp) yourself
3. CHECK: Did executor follow "Understand Flow"?
   → Review their fix approach — if trial-and-error, REJECT
4. CHECK: Changed 3+ files?
   → Trigger Oracle ACC gate (see §8)
```

**Rubber-stamping executor results is the #1 loop-mode failure.** Always inspect.
<!-- /MASTER:skill-ulw-loop-evidence-collection -->

---

## 7. Definition of Done (MANDATORY GATES)

<!-- MASTER:skill-ulw-loop-definition-of-done -->
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
<!-- /MASTER:skill-ulw-loop-definition-of-done -->

<!-- MASTER:skill-ulw-loop-stop-conditions -->
### Stop Conditions (When to EXIT the Loop)

| Condition                                   | Action                                                   |
| ------------------------------------------- | -------------------------------------------------------- |
| All TODO items completed + all gates passed | Exit with completion promise                             |
| Same error 5+ times across attempts         | ASK USER — confirm: continue / skip / different approach |
| Executor stuck in fix→test→fail loop        | STOP executor, invoke stuck-recovery                     |
| Blocking question requires user input       | STOP — ask user, resume after answer                     |
| Max iterations reached                      | Exit with progress summary                               |
<!-- /MASTER:skill-ulw-loop-stop-conditions -->

---

## 8. Oracle ACC Gate in Loops

<!-- MASTER:skill-ulw-loop-oracle-gate -->
Oracle consultation in loop mode must balance **quality vs cost**.

### When to Invoke Oracle

| Scenario | Oracle Required? | Rationale |
| -------- | ---------------- | --------- |
| Single task, 1-2 files changed | NO | Self-evaluation sufficient |
| Single task, 3+ files changed | YES (per-task) | Significant change needs expert review |
| Batch of small tasks (all <3 files) | YES (per-batch, after git commit) | Batch review is cost-effective |
| Schema/migration change | YES (always) | High-risk change, regardless of file count |
| Architecture/refactor task | YES (always) | Structural decisions need validation |

### Oracle Consultation in Loop Context

```
task(
  subagent_type="oracle",
  load_skills=["oracle-rules"],
  session_id="[previous session if re-submitting]",
  prompt="
**LOOP MODE EVALUATION — Task N**

### Task Context: [plan task N or direct user request]
### Changes Made: [files changed + summary]
### Evidence: [build/test/lint results]
### Executor Approach: [traced flow? or trial-and-error?]

ACC or REJECT.
  "
)
```

### Fallback (Oracle unavailable)

1. Retry with `session_id`
2. If unavailable → use `category="deep"` with Oracle-style evaluation prompt
3. If deep also fails → thorough self-review + document Oracle was unavailable
4. **NEVER skip evaluation entirely** — at minimum, self-review with zero-suspicion gate
<!-- /MASTER:skill-ulw-loop-oracle-gate -->

---

## 9. Session Continuity (CRITICAL in Loops)

<!-- MASTER:skill-ulw-loop-session-continuity -->
Long-running loops MUST use `session_id` for delegation:

- **Every** delegation returns a `session_id` — STORE IT
- If task fails → retry with `session_id`, NOT a fresh delegation
- If follow-up needed → use `session_id` to preserve full context

```
❌ WRONG: Task failed → new delegation from scratch (loses context)
✅ CORRECT: Task failed → session_id="ses_xxx", prompt="Fix: [error]"
```
<!-- /MASTER:skill-ulw-loop-session-continuity -->

---

## 10. Failure Handling in Loops

<!-- MASTER:failure-counter -->
**Failure counter tracks SAME error/issue.** If a DIFFERENT error occurs, reset counter to 1.

| Count | Action                                                                                |
| ----- | ------------------------------------------------------------------------------------- |
| 1     | Fix directly, record error. Document what was tried.                                  |
| 2     | **STOP fixing.** FULL RESEARCH: trace flow + docs + SDK + references. Fix with evidence. |
| 3-4   | Continue fixing, but MUST use research from step 2. No guessing.                      |
| 5+    | **ASK USER.** Confirm: continue / skip / try different approach. Full failure log.    |

**Reset rule:** Different error = new counter starting at 1. Same error repeating = increment counter.
<!-- /MASTER:failure-counter -->

---

## 11. Loop Iteration Checklist

<!-- MASTER:skill-ulw-loop-iteration-checklist -->
At the START of each loop iteration:

```
1. Check: any background tasks completed? → Collect + verify results
2. Check: any slots free? → Delegate next task from queue
3. Check: any verification pending? → Run Evaluation 1 + 2
4. Check: all tasks done? → Run final gates → exit if passed
5. Check: stuck anywhere? → Apply stuck-recovery protocol
6. Check: UI bugs discovered during verification? → CRITICAL → add to queue immediately, WARNING → log for later batch
```
<!-- /MASTER:skill-ulw-loop-iteration-checklist -->

---

## 12. Anti-Patterns in Loop Mode

<!-- MASTER:skill-ulw-loop-anti-patterns -->
| Forbidden                                  | Do Instead                                       |
| ------------------------------------------ | ------------------------------------------------ |
| Mark task done without verification        | Run both evaluations first                       |
| Claim "done" without self-check            | Ask the 4 self-check questions                   |
| Assume task is complete without re-reading | Re-read original task requirement                |
| Wait idle for all 3 delegations            | Process completions as they arrive               |
| Skip git commit between batches            | Commit after each completed batch                |
| Ignore UI bugs found during verification   | CRITICAL → add to queue, WARNING → log for later |
| Restart delegation from scratch on failure | Use session_id to continue                       |
| Let loop spin > 5 failures on same issue   | STOP and escalate to user                        |
| Accept executor results without evidence   | Demand evidence in required format (see §6)      |
| Skip Oracle ACC for 3+ file changes       | Always invoke Oracle for significant changes     |
| Delegate without 6-section prompt          | Use full prompt structure for every delegation   |
| Rubber-stamp executor "done" reports       | Re-run deterministic checks yourself             |
<!-- /MASTER:skill-ulw-loop-anti-patterns -->
