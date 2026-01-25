-- ============================================
-- Migration: Fix ALL Ads Tables Schema - Add Missing Columns
-- Database: omni_main
-- Date: 2026-01-25
-- ============================================

-- Function to add column if not exists
CREATE OR REPLACE FUNCTION add_column_if_not_exists(
    p_schema TEXT, 
    p_table TEXT, 
    p_column TEXT, 
    p_type TEXT,
    p_default TEXT DEFAULT NULL
)
RETURNS VOID AS $$
DECLARE
    col_exists BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = p_schema 
        AND table_name = p_table 
        AND column_name = p_column
    ) INTO col_exists;
    
    IF NOT col_exists THEN
        IF p_default IS NOT NULL THEN
            EXECUTE format('ALTER TABLE %I.%I ADD COLUMN %I %s DEFAULT %s', 
                p_schema, p_table, p_column, p_type, p_default);
        ELSE
            EXECUTE format('ALTER TABLE %I.%I ADD COLUMN %I %s', 
                p_schema, p_table, p_column, p_type);
        END IF;
        RAISE NOTICE 'Added column %.%.%', p_schema, p_table, p_column;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Function to fix all ads tables in a tenant schema
CREATE OR REPLACE FUNCTION fix_all_ads_tables(schema_name TEXT)
RETURNS VOID AS $$
BEGIN
    EXECUTE format('SET search_path TO %I, public', schema_name);

    -- ============================================
    -- Fix shopee_ads_upload_batches
    -- ============================================
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = schema_name AND table_name = 'shopee_ads_upload_batches') THEN
        PERFORM add_column_if_not_exists(schema_name, 'shopee_ads_upload_batches', 'period_label', 'VARCHAR(20)', '''''' );
        PERFORM add_column_if_not_exists(schema_name, 'shopee_ads_upload_batches', 'period_start', 'VARCHAR(20)', '''''' );
        PERFORM add_column_if_not_exists(schema_name, 'shopee_ads_upload_batches', 'period_end', 'VARCHAR(20)', '''''' );
        EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_ads_batch_period ON %I.shopee_ads_upload_batches(period_label)', schema_name);
    END IF;

    -- ============================================
    -- Fix shopee_ads_product_data
    -- ============================================
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = schema_name AND table_name = 'shopee_ads_product_data') THEN
        PERFORM add_column_if_not_exists(schema_name, 'shopee_ads_product_data', 'period_label', 'VARCHAR(20)', NULL);
        PERFORM add_column_if_not_exists(schema_name, 'shopee_ads_product_data', 'period_start', 'VARCHAR(20)', NULL);
        PERFORM add_column_if_not_exists(schema_name, 'shopee_ads_product_data', 'period_end', 'VARCHAR(20)', NULL);
        EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_ads_data_period ON %I.shopee_ads_product_data(period_label)', schema_name);
    END IF;

    -- ============================================
    -- Fix tiktok_ads_upload_batches
    -- ============================================
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = schema_name AND table_name = 'tiktok_ads_upload_batches') THEN
        PERFORM add_column_if_not_exists(schema_name, 'tiktok_ads_upload_batches', 'period_label', 'VARCHAR(20)', '''''' );
        PERFORM add_column_if_not_exists(schema_name, 'tiktok_ads_upload_batches', 'period_start', 'VARCHAR(20)', '''''' );
        PERFORM add_column_if_not_exists(schema_name, 'tiktok_ads_upload_batches', 'period_end', 'VARCHAR(20)', '''''' );
        EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_batch_period ON %I.tiktok_ads_upload_batches(period_label)', schema_name);
    ELSE
        -- Create if not exists
        EXECUTE format('
            CREATE TABLE %I.tiktok_ads_upload_batches (
                id SERIAL PRIMARY KEY,
                tenant_id VARCHAR(255) NOT NULL,
                filename VARCHAR(500) NOT NULL,
                period_start VARCHAR(20) NOT NULL DEFAULT '''',
                period_end VARCHAR(20) NOT NULL DEFAULT '''',
                period_label VARCHAR(20) NOT NULL DEFAULT '''',
                total_rows INTEGER DEFAULT 0,
                uploaded_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
                created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
            )', schema_name);
        EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_batch_tenant ON %I.tiktok_ads_upload_batches(tenant_id)', schema_name);
        EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_batch_period ON %I.tiktok_ads_upload_batches(period_label)', schema_name);
    END IF;

    -- ============================================
    -- Fix tiktok_ads_creative_data
    -- ============================================
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = schema_name AND table_name = 'tiktok_ads_creative_data') THEN
        PERFORM add_column_if_not_exists(schema_name, 'tiktok_ads_creative_data', 'period_label', 'VARCHAR(20)', NULL);
        PERFORM add_column_if_not_exists(schema_name, 'tiktok_ads_creative_data', 'period_start', 'VARCHAR(20)', NULL);
        PERFORM add_column_if_not_exists(schema_name, 'tiktok_ads_creative_data', 'period_end', 'VARCHAR(20)', NULL);
        EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_data_period ON %I.tiktok_ads_creative_data(period_label)', schema_name);
    ELSE
        EXECUTE format('
            CREATE TABLE %I.tiktok_ads_creative_data (
                id SERIAL PRIMARY KEY,
                batch_id INTEGER,
                tenant_id VARCHAR(255) NOT NULL,
                creative_id VARCHAR(255),
                creative_name TEXT,
                product_id VARCHAR(255),
                product_name TEXT,
                status VARCHAR(100),
                impressions INTEGER DEFAULT 0,
                clicks INTEGER DEFAULT 0,
                ctr DOUBLE PRECISION DEFAULT 0,
                cost DOUBLE PRECISION DEFAULT 0,
                conversions INTEGER DEFAULT 0,
                conversion_rate DOUBLE PRECISION DEFAULT 0,
                revenue DOUBLE PRECISION DEFAULT 0,
                roi DOUBLE PRECISION DEFAULT 0,
                cpa DOUBLE PRECISION DEFAULT 0,
                video_views INTEGER DEFAULT 0,
                vtr DOUBLE PRECISION DEFAULT 0,
                period_label VARCHAR(20),
                period_start VARCHAR(20),
                period_end VARCHAR(20)
            )', schema_name);
        EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_data_tenant ON %I.tiktok_ads_creative_data(tenant_id)', schema_name);
        EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_data_product ON %I.tiktok_ads_creative_data(product_id)', schema_name);
        EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_data_period ON %I.tiktok_ads_creative_data(period_label)', schema_name);
    END IF;

    -- ============================================
    -- Create tiktok_ads_product_summaries if not exists
    -- ============================================
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = schema_name AND table_name = 'tiktok_ads_product_summaries') THEN
        EXECUTE format('
            CREATE TABLE %I.tiktok_ads_product_summaries (
                id SERIAL PRIMARY KEY,
                tenant_id VARCHAR(255) NOT NULL,
                product_id VARCHAR(255) NOT NULL,
                product_name TEXT,
                total_cost DOUBLE PRECISION DEFAULT 0,
                total_revenue DOUBLE PRECISION DEFAULT 0,
                total_conv INTEGER DEFAULT 0,
                avg_roi DOUBLE PRECISION DEFAULT 0,
                avg_cpa DOUBLE PRECISION DEFAULT 0,
                total_creatives INTEGER DEFAULT 0,
                period_label VARCHAR(20) NOT NULL DEFAULT '''',
                created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
                UNIQUE(tenant_id, product_id, period_label)
            )', schema_name);
        EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_summary_tenant ON %I.tiktok_ads_product_summaries(tenant_id)', schema_name);
    END IF;

    -- ============================================
    -- Create tiktok_ads_ml_predictions if not exists
    -- ============================================
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = schema_name AND table_name = 'tiktok_ads_ml_predictions') THEN
        EXECUTE format('
            CREATE TABLE %I.tiktok_ads_ml_predictions (
                id SERIAL PRIMARY KEY,
                tenant_id VARCHAR(255) NOT NULL,
                product_id VARCHAR(255) NOT NULL,
                predicted_roi DOUBLE PRECISION DEFAULT 0,
                predicted_conv INTEGER DEFAULT 0,
                confidence DOUBLE PRECISION DEFAULT 0,
                recommend_budget DOUBLE PRECISION DEFAULT 0,
                predicted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
                valid_until TIMESTAMPTZ,
                created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
            )', schema_name);
        EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_ml_tenant ON %I.tiktok_ads_ml_predictions(tenant_id)', schema_name);
    END IF;

    RAISE NOTICE '✅ Schema fixed: %', schema_name;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Execute for all tenant schemas
-- ============================================
DO $$
DECLARE
    tenant_schema TEXT;
BEGIN
    FOR tenant_schema IN 
        SELECT schema_name 
        FROM information_schema.schemata 
        WHERE schema_name LIKE 'tenant_%'
    LOOP
        PERFORM fix_all_ads_tables(tenant_schema);
    END LOOP;
    
    RAISE NOTICE '========================================';
    RAISE NOTICE '✅ ALL ads analytics tables migrated!';
    RAISE NOTICE '========================================';
END;
$$;

-- Cleanup functions
DROP FUNCTION IF EXISTS fix_all_ads_tables(TEXT);
DROP FUNCTION IF EXISTS add_column_if_not_exists(TEXT, TEXT, TEXT, TEXT, TEXT);
