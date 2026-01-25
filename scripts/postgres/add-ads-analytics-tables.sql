-- ============================================
-- Migration: Add Ads Analytics Tables for Shopee & TikTok
-- Database: omni_main
-- Target Schema: All tenant schemas (tenant_*)
-- Date: 2026-01-25
-- ============================================

-- Function to create ads analytics tables in a specific tenant schema
CREATE OR REPLACE FUNCTION create_ads_tables(schema_name TEXT)
RETURNS VOID AS $$
BEGIN
    -- Set search path to target schema
    EXECUTE format('SET search_path TO %I, public', schema_name);

    -- ============================================
    -- 1. Shopee Ads Upload Batch Table
    -- Tracks each CSV file upload
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_ads_upload_batches (
            id SERIAL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            filename VARCHAR(500) NOT NULL,
            period_start VARCHAR(20) NOT NULL,
            period_end VARCHAR(20) NOT NULL,
            period_label VARCHAR(20) NOT NULL,
            total_rows INTEGER DEFAULT 0,
            uploaded_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
        )', schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_ads_batch_tenant ON %I.shopee_ads_upload_batches(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_ads_batch_period ON %I.shopee_ads_upload_batches(period_label)', schema_name);

    -- ============================================
    -- 2. Shopee Ads Product Data Table
    -- Raw product-level ads data from CSV
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_ads_product_data (
            id SERIAL PRIMARY KEY,
            batch_id INTEGER NOT NULL,
            tenant_id VARCHAR(255) NOT NULL,
            product_id VARCHAR(255),
            product_name TEXT,
            status VARCHAR(100),
            bidding_mode VARCHAR(100),
            placement VARCHAR(100),
            impressions INTEGER DEFAULT 0,
            clicks INTEGER DEFAULT 0,
            ctr DOUBLE PRECISION DEFAULT 0,
            conversions INTEGER DEFAULT 0,
            conversion_rate DOUBLE PRECISION DEFAULT 0,
            cost DOUBLE PRECISION DEFAULT 0,
            revenue DOUBLE PRECISION DEFAULT 0,
            roas DOUBLE PRECISION DEFAULT 0,
            acos DOUBLE PRECISION DEFAULT 0,
            units_sold INTEGER DEFAULT 0,
            cost_per_conv DOUBLE PRECISION DEFAULT 0,
            direct_conversions INTEGER DEFAULT 0,
            direct_revenue DOUBLE PRECISION DEFAULT 0,
            direct_roas DOUBLE PRECISION DEFAULT 0,
            direct_acos DOUBLE PRECISION DEFAULT 0,
            direct_cost_per_conv DOUBLE PRECISION DEFAULT 0,
            period_label VARCHAR(20),
            period_start VARCHAR(20),
            period_end VARCHAR(20),
            CONSTRAINT fk_shopee_ads_batch FOREIGN KEY (batch_id)
                REFERENCES %I.shopee_ads_upload_batches(id) ON DELETE CASCADE
        )', schema_name, schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_ads_data_tenant ON %I.shopee_ads_product_data(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_ads_data_batch ON %I.shopee_ads_product_data(batch_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_ads_data_product ON %I.shopee_ads_product_data(product_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_ads_data_period ON %I.shopee_ads_product_data(period_label)', schema_name);

    -- ============================================
    -- 3. TikTok Ads Upload Batch Table
    -- Tracks each Excel file upload
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_ads_upload_batches (
            id SERIAL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            filename VARCHAR(500) NOT NULL,
            period_start VARCHAR(20) NOT NULL,
            period_end VARCHAR(20) NOT NULL,
            period_label VARCHAR(20) NOT NULL,
            total_rows INTEGER DEFAULT 0,
            uploaded_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
        )', schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_batch_tenant ON %I.tiktok_ads_upload_batches(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_batch_period ON %I.tiktok_ads_upload_batches(period_label)', schema_name);

    -- ============================================
    -- 4. TikTok Ads Creative Data Table
    -- Creative-level ads data from Excel
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_ads_creative_data (
            id SERIAL PRIMARY KEY,
            batch_id INTEGER NOT NULL,
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
            period_end VARCHAR(20),
            CONSTRAINT fk_tiktok_ads_batch FOREIGN KEY (batch_id)
                REFERENCES %I.tiktok_ads_upload_batches(id) ON DELETE CASCADE
        )', schema_name, schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_data_tenant ON %I.tiktok_ads_creative_data(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_data_batch ON %I.tiktok_ads_creative_data(batch_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_data_creative ON %I.tiktok_ads_creative_data(creative_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_data_product ON %I.tiktok_ads_creative_data(product_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_data_period ON %I.tiktok_ads_creative_data(period_label)', schema_name);

    -- ============================================
    -- 5. TikTok Ads Product Summary Table
    -- Aggregated product-level summary
    -- ============================================
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
            period_label VARCHAR(20) NOT NULL,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(tenant_id, product_id, period_label)
        )', schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_summary_tenant ON %I.tiktok_ads_product_summaries(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_summary_product ON %I.tiktok_ads_product_summaries(product_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_summary_period ON %I.tiktok_ads_product_summaries(period_label)', schema_name);

    -- ============================================
    -- 6. TikTok Ads ML Prediction Table
    -- Machine learning predictions for ads performance
    -- ============================================
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
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_ads_ml_valid ON %I.tiktok_ads_ml_predictions(valid_until)', schema_name);

    RAISE NOTICE 'Ads analytics tables created successfully in schema: %', schema_name;
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
        RAISE NOTICE 'Creating ads analytics tables in schema: %', tenant_schema;
        PERFORM create_ads_tables(tenant_schema);
    END LOOP;
    
    RAISE NOTICE '✅ Ads analytics tables migration completed for all tenant schemas';
END;
$$;

-- ============================================
-- Cleanup: Drop the function after use (optional)
-- ============================================
-- DROP FUNCTION IF EXISTS create_ads_tables(TEXT);
