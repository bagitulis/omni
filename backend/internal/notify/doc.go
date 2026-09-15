// Package notify is the canonical notification pipeline for Omni.
//
// It orchestrates three concerns behind one narrow API (Bus.Emit):
//
//   1. Persistence — Postgres via repositories.NotificationRepository,
//      with dedup UPSERT when a producer supplies a dedup_key.
//   2. Fanout — Envelope broadcast to a Fanout implementation
//      (InProcess for dev / single-replica, Redis Pub/Sub for multi-replica).
//   3. Safety — SanitizeForUser strips SQL/stack/token patterns from title
//      and message BEFORE persistence; SafeActionURL rejects any URL that
//      is not a path relative to the SPA.
//
// The package deliberately owns NO transport (SSE/WS) and NO scheduling
// (cron). Those live in handlers and services and consume Bus.
//
// The legacy services.NotificationService is being migrated onto this Bus
// during the V2 rollout; both coexist until the SSE endpoint is removed.
package notify
