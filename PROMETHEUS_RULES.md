# PROMETHEUS PLANNING RULES

> **STATUS: MANDATORY**
>
> This file contains the planning rules for AI Prometheus.
> For detailed implementation guidelines, code patterns, and architecture reference, see **AGENTS.md**.

---

## COMPLETION BLOCK (Required at Plan Start)

Before submitting any plan, Prometheus MUST include this block:

```
[PROMETHEUS PRE-PLANNING COMPLETED]
✅ Read AGENTS.md - Critical Rules section
✅ Read AGENTS.md - Architecture Patterns section
✅ Identified all affected files
✅ Multi-tenancy check completed
✅ Evidence type identified for task
```

---

## 1. PRE-PLANNING CHECKLIST

| Step | Check | Notes |
|------|-------|-------|
| 1 | Read AGENTS.md completely | Focus on Critical Rules & Architecture |
| 2 | Identify ALL files to modify | List every file in plan |
| 3 | Database changes? | If yes → migration required |
| 4 | Multi-tenant involved? | If yes → tenant_id validation required |
| 5 | Estimate line counts | Max 300 per file (models: 500) |
| 6 | Determine evidence type | Unit→Test, Integration→Docker/Test, Full→Both |

---

## 2. PLAN TEMPLATE

```markdown
## Task: [Name]

[PROMETHEUS PRE-PLANNING COMPLETED]
✅ Read AGENTS.md - Critical Rules section
✅ Read AGENTS.md - Architecture Patterns section
✅ Identified all affected files
✅ Multi-tenancy check completed
✅ Evidence type identified: [Unit/Integration/Full Feature]

### Affected Files
| File | Estimated Lines | Status |
|------|-----------------|--------|
| [path] | ~X lines | ✅ OK/<300 |

### Implementation Phases
1. **Phase 1: Analysis** - [tasks]
2. **Phase 2: Implementation** - [tasks]
3. **Phase 3: Cleanup** - DRY, SRP, <300 lines
4. **Phase 4: Testing** - go build && go test
5. **Phase 5: Docker** - build.py smart (if needed)

### Success Criteria
- [ ] go build ./... passes
- [ ] go test ./... passes
- [ ] Evidence collected (appropriate for task type)
```

---

## 3. QUALITY GATES

Plan is VALID only if ALL gates pass:

| # | Gate | Requirement |
|---|------|-------------|
| 1 | File Size | All files < 300 lines (models: 500) |
| 2 | Architecture | Handler → Service → Repository |
| 3 | JSON Tags | All snake_case |
| 4 | Testing | go build + go test included |
| 5 | Tenant | No default tenant, explicit validation |
| 6 | Cleanup | DRY, SRP, OOP phase included |
| 7 | Success Criteria | Clearly defined |
| 8 | Evidence Type | Specified (matches task complexity) |

**If ANY gate fails → revise plan before execution.**

---

## 4. ANTI-PATTERNS (FORBIDDEN)

| # | Don't | Do Instead |
|---|-------|------------|
| 1 | Plan without file size limit | Specify "max 300 lines" |
| 2 | Business logic in Handler | Put in Service layer |
| 3 | Skip tenant_id validation | Validate all protected endpoints |
| 4 | camelCase JSON response | Use snake_case |
| 5 | Skip testing phase | Always include go build + go test |
| 6 | Assume default tenant | Explicit error if missing |
| 7 | Skip cleanup phase | Always include DRY/SRP review |

---

## 5. EVIDENCE REQUIREMENTS

| Task Type | Required Evidence |
|-----------|-------------------|
| Unit Test / Code Only | Test output only |
| Integration / API | Docker log OR Test output |
| Full Feature | Docker log AND Test output |
| Documentation | Visual confirmation |

**Evidence must directly prove the specific problem was fixed.**

---

## 6. CRITICAL RULES REFERENCE

These rules from AGENTS.md CANNOT be violated:

1. **❌ NO FALSE POSITIVES** - Never return `success: true` with errors
2. **❌ NO ALIASES** - Fix names directly, no workarounds
3. **❌ NO DEFAULT TENANT** - Always validate, error if missing
4. **🎯 JSON = snake_case** - All API responses
5. **📏 MAX 300 LINES** - Per file (models: 500, migrations: unlimited)
6. **🔐 context.Context** - All DB/network operations
7. **🏗️ CLEAN ARCHITECTURE** - Handler → Service → Repository
8. **📝 STRUCTURED LOGGING** - zerolog only, no fmt.Printf
9. **🗄️ MIGRATIONS REQUIRED** - No raw DDL changes
10. **🧪 100% TEST SUCCESS** - go build && go test must pass
11. **🔐 GIT RESTRICTED** - Only add, commit, push (carefully)

---

## QUICK REFERENCE

```
Framework:    GIN (not Fiber)
Database:     PostgreSQL (multi-tenant schemas)
ORM:          GORM
Logging:      zerolog
JSON:         snake_case
Architecture: Handler → Service → Repository
Testing:      go build ./... && go test ./...
Docker:       build.py smart (default)
```

---

**File Version:** 2.0  
**Last Updated:** 2026-02-02  
**Status:** ACTIVE - MANDATORY COMPLIANCE
