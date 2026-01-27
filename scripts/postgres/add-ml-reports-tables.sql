-- ============================================
-- Migration: Add ML Reports Tables
-- Database: omni_main
-- Target Schema: All tenant schemas (tenant_*)
-- Date: 2026-01-27
-- Description: Tables for storing ML-generated reports and async job tracking
-- ============================================

-- Function to create ML report tables in a specific tenant schema
CREATE OR REPLACE FUNCTION create_ml_report_tables(schema_name TEXT)
RETURNS VOID AS $$
BEGIN
    -- Set search path to target schema
    EXECUTE format('SET search_path TO %I, public', schema_name);

    -- ============================================
    -- 1. ML Reports Table
    -- Stores generated HTML reports from Python ML scripts
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.ml_reports (
            id SERIAL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            platform VARCHAR(50) NOT NULL CHECK (platform IN (''shopee'', ''tiktok'')),
            report_type VARCHAR(50) NOT NULL CHECK (report_type IN (''full'', ''executive'', ''quick'')),
            period_start TIMESTAMPTZ,
            period_end TIMESTAMPTZ,
            period_label VARCHAR(50),
            file_path VARCHAR(500) NOT NULL,
            file_name VARCHAR(255) NOT NULL,
            file_size BIGINT DEFAULT 0,
            status VARCHAR(50) DEFAULT ''completed'' CHECK (status IN (''pending'', ''completed'', ''failed'')),
            error_msg TEXT,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
        )', schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_ml_reports_tenant ON %I.ml_reports(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_ml_reports_platform ON %I.ml_reports(platform)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_ml_reports_period ON %I.ml_reports(period_label)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_ml_reports_status ON %I.ml_reports(status)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_ml_reports_created ON %I.ml_reports(created_at DESC)', schema_name);

    -- ============================================
    -- 2. ML Jobs Table
    -- Tracks async ML job execution
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.ml_jobs (
            id SERIAL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            job_type VARCHAR(100) NOT NULL,
            platform VARCHAR(50) NOT NULL CHECK (platform IN (''shopee'', ''tiktok'')),
            status VARCHAR(50) DEFAULT ''pending'' CHECK (status IN (''pending'', ''running'', ''completed'', ''failed'', ''cancelled'')),
            progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
            result_id INTEGER,
            error_msg TEXT,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            started_at TIMESTAMPTZ,
            completed_at TIMESTAMPTZ,
            CONSTRAINT fk_ml_job_result FOREIGN KEY (result_id)
                REFERENCES %I.ml_reports(id) ON DELETE SET NULL
        )', schema_name, schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_ml_jobs_tenant ON %I.ml_jobs(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_ml_jobs_platform ON %I.ml_jobs(platform)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_ml_jobs_status ON %I.ml_jobs(status)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_ml_jobs_created ON %I.ml_jobs(created_at DESC)', schema_name);

    RAISE NOTICE 'ML report tables created successfully in schema: %', schema_name;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Execute for all tenant schemas
-- ============================================
DO $$
DECLARE
    tenant_schema TEXT;
BEGIN
    -- Loop through all tenant schemas
    FOR tenant_schema IN 
        SELECT schema_name 
        FROM information_schema.schemata 
        WHERE schema_name LIKE 'tenant_%'
    LOOP
        RAISE NOTICE 'Creating ML report tables in schema: %', tenant_schema;
        PERFORM create_ml_report_tables(tenant_schema);
    END LOOP;
    
    RAISE NOTICE '✅ ML report tables migration completed for all tenant schemas';
END;
$$;

-- ============================================
-- Cleanup: Drop the function after use (optional)
-- ============================================
-- DROP FUNCTION IF EXISTS create_ml_report_tables(TEXT);
