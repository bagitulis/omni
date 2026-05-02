-- ============================================================
-- Drop orphaned analytics/ads tables after migration to ads-analytics
-- Run: docker cp scripts/drop-analytics-tables.sql omni-postgres:/tmp/ && docker exec omni-postgres psql -U omni -d omni_main -f /tmp/drop-analytics-tables.sql
-- ============================================================

DO $$
DECLARE
    _schema TEXT;
    _table TEXT;
    _mv TEXT;
    _patterns TEXT[] := ARRAY[
        'ml_jobs', 'ml_reports', 'ml_calculation_status', 'ml_score_cache',
        'analytics_cache_metadata',
        'shopee_ads_product_data', 'shopee_ads_upload_batches',
        'tiktok_ads_creative_data', 'tiktok_ads_ml_predictions',
        'tiktok_ads_product_name', 'tiktok_ads_product_names',
        'tiktok_ads_product_summaries', 'tiktok_ads_upload_batches'
    ];
    _mv_patterns TEXT[] := ARRAY[
        'mv_ml_portfolio_summary', 'mv_ml_product_analysis',
        'mv_shopee_ads_product_analysis', 'mv_shopee_ads_summary',
        'mv_tiktok_ads_period_summary', 'mv_tiktok_ads_summary'
    ];
BEGIN
    -- Drop materialized views and tables from ALL tenant and system schemas
    FOR _schema IN
        SELECT schema_name FROM information_schema.schemata
        WHERE schema_name LIKE 'tenant_%' OR schema_name = 'system'
    LOOP
        FOREACH _mv IN ARRAY _mv_patterns LOOP
            IF EXISTS (
                SELECT 1 FROM pg_matviews
                WHERE schemaname = _schema AND matviewname = _mv
            ) THEN
                EXECUTE format('DROP MATERIALIZED VIEW IF EXISTS %I.%I CASCADE', _schema, _mv);
                RAISE NOTICE 'Dropped materialized view: %.%', _schema, _mv;
            END IF;
        END LOOP;

        FOREACH _table IN ARRAY _patterns LOOP
            IF EXISTS (
                SELECT 1 FROM pg_tables
                WHERE schemaname = _schema AND tablename = _table
            ) THEN
                EXECUTE format('DROP TABLE IF EXISTS %I.%I CASCADE', _schema, _table);
                RAISE NOTICE 'Dropped table: %.%', _schema, _table;
            END IF;
        END LOOP;
    END LOOP;
END $$;

-- Verify: list any remaining analytics/ads tables (should be empty)
SELECT schemaname, tablename
FROM pg_tables
WHERE (schemaname LIKE 'tenant_%' OR schemaname = 'system')
  AND (tablename LIKE '%ads%' OR tablename LIKE 'ml_%' OR tablename LIKE 'analytics_cache%')
ORDER BY schemaname, tablename;
