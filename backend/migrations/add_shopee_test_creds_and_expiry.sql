-- Phase 8: extend credential_app_configs to cover Shopee test credentials,
-- partner-key expiry dates, active environment switch (live vs test), and a
-- generic app-status flag (all platforms).
--
-- Applied automatically by GORM AutoMigrate on next backend start via the
-- extended CredentialAppConfig model. This file exists for ops visibility and
-- as a safety net for a manual re-apply against restored backups.
--
-- All columns are NULLABLE / have safe defaults so the migration is
-- backward-compatible: existing Lazada + TikTok rows are unaffected.

ALTER TABLE credential_app_configs
    ADD COLUMN IF NOT EXISTS test_partner_id             BIGINT      NULL,
    ADD COLUMN IF NOT EXISTS test_partner_key            TEXT        NULL,
    ADD COLUMN IF NOT EXISTS partner_key_expires_at      TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS test_partner_key_expires_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS app_status                  VARCHAR(16) NOT NULL DEFAULT 'online',
    ADD COLUMN IF NOT EXISTS active_partner_env          VARCHAR(8)  NOT NULL DEFAULT 'live';

-- Optional CHECK constraints for the two enum-ish columns. Kept as separate
-- statements so a re-run can `IF NOT EXISTS`-skip cleanly.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'credential_app_configs_app_status_check') THEN
        ALTER TABLE credential_app_configs
            ADD CONSTRAINT credential_app_configs_app_status_check
            CHECK (app_status IN ('online', 'offline'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'credential_app_configs_active_env_check') THEN
        ALTER TABLE credential_app_configs
            ADD CONSTRAINT credential_app_configs_active_env_check
            CHECK (active_partner_env IN ('live', 'test'));
    END IF;
END$$;
