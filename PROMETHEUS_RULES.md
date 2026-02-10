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

| #   | Step                        | Description                                         |
| --- | --------------------------- | --------------------------------------------------- |
| 1   | Read AGENTS.md              | Focus on Critical Rules & Architecture              |
| 2   | Identify files              | List ALL files to be modified                       |
| 3   | Check database              | Migration needed?                                   |
| 4   | Check multi-tenant          | tenant_id validation needed?                        |
| 5   | Line estimation             | ~300 per code file (quality signal, not hard limit) |
| 6   | Determine evidence          | Unit→Test, Integration→Docker/Test, Full→Both       |
| 7   | **IMPACT ANALYSIS**         | **MANDATORY - See section below**                   |
| 8   | **EXTERNAL RESEARCH NEEDS** | **MANDATORY - Identify required SDK/docs**          |

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
│ 3️⃣ EXTERNAL (ONLY IF STUCK - LAST RESORT)                              │
│    @librarian → Search official docs ONLY if not found locally           │
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

### 2.2 Research Phase in TODO LIST (MANDATORY)

Every plan involving platform integration MUST include:

```markdown
### TODO LIST

1. [ ] **[Phase 0] External Research** ⚠️ MANDATORY
   - [ ] 🔍 Search for official documentation for the API endpoints used
   - [ ] 🔍 Search for implementation examples on GitHub (grep.app / librarian)
   - [ ] 🔍 Verify request/response format from official docs
   - [ ] 🔍 Identify authentication flow (OAuth, API Key, etc.)
   - [ ] 🔍 Check rate limiting & error codes

2. [ ] **[Phase 1] Analysis** (existing)
       ...
```

### 2.3 When the Librarian Agent is MANDATORY

| Trigger                                    | Action Required                                       |
| ------------------------------------------ | ----------------------------------------------------- |
| Involves platform API (Shopee/Lazada/etc.) | `@librarian` - search official docs & examples        |
| Error from external API                    | `@librarian` - search error code meaning & solution   |
| Unclear request/response format            | `@librarian` - search official API spec               |
| OAuth/Authentication issues                | `@librarian` - search auth flow documentation         |
| Rate limiting/throttling                   | `@librarian` - search best practices & retry strategy |
| Unfamiliar Go library                      | `@librarian` - search usage examples on GitHub        |

### 2.4 Stuck Recovery Protocol (MANDATORY for Executor)

```
┌─────────────────────────────────────────────────────────────────────────┐
│ WHEN AI EXECUTOR IS STUCK (Repeated errors / No progress > 10 minutes)  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  STEP 1: STOP - Do not keep trying without references!                  │
│                                                                         │
│  STEP 2: IDENTIFY - Categorize the problem:                              │
│    □ External API error → Search in official docs                         │
│    □ Format mismatch → Search for implementation examples                │
│    □ Auth failed → Search for auth flow documentation                    │
│    □ Logic error → Search for existing patterns in codebase              │
│                                                                         │
│  STEP 3: RESEARCH - Use the right tools:                                 │
│    • @librarian → For external docs & OSS examples                       │
│    • @explore   → For existing patterns in this codebase                  │
│    • @oracle    → For architecture/design decisions                      │
│                                                                         │
│  STEP 4: IMPLEMENT - After obtaining clear references                    │
│                                                                         │
│  ⛔ ANTI-PATTERN:                                                        │
│    • Continued trial-and-error without reading docs                       │
│    • Blindly guessing request/response formats                           │
│    • Copy-pasting without understanding context                          │
│    • Bypassing auth/validation to "just try it"                          │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.5 TRACE FLOW BEFORE FIX Protocol (ANTI-LOOPING)

> **⚠️ CRITICAL:** AI often loops execution-testing without tracing the flow.
> This MUST be done BEFORE attempting any fix!

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 🔴 FORBIDDEN: Direct fix → test → fail → fix again → test → fail...     │
│ 🟢 MANDATORY: Trace Flow → Identify Root Cause → Targeted Fix           │
└─────────────────────────────────────────────────────────────────────────┘
```

#### A. Data Flow Tracing (MANDATORY before fix)

```
┌─────────────────────────────────────────────────────────────────────────┐
│ TRACE DATA FLOW - Follow the data journey from START to ERROR           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  1. FRONTEND → What is sent?                                            │
│     • Request body format                                               │
│     • Headers (Authorization, Content-Type)                             │
│     • URL params & query strings                                        │
│                                                                         │
│  2. BACKEND HANDLER → What is received?                                 │
│     • Request parsing successful?                                       │
│     • Validation passed?                                                │
│     • tenant_id exists?                                                 │
│                                                                         │
│  3. SERVICE LAYER → Logic running correctly?                            │
│     • Input to service correct?                                         │
│     • Business logic executed?                                          │
│     • External API call (if any) successful?                            │
│                                                                         │
│  4. REPOSITORY → Database operation correct?                             │
│     • Query executed?                                                   │
│     • Data returned?                                                    │
│     • Connection OK?                                                    │
│                                                                         │
│  5. RESPONSE → What is returned?                                        │
│     • Response format correct?                                          │
│     • Data matches expectations?                                        │
│     • Informative error message?                                        │
│                                                                         │
│  📍 IDENTIFY: At which layer is the data FIRST incorrect?               │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### B. Must Answer Before Fix

| Question                               | Must be Answered                          |
| -------------------------------------- | ----------------------------------------- |
| In which layer did the error occur?    | Handler / Service / Repository / External |
| What is the exact error message?       | Copy paste exact message                  |
| What data entered that layer?          | Log / debug print the result              |
| What data should have entered?         | Expected format from docs/spec            |
| Where did the difference first appear? | **THIS IS THE ROOT CAUSE**                |

#### C. Trace Flow in TODO LIST

```markdown
### TODO LIST

1. [ ] **[Phase 0.5] TRACE FLOW** ⚠️ BEFORE ANY FIX
   - [ ] Trace: What does the Frontend send? (check Network tab / curl)
   - [ ] Trace: What does the Handler receive? (add temporary logs)
   - [ ] Trace: What does the Service process? (log input/output)
   - [ ] Trace: What does the Repository query? (log SQL query)
   - [ ] Trace: What response is returned?
   - [ ] IDENTIFY: Which layer is incorrect first? → **[WRITE HERE]**

2. [ ] **[Phase 1] Fix Based on Trace**
   - [ ] Fix ONLY in the identified layer
   - [ ] Do not fix randomly across all layers
```

#### D. Trace Flow Example

##### ❌ INCORRECT (Direct Fix Without Trace)

```
Error: Order does not appear on frontend

Fix attempt 1: Change query in repository → FAILED
Fix attempt 2: Change response format in handler → FAILED
Fix attempt 3: Change parsing in frontend → FAILED
Fix attempt 4: Change database schema → FAILED
... (continuous looping)
```

##### ✅ CORRECT (Trace First, Fix Once)

```
Error: Order does not appear on frontend

TRACE FLOW:
1. Frontend request: GET /api/orders?tenant_id=xxx ✅
2. Handler receive: tenant_id = "xxx" ✅
3. Service call: GetOrders(ctx, "xxx") ✅
4. Repository query: SELECT * FROM orders WHERE tenant_id = ?
   → Result: 0 rows ❌ ← FIRST PROBLEM HERE
5. Check database: Data EXISTS but tenant_id = "yyy" not "xxx"

ROOT CAUSE: Stored tenant_id differs from the queried one
SOLUTION: Fix in one place - data ingestion (not in query/response)
```

#### E. FORBIDDEN Anti-Patterns

| ❌ Do Not Do                         | ✅ What Should be Done                     |
| ------------------------------------ | ------------------------------------------ |
| Directly edit code without trace     | Trace flow first, identify root cause      |
| Fix in all layers simultaneously     | Fix ONLY in the problematic layer          |
| Loop: fix → test → fail → fix → test | Trace → identify → precise fix → test ONCE |
| Guessing error location              | Follow data from start to error            |
| Delete error handling to "bypass"    | Fix actual cause, do not hide error        |

### 2.5 Research Requirements in Pre-Planning

Add to Pre-Planning Verification:

```markdown
### Pre-Planning Verification

- AGENTS.md has been read: [Yes/No]
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

### 2.6 Research Phase Example

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
- Identified files: [list files]
- Database changes: [Yes/No - if yes, migration required]
- Multi-tenant: [Yes/No - if yes, tenant_id validation required]
- Evidence type: [Unit/Integration/Full Feature]
- **Impact analysis has been performed: [Yes/No]**

### Impact Analysis of Changes

[Must be filled - see template in Section 2]

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
   - [ ] Review files > ~300 lines for SRP/DRY/OOP violations
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
- [ ] Files > ~300 lines reviewed for SRP/DRY/OOP

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

| #   | Gate                | Requirement                                 |
| --- | ------------------- | ------------------------------------------- |
| 1   | Todo List present   | Checklist format [ ] that can be executed   |
| 2   | **Impact Analysis** | **MANDATORY for every file modified**       |
| 3   | File Size           | ~300 lines quality signal for code files    |
| 4   | Architecture        | Handler → Service → Repository              |
| 5   | JSON Tags           | All snake_case                              |
| 6   | Testing Phase       | go build + go test present in todo          |
| 7   | Tenant Check        | Validate tenant_id if endpoint is protected |
| 8   | Cleanup Phase       | DRY, SRP review present in todo             |
| 9   | Evidence Type       | Mentioned in plan                           |

**If any gate FAILS → revise plan before execution.**

---

## 6. ANTI-PATTERNS (FORBIDDEN)

| #   | Do Not Do                               | Do                                                          |
| --- | --------------------------------------- | ----------------------------------------------------------- |
| 1   | Long prose/paragraph output             | Output TODO LIST with [ ]                                   |
| 2   | **Skip impact analysis**                | **MANDATORY impact analysis for every change**              |
| 3   | Skip file size review                   | Review SRP/DRY/OOP if file > ~300 lines                     |
| 4   | Business logic in Handler               | Direct to Service layer                                     |
| 5   | Skip tenant_id validation               | Always validate in protected endpoints                      |
| 6   | camelCase in JSON response              | Use snake_case                                              |
| 7   | Skip testing phase                      | MANDATORY go build + go test                                |
| 8   | Assume default tenant                   | Explicit error if missing                                   |
| 9   | Skip cleanup phase                      | MANDATORY DRY/SRP review                                    |
| 10  | Change API without checking frontend    | Check all consumers in frontend                             |
| 11  | Change DB schema without migration plan | Always include migration steps                              |
| 12  | Skip DB backup after schema changes     | MANDATORY: `python build.py backup` after migration applied |

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

## 8. CRITICAL RULES (from AGENTS.md)

These rules MUST NOT be violated:

1. **❌ NO FALSE POSITIVES** - Do not return `success: true` if there is an error
2. **❌ NO ALIASES** - Fix names directly, no workarounds
3. **❌ NO DEFAULT TENANT** - Always validate, error if missing
4. **🎯 JSON = snake_case** - All API responses
5. **📏 ~300 LINES QUALITY SIGNAL** - Review SRP/DRY/OOP if code file exceeds (docs/config: no limit)
6. **🔐 context.Context** - All DB/network operations
7. **🏗️ CLEAN ARCHITECTURE** - Handler → Service → Repository
8. **📝 STRUCTURED LOGGING** - zerolog only, not fmt.Printf
9. **🗄️ MIGRATIONS REQUIRED** - No raw DDL changes
10. **🧪 100% TEST SUCCESS** - go build && go test must pass
11. **🔐 GIT RESTRICTED** - Only add, commit, push (be careful)

---

## QUICK REFERENCE

```
Framework:    GIN (not Fiber)
Database:     PostgreSQL (multi-tenant schemas)
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
- Identified files: order_handler.go, order_service.go, export_utils.go
- Database changes: No
- Multi-tenant: Yes - tenant_id validation required
- Evidence type: Integration

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
   - [ ] Verify files > ~300 lines are clean (SRP/DRY/OOP)
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

## 9. DELEGATION RULES (SPEEDING UP EXECUTION)

> **Delegate to sub-agents for parallel processing and focused expertise.**
> **MAXIMUM 2 parallel delegations** to avoid overload.

### 8.1 When Delegation is MANDATORY

| Situation                            | Delegate To                                    | Reason                           |
| ------------------------------------ | ---------------------------------------------- | -------------------------------- |
| Search SDK docs / official API       | `@librarian`                                   | Expertise in external references |
| Search existing patterns in codebase | `@explore`                                     | Faster contextual grep           |
| Architecture/design question         | `@oracle`                                      | High-IQ reasoning                |
| UI/UX / Frontend work                | `delegate_task(category="visual-engineering")` | Frontend specialist              |
| Complex logic problem                | `delegate_task(category="ultrabrain")`         | Deep reasoning                   |
| Quick/trivial fix                    | `delegate_task(category="quick")`              | Fast execution                   |

### 8.2 Delegation Strategy

```
┌─────────────────────────────────────────────────────────────────────────┐
│ PARALLEL DELEGATION (Speeding Up Research)                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  When research is needed, fire 2 agents in PARALLEL:                    │
│                                                                         │
│  // Example: Fix Shopee API error                                       │
│  @librarian: "Search Shopee GetOrderList API docs, request/response"      │
│  @explore: "Search for existing shopee API pattern in codebase this"      │
│                                                                         │
│  → Both run in parallel, results merged for fix                         │
│                                                                         │
│  ⚠️ MAXIMUM 2 parallel to avoid overload                                 │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 8.3 Delegation in TODO LIST

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

### 8.4 Delegation Format

```markdown
## Delegation Request

**Agent:** @librarian / @explore / @oracle
**Task:** [specific question/search]
**Expected Output:** [what is expected]
**Context:** [relevant background info]
```

### 8.5 Effective Delegation Example

#### ❌ INCORRECT (No Delegation, All by Self)

```
Task: Fix Shopee order sync

*try to fix self*
*fail*
*try again*
*fail*
*try again*
... (wasting time without reference)
```

#### ✅ CORRECT (Delegate for Research)

```
Task: Fix Shopee order sync

## Fix Attempt #1
**Failure Count:** 1
*failed - error: invalid signature*

## Fix Attempt #2 - TRACE FLOW + DELEGATION
**Failure Count:** 2

**Parallel Delegation:**
🔀 @librarian: "Search for Shopee API signature generation docs,
               including parameter order and hash algorithm"
🔀 @explore: "Search for existing shopee signature generation in codebase"

**Results:**
- @librarian: Signature = SHA256(base_string + secret),
              base_string must be sorted by key
- @explore: File `internal/shopee/auth.go` line 45 has existing impl

**Root Cause:** Parameters not sorted before hash
**Fix:** Update signature generation according to docs

*test* → SUCCESS
```

### 8.6 Delegation Decision Tree

```
┌─────────────────────────────────────────────────────────────────────────┐
│ WHEN TO DELEGATE vs DO IT YOURSELF                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  DELEGATE if:                                                           │
│  ├─ Need external docs (SDK, API) → @librarian                         │
│  ├─ Need to find patterns in codebase → @explore                        │
│  ├─ Need architecture decision → @oracle                                │
│  ├─ Frontend/UI work → delegate_task(visual-engineering)                │
│  ├─ Failure >= 2 and need research → @librarian + @explore             │
│  └─ Task can be paralleled → fire 2 agents at once                      │
│                                                                         │
│  DO IT YOURSELF if:                                                    │
│  ├─ Simple clear edit                                                   │
│  ├─ Already have enough references                                      │
│  ├─ Trivial task (typo fix, formatting)                                 │
│  └─ Delegation overhead > benefit                                        │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 10. FAILURE COUNTER RULE (FOR EXECUTOR)

> **⚠️ THIS RULE MUST BE FOLLOWED BY AI EXECUTOR (Sisyphus/Builder)**
>
> AI often does not realize it is stuck/looping. This rule FORCES awareness.

### 9.1 Failure Definition

| Condition                          | Count |
| ---------------------------------- | ----- |
| `go build` failed after edit       | +1    |
| `go test` failed after fix         | +1    |
| Same error appears again after fix | +1    |
| API call fails with the same error | +1    |
| Fix does not resolve the problem   | +1    |

### 9.2 Mandatory Action by Failure Count

```
┌─────────────────────────────────────────────────────────────────────────┐
│ FAILURE COUNT → MANDATORY ACTION (CANNOT BE IGNORED)                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  FAILURE = 1 (First time failed)                                        │
│  → Allowed to try fix directly                                          │
│  → BUT record: "Failure #1: [error message]"                             │
│                                                                         │
│  FAILURE = 2 (Second time failed) ⚠️ TRACE FLOW ACTIVATED                │
│  → STOP! Do not fix directly again                                      │
│  → MANDATORY: Trace flow from frontend → database                       │
│  → MANDATORY: Identify which layer has the root cause                   │
│  → Write: "Failure #2 - TRACE FLOW protocol activated"                   │
│                                                                         │
│  FAILURE >= 3 (Failed 3x or more) 🚨 RESEARCH ACTIVATED                 │
│  → TOTAL STOP! No code editing without research                         │
│  → MANDATORY: @librarian to search for SDK docs / official API docs     │
│  → MANDATORY: @explore to search for existing patterns in the codebase   │
│  → MANDATORY: @oracle if architecture/design issue                       │
│  → Write: "Failure #3+ - RESEARCH protocol activated"                    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 9.3 Mandatory Format in Every Fix Attempt

```markdown
## Fix Attempt #[N]

**Failure Count:** [current count]
**Previous Error:** [previous error]
**Hypothesis:** [why this will succeed]
**Action:** [what will be done]

[If failure >= 2, MANDATORY to add:]
**Trace Flow Result:**

- Frontend sends: [what]
- Handler receives: [what]
- Service processes: [what]
- Repository queries: [what]
- Root cause identified: [in which layer]

[If failure >= 3, MANDATORY to add:]
**Research Result:**

- @librarian found: [research results]
- @explore found: [existing pattern]
- Reference: [link/source]
```

---

## 11. EXECUTOR QUICK REFERENCE (CHEAT SHEET)

> **Print this and follow every task execution**

```
┌─────────────────────────────────────────────────────────────────────────┐
│ EXECUTOR CHECKLIST - BEFORE STARTING                                    │
├─────────────────────────────────────────────────────────────────────────┤
│ [ ] Read AGENTS.md (Critical Rules section)                             │
│ [ ] Read PROMETHEUS_RULES.md (Section 2: Research & Trace Flow)         │
│ [ ] Identify task type: BUG FIX or NEW FEATURE?                         │
│     • Bug fix → Get ready to trace flow if failure >= 2                  │
│     • New feature with external API → Research first (delegate!)        │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ EXECUTOR CHECKLIST - DURING EXECUTION                                   │
├─────────────────────────────────────────────────────────────────────────┤
│ [ ] Track failure count (MANDATORY!)                                    │
│ [ ] Failure = 1 → Allowed to fix directly, RECORD error                  │
│ [ ] Failure = 2 → STOP, trace flow + delegate @explore                  │
│ [ ] Failure >= 3 → STOP, delegate @librarian + @explore (PARALLEL)      │
│ [ ] Do not loop fix-test without trace/research!                        │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ 🔴 LOCAL SDK PRIORITY (SEARCH HERE FIRST!)                              │
├─────────────────────────────────────────────────────────────────────────┤
│ Shopee  → backend/shopee-sdk/     (orders.go, products.go, client.go)   │
│ Lazada  → backend/lazada-sdk/     (order.go, product.go, auth.go)       │
│ Lazada  → backend/lazada_sdk/iop-sdk-go/  (Official IOP SDK)            │
│ TikTok  → backend/tiktok_sdk/     (100+ files, comprehensive!)          │
│                                                                         │
│ MANDATORY ORDER:                                                        │
│ 1️⃣ @explore: "Search in backend/*sdk*/" → LOCAL FIRST                   │
│ 2️⃣ @explore: "Search in internal/" → EXISTING PATTERN                  │
│ 3️⃣ @librarian: External docs → ONLY IF STUCK                            │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ DELEGATION RULES (SPEEDING UP)                                          │
├─────────────────────────────────────────────────────────────────────────┤
│ ⚠️ MAXIMUM 2 PARALLEL DELEGATIONS                                       │
│                                                                         │
│ @explore    → Existing pattern in codebase + LOCAL SDK                   │
│ @librarian  → External docs (LAST RESORT - ONLY IF NOT LOCAL)           │
│ @oracle     → Architecture decision                                      │
│                                                                         │
│ delegate_task(category="visual-engineering") → Frontend/UI work          │
│ delegate_task(category="ultrabrain") → Complex logic                     │
│ delegate_task(category="quick") → Trivial tasks                          │
│                                                                         │
│ PARALLEL EXAMPLE:                                                        │
│ 🔀 @explore: "Search for GetOrderList implementation in backend/shopee-sdk/"│
│ 🔀 @explore: "Search for existing shopee order pattern in internal/"      │
│ → Wait for result → Merge → Fix                                         │
│                                                                         │
│ ONLY IF NOT FOUND:                                                      │
│ 🔀 @librarian: "Search official Shopee API docs for error code X"        │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ EXECUTOR CHECKLIST - BEFORE FINISHING                                   │
├─────────────────────────────────────────────────────────────────────────┤
│ [ ] go build ./... passes                                               │
│ [ ] go test ./... passes                                                │
│ [ ] No looping (max 2 fix attempts without trace)                       │
│ [ ] Evidence according to task type has been collected                  │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ PLATFORM SDK REFERENCES                                                 │
├─────────────────────────────────────────────────────────────────────────┤
│ Shopee:    https://open.shopee.com/documents                            │
│ Lazada:    https://open.lazada.com/doc/api.htm                          │
│ TikTok:    https://partner.tiktokshop.com/doc                           │
│ Tokopedia: https://developer.tokopedia.com/                             │
└─────────────────────────────────────────────────────────────────────────┘
```

---

**File Version:** 3.0  
**Last Updated:** 2026-02-02  
**Status:** ACTIVE - MANDATORY COMPLIANCE
**Changes:** Added Delegation Rules, Failure Counter Rule, Trace Flow Protocol, Research Protocol, Executor Quick Reference
