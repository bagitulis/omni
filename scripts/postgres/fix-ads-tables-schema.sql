-- ============================================
-- Migration: Fix Ads Tables Schema - Add Missing Columns
-- Database: omni_main
-- Target Schema: All tenant schemas (tenant_*)
-- Date: 2026-01-25
-- ============================================

-- Function to fix ads tables schema in a specific tenant schema
CREATE OR REPLACE FUNCTION fix_ads_tables_schema(schema_name TEXT)
RETURNS VOID AS $$
BEGIN
    -- Set search path
    EXECUTE format('SET search_path TO %I, public', schema_name);

    -- ============================================
    -- Fix shopee_ads_upload_batches
    -- ============================================
    -- Add period_label if not exists
    EXECUTE format('
        DO $inner$
        BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM information_schema.columns 
                WHERE table_schema = %L AND table_name = ''shopee_ads_upload_batches'' AND column_name = ''period_label''
            ) THEN
                ALTER TABLE %I.shopee_ads_upload_batches ADD COLUMN period_label VARCHAR(20) NOT NULL DEFAULT '''';
            END IF;
        END $inner$;
    ', schema_name, schema_name);

    -- Add period_start if not exists
    EXECUTE format('
        DO $inner$
        BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM information_schema.columns 
                WHERE table_schema = %L AND table_name = ''shopee_ads_upload_batches'' AND column_name = ''period_start''
            ) THEN
                ALTER TABLE %I.shopee_ads_upload_batches ADD COLUMN period_start VARCHAR(20) NOT NULL DEFAULT '''';
            END IF;
        END $inner$;
    ', schema_name, schema_name);

    -- Add period_end if not exists
    EXECUTE format('
        DO $inner$
        BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM information_schema.columns 
                WHERE table_schema = %L AND table_name = ''shopee_ads_upload_batches'' AND column_name = ''period_end''
            ) THEN
                ALTER TABLE %I.shopee_ads_upload_batches ADD COLUMN period_end VARCHAR(20) NOT NULL DEFAULT '''';
            END IF;
        END $inner$;
    ', schema_name, schema_name);

    -- Create index on period_label
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_ads_batch_period ON %I.shopee_ads_upload_batches(period_label)', schema_name);

    -- ============================================
    -- Fix shopee_ads_product_data - Add missing columns
    -- ============================================
    EXECUTE format('
        DO $inner$
        BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM information_schema.columns 
                WHERE table_schema = %L AND table_name = ''shopee_ads_product_data'' AND column_name = ''period_label''
            ) THEN
                ALTER TABLE %I.shopee_ads_product_data ADD COLUMN period_label VARCHAR(20);
            END IF;
        END $inner$;
    ', schema_name, schema_name);

    EXECUTE format('
        DO $inner$
        BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM information_schema.columns 
                WHERE table_schema = %L AND table_name = ''shopee_ads_product_data'' AND column_name = ''period_start''
            ) THEN
                ALTER TABLE %I.shopee_ads_product_data ADD COLUMN period_start VARCHAR(20);
            END IF;
        END $inner$;
    ', schema_name, schema_name);

    EXECUTE format('
        DO $inner$
        BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM information_schema.columns 
                WHERE table_schema = %L AND table_name = ''shopee_ads_product_data'' AND column_name = ''period_end''
            ) THEN
                ALTER TABLE %I.shopee_ads_product_data ADD COLUMN period_end VARCHAR(20);
            END IF;
        END $inner$;
    ', schema_name, schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_ads_data_period ON %I.shopee_ads_product_data(period_label)', schema_name);

    -- ============================================
    -- Create TikTok tables if not exist
    -- ============================================
    -- tiktok_ads_upload_batches
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_ads_upload_batches (
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

    -- tiktok_ads_creative_data
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_ads_creative_data (
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
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_data_batch ON %I.tiktok_ads_creative_data(batch_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_data_product ON %I.tiktok_ads_creative_data(product_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_data_period ON %I.tiktok_ads_creative_data(period_label)', schema_name);

    -- tiktok_ads_product_summaries
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_ads_product_summaries (
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
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_summary_product ON %I.tiktok_ads_product_summaries(product_id)', schema_name);

    -- tiktok_ads_ml_predictions
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_ads_ml_predictions (
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
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_ml_product ON %I.tiktok_ads_ml_predictions(product_id)', schema_name);

    RAISE NOTICE 'Schema fixed for: %', schema_name;
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
        RAISE NOTICE 'Fixing ads tables schema in: %', tenant_schema;
        PERFORM fix_ads_tables_schema(tenant_schema);
    END LOOP;
    
    RAISE NOTICE '✅ Ads analytics tables schema fix completed for all tenant schemas';
END;
$$;

-- Cleanup
DROP FUNCTION IF EXISTS fix_ads_tables_schema(TEXT);
