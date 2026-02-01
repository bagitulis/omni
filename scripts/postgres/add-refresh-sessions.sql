-- Migration: Add refresh_sessions table for secure token management
-- Run this in BOTH system schema AND each tenant schema
-- Date: 2025-02-01

-- ================================================
-- Create in system schema (for developer accounts)
-- ================================================
CREATE TABLE IF NOT EXISTS system.refresh_sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    tenant_id VARCHAR(100),
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    replaced_by_hash VARCHAR(64),
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for system schema
CREATE INDEX IF NOT EXISTS idx_system_refresh_sessions_user_id ON system.refresh_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_system_refresh_sessions_expires_at ON system.refresh_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_system_refresh_sessions_token_hash ON system.refresh_sessions(token_hash);

-- ================================================
-- Function to create refresh_sessions in tenant schemas
-- ================================================
CREATE OR REPLACE FUNCTION create_tenant_refresh_sessions(schema_name TEXT) RETURNS VOID AS $$
BEGIN
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.refresh_sessions (
            id VARCHAR(36) PRIMARY KEY,
            user_id VARCHAR(36) NOT NULL,
            tenant_id VARCHAR(100),
            token_hash VARCHAR(64) NOT NULL UNIQUE,
            expires_at TIMESTAMP NOT NULL,
            revoked_at TIMESTAMP,
            replaced_by_hash VARCHAR(64),
            ip_address VARCHAR(45),
            user_agent TEXT,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            last_used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );

        CREATE INDEX IF NOT EXISTS idx_%I_refresh_sessions_user_id ON %I.refresh_sessions(user_id);
        CREATE INDEX IF NOT EXISTS idx_%I_refresh_sessions_expires_at ON %I.refresh_sessions(expires_at);
    ', schema_name, schema_name, schema_name, schema_name, schema_name);
END;
$$ LANGUAGE plpgsql;

-- ================================================
-- Apply to existing tenant schemas
-- ================================================
DO $$
DECLARE
    schema_rec RECORD;
BEGIN
    FOR schema_rec IN
        SELECT schema_name
        FROM information_schema.schemata
        WHERE schema_name LIKE 'tenant_%'
    LOOP
        PERFORM create_tenant_refresh_sessions(schema_rec.schema_name);
        RAISE NOTICE 'Created refresh_sessions table in schema: %', schema_rec.schema_name;
    END LOOP;
END $$;

-- ================================================
-- Grant permissions (adjust as needed)
-- ================================================
-- GRANT SELECT, INSERT, UPDATE, DELETE ON system.refresh_sessions TO omni_app;

COMMENT ON TABLE system.refresh_sessions IS 'Stores refresh token sessions for secure token rotation and revocation';
