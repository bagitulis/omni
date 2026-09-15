-- Notifications V2 — drop legacy `read` column.
--
-- Applied MANUALLY by ops, NOT by GORM AutoMigrate (per project policy:
-- AutoMigrate is additive-only; DROP requires explicit intent).
--
-- Pre-conditions before running this file:
--   1. All application code writes read state ONLY via the
--      `notification_reads` join table (see backend/internal/repositories/
--      notification_repository_v2.go: MarkReadForUser, MarkAllReadForUser,
--      BulkMarkReadForUser).
--   2. Legacy repo methods MarkAsRead / MarkAllAsRead / UnreadCount (which
--      still touch the `read` column) are no longer called from any handler
--      or job in this release.
--   3. Backfill sanity check has been run once per tenant schema — see
--      NOTES below for the query and expected outcome.
--   4. Deploy has been stable in production for at least 1 week without
--      regression on notification read state.
--
-- Rollback: re-add the column. Existing per-user reads in
-- `notification_reads` remain intact; the column will be `false` for every
-- row and code paths that read from `notifications.read` will treat all
-- rows as unread until backfilled.
--
--   ALTER TABLE notifications ADD COLUMN read BOOLEAN NOT NULL DEFAULT FALSE;
--   UPDATE notifications n
--   SET read = TRUE
--   WHERE EXISTS (
--     SELECT 1 FROM notification_reads r
--     WHERE r.notification_id = n.id
--   );
--
-- Run PER TENANT SCHEMA. Example:
--
--   SET search_path TO tenant_<tenant_id>;
--   \i backend/migrations/notifications_v2_drop_legacy_read.sql
--
-- Or scripted across every tenant schema; see docs/deployment/README for
-- the standard multi-tenant apply loop.

BEGIN;

-- Sanity check: any tenant whose `read` column and `notification_reads` join
-- table disagree (i.e. read=true rows without a join row) is a candidate for
-- one-time backfill BEFORE we drop the column. Log-only; does not raise.
DO $$
DECLARE
    orphaned_reads INTEGER;
BEGIN
    SELECT COUNT(*) INTO orphaned_reads
    FROM notifications n
    WHERE n.read = TRUE
      AND NOT EXISTS (
        SELECT 1 FROM notification_reads r
        WHERE r.notification_id = n.id
      );
    RAISE NOTICE 'notifications_v2: % legacy read=true rows without notification_reads entries (backfilling as user_id=0 sentinel)', orphaned_reads;
END$$;

-- Backfill legacy read state into notification_reads for user_id=0 (sentinel
-- meaning "read before per-user tracking existed"). Idempotent.
INSERT INTO notification_reads (notification_id, user_id, read_at)
SELECT n.id, 0, n.created_at
FROM notifications n
WHERE n.read = TRUE
ON CONFLICT (notification_id, user_id) DO NOTHING;

-- Drop the column. Fast (metadata-only in Postgres) on tables up to millions
-- of rows.
ALTER TABLE notifications DROP COLUMN IF EXISTS read;

COMMIT;

-- After running this file: bump the model in
-- backend/internal/models/notification.go to remove the `Read` field, then
-- delete the legacy `MarkAsRead`, `MarkAllAsRead`, and `UnreadCount` methods
-- from notification_repository.go (all callers should already be gone).
