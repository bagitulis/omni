# PROMETHEUS PLANNING RULES

> **STATUS: MANDATORY**
>
> Prometheus is the AI planner. Primary output = **TODO LIST for execution**.
> For implementation details, code patterns, and architecture, see **AGENTS.md**.
> You may delegate to subagents (@explore, @librarian) for research, but Prometheus must review results.

---

## PROMETHEUS OBJECTIVES

**Input:** User request
**Output:** TODO LIST ready for execution by Sisyphus/Builder

Prometheus DOES NOT perform tasks. Prometheus CREATES A PLAN in the form of a TODO LIST.

---

## 1. PRE-PLANNING (Before Creating a Plan)

Before creating a plan, Prometheus MUST:

| #   | Step                        | Description                                                         |
| --- | --------------------------- | ------------------------------------------------------------------- |
| 1   | Read AGENTS.md              | Focus on Critical Rules & Architecture                              |
| 2   | **FLOW MAP**                | **MANDATORY - Map the data flow (see §1.1)**                        |
| 3   | Identify files              | List ALL files to be modified                                       |
| 4   | Check database              | Migration needed?                                                   |
| 5   | Check multi-tenant          | tenant_id validation needed?                                        |
| 6   | Line estimation             | ~300 per code file (MUST refactor if exceeded, 400+ NOT acceptable) |
| 7   | Determine evidence          | Unit→Test, Integration→Docker/Test, Full→Both                       |
| 8   | **IMPACT ANALYSIS**         | **MANDATORY - See section below**                                   |
| 9   | **EXTERNAL RESEARCH NEEDS** | **MANDATORY - Identify required SDK/docs**                          |

### 1.1 FLOW MAP (MANDATORY)

> **⚠️ CRITICAL:** Every plan MUST begin with a Flow Map that shows the data path for the feature/bug.
> If the flow is unknown, incomplete, or messy → fixing/creating it becomes the FIRST task.

#### Flow Map Template (MUST be filled in every plan)

```markdown
### Flow Map

**Entry Point:** [UI route/page | webhook | cron job | external trigger]

**Frontend Path:**
Page/Component → State/Store → API Client → Endpoint Called

**Backend Path:**
Router → Handler → Service → Repository → DB Tables/Queries

**Data Contracts:**

- Request: [JSON keys, snake_case] → [validation rules]
- Response: [JSON keys, snake_case] → [error shapes]

**External Dependencies:**

- SDK/API: [platform + local SDK path, e.g., backend/shopee-sdk/orders.go]
- Rate limits / retries: [if relevant]

**Breakpoints (top 3 likely failure points):**

1. [layer + what could fail + how to observe]
2. [layer + what could fail + how to observe]
3. [layer + what could fail + how to observe]
```

#### Flow Map Rules

| Rule                         | Description                                                                                 |
| ---------------------------- | ------------------------------------------------------------------------------------------- |
| **Every plan has one**       | No exceptions — even trivial tasks have a flow                                              |
| **Topic-aware**              | Flow path depends on what you're working on (see table below)                               |
| **Executor references flow** | Each executor task must state which flow node it modifies                                   |
| **Missing flow = fix first** | If flow can't be written → add Flow Repair tasks BEFORE feature tasks                       |
| **Broken flow = fix first**  | If flow violates architecture (biz logic in handler, direct DB in handler) → refactor first |

#### Topic-Specific Flow Paths

| Topic                    | Relevant Flow                                            |
| ------------------------ | -------------------------------------------------------- |
| **Platform integration** | Platform API → SDK → Handler → Service → Repository → DB |
| **Database/Schema**      | Migration → Schema → Repository → Service → Handler      |
| **Backend API**          | Handler → Service → Repository → DB → Response           |
| **Frontend**             | Component → API call → Response → State → Render         |
| **Full-stack feature**   | DB schema → Repo → Service → Handler → API → Frontend UI |

#### Flow Repair (When Flow is Missing or Messy)

If Prometheus cannot write the Flow Map because:

- The flow **doesn't exist yet** → Add "Create Flow" tasks first
- The flow **violates architecture** (e.g., business logic in handler, direct DB calls from handler) → Add refactoring tasks first
- The flow **is inconsistent** (e.g., some endpoints use service layer, others skip it) → Add alignment tasks first

```markdown
### TODO LIST

0. [ ] **[Phase -1] Flow Repair** ⚠️ BEFORE FEATURE WORK
   - [ ] Identify architecture violations in current flow
   - [ ] Move business logic from Handler → Service
   - [ ] Add missing Service/Repository layers
   - [ ] Verify flow follows: Handler → Service → Repository → DB
   - [ ] Update Flow Map after repair

1. [ ] **[Phase 1] Implementation** (uses repaired flow)
       ...
```

---

## 2. EXTERNAL REFERENCE & RESEARCH PROTOCOL (MANDATORY)

> **⚠️ CRITICAL:** AI executors OFTEN fail because they do not look for correct references.
>
> Prometheus MUST include research requirements in every plan involving external APIs/SDKs.

### 2.1 LOCAL SDK References (FIRST PRIORITY) 🔴

> **⚠️ CRITICAL:** AI MUST search in local SDK folders FIRST before looking for external docs!
> Local SDKs already have complete implementations - DO NOT skip!

#### 📁 Local SDK Folders (SEARCH HERE FIRST!)

| Platform   | Local Path                       | Contents                            |
| ---------- | -------------------------------- | ----------------------------------- |
| **Shopee** | `backend/shopee-sdk/`            | client, orders, products, logistics |
| **Lazada** | `backend/lazada-sdk/`            | client, order, product, auth        |
| **Lazada** | `backend/lazada_sdk/iop-sdk-go/` | Official IOP SDK                    |
| **TikTok** | `backend/tiktok_sdk/`            | Official SDK (100+ files)           |

#### 🔍 Research Priority Order (THIS ORDER IS MANDATORY!)

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 1️⃣ LOCAL FIRST (HIGHEST PRIORITY)                                       │
│    @explore → Search in local SDK folders: backend/*sdk*/                │
│    Example: "search for GetOrderList implementation in backend/shopee-sdk/"│
├─────────────────────────────────────────────────────────────────────────┤
│ 2️⃣ CODEBASE (SECONDARY)                                                 │
│    @explore → Search for existing implementation in internal/            │
│    Example: "search for how shopee orders are saved to database"         │
├─────────────────────────────────────────────────────────────────────────┤
│ 3️⃣ EXTERNAL (WHEN LOCAL IS INSUFFICIENT)                                │
│    @librarian → Search official docs if local SDK doesn't cover it       │
│    Use cases: error codes, API changes, auth flow clarification,         │
│    rate limits, new endpoints not yet in local SDK                       │
│    Example: "search official docs for error code XXXX"                   │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 📖 External Docs (Backup Reference)

| Platform      | Official Documentation              | When to Use              |
| ------------- | ----------------------------------- | ------------------------ |
| **Shopee**    | https://open.shopee.com/documents   | Error codes, API changes |
| **Lazada**    | https://open.lazada.com/doc/api.htm | Error codes, API changes |
| **TikTok**    | https://partner.tiktokshop.com/doc  | Error codes, API changes |
| **Tokopedia** | https://developer.tokopedia.com/    | Error codes, API changes |

### 2.2 Research Phase in TODO LIST (MANDATORY for Platform Integrations)

Every plan involving platform integration MUST include a research phase.
**Research priority: local SDK → existing codebase → external docs (when local is insufficient).**

```markdown
### TODO LIST

1. [ ] **[Phase 0] External Research** ⚠️ MANDATORY (LOCAL SDK → CODEBASE → EXTERNAL)
   - [ ] 🔍 Search local SDK for existing implementation (`@explore` → `backend/*sdk*/`)
   - [ ] 🔍 Search existing patterns in codebase (@explore → internal/)
   - [ ] 🔍 Verify request/response format from local SDK or official docs
   - [ ] 🔍 Identify authentication flow (OAuth, API Key, etc.)
   - [ ] 🔍 ONLY if local insufficient → Search official docs & GitHub examples (@librarian)

2. [ ] **[Phase 1] Analysis** (existing)
       ...
```

### 2.3 When the Librarian Agent is REQUIRED

> **Priority**: Always search local SDK (`@explore`) first. Use `@librarian` only when local sources don't cover the need.

| Trigger                                              | Action Required                                       |
| ---------------------------------------------------- | ----------------------------------------------------- |
| Local SDK doesn't cover the platform API need        | `@librarian` - search official docs & examples        |
| Error from external API (code not in local SDK)      | `@librarian` - search error code meaning & solution   |
| Unclear request/response format after checking local | `@librarian` - search official API spec               |
| OAuth/Authentication issues not covered locally      | `@librarian` - search auth flow documentation         |
| Rate limiting/throttling (not documented locally)    | `@librarian` - search best practices & retry strategy |
| Unfamiliar Go library (no local examples)            | `@librarian` - search usage examples on GitHub        |

### 2.4 Stuck Recovery Protocol

> Executor stuck-recovery details are in **EXECUTOR_RULES.md §6-§7** and the `stuck-recovery` skill.
> Prometheus should include research phases in plans so executors have references BEFORE they get stuck.

### 2.5 TRACE FLOW — Plan Directive for Bug Fixes

> **For bug fix tasks, Prometheus MUST include a Trace Flow phase in the plan.**
> Executors follow detailed trace flow protocol in **EXECUTOR_RULES.md §6**.

#### Include in Bug Fix Plans:

```markdown
### TODO LIST

1. [ ] **[Phase 0.5] TRACE FLOW** ⚠️ BEFORE ANY FIX
   - [ ] Identify which layers are involved in the bug
   - [ ] Trace data flow through each relevant layer
   - [ ] IDENTIFY: Which layer is incorrect first? → **[WRITE HERE]**

2. [ ] **[Phase 1] Fix Based on Trace**
   - [ ] Fix ONLY in the identified layer
   - [ ] Do not fix randomly across all layers
```

#### Pre-Planning Verification (Bug Fix):

| Question                                | Must Answer                        |
| --------------------------------------- | ---------------------------------- |
| Is this a bug fix?                      | If yes → Trace Flow phase REQUIRED |
| Which layers are potentially involved?  | List them in the plan              |
| Does executor need research references? | Include Phase 0 delegation if yes  |

### 2.6 Research Requirements in Pre-Planning

Add to Pre-Planning Verification:

```markdown
### Pre-Planning Verification

- AGENTS.md has been read: [Yes/No]
- **Flow Map completed: [Yes/No — if No, add Flow Repair phase]**
- Identified files: [list files]
- Database changes: [Yes/No]
- Multi-tenant: [Yes/No]
- Evidence type: [Unit/Integration/Full Feature]
- Impact analysis has been performed: [Yes/No]
- **External Research Required: [Yes/No]**
  - Platform APIs: [Shopee/Lazada/TikTok/etc - specify]
  - Reference docs: [link to official docs]
  - Research agent: [@librarian/@explore - specify which will be used]
- **Trace Flow Required: [Yes/No]**
  - If BUG FIX → MANDATORY Yes, trace flow before fix
  - If NEW FEATURE → Not mandatory, but recommended
```

### 2.7 Research Phase Example

#### ❌ INCORRECT (Skip Research)

```
## Task: Fix Shopee Order Sync

### TODO LIST
1. [ ] Debug order_sync.go
2. [ ] Fix error
3. [ ] Test
```

#### ✅ CORRECT (With Research Phase)

```
## Task: Fix Shopee Order Sync

### Pre-Planning Verification
- External Research Required: **Yes**
  - Platform APIs: Shopee Order API
  - LOCAL SDK: backend/shopee-sdk/orders.go (SEARCH HERE FIRST!)
  - Research agent: @explore for LOCAL SDK + existing pattern

### TODO LIST

1. [ ] **[Phase 0] Research** ⚠️ MANDATORY (MANDATORY ORDER!)
   - [ ] @explore: Search for GetOrderList in backend/shopee-sdk/orders.go (LOCAL FIRST!)
   - [ ] @explore: Search for existing shopee order pattern in internal/
   - [ ] Verify auth flow from existing implementation
   - [ ] ONLY if not found → @librarian: Search official docs

2. [ ] **[Phase 1] Analysis**
   - [ ] Read error message carefully
   - [ ] Compare with implementation in local SDK
   - [ ] Identify mismatch

3. [ ] **[Phase 2] Fix Based on Research**
   - [ ] Update request format according to local SDK pattern
   - [ ] Update response parsing according to actual format
   - [ ] Add proper error handling for Shopee error codes

4. [ ] **[Phase 3] Testing**
   - [ ] go build && go test
   - [ ] Verify data enters database
```

---

## 3. IMPACT ANALYSIS OF CHANGES (MANDATORY)

> **⚠️ VERY IMPORTANT:** Every change MUST have its impact analyzed comprehensively before execution.

### 3.1 Impact Analysis Checklist

For EVERY file to be modified, Prometheus MUST analyze:

#### A. Impact on Backend

| Question                            | Must be Answered                         |
| ----------------------------------- | ---------------------------------------- |
| Which function/method is changing?  | List all functions                       |
| Who calls this function?            | Search for all callers (use grep/LSP)    |
| Does the function signature change? | If yes, all callers must be updated      |
| Does the return type change?        | If yes, all consumers must be updated    |
| Is any interface affected?          | If yes, all implementors must be updated |

#### B. Impact on Frontend

| Question                           | Must be Answered                             |
| ---------------------------------- | -------------------------------------------- |
| Which API endpoint is changing?    | List all endpoints                           |
| Which component consumes this API? | Search for all components that fetch         |
| Does the response format change?   | If yes, all consumers must be updated        |
| Do props/state change?             | If yes, parent/child components are affected |
| Does any shared component change?  | If yes, all user components are affected     |

#### C. Impact on Database

| Question                                | Must be Answered          |
| --------------------------------------- | ------------------------- |
| Which table is changing?                | List all tables           |
| Which column is added/changed/deleted?  | Change details            |
| Is any foreign key affected?            | Check relations           |
| Is any index in need of update?         | Performance consideration |
| Does existing data need to be migrated? | Data migration plan       |

#### D. Integration Impact

| Question                          | Must be Answered                       |
| --------------------------------- | -------------------------------------- |
| API contract changed?             | Frontend must sync                     |
| Is it a breaking change?          | If yes, there must be a migration path |
| Need to update API documentation? | Swagger/OpenAPI                        |
| Is any other service affected?    | Microservice dependencies              |

### Impact Analysis Template (MANDATORY in Plan)

```markdown
### Impact Analysis of Changes

#### File: `[path/to/file]`

- **Changes:** [brief description]
- **Affected Caller/Consumer:**
  - `file1.go` - function X calls the modified function
  - `Component.vue` - consumes the modified API
- **Breaking change:** [Yes/No]
- **Action required:**
  - [ ] Update caller in file1.go
  - [ ] Update Component.vue to handle new response
```

### Impact Analysis Example

#### ❌ INCORRECT (Without Impact Analysis)

```
## Task: Change order response format

### TODO LIST
1. [ ] Change response in order_handler.go
2. [ ] Done
```

#### ✅ CORRECT (With Impact Analysis)

```
## Task: Change order response format

### Impact Analysis of Changes

#### File: `internal/handlers/order_handler.go`
- **Changes:** Change field `orderSn` to `order_sn` (snake_case)
- **Affected Caller/Consumer:**
  - `frontend/src/api/order.ts` - parsing response
  - `frontend/src/views/OrderList.vue` - display in table
  - `frontend/src/views/OrderDetail.vue` - display detail
- **Breaking change:** Yes - frontend expects `orderSn`
- **Action required:**
  - [ ] Update order.ts interface
  - [ ] Update OrderList.vue template binding
  - [ ] Update OrderDetail.vue template binding
  - [ ] Verify table has no empty columns after changes

### TODO LIST
1. [ ] **[Phase 1] Analysis**
   - [ ] Grep all usages of `orderSn` in frontend
   - [ ] List all affected components

2. [ ] **[Phase 2] Backend**
   - [ ] Change response format in order_handler.go

3. [ ] **[Phase 3] Frontend**
   - [ ] Update interface in order.ts
   - [ ] Update OrderList.vue
   - [ ] Update OrderDetail.vue

4. [ ] **[Phase 4] Testing**
   - [ ] Verify data appears in table (no empty columns)
   - [ ] go build && go test
```

---

## 4. OUTPUT FORMAT: TODO LIST

**MANDATORY:** Prometheus must output in a TODO LIST format that can be directly executed.

### Template Todo List

```markdown
## Task: [Task Name]

### Pre-Planning Verification

- AGENTS.md has been read: [Yes/No]
- **Flow Map completed: [Yes/No — if No, add Flow Repair phase]**
- Identified files: [list files]
- Database changes: [Yes/No - if yes, migration required]
- Multi-tenant: [Yes/No - if yes, tenant_id validation required]
- Evidence type: [Unit/Integration/Full Feature]
- **Impact analysis has been performed: [Yes/No]**

### Flow Map

[MANDATORY - see §1.1 for template]

### Impact Analysis of Changes

[Must be filled - see template in Section 3]

### TODO LIST

1. [ ] **[Phase 1] Analysis**
   - [ ] Read file X to understand existing structure
   - [ ] Identify the pattern used
   - [ ] Check dependencies

2. [ ] **[Phase 2] Implementation**
   - [ ] Create/edit file: `path/to/file.go` (~X lines)
   - [ ] Implement function X in service layer
   - [ ] Implement handler Y
   - [ ] ...etc

3. [ ] **[Phase 3] Cleanup**
   - [ ] Review files > ~300 lines for SRP/DRY/OOP violations (400+ NOT acceptable)
   - [ ] Remove duplicate code
   - [ ] Remove dead code
   - [ ] Apply DRY & SRP

4. [ ] **[Phase 4] Testing**
   - [ ] Backend: go build ./... + go test ./...
   - [ ] Frontend (if changed): npm run build + npm run lint
   - [ ] Fix if there are errors

5. [ ] **[Phase 5] Finalization**
   - [ ] Collect evidence according to task type
   - [ ] Apply Docker if needed: `python build.py smart`
   - [ ] If schema changed: `python build.py backup` (AFTER migration applied)

### Affected Files

| File              | Estimated Lines | Action        |
| ----------------- | --------------- | ------------- |
| `path/to/file.go` | ~150 lines      | Create/Modify |
| ...               | ...             | ...           |

### Success Criteria

#### Build & Test

- [ ] Backend: go build ./... passes + go test ./... passes
- [ ] Frontend (if changed): npm run build passes
- [ ] Docker: `python build.py smart` succeeds (if deploying)
- [ ] Schema changes: `python build.py backup` executed AFTER migration applied
- [ ] Files > ~300 lines reviewed for SRP/DRY/OOP (400+ NOT acceptable)

#### Code Quality (according to AGENTS.md)

- [ ] Format code according to AGENTS.md (snake_case JSON, architecture pattern)
- [ ] No duplicate/dead code
- [ ] No false positives (success: true only for success)

#### Frontend (if there are frontend changes)

- [ ] UI/UX layout not messy (Prometheus verifies by reviewing code structure — executors verify via Playwright and MUST report any UI bugs found)
- [ ] Component structure neat and reusable
- [ ] Responsive design maintained

#### Integration (Backend + Frontend + Database)

- [ ] Data appears in frontend table (no empty columns)
- [ ] API response according to format (snake_case)
- [ ] Database query returns correct data

#### Evidence

- [ ] Docker log shows successful operation with specific data
- [ ] Test output shows all tests PASS
```

---

## 5. QUALITY GATES

Plan is VALID only if ALL gates are met:

| #   | Gate                | Requirement                                           |
| --- | ------------------- | ----------------------------------------------------- |
| 1   | Todo List present   | Checklist format [ ] that can be executed             |
| 2   | **Flow Map**        | **MANDATORY - data path declared end-to-end**         |
| 3   | **Impact Analysis** | **MANDATORY for every file modified**                 |
| 4   | File Size           | ~300 lines MUST trigger refactor, 400+ NOT acceptable |
| 5   | Architecture        | Handler → Service → Repository                        |
| 6   | JSON Tags           | All snake_case                                        |
| 7   | Testing Phase       | go build + go test present in todo                    |
| 8   | Tenant Check        | Validate tenant_id if endpoint is protected           |
| 9   | Cleanup Phase       | DRY, SRP review present in todo                       |
| 10  | Evidence Type       | Mentioned in plan                                     |

**If any gate FAILS → revise plan before execution.**

---

## 6. ANTI-PATTERNS (FORBIDDEN)

| #   | Do Not Do                               | Do                                                                 |
| --- | --------------------------------------- | ------------------------------------------------------------------ |
| 1   | Long prose/paragraph output             | Output TODO LIST with [ ]                                          |
| 2   | **Skip Flow Map**                       | **MANDATORY flow map for every plan**                              |
| 3   | **Skip impact analysis**                | **MANDATORY impact analysis for every change**                     |
| 4   | Skip file size review                   | MUST review SRP/DRY/OOP if file > ~300 lines (400+ NOT acceptable) |
| 5   | Business logic in Handler               | Direct to Service layer                                            |
| 6   | Skip tenant_id validation               | Always validate in protected endpoints                             |
| 7   | camelCase in JSON response              | Use snake_case                                                     |
| 8   | Skip testing phase                      | MANDATORY go build + go test                                       |
| 9   | Assume default tenant                   | Explicit error if missing                                          |
| 10  | Skip cleanup phase                      | MANDATORY DRY/SRP review                                           |
| 11  | Change API without checking frontend    | Check all consumers in frontend                                    |
| 12  | Change DB schema without migration plan | Always include migration steps                                     |
| 13  | Skip DB backup after schema changes     | MANDATORY: `python build.py backup` after migration applied        |
| 14  | Plan on broken/missing flow             | Fix/create flow BEFORE feature work                                |

---

## 7. EVIDENCE REQUIREMENTS

| Task Type             | Required Evidence          |
| --------------------- | -------------------------- |
| Unit Test / Code Only | Test output only           |
| Integration / API     | Docker log OR Test output  |
| Full Feature          | Docker log AND Test output |
| Documentation         | Visual confirmation        |

**Evidence must prove that the specific problem has been resolved.**

---

## 8. CRITICAL RULES

> **Canonical source: AGENTS.md §1-§5.** Do NOT duplicate rules here — follow AGENTS.md directly.
>
> Key invariants: English-only, no false positives, no default tenant, snake_case JSON,
> ~300 lines MUST trigger refactor (400+ NOT acceptable), Handler→Service→Repository architecture,
> git add/commit/push only (no destructive commands).

### Test Policy

- Run `go build ./...` + `go test ./...` (backend) and `npm run build` + `npm run lint` (frontend, if changed)
- **No test file exists for changed code** → MUST create test file with meaningful tests covering the changes
- **Pre-existing test failures**: Document them explicitly, prove your changes did NOT introduce or worsen them
- New test failures from your changes → MUST fix before task is complete

---

## QUICK REFERENCE

```
Framework:    GIN (not Fiber)
Database:     SQLite (dev) / PostgreSQL (prod, multi-tenant)
ORM:          GORM
Logging:      zerolog
JSON:         snake_case
Architecture: Handler → Service → Repository
Testing:      go build ./... && go test ./... (backend)
              npm run build && npm run lint (frontend)
Docker:       python build.py smart (default)
DB Backup:    python build.py backup (AFTER schema changes)
```

---

## PROMETHEUS OUTPUT EXAMPLE

### ❌ INCORRECT (Prose/paragraph)

```
For this task, we need to do several things. First, we will
read existing files. Then we implement new features. After that
we do testing and cleanup...
```

### ✅ CORRECT (TODO List)

```markdown
## Task: Add Order Export Feature

### Pre-Planning Verification

- AGENTS.md has been read: Yes
- **Flow Map completed: Yes**
- Identified files: order_handler.go, order_service.go, export_utils.go
- Database changes: No
- Multi-tenant: Yes - tenant_id validation required
- Evidence type: Integration

### Flow Map

**Entry Point:** UI → Order List page → "Export" button click

**Frontend Path:**
OrderList.tsx → exportOrders() → API client → GET /api/orders/export

**Backend Path:**
Router → OrderHandler.HandleExportOrders → OrderService.ExportOrders → OrderRepository.GetOrders → orders table

**Data Contracts:**

- Request: `GET /api/orders/export?tenant_id=xxx&format=csv`
- Response: CSV file download (Content-Type: text/csv)

**External Dependencies:** None

**Breakpoints:**

1. Handler: tenant_id validation might be missing
2. Service: Large dataset could cause memory issues
3. Repository: Query might not filter by date range

### TODO LIST

1. [ ] **[Phase 1] Analysis**
   - [ ] Read `internal/handlers/order_handler.go`
   - [ ] Read `internal/services/order_service.go`
   - [ ] Identify existing export patterns

2. [ ] **[Phase 2] Implementation**
   - [ ] Add `ExportOrders` method in `order_service.go` (~50 lines)
   - [ ] Add `HandleExportOrders` handler in `order_handler.go` (~30 lines)
   - [ ] Create `internal/utils/export_utils.go` (~100 lines)

3. [ ] **[Phase 3] Cleanup**
   - [ ] Verify files > ~300 lines are clean (SRP/DRY/OOP, 400+ NOT acceptable)
   - [ ] Apply DRY - extract common export logic

4. [ ] **[Phase 4] Testing**
   - [ ] Run: go build ./...
   - [ ] Run: go test ./...

5. [ ] **[Phase 5] Finalization**
   - [ ] Test endpoint via curl/Postman
   - [ ] Collect Docker log as evidence

### Affected Files

| File                                 | Estimation | Action |
| ------------------------------------ | ---------- | ------ |
| `internal/handlers/order_handler.go` | +30 lines  | Modify |
| `internal/services/order_service.go` | +50 lines  | Modify |
| `internal/utils/export_utils.go`     | ~100 lines | Create |

### Success Criteria

#### Build & Test

- [ ] go build ./... passes
- [ ] go test ./... passes

#### Code Quality

- [ ] Format according to AGENTS.md (snake_case JSON)
- [ ] No duplicate/dead code

#### Integration

- [ ] Export endpoint returns valid CSV/Excel
- [ ] Data appears complete (no empty columns)
- [ ] Docker log shows export successful with record count
```

---

## 9. DELEGATION GUIDELINES FOR PLANS

> **Prometheus plans SHOULD include delegation directives when appropriate.**
> **MAXIMUM 3 parallel delegations** to balance throughput and system load.
> For detailed executor rules (failure counter, trace flow, SDK priority), see **EXECUTOR_RULES.md**.

### 9.1 Delegation Quick Reference

| Situation                            | Delegate To                                    |
| ------------------------------------ | ---------------------------------------------- |
| Search SDK docs / official API       | `@librarian`                                   |
| Search existing patterns in codebase | `@explore`                                     |
| Architecture/design question         | `@oracle`                                      |
| UI/UX / Frontend work                | `delegate_task(category="visual-engineering")` |
| Complex logic problem                | `delegate_task(category="ultrabrain")`         |
| Quick/trivial fix                    | `delegate_task(category="quick")`              |

### 9.2 Delegation in TODO LIST

When plans require research or parallel work, include delegation phases:

```markdown
### TODO LIST

1. [ ] **[Phase 0] Parallel Research** ⚠️ DELEGATE
   - [ ] 🔀 PARALLEL #1: @librarian → Search [specific docs]
   - [ ] 🔀 PARALLEL #2: @explore → Search [existing pattern]
   - [ ] ⏳ Wait for results, merge findings

2. [ ] **[Phase 0.5] Trace Flow** (if bug fix)
   - [ ] Trace based on research results
3. [ ] **[Phase 1] Implementation**
   - [ ] Fix based on research + trace
```

### 9.3 Delegation Format

```markdown
## Delegation Request

**Agent:** @librarian / @explore / @oracle
**Task:** [specific question/search]
**Expected Output:** [what is expected]
**Context:** [relevant background info]
```

> **Note:** Executors follow their own failure counter and trace flow protocols.
> See **EXECUTOR_RULES.md §6-7** for details. Prometheus does NOT need to duplicate those rules.

---

**File Version:** 4.0  
**Last Updated:** 2026-02-11  
**Status:** ACTIVE - MANDATORY COMPLIANCE
**Changes:** v4.0 - Slimmed down executor-facing content (§9-11 → §9 only), moved to EXECUTOR_RULES.md. Fixed MAXIMUM 3 parallel delegations.
