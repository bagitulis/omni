-- ============================================
-- Migration: Add progress tracking columns to jobs table
-- Date: 2026-01-26
-- Purpose: Enable background job progress tracking for long-running operations
-- ============================================

-- Add progress columns to jobs table in all tenant schemas
DO $$
DECLARE
    schema_record RECORD;
BEGIN
    FOR schema_record IN 
        SELECT schema_name 
        FROM information_schema.schemata 
        WHERE schema_name LIKE 'tenant_%'
    LOOP
        -- Add progress_percent column
        EXECUTE format('
            ALTER TABLE %I.jobs 
            ADD COLUMN IF NOT EXISTS progress_percent INTEGER DEFAULT 0
        ', schema_record.schema_name);
        
        -- Add progress_message column
        EXECUTE format('
            ALTER TABLE %I.jobs 
            ADD COLUMN IF NOT EXISTS progress_message VARCHAR(255)
        ', schema_record.schema_name);
        
        -- Add total_items column
        EXECUTE format('
            ALTER TABLE %I.jobs 
            ADD COLUMN IF NOT EXISTS total_items INTEGER DEFAULT 0
        ', schema_record.schema_name);
        
        -- Add processed_items column
        EXECUTE format('
            ALTER TABLE %I.jobs 
            ADD COLUMN IF NOT EXISTS processed_items INTEGER DEFAULT 0
        ', schema_record.schema_name);
        
        -- Add result_data column for storing job result
        EXECUTE format('
            ALTER TABLE %I.jobs 
            ADD COLUMN IF NOT EXISTS result_data JSONB
        ', schema_record.schema_name);
        
        RAISE NOTICE 'Added progress columns to %.jobs', schema_record.schema_name;
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
