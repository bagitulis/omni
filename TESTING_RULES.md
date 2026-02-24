# TESTING RULES — Testing Constitution

> **STATUS: MANDATORY** | **Version: 1.0** | **Updated: 2026-02-24**
>
> This file is the **IMMUTABLE TESTING CONSTITUTION** for ALL agents.
> These rules CANNOT be overridden. Test failures are BUGS, not test problems.

---

## 1. Zero-Tolerance Policy

Test failures have exactly 2 valid resolutions:

1. **Bug in code** → Fix the code. Never weaken the assertion.
2. **Misconfigured test setup** → Fix the test config/setup (not the assertion threshold).

### BANNED Actions (ZERO EXCEPTIONS)

| Banned Action                                                                    | Why                                                                     |
| -------------------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| Lowering assertion thresholds to pass a bug                                      | Hides the bug; defeats the purpose of testing                           |
| `test.Skip()` / `t.Skip()` without JIRA/ticket comment AND written justification | Silently hides failures                                                 |
| Changing a test assertion so it passes (without justification)                   | If test is wrong, revise WITH documented justification                  |
| Tests hitting external Shopee/Lazada/TikTok APIs directly                        | Must mock via `page.route()` — external APIs are flaky and rate-limited |
| Adding new tests for `frontend-vue/` (legacy)                                    | Legacy app is frozen — no new test investment                           |
| Rewriting existing solid backend unit tests                                      | If backend unit tests are passing, do NOT touch them                    |

---

## 2. Test Failure Protocol

When a test fails, follow this protocol in order:

1. **Read the error message carefully** — understand what assertion failed and why
2. **Reproduce manually** — confirm the failure is real, not a flaky environment issue
3. **Trace root cause** — identify the actual cause (code bug, config error, data issue)
4. **Fix the CAUSE, not the SYMPTOM** — never patch the test to hide the failure
5. **Verify fix does not break other tests** — run the full test suite after fixing

---

## 3. Coverage Baseline (Measured 2026-02-24)

### Backend Coverage (Go)

Measured via `go test -cover ./...` in `backend/`.
Full report: `backend/tests/coverage/baseline-report.txt`

| Package                                    | Coverage |
| ------------------------------------------ | -------- |
| `internal/utils/http`                      | 90.9%    |
| `internal/utils`                           | 90.3%    |
| `internal/utils/logger`                    | 75.0%    |
| `internal/services/cache`                  | 85.5%    |
| `internal/models`                          | 63.3%    |
| `internal/services/route`                  | 52.3%    |
| `internal/services/analytics/intelligence` | 37.9%    |
| `internal/handlers`                        | 36.0%    |
| `internal/middleware`                      | 29.0%    |
| `internal/services/ads`                    | 26.9%    |
| Most `services/*` packages                 | 0.0%     |
| Most `pkg/*` packages                      | 0.0%     |

**Overall backend average: ~15–20%** (many packages have 0% coverage)

### Frontend Coverage (React/TypeScript)

Measured via `npx vitest run --coverage` in `frontend/`.
Full report: `frontend/coverage/baseline-report.txt`
Raw per-file data: `frontend/coverage/coverage-final.json`

| Metric     | Coverage                |
| ---------- | ----------------------- |
| Statements | 27.0% (11,258 / 41,655) |
| Functions  | 36.8% (299 / 812)       |
| Branches   | 70.2% (1,526 / 2,173)   |

Test results: **44 files pass, 3 files fail** (pre-existing failures, not caused by this plan).

### Coverage Targets by Phase

| Phase        | Backend (statements) | Frontend (statements) | Timeline   |
| ------------ | -------------------- | --------------------- | ---------- |
| **Baseline** | ~15–20%              | 27%                   | 2026-02-24 |
| **Phase 1**  | 27% (baseline +10pp) | 37% (baseline +10pp)  | +3 months  |
| **Phase 2**  | 42% (Phase 1 +15pp)  | 52% (Phase 1 +15pp)   | +6 months  |
| **Phase 3**  | 100% (long-term)     | 100% (long-term)      | Long-term  |

> **Rule**: Do NOT set phase targets ABOVE the measured baseline — targets must be achievable.
> Phase 1 = baseline + ~10pp. Phase 2 = Phase 1 + ~15pp. Phase 3 = 100% long-term goal.

---

## 4. Forbidden Patterns

The following patterns are BANNED in ALL test files (backend and frontend):

| Pattern                                                          | Reason                                                 |
| ---------------------------------------------------------------- | ------------------------------------------------------ |
| `test.Skip()` / `t.Skip()` without justification + ticket number | Silent test suppression                                |
| `xfail` or `xit` without justification                           | Silent test suppression                                |
| `//nolint` without explanation                                   | Hides lint violations silently                         |
| `as any` or `@ts-ignore` in test files                           | Type safety is required in tests too                   |
| Hardcoded credentials (username/password) in test files          | Security risk — use env vars instead                   |
| Tests that depend on state from previous tests                   | Must be isolated — each test must set up its own state |
| Test files touching `frontend-vue/` (legacy)                     | Legacy app is frozen — no new tests                    |

---

## 5. Lighthouse Budget Immutability

Lighthouse performance budgets are IMMUTABLE — they can never be lowered to make a failing test pass.

### Rules

- Lighthouse score thresholds are **READ-ONLY** after initial setup — never lower them
- If a Lighthouse test fails: **fix the code/page** to meet the threshold
- Adding new routes to Lighthouse tests: **OK**
- Lowering existing thresholds: **BANNED**

### Default Thresholds

| Category       | Minimum Score |
| -------------- | ------------- |
| Performance    | 90            |
| Accessibility  | 90            |
| Best Practices | 90            |
| SEO            | 60            |

Thresholds are defined in `backend/tests/e2e/lighthouse/config.go` — treat as read-only after initial setup.

---

## 6. Test Naming Conventions

Consistent naming makes tests searchable and self-documenting.

| Framework      | Convention                                    | Example                                        |
| -------------- | --------------------------------------------- | ---------------------------------------------- |
| Go (`testing`) | `TestFunctionName_Condition_ExpectedBehavior` | `TestCreateOrder_DuplicateSKU_ReturnsConflict` |
| React/Vitest   | `'does X when Y'` or `'renders X correctly'`  | `'renders empty state when no products'`       |
| Playwright E2E | `'user can do X'` or `'page shows Y when Z'`  | `'user can filter products by platform'`       |

### Test File Naming

| Language       | Convention                                         |
| -------------- | -------------------------------------------------- |
| Go             | `*_test.go` (same package as code under test)      |
| TypeScript     | `*.test.ts` / `*.spec.ts` (co-located with source) |
| Playwright E2E | `*.spec.ts` under `frontend/e2e/`                  |

---

## 7. Test Tiers

| Tier        | Scope                       | Max Duration | Framework                         |
| ----------- | --------------------------- | ------------ | --------------------------------- |
| Unit        | Isolated function/component | < 1s         | Go `testing` + `testify` / Vitest |
| Integration | Service + Repository        | < 10s        | Go `testing` + Docker DB          |
| E2E         | Full stack via browser      | < 30s        | Playwright                        |
| Lighthouse  | Performance + A11y + SEO    | < 60s        | Go runner + Chromium              |

### Tier Rules

- **Unit tests**: No network, no DB, no filesystem. Pure logic only.
- **Integration tests**: May use a test DB (Docker). No external API calls.
- **E2E tests**: Full browser automation. Must mock external APIs via `page.route()`.
- **Lighthouse tests**: Must run against a fully running frontend (`http://localhost:5174`).

---

## 8. Skipped Routes List

These parameterized routes are explicitly excluded from automated Lighthouse/E2E testing until data seeding infrastructure is in place.

| Route                      | Reason                                             | Status       |
| -------------------------- | -------------------------------------------------- | ------------ |
| `/products/:id/edit`       | Requires valid product ID (data seeding needed)    | ⏳ Follow-up |
| `/products/import`         | Requires file upload fixture                       | ⏳ Follow-up |
| `/order-manager/:platform` | Requires valid platform slug (data seeding needed) | ⏳ Follow-up |

---

## 9. Known Issues Found During Planning

These issues were identified during the testing overhaul planning phase and must be tracked to resolution.

| File                                     | Bug                                                                   | Severity | Fix Task        |
| ---------------------------------------- | --------------------------------------------------------------------- | -------- | --------------- |
| `backend/tests/e2e/lighthouse/routes.go` | Routes mismatch vs App.tsx — hits routes that don't exist             | HIGH     | Task 3, Task 11 |
| `backend/tests/e2e/lighthouse/config.go` | `FrontendURL = http://localhost:80` should be `http://localhost:5174` | HIGH     | Task 11         |
| `backend/tests/e2e/lighthouse/config.go` | Hardcoded credentials (`yumna/password123`) — should use env vars     | MEDIUM   | Task 11         |
| `frontend/e2e/f3-qa-replay.spec.ts` etc. | 3 old E2E specs have syntax errors                                    | HIGH     | Task 2 (delete) |

---

## 10. Bugs Found During Coverage Work

> Discovered while writing tests to increase coverage. Auto-logged by agents.
> **Status values**: `AUTO-FIXED` | `NEEDS-CONFIRMATION` | `DEFERRED`

| Date | File / Package | Bug Type | Description | Severity | Status |
| ---- | -------------- | -------- | ----------- | -------- | ------ |

---
_This document is part of the OMNI project testing infrastructure. All agents MUST follow these rules without exception._
