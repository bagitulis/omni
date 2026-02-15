-- Migration: Add sync_status column to inventory_records
-- Date: 2026-02-15
-- Purpose: Support per-item sync status filtering

-- Add sync_status column to all tenant schemas
DO $$
DECLARE
    tenant_schema text;
BEGIN
    FOR tenant_schema IN 
        SELECT DISTINCT table_schema 
        FROM information_schema.tables 
        WHERE table_name = 'inventory_records' 
        AND table_schema LIKE 'tenant_%'
    LOOP
        -- Check if column already exists
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = tenant_schema 
            AND table_name = 'inventory_records' 
            AND column_name = 'sync_status'
        ) THEN
            EXECUTE format('ALTER TABLE %I.inventory_records ADD COLUMN sync_status VARCHAR(50)', tenant_schema);
            RAISE NOTICE 'Added sync_status column to %.inventory_records', tenant_schema;
        ELSE
            RAISE NOTICE 'sync_status column already exists in %.inventory_records', tenant_schema;
        END IF;
    END LOOP;
END $$;
