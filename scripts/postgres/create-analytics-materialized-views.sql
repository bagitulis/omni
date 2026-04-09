-- ============================================
-- Migration: Create Analytics Materialized Views for Performance
-- Database: omni_main
-- Target Schema: All tenant schemas (tenant_*)
-- Date: 2026-01-29
-- Purpose: Pre-aggregate analytics data for fast dashboard loading
-- ============================================

-- Function to create materialized views in a specific tenant schema
CREATE OR REPLACE FUNCTION create_analytics_mvs(schema_name TEXT)
RETURNS VOID AS $$
BEGIN
    -- Set search path to target schema
    EXECUTE format('SET search_path TO %I, public', schema_name);

    -- ============================================
    -- 1. Cache Metadata Table
    -- Tracks last refresh time for each MV
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.analytics_cache_metadata (
            id SERIAL PRIMARY KEY,
            view_name VARCHAR(255) NOT NULL UNIQUE,
            last_refresh_at TIMESTAMPTZ,
            refresh_duration_ms INTEGER,
            row_count INTEGER,
            status VARCHAR(50) DEFAULT ''pending'',
            error_message TEXT,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
        )', schema_name);

    -- ============================================
    -- 2. ML Product Analysis MV
    -- Pre-aggregated product-level metrics for ML Dashboard
    -- ============================================
    EXECUTE format('DROP MATERIALIZED VIEW IF EXISTS %I.mv_ml_product_analysis CASCADE', schema_name);
    
    EXECUTE format('
        CREATE MATERIALIZED VIEW %I.mv_ml_product_analysis AS
        SELECT 
            tenant_id,
            product_id,
            MAX(COALESCE(video_title, product_id)) as product_name,
            MAX(creative_type) as creative_type,
            COALESCE(SUM(cost), 0) as total_cost,
            COALESCE(SUM(gross_revenue), 0) as total_revenue,
            COALESCE(SUM(orders_sku), 0) as total_orders,
            COALESCE(SUM(impressions), 0) as total_impressions,
            COALESCE(SUM(clicks), 0) as total_clicks,
            COUNT(*) as record_count,
            COUNT(DISTINCT period_label) as period_count,
            CASE WHEN SUM(cost) > 0 THEN SUM(gross_revenue)::float / SUM(cost) ELSE 0 END as roas,
            CASE WHEN SUM(impressions) > 0 THEN (SUM(clicks)::float / SUM(impressions)) * 100 ELSE 0 END as ctr,
            CASE WHEN SUM(clicks) > 0 THEN (SUM(orders_sku)::float / SUM(clicks)) * 100 ELSE 0 END as conversion_rate,
            MIN(period_start) as first_period,
            MAX(period_end) as last_period
        FROM %I.tiktok_ads_creative_data
        GROUP BY tenant_id, product_id
    ', schema_name, schema_name);

    EXECUTE format('CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_ml_product_tenant_product 
        ON %I.mv_ml_product_analysis(tenant_id, product_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_mv_ml_product_roas 
        ON %I.mv_ml_product_analysis(roas DESC)', schema_name);

    -- ============================================
    -- 3. ML Portfolio Summary MV
    -- Per-tenant portfolio health metrics
    -- ============================================
    EXECUTE format('DROP MATERIALIZED VIEW IF EXISTS %I.mv_ml_portfolio_summary CASCADE', schema_name);
    
    EXECUTE format('
        CREATE MATERIALIZED VIEW %I.mv_ml_portfolio_summary AS
        SELECT 
            tenant_id,
            COUNT(*) as total_products,
            COUNT(CASE WHEN roas >= 3 THEN 1 END) as high_performers,
            COUNT(CASE WHEN roas >= 1 AND roas < 3 THEN 1 END) as medium_performers,
            COUNT(CASE WHEN roas < 1 AND total_cost > 0 THEN 1 END) as low_performers,
            COUNT(CASE WHEN total_cost = 0 THEN 1 END) as no_spend,
            COALESCE(SUM(total_cost), 0) as total_cost,
            COALESCE(SUM(total_revenue), 0) as total_revenue,
            COALESCE(SUM(total_orders), 0) as total_orders,
            COALESCE(SUM(total_impressions), 0) as total_impressions,
            COALESCE(SUM(total_clicks), 0) as total_clicks,
            SUM(total_revenue) - SUM(total_cost) as total_profit,
            CASE WHEN SUM(total_cost) > 0 THEN SUM(total_revenue)::float / SUM(total_cost) ELSE 0 END as overall_roas,
            CASE WHEN SUM(total_impressions) > 0 THEN (SUM(total_clicks)::float / SUM(total_impressions)) * 100 ELSE 0 END as avg_ctr,
            AVG(roas) as avg_product_roas
        FROM %I.mv_ml_product_analysis
        GROUP BY tenant_id
    ', schema_name, schema_name);

    EXECUTE format('CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_portfolio_tenant 
        ON %I.mv_ml_portfolio_summary(tenant_id)', schema_name);

    -- ============================================
    -- 4. TikTok Ads Dashboard Summary MV
    -- Aggregated metrics for TikTok Ads page
    -- ============================================
    EXECUTE format('DROP MATERIALIZED VIEW IF EXISTS %I.mv_tiktok_ads_summary CASCADE', schema_name);
    
    EXECUTE format('
        CREATE MATERIALIZED VIEW %I.mv_tiktok_ads_summary AS
        SELECT 
            tenant_id,
            COALESCE(SUM(cost), 0) as total_cost,
            COALESCE(SUM(gross_revenue), 0) as total_revenue,
            COALESCE(SUM(orders_sku), 0) as total_orders,
            COALESCE(SUM(impressions), 0) as total_impressions,
            COALESCE(SUM(clicks), 0) as total_clicks,
            COUNT(*) as total_records,
            COUNT(DISTINCT product_id) as total_products,
            COUNT(DISTINCT campaign_id) as total_campaigns,
            CASE WHEN SUM(cost) > 0 THEN SUM(gross_revenue)::float / SUM(cost) ELSE 0 END as overall_roas,
            CASE WHEN SUM(impressions) > 0 THEN (SUM(clicks)::float / SUM(impressions)) * 100 ELSE 0 END as avg_ctr,
            CASE WHEN SUM(orders_sku) > 0 THEN SUM(cost) / SUM(orders_sku) ELSE 0 END as avg_cpo
        FROM %I.tiktok_ads_creative_data
        GROUP BY tenant_id
    ', schema_name, schema_name);

    EXECUTE format('CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_tiktok_summary_tenant 
        ON %I.mv_tiktok_ads_summary(tenant_id)', schema_name);

    -- ============================================
    -- 5. TikTok Ads Period Summary MV
    -- Per-period aggregation for trend charts
    -- ============================================
    EXECUTE format('DROP MATERIALIZED VIEW IF EXISTS %I.mv_tiktok_ads_period_summary CASCADE', schema_name);
    
    EXECUTE format('
        CREATE MATERIALIZED VIEW %I.mv_tiktok_ads_period_summary AS
        SELECT 
            tenant_id,
            period_label,
            MIN(period_start) as period_start,
            MAX(period_end) as period_end,
            COALESCE(SUM(cost), 0) as total_cost,
            COALESCE(SUM(gross_revenue), 0) as total_revenue,
            COALESCE(SUM(orders_sku), 0) as total_orders,
            COALESCE(SUM(impressions), 0) as total_impressions,
            COALESCE(SUM(clicks), 0) as total_clicks,
            COUNT(DISTINCT product_id) as active_products,
            CASE WHEN SUM(cost) > 0 THEN SUM(gross_revenue)::float / SUM(cost) ELSE 0 END as roas
        FROM %I.tiktok_ads_creative_data
        GROUP BY tenant_id, period_label
        ORDER BY period_label DESC
    ', schema_name, schema_name);

    EXECUTE format('CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_tiktok_period_tenant_period 
        ON %I.mv_tiktok_ads_period_summary(tenant_id, period_label)', schema_name);

    -- ============================================
    -- 6. Shopee Ads Dashboard Summary MV
    -- ============================================
    EXECUTE format('DROP MATERIALIZED VIEW IF EXISTS %I.mv_shopee_ads_summary CASCADE', schema_name);
    
    EXECUTE format('
        CREATE MATERIALIZED VIEW %I.mv_shopee_ads_summary AS
        SELECT 
            tenant_id,
            COALESCE(SUM(cost), 0) as total_cost,
            COALESCE(SUM(revenue), 0) as total_revenue,
            COALESCE(SUM(direct_revenue), 0) as total_direct_revenue,
            COALESCE(SUM(units_sold), 0) as total_units_sold,
            COALESCE(SUM(conversions), 0) as total_conversions,
            COALESCE(SUM(impressions), 0) as total_impressions,
            COALESCE(SUM(clicks), 0) as total_clicks,
            COUNT(*) as total_records,
            COUNT(DISTINCT product_id) as total_products,
            CASE WHEN SUM(cost) > 0 THEN SUM(revenue)::float / SUM(cost) ELSE 0 END as overall_roas,
            CASE WHEN SUM(cost) > 0 THEN SUM(direct_revenue)::float / SUM(cost) ELSE 0 END as direct_roas,
            CASE WHEN SUM(impressions) > 0 THEN (SUM(clicks)::float / SUM(impressions)) * 100 ELSE 0 END as avg_ctr
        FROM %I.shopee_ads_product_data
        GROUP BY tenant_id
    ', schema_name, schema_name);

    EXECUTE format('CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_shopee_summary_tenant 
        ON %I.mv_shopee_ads_summary(tenant_id)', schema_name);

    -- ============================================
    -- 7. Shopee Ads Product Analysis MV
    -- ============================================
    EXECUTE format('DROP MATERIALIZED VIEW IF EXISTS %I.mv_shopee_ads_product_analysis CASCADE', schema_name);
    
    EXECUTE format('
        CREATE MATERIALIZED VIEW %I.mv_shopee_ads_product_analysis AS
        SELECT 
            tenant_id,
            product_id,
            MAX(product_name) as product_name,
            COALESCE(SUM(cost), 0) as total_cost,
            COALESCE(SUM(revenue), 0) as total_revenue,
            COALESCE(SUM(direct_revenue), 0) as total_direct_revenue,
            COALESCE(SUM(units_sold), 0) as total_units_sold,
            COALESCE(SUM(conversions), 0) as total_conversions,
            COALESCE(SUM(impressions), 0) as total_impressions,
            COALESCE(SUM(clicks), 0) as total_clicks,
            COUNT(DISTINCT period_label) as period_count,
            CASE WHEN SUM(cost) > 0 THEN SUM(revenue)::float / SUM(cost) ELSE 0 END as roas,
            CASE WHEN SUM(cost) > 0 THEN SUM(direct_revenue)::float / SUM(cost) ELSE 0 END as direct_roas
        FROM %I.shopee_ads_product_data
        GROUP BY tenant_id, product_id
    ', schema_name, schema_name);

    EXECUTE format('CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_shopee_product_tenant_product 
        ON %I.mv_shopee_ads_product_analysis(tenant_id, product_id)', schema_name);

    -- ============================================
    -- Initialize cache metadata
    -- ============================================
    EXECUTE format('
        INSERT INTO %I.analytics_cache_metadata (view_name, status)
        VALUES 
            (''mv_ml_product_analysis'', ''initialized''),
            (''mv_ml_portfolio_summary'', ''initialized''),
            (''mv_tiktok_ads_summary'', ''initialized''),
            (''mv_tiktok_ads_period_summary'', ''initialized''),
            (''mv_shopee_ads_summary'', ''initialized''),
            (''mv_shopee_ads_product_analysis'', ''initialized'')
        ON CONFLICT (view_name) DO UPDATE SET 
            last_refresh_at = CURRENT_TIMESTAMP,
            status = ''refreshed'',
            updated_at = CURRENT_TIMESTAMP
    ', schema_name);

    RAISE NOTICE 'Analytics materialized views created successfully in schema: %', schema_name;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Function to refresh all MVs in a tenant schema
-- ============================================
CREATE OR REPLACE FUNCTION refresh_analytics_mvs(schema_name TEXT)
RETURNS TABLE(view_name TEXT, refresh_time_ms INTEGER, row_count INTEGER) AS $$
DECLARE
    start_time TIMESTAMP;
    end_time TIMESTAMP;
    duration_ms INTEGER;
    rows INTEGER;
BEGIN
    EXECUTE format('SET search_path TO %I, public', schema_name);

    -- Refresh mv_ml_product_analysis
    start_time := clock_timestamp();
    EXECUTE format('REFRESH MATERIALIZED VIEW CONCURRENTLY %I.mv_ml_product_analysis', schema_name);
    end_time := clock_timestamp();
    duration_ms := EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
    EXECUTE format('SELECT COUNT(*) FROM %I.mv_ml_product_analysis', schema_name) INTO rows;
    EXECUTE format('UPDATE %I.analytics_cache_metadata SET last_refresh_at = CURRENT_TIMESTAMP, 
        refresh_duration_ms = $1, row_count = $2, status = ''refreshed'', updated_at = CURRENT_TIMESTAMP
        WHERE view_name = ''mv_ml_product_analysis''', schema_name) USING duration_ms, rows;
    view_name := 'mv_ml_product_analysis'; refresh_time_ms := duration_ms; row_count := rows;
    RETURN NEXT;

    -- Refresh mv_ml_portfolio_summary (depends on mv_ml_product_analysis)
    start_time := clock_timestamp();
    EXECUTE format('REFRESH MATERIALIZED VIEW CONCURRENTLY %I.mv_ml_portfolio_summary', schema_name);
    end_time := clock_timestamp();
    duration_ms := EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
    EXECUTE format('SELECT COUNT(*) FROM %I.mv_ml_portfolio_summary', schema_name) INTO rows;
    EXECUTE format('UPDATE %I.analytics_cache_metadata SET last_refresh_at = CURRENT_TIMESTAMP, 
        refresh_duration_ms = $1, row_count = $2, status = ''refreshed'', updated_at = CURRENT_TIMESTAMP
        WHERE view_name = ''mv_ml_portfolio_summary''', schema_name) USING duration_ms, rows;
    view_name := 'mv_ml_portfolio_summary'; refresh_time_ms := duration_ms; row_count := rows;
    RETURN NEXT;

    -- Refresh mv_tiktok_ads_summary
    start_time := clock_timestamp();
    EXECUTE format('REFRESH MATERIALIZED VIEW CONCURRENTLY %I.mv_tiktok_ads_summary', schema_name);
    end_time := clock_timestamp();
    duration_ms := EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
    EXECUTE format('SELECT COUNT(*) FROM %I.mv_tiktok_ads_summary', schema_name) INTO rows;
    EXECUTE format('UPDATE %I.analytics_cache_metadata SET last_refresh_at = CURRENT_TIMESTAMP, 
        refresh_duration_ms = $1, row_count = $2, status = ''refreshed'', updated_at = CURRENT_TIMESTAMP
        WHERE view_name = ''mv_tiktok_ads_summary''', schema_name) USING duration_ms, rows;
    view_name := 'mv_tiktok_ads_summary'; refresh_time_ms := duration_ms; row_count := rows;
    RETURN NEXT;

    -- Refresh mv_tiktok_ads_period_summary
    start_time := clock_timestamp();
    EXECUTE format('REFRESH MATERIALIZED VIEW CONCURRENTLY %I.mv_tiktok_ads_period_summary', schema_name);
    end_time := clock_timestamp();
    duration_ms := EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
    EXECUTE format('SELECT COUNT(*) FROM %I.mv_tiktok_ads_period_summary', schema_name) INTO rows;
    EXECUTE format('UPDATE %I.analytics_cache_metadata SET last_refresh_at = CURRENT_TIMESTAMP, 
        refresh_duration_ms = $1, row_count = $2, status = ''refreshed'', updated_at = CURRENT_TIMESTAMP
        WHERE view_name = ''mv_tiktok_ads_period_summary''', schema_name) USING duration_ms, rows;
    view_name := 'mv_tiktok_ads_period_summary'; refresh_time_ms := duration_ms; row_count := rows;
    RETURN NEXT;

    -- Refresh mv_shopee_ads_summary
    start_time := clock_timestamp();
    EXECUTE format('REFRESH MATERIALIZED VIEW CONCURRENTLY %I.mv_shopee_ads_summary', schema_name);
    end_time := clock_timestamp();
    duration_ms := EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
    EXECUTE format('SELECT COUNT(*) FROM %I.mv_shopee_ads_summary', schema_name) INTO rows;
    EXECUTE format('UPDATE %I.analytics_cache_metadata SET last_refresh_at = CURRENT_TIMESTAMP, 
        refresh_duration_ms = $1, row_count = $2, status = ''refreshed'', updated_at = CURRENT_TIMESTAMP
        WHERE view_name = ''mv_shopee_ads_summary''', schema_name) USING duration_ms, rows;
    view_name := 'mv_shopee_ads_summary'; refresh_time_ms := duration_ms; row_count := rows;
    RETURN NEXT;

    -- Refresh mv_shopee_ads_product_analysis
    start_time := clock_timestamp();
    EXECUTE format('REFRESH MATERIALIZED VIEW CONCURRENTLY %I.mv_shopee_ads_product_analysis', schema_name);
    end_time := clock_timestamp();
    duration_ms := EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
    EXECUTE format('SELECT COUNT(*) FROM %I.mv_shopee_ads_product_analysis', schema_name) INTO rows;
    EXECUTE format('UPDATE %I.analytics_cache_metadata SET last_refresh_at = CURRENT_TIMESTAMP, 
        refresh_duration_ms = $1, row_count = $2, status = ''refreshed'', updated_at = CURRENT_TIMESTAMP
        WHERE view_name = ''mv_shopee_ads_product_analysis''', schema_name) USING duration_ms, rows;
    view_name := 'mv_shopee_ads_product_analysis'; refresh_time_ms := duration_ms; row_count := rows;
    RETURN NEXT;
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
        RAISE NOTICE 'Creating analytics MVs in schema: %', tenant_schema;
        PERFORM create_analytics_mvs(tenant_schema);
    END LOOP;
    
    RAISE NOTICE 'Analytics materialized views migration completed for all tenant schemas';
END;
$$;
