-- Migration: Create marketplace_sync_history table
-- Date: 2026-02-17
-- Purpose: Track marketplace sync operations (stock, price, wholesale, MPQ, clone) per SKU per platform

-- Add marketplace_sync_history table to all tenant schemas
DO $$
DECLARE
    tenant_schema text;
BEGIN
    FOR tenant_schema IN 
        SELECT DISTINCT table_schema 
        FROM information_schema.tables 
        WHERE table_schema LIKE 'tenant_%'
    LOOP
        -- Create table if not exists
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_schema = tenant_schema 
            AND table_name = 'marketplace_sync_history'
        ) THEN
            EXECUTE format('
                CREATE TABLE %I.marketplace_sync_history (
                    id TEXT PRIMARY KEY,
                    tenant_id TEXT NOT NULL,
                    sku TEXT NOT NULL,
                    platform TEXT NOT NULL CHECK (platform IN (''shopee'', ''tiktok'', ''lazada'')),
                    operation TEXT NOT NULL CHECK (operation IN (''stock_update'', ''price_update'', ''wholesale_update'', ''mpq_update'', ''clone'')),
                    status TEXT NOT NULL CHECK (status IN (''success'', ''failed'', ''partial'')),
                    request_data TEXT,
                    response_data TEXT,
                    error_message TEXT,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )', tenant_schema);

            EXECUTE format('CREATE INDEX idx_msh_tenant_id ON %I.marketplace_sync_history(tenant_id)', tenant_schema);
            EXECUTE format('CREATE INDEX idx_msh_sku ON %I.marketplace_sync_history(sku)', tenant_schema);
            EXECUTE format('CREATE INDEX idx_msh_platform ON %I.marketplace_sync_history(platform)', tenant_schema);
            EXECUTE format('CREATE INDEX idx_msh_operation ON %I.marketplace_sync_history(operation)', tenant_schema);
            EXECUTE format('CREATE INDEX idx_msh_created_at ON %I.marketplace_sync_history(created_at)', tenant_schema);

            RAISE NOTICE 'Created marketplace_sync_history table in schema %', tenant_schema;
        ELSE
            RAISE NOTICE 'marketplace_sync_history table already exists in schema %', tenant_schema;
        END IF;
    END LOOP;
END $$;
