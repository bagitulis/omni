# Realtime, E2E Guardrails, and Platform API Drift — Design

**Status:** Approved (Phase 0 in execution)
**Date:** 2026-09-15
**Author:** Droid (session)
**Scope:** 6-phase remediation covering E2E test suite health, WebSocket realtime layer, Lazada PII masking (breaking, live 2026-07-01), TikTok API version drift, and CI hardening.

Prerequisites for every change in this spec: TDD (RED → GREEN → REFACTOR) with negative and error paths, dry-run before destructive ops, CodeGraph `explore`/`impact` before touching any symbol, evidence saved under `.sisyphus/evidence/`.

---

## 1. Motivation

### 1.1 E2E test suite (audit summary — full report in session log)

- 18 Playwright specs; 24 tests in 3 specs are `test.skip` placeholders (`credential-lifecycle`, `booking-integration`, `developer-context`).
- `report-escrow.spec.ts:9` defaults `BASE_URL` to `http://localhost:5176` while `playwright.config.ts:6` sets `baseURL: "http://localhost:5174"`; without the env, every navigation is against a dead port.
- `frontend/e2e/helpers/db-reset.ts` is a misnomer: its own comment says "Does NOT reset database". Ten spec files import it under the DB-reset name.
- Five specs contain fallback selectors that silently short-circuit (`if (count === 0) return;`) so the test can never fail on the flow it claims to cover (orders export/print, marketplace sync button, marketplace 401 catch-all, route-mapping refresh, script-monitor forms).
- `frontend/playwright.config.ts` has no `webServer`, no `retries`, no `globalSetup` — tests assume `:5174` is already up.
- `.github/workflows/ci.yml`: **Playwright is never invoked**. `test-frontend` runs `npm run lint` + `npm test` (Vitest). `runtime-strict` runs Go E2E only.
- Backend Go E2E: `backend/tests/e2e/lighthouse/test-results/lighthouse-report.md` shows every metric as `0 / ❌ Fail` (stale 2026-02-04, runner broken and unnoticed).
- Feature coverage gaps: Extensions module (3 pages shipped 2026-09-14), User Management, Product Edit, Product Import — no E2E specs.

### 1.2 WebSocket surface (audit summary)

- `backend/internal/extensions/` ships a battle-tested WebSocket hub, but it is **purpose-built for extension RPC** (keyed by `extensionID`, per-key in-flight cap, result-channel correlation). `BroadcastToDashboards` is an explicit no-op (`hub.go:658`).
- Frontend has **zero** WebSocket usage; live updates today are TanStack Query polls + one SSE endpoint (`/api/notifications/stream` with ticket auth).
- Nginx `tunnel.conf:115-144` handles Upgrade correctly on `/api` but `proxy_read_timeout 60s` is too tight for dashboard sockets.
- Extensions hub's own docstring states the split: extension traffic goes through it, dashboards fan out through SSE. Extending the hub to per-tenant broadcast would break its RPC contract.
- Decision: **new package `backend/internal/realtime/`** that reuses the hub's event-loop pattern (single goroutine, buffered channels, sweeper) but is JWT-authed, keyed by `tenantID`, and topic-subscribe. SSE stays until the FE client is proven.

### 1.3 Platform API drift (audit summary)

- **Lazada — CRITICAL breaking, live 2026-07-01**: buyer PII is masked in `GET /orders/get` responses (buyer name, phone, address). Codebase renders these fields raw. Exception: `Delivery-by-Seller (DBS)` application per seller. Reference: `open.lazada.com/apps/announcement/detail?docId=145548`; independent write-up: Zetpy 2026-06-26.
- **Shopee**: `v2.video.edit_video_info` / `v2.video.get_video_detail` gained mandatory `aigc_label` on 2026-09-03 (Announcement 1554). Codebase does not use Video API — safe. Internal partner-key rotation deadline `29/05/2026` (`_bmad-output/planning-artifacts/prd.md:152`) has passed; needs verification.
- **TikTok**: legacy V1 already deprecated (codebase on `/api/v2/token/*` for auth and versioned `/202309/`, `/202501/`, `/202502/` for shop APIs). `GET /product/202309/products/search` (`pkg/tiktok/api.go:87`) still called; `POST /product/202502/products/search` (`pkg/tiktok/product.go:248`) also called; check if 202309 search is deprecated. `Get Payments API` — official migration 202309 → 202605 announced Aug 2026 (codebase doesn't use `/payments/`, safe for that endpoint). `CategoryVersion: "v1"` hardcoded for ID in `pkg/tiktok/category.go:63` — check whether TikTok pushed v2 for Indonesia. New **Inventory Update Webhook** available (Nov 2025) — not consumed.
- **Google**: `google.golang.org/api v0.260.0`, Sheets v4, `drive.readonly` — no drift; `drive.readonly` remains a restricted scope requiring app verification (policy, not API).

---

## 2. Roadmap (6 phases)

Each phase is merge-able independently. Non-goals in every phase: no stack replacement, no arch refactor, no unrelated cleanup.

| Phase | Name | Blast radius | Est. |
|---|---|---|---|
| 0 | E2E Guardrails | `frontend/playwright.config.ts`, `frontend/e2e/**`, `.github/workflows/ci.yml` | 1 session |
| 1 | Lazada PII masking (breaking, live 2026-07-01) | `backend/pkg/lazada/`, `internal/services/lazada/`, `internal/handlers/lazada/`, `frontend/src/pages/orders/`, DB migration for `unmasked_available` column | 1-2 sessions |
| 2 | Unskip 24 placeholder E2E + remove escape hatches | `frontend/e2e/*.spec.ts` (5 files); no production code | 1 session |
| 3 | New E2E specs for 7 gap pages (Extensions/Users/Product Edit/Import) | `frontend/e2e/*.spec.ts` (7 new files); no production code | 1 session |
| 4 | Realtime hub v1: `internal/realtime/` + `useRealtime` hook + notifications-over-WS | new backend package, new frontend hook, nginx timeout | 2 sessions |
| 5 | Realtime hub v2: topics `orders/updated`, `inventory/updated`, `sync/progress`, `chat/hint` + publisher wiring | `services/sync/`, `services/inventory/`, `services/orders/`, `services/autofunction/`, FE stores | 2-3 sessions |
| 6 | TikTok API drift audit + Inventory Update Webhook adoption + CategoryVersion verification | `backend/pkg/tiktok/`, `services/webhooks/tiktok_processor.go`, `services/tiktok/` | 1-2 sessions |

Dependency notes: Phase 0 unblocks Phases 2–3 verification. Phase 1 depends on Phase 0's CI. Phases 4–5 depend on Phase 0 for E2E validation and are independent of the Lazada/TikTok phases. Phase 6 can run in parallel with any of 4/5.

---

## 3. Phase 0 — E2E Guardrails (in execution)

### 3.1 Goal

Every subsequent E2E fix is verifiable in CI on every PR. `playwright test` must be able to run without a human pre-booting the frontend and without silently masking flakes.

### 3.2 Tasks

- **0.1** Update `frontend/playwright.config.ts`:
  - Add `webServer` (dev-only via `!process.env.CI`; CI still expects Docker-served `:5174` from the `runtime-strict` compose stack).
  - Add `retries: process.env.CI ? 2 : 0`.
  - Keep single `chromium` project.
- **0.2** Rename `frontend/e2e/helpers/db-reset.ts` → `frontend/e2e/helpers/clearBrowserState.ts`; rename exported `resetTestState` → `clearBrowserState` and `resetAndNavigate` → `clearAndNavigate` (keeps the "clear state" contract in the name); update 10 caller specs.
- **0.3** Fix `frontend/e2e/report-escrow.spec.ts:9`: drop `BASE_URL || "http://localhost:5176"`; use the Playwright config baseURL via `test.use({ baseURL })` or the built-in relative `page.goto("/login")` pattern (rest of suite already uses relative paths). Removes the port-5176 landmine and the split default.
- **0.4** Convert 5 escape hatches to `test.fixme` with issue references so they surface as skipped-with-reason instead of green false positives:
  - `orders.spec.ts` export/print fallback
  - `marketplace.spec.ts` sync-button fallback
  - `marketplace.spec.ts` `.catch(() => {})` 401 assertion swallow
  - `route-mapping.spec.ts` refresh-button fallback
  - `script-monitor.spec.ts` form/list fallback
- **0.5** Add non-blocking `test-e2e` job to `.github/workflows/ci.yml`:
  - Triggers on push to main and manual dispatch (same as `runtime-strict`).
  - `needs: [runtime-strict]` so the compose stack is already up (backend `:3000`, frontend `:5174`).
  - Installs Chromium (`npx playwright install --with-deps chromium`), runs `npx playwright test --reporter=list,html`, uploads `frontend/playwright-report/` on failure.
  - `continue-on-error: true` (non-blocking during stabilisation per user pref; convert to required in Phase 4).

### 3.3 TDD strategy

Config and CI changes are setup-heavy; TDD is applied differently than for behaviour code. The RED evidence is proof that the current state fails the requirement; GREEN is proof it now meets it. All evidence saved to `.sisyphus/evidence/phase0-*.txt`.

| Change | RED evidence | GREEN evidence |
|---|---|---|
| `webServer` config | Grep `webServer` in current `playwright.config.ts` returns nothing | Grep returns the new block; `npx playwright test --list` from a clean shell resolves specs without pre-existing dev server |
| `retries` config | Grep `retries` in current config returns nothing | Grep returns `retries: process.env.CI ? 2 : 0` |
| Helper rename | `rg 'db-reset' frontend/` shows 10 imports; deleting the file breaks build (dry-run in a scratch worktree, not committed) | `rg 'db-reset' frontend/` returns nothing; `rg 'clearBrowserState' frontend/` returns 10+1 hits (source + callers); `npm run build` and `npx playwright test --list` both green |
| `report-escrow.spec.ts` BASE_URL | Grep shows `http://localhost:5176` default and `BASE_URL + "/…"` concatenation | Grep for `5176` returns nothing; the file uses relative `page.goto("/report/shopee")` |
| Escape hatches → `test.fixme` | Grep `if \(.+count === 0\) return;` in `frontend/e2e/` returns 5 matches | Grep returns 0 matches; `rg 'test\.fixme' frontend/e2e/` shows the 5 replacements with issue refs |
| CI job | Grep `test-e2e` in `.github/workflows/ci.yml` returns nothing | Grep shows the job definition; job runs on Draft PR |

### 3.4 Dry run (non-destructive but sensitive)

Fase 0 touches config, tests, and CI — no DB/tenant/credential ops. Dry-run checks:

- Before edit: `git status` + `git diff --stat` (baseline clean).
- Before rename: run `rg 'db-reset|resetTestState|resetAndNavigate' frontend/` and confirm every match is one we plan to touch (no library or docs reference).
- After edits: `git diff --stat` shows only planned files.
- Before commit: `cd frontend; npm run build && npm run lint && npx playwright test --list`.

No production data touched. No commits until every gate is green.

### 3.5 Definition of done (Phase 0)

- [ ] `docs/superpowers/specs/2026-09-15-realtime-e2e-platform-drift-design.md` (this file) committed.
- [ ] `frontend/playwright.config.ts` has `webServer` (dev-only) and `retries: process.env.CI ? 2 : 0`.
- [ ] `frontend/e2e/helpers/clearBrowserState.ts` exists; `db-reset.ts` gone; 10 caller specs updated.
- [ ] `frontend/e2e/report-escrow.spec.ts` uses config baseURL, no `:5176` anywhere.
- [ ] 5 escape hatches replaced with `test.fixme` + issue reference.
- [ ] `.github/workflows/ci.yml` has non-blocking `test-e2e` job needing `runtime-strict`.
- [ ] `.sisyphus/evidence/phase0-*.txt` captures RED and GREEN evidence.
- [ ] `cd frontend; npm run build && npm run lint && npx playwright test --list` all green locally.
- [ ] One commit (or one per user request), pushed.

---

## 4. Phase 1 — Lazada PII masking (outline)

Detail spec drafted at Phase 1 kickoff; approach summary:

- Introduce typed `MaskedField[T]` in `pkg/lazada/types.go` capturing `{ value T; masked bool; unmasked_available bool }`.
- `pkg/lazada/api.go GetOrders` parses new masked payloads (Lazada returns strings like `J***s` and `+62812***56`) and populates the masked flag.
- Repository stores the raw payload; service layer surfaces `MaskedField` to the handler; handler returns snake_case JSON with `masked: true` and `unmasked_available: false`.
- Frontend `types/order.ts` gains `masked` fields; `frontend/src/pages/orders/OrderDetailDrawer.tsx` renders masked strings verbatim and shows a badge + link "Apply for Delivery-by-Seller unmasking" when `unmasked_available === false`.
- DB migration: add `lazada_masking_status jsonb` column to `orders` (or side table if migrations prefer). Auto-backfill on next sync.
- TDD: fixtures with real Lazada masked payload; assert Go parses, service passes through, handler returns snake_case, frontend renders as expected.
- Dry-run: sync one tenant's orders with `--dry-run` flag (add flag to sync CLI) → report which orders would flip from unmasked → masked without writing.

Test coverage: happy path (all masked), partial mask (some fields masked, others not on DBS orders), malformed asterisk pattern, backward compat with pre-mask cached orders.

---

## 5. Phase 2 — Unskip placeholder E2E + remove escape hatches (outline)

- Implement bodies for the 24 `test.skip` tests across `credential-lifecycle.spec.ts`, `booking-integration.spec.ts`, `developer-context.spec.ts` following the pattern established by `notifications.spec.ts` (mocked API via `page.route`, no live backend deps).
- Remove the 5 escape hatches now that the referenced controls actually exist (or convert to real assertions if the flow is genuinely conditional).
- Fix hardcoded English copy in `dashboard.spec.ts`, `settings.spec.ts` (use `data-testid` where the DOM supports it; add testids where it doesn't — minimal, additive).

TDD: each implementation gets its `test.fail()` variant asserting the mock is honoured (RED → GREEN pattern for E2E is confirming the mocked route was hit and the DOM changed as expected).

---

## 6. Phase 3 — New E2E specs (outline)

- `extensions.spec.ts`, `extensions-scraper.spec.ts`, `extensions-results.spec.ts` — smoke + pairing UI + mocked WS handshake (uses `route` handlers to fake the pairing API).
- `users.spec.ts` — user list + role gate + CRUD forms.
- `product-edit.spec.ts` — edit form, price/stock modals, save flow.
- `product-import.spec.ts` — CSV upload, validation errors, progress display.

Use `clearBrowserState` (renamed helper) and the mocking pattern from `notifications.spec.ts`. No live backend.

---

## 7. Phase 4 — Realtime hub v1 (outline)

### 7.1 Package layout (`backend/internal/realtime/`)

- `hub.go` — single-goroutine event loop. Keyed by `(tenantID, clientID)` where `clientID` is a uuid per WS connection.
- `client.go` — per-client state: send channel (buffered, cap 64), subscribed topics (`map[string]struct{}`), tenantID.
- `topic.go` — topic registry (`map[topic]map[clientID]*client`). Publish walks the inner map and non-blocking sends to each client (drop with metric on full send buffer; do not stall the loop).
- `conn.go` — gorilla upgrader with dashboard-origin allowlist from `main.go` (populated by env `FRONTEND_ORIGINS`). Handshake reads first frame `{type:"auth", action:"connect", payload:{token:"<jwt>"}}`.
- `auth.go` — pluggable `Authenticate func(token string) (tenantID, userID string, err error)`. Wired to the existing JWT service.
- `envelope.go` — `WSMessage { type, action, topic, id, tenant_id, sequence, payload }`.
- `paths.go` — `const RealtimePath = "/api/realtime/ws"`.
- `errors.go` — typed errors matching extensions hub style.
- `test_helpers_test.go`, `hub_test.go`, `conn_test.go`, `topic_test.go`, `auth_test.go`, `isolation_test.go`, `mv3_spike_equivalent_test.go` (browser tab churn), `zz_gap_probe_test.go`.

### 7.2 Frontend

- `frontend/src/lib/realtime.ts` — connection state machine (`disconnected` → `connecting` → `authenticating` → `connected` → `reconnecting`), exponential backoff `1s → 30s`.
- `frontend/src/hooks/useRealtime.ts` — subscribe hook: `useRealtime("orders/updated", (msg) => …)`.
- `frontend/src/stores/realtimeStore.ts` (Zustand) — singleton connection.
- Migrate `NotificationContext.tsx` to consume WS `notifications/updated` topic; keep SSE as fallback for 1 sprint, delete in Phase 5.

### 7.3 Nginx

- `nginx/conf.d/tunnel.conf`: introduce a `location = /api/realtime/ws { … proxy_read_timeout 3600s; }` block; keep the general `/api` block as-is.

### 7.4 TDD

Every hub behaviour has a `RED` test:
- publish before subscribe → drop
- publish with no subscribers → no panic
- subscriber leaves mid-publish → no leak (goroutine, channel)
- send buffer full → drop with metric, connection stays open
- auth failure closes connection with code 4401
- cross-tenant isolation: publish to tenant A never reaches tenant B
- reconnect with `since` sequence → replay (Phase 5 adds replay; Phase 4 rejects with `unsupported` code so FE knows to full-refresh)

Dry run: not applicable (net-new package, no destructive ops), but a canary rollout plan: FE feature-flag `VITE_REALTIME_ENABLED`, default off, enable per-tenant via `global_config`.

---

## 8. Phase 5 — Realtime hub v2 + publisher wiring (outline)

- Publisher API: `realtime.Publish(ctx, tenantID, topic string, payload any)` — safe to call from any service; enqueues to hub via a buffered `publishCh`.
- Wire into:
  - `services/sync/order_sync_service.go` — after each successful order upsert → publish `orders/updated` with `{order_sn, platform, changed_fields}`.
  - `services/inventory/stock_orchestrator.go` — after successful stock update → publish `inventory/updated`.
  - `services/autofunction/executor.go` — periodic `sync/progress` with `{job_id, percent, current, total}`.
  - `services/webhooks/*_processor.go` — after webhook persists, publish equivalent topic (dedupe on FE via `id + sequence`).
  - `services/chat/*` (future) — `chat/hint` for typing / new-message indicator.
- Sequence numbers per (tenant, topic) — in-memory monotonic counter with periodic checkpoint to Redis; on reconnect FE sends `since: N` and hub replays from a ring buffer (default 512 entries per topic).
- FE stores subscribe and invalidate TanStack Query caches instead of adding parallel state:
  - `orders/updated` → `queryClient.invalidateQueries({ queryKey: ["orders"] })`.
  - `inventory/updated` → invalidate `["inventory"]`.

TDD: each publisher wiring gets a test that fakes the hub and asserts publish was called with the right topic + payload; hub tests cover fan-out and replay.

Dry run: enable per-tenant via `global_config`; monitor DB queries for regression (invalidation should not increase QPS beyond baseline + realtime save).

---

## 9. Phase 6 — TikTok API drift (outline)

- Audit `codegraph explore` for every `/202309/` call site; compare against TikTok changelog. Any endpoint with a newer version and a documented sunset date gets a migration task with a dedicated PR.
- Adopt **Inventory Update Webhook** (Nov 2025): register a new handler under `internal/handlers/webhooks/tiktok_inventory.go`, wire in `internal/services/webhooks/tiktok_processor.go`, dispatch as a job that invalidates the inventory cache and (if Phase 5 done) publishes `inventory/updated`.
- Verify `CategoryVersion: "v1"` for Indonesia in `pkg/tiktok/category.go:63`: probe `RecommendCategory` in sandbox with both v1 and v2; if v2 is now accepted for ID, switch and add unit-test fixture.
- Verify `GET /product/202309/products/search` (`api.go:87`) is still supported alongside the newer `POST /product/202502/products/search` (`product.go:248`); if 202309 is deprecated, migrate the callers using the same request shape.
- No changes to Payments API (codebase does not use it).

TDD: fixture-driven — record real TikTok responses from a sandbox shop (redacted), assert parsing on both versions during migration windows.

Dry run: run the sandbox migration in a scratch tenant; commit only after two consecutive successful syncs.

---

## 10. Risk register

| Risk | Mitigation |
|---|---|
| Playwright CI job flakes on cold Docker compose start | `runtime-strict` already waits for backend + frontend readiness; retries: 2; non-blocking during Phase 0 |
| Lazada masking breaks report exports that assume raw address | Phase 1 keeps masked strings verbatim and exposes `unmasked_available`; docs get an updated note |
| Realtime hub v1 stalls or leaks under load | Isolation + gap-probe test files mirror the extensions-hub coverage; hub goroutine is drained in `Stop()` with test proving Stop-without-Run doesn't hang |
| Nginx `proxy_read_timeout` change breaks other `/api` paths | Add a dedicated `location = /api/realtime/ws` block; do not touch the general `/api` block |
| TikTok 202309 → newer version silently drops fields | Fixture-driven diff test: assert every field present in the 202309 response is present in the newer response before flipping |

---

## 11. Out of scope

- Redis-backed message replay beyond 512-entry in-memory ring (defer to Phase 6+).
- Prometheus metrics scaffolding (backlog).
- Migration to any managed WS gateway (self-host stays).
- Google Sheets webhook (Sheets doesn't push; polling remains).
- Shopee Video API (codebase does not use it; `aigc_label` change is noted only for tracking).
