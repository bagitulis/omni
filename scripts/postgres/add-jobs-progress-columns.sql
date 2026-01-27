-- ============================================
-- Migration: Add progress tracking columns to jobs table
-- Date: 2026-01-26
-- Purpose: Enable background job progress tracking for long-running operations
-- ============================================

-- Add progress columns to jobs table in all tenant schemas
DO $$
DECLARE
    schema_record RECORD;
    qualified_table TEXT;
BEGIN
    FOR schema_record IN 
        SELECT schema_name 
        FROM information_schema.schemata 
        WHERE schema_name LIKE 'tenant_%'
    LOOP
        qualified_table := format('%I.jobs', schema_record.schema_name);

        -- Create table if missing, with full progress columns (avoids partial auto-create during restore)
        IF to_regclass(qualified_table) IS NULL THEN
            EXECUTE format('
                CREATE TABLE %s (
                    id VARCHAR(255) NOT NULL PRIMARY KEY,
                    type VARCHAR(255) NOT NULL,
                    status VARCHAR(50) NOT NULL,
                    priority VARCHAR(50) DEFAULT ''normal'',
                    data JSONB NOT NULL,
                    error_message TEXT,
                    progress_percent INTEGER DEFAULT 0,
                    progress_message TEXT DEFAULT '''',
                    total_items INTEGER DEFAULT 0,
                    processed_items INTEGER DEFAULT 0,
                    result_data JSONB,
                    started_at TIMESTAMPTZ,
                    completed_at TIMESTAMPTZ,
                    created_at TIMESTAMPTZ DEFAULT now(),
                    updated_at TIMESTAMPTZ DEFAULT now()
                )', qualified_table);
            RAISE NOTICE 'Created %', qualified_table;
        END IF;

        -- Add/ensure progress columns (idempotent)
        EXECUTE format('ALTER TABLE %s ADD COLUMN IF NOT EXISTS progress_percent INTEGER DEFAULT 0', qualified_table);
        EXECUTE format('ALTER TABLE %s ADD COLUMN IF NOT EXISTS progress_message TEXT DEFAULT ''''', qualified_table);
        EXECUTE format('ALTER TABLE %s ADD COLUMN IF NOT EXISTS total_items INTEGER DEFAULT 0', qualified_table);
        EXECUTE format('ALTER TABLE %s ADD COLUMN IF NOT EXISTS processed_items INTEGER DEFAULT 0', qualified_table);
        EXECUTE format('ALTER TABLE %s ADD COLUMN IF NOT EXISTS result_data JSONB', qualified_table);

        -- Normalize column types to match Go models (jsonb/text)
        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = schema_record.schema_name 
              AND table_name = 'jobs'
              AND column_name = 'result_data'
              AND data_type <> 'jsonb'
        ) THEN
            EXECUTE format('ALTER TABLE %s ALTER COLUMN result_data DROP DEFAULT', qualified_table);
            EXECUTE format('ALTER TABLE %s ALTER COLUMN result_data TYPE JSONB USING NULLIF(result_data, '''')::jsonb', qualified_table);
            RAISE NOTICE 'Normalized % result_data to jsonb', qualified_table;
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = schema_record.schema_name 
              AND table_name = 'jobs'
              AND column_name = 'progress_message'
              AND data_type <> 'text'
        ) THEN
            EXECUTE format('ALTER TABLE %s ALTER COLUMN progress_message TYPE TEXT', qualified_table);
            RAISE NOTICE 'Normalized % progress_message to text', qualified_table;
        END IF;

        RAISE NOTICE 'Ensured progress columns on %', qualified_table;
    END LOOP;
END $$;

-- Verify the changes
SELECT 
    table_schema,
    column_name,
    data_type
FROM information_schema.columns 
WHERE table_name = 'jobs' 
    AND table_schema LIKE 'tenant_%'
    AND column_name IN ('progress_percent', 'progress_message', 'total_items', 'processed_items', 'result_data')
ORDER BY table_schema, column_name;
