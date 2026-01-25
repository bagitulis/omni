-- ============================================
-- Migration: Add Analytics Tables for Shopee & TikTok Escrow
-- Database: omni_main
-- Target Schema: tenant_yumna_bertigamart (and other tenant schemas)
-- Date: 2026-01-25
-- ============================================

-- Function to create analytics tables in a specific tenant schema
CREATE OR REPLACE FUNCTION create_analytics_tables(schema_name TEXT)
RETURNS VOID AS $$
BEGIN
    -- Set search path to target schema
    EXECUTE format('SET search_path TO %I, public', schema_name);

    -- ============================================
    -- 1. Analytics Settings Table
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.analytics_settings (
            id VARCHAR(255) PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            platform VARCHAR(50) DEFAULT ''shopee'',
            price_column VARCHAR(100) DEFAULT ''HARGA'',
            formula_deduction DOUBLE PRECISION DEFAULT 1500,
            formula_multiplier DOUBLE PRECISION DEFAULT 0.84,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(tenant_id, platform)
        )', schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_analytics_settings_tenant ON %I.analytics_settings(tenant_id)', schema_name);

    -- ============================================
    -- 2. Shopee Escrow Sync Table
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_escrow_sync (
            id VARCHAR(255) PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            month INTEGER NOT NULL,
            year INTEGER NOT NULL,
            total_orders INTEGER DEFAULT 0,
            synced_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(tenant_id, month, year)
        )', schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_escrow_sync_tenant ON %I.shopee_escrow_sync(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_escrow_sync_month_year ON %I.shopee_escrow_sync(month, year)', schema_name);

    -- ============================================
    -- 3. Shopee Escrow Order Table
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_escrow_orders (
            id VARCHAR(255) PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            order_sn VARCHAR(255) NOT NULL,
            month INTEGER NOT NULL,
            year INTEGER NOT NULL,
            order_date TIMESTAMPTZ,
            buyer_user_name VARCHAR(255),
            escrow_amount DOUBLE PRECISION DEFAULT 0,
            commission_fee DOUBLE PRECISION DEFAULT 0,
            service_fee DOUBLE PRECISION DEFAULT 0,
            seller_processing_fee DOUBLE PRECISION DEFAULT 0,
            buyer_paid_shipping_fee DOUBLE PRECISION DEFAULT 0,
            actual_shipping_fee DOUBLE PRECISION DEFAULT 0,
            shopee_shipping_rebate DOUBLE PRECISION DEFAULT 0,
            estimated_shipping_fee DOUBLE PRECISION DEFAULT 0,
            buyer_total_amount DOUBLE PRECISION DEFAULT 0,
            buyer_payment_method VARCHAR(100),
            raw_order_income TEXT,
            raw_buyer_payment_info TEXT,
            synced_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(tenant_id, order_sn)
        )', schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_escrow_orders_tenant ON %I.shopee_escrow_orders(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_escrow_orders_month_year ON %I.shopee_escrow_orders(month, year)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_escrow_orders_order_sn ON %I.shopee_escrow_orders(order_sn)', schema_name);

    -- ============================================
    -- 4. Shopee Escrow Item Table
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_escrow_items (
            id VARCHAR(255) PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            escrow_order_id VARCHAR(255) NOT NULL,
            order_id VARCHAR(255),
            order_sn VARCHAR(255),
            month INTEGER,
            year INTEGER,
            item_id BIGINT,
            model_id BIGINT,
            sku VARCHAR(255),
            model_sku VARCHAR(255),
            item_name TEXT,
            model_name VARCHAR(500),
            quantity INTEGER DEFAULT 0,
            original_price DOUBLE PRECISION DEFAULT 0,
            selling_price DOUBLE PRECISION DEFAULT 0,
            discounted_price DOUBLE PRECISION DEFAULT 0,
            seller_discount DOUBLE PRECISION DEFAULT 0,
            shopee_discount DOUBLE PRECISION DEFAULT 0,
            discount_from_coin DOUBLE PRECISION DEFAULT 0,
            discount_from_voucher_seller DOUBLE PRECISION DEFAULT 0,
            discount_from_voucher_shopee DOUBLE PRECISION DEFAULT 0,
            ams_commission_fee DOUBLE PRECISION DEFAULT 0,
            seller_order_processing_fee DOUBLE PRECISION DEFAULT 0,
            raw_item_data TEXT,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            CONSTRAINT fk_shopee_escrow_order FOREIGN KEY (escrow_order_id) 
                REFERENCES %I.shopee_escrow_orders(id) ON DELETE CASCADE
        )', schema_name, schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_escrow_items_tenant ON %I.shopee_escrow_items(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_escrow_items_order_id ON %I.shopee_escrow_items(escrow_order_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_escrow_items_sku ON %I.shopee_escrow_items(sku)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_shopee_escrow_items_model_sku ON %I.shopee_escrow_items(model_sku)', schema_name);

    -- ============================================
    -- 5. TikTok Escrow Sync Table
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_escrow_sync (
            id VARCHAR(255) PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            month INTEGER NOT NULL,
            year INTEGER NOT NULL,
            total_orders INTEGER DEFAULT 0,
            synced_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(tenant_id, month, year)
        )', schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_escrow_sync_tenant ON %I.tiktok_escrow_sync(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_escrow_sync_month_year ON %I.tiktok_escrow_sync(month, year)', schema_name);

    -- ============================================
    -- 6. TikTok Escrow Order Table
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_escrow_orders (
            id VARCHAR(255) PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            order_id VARCHAR(255) NOT NULL,
            month INTEGER NOT NULL,
            year INTEGER NOT NULL,
            order_status VARCHAR(100),
            order_date TIMESTAMPTZ,
            buyer_name VARCHAR(255),
            transaction_id VARCHAR(255),
            transaction_type VARCHAR(100),
            statement_time TIMESTAMPTZ,
            total_settlement_amount DOUBLE PRECISION DEFAULT 0,
            product_revenue DOUBLE PRECISION DEFAULT 0,
            platform_commission DOUBLE PRECISION DEFAULT 0,
            transaction_fee DOUBLE PRECISION DEFAULT 0,
            shipping_fee_customer_paid DOUBLE PRECISION DEFAULT 0,
            shipping_fee_actual DOUBLE PRECISION DEFAULT 0,
            shipping_fee_platform_discount DOUBLE PRECISION DEFAULT 0,
            seller_shipping_discount DOUBLE PRECISION DEFAULT 0,
            refund_amount DOUBLE PRECISION DEFAULT 0,
            adjustment DOUBLE PRECISION DEFAULT 0,
            buyer_total_amount DOUBLE PRECISION DEFAULT 0,
            currency VARCHAR(10) DEFAULT ''IDR'',
            raw_transaction_data TEXT,
            raw_order_data TEXT,
            synced_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(tenant_id, order_id)
        )', schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_escrow_orders_tenant ON %I.tiktok_escrow_orders(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_escrow_orders_month_year ON %I.tiktok_escrow_orders(month, year)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_escrow_orders_order_id ON %I.tiktok_escrow_orders(order_id)', schema_name);

    -- ============================================
    -- 7. TikTok Escrow Item Table
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_escrow_items (
            id VARCHAR(255) PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            escrow_order_id VARCHAR(255) NOT NULL,
            order_id VARCHAR(255),
            product_id VARCHAR(255),
            product_name TEXT,
            sku_id VARCHAR(255),
            seller_sku VARCHAR(255),
            quantity INTEGER DEFAULT 0,
            sale_price DOUBLE PRECISION DEFAULT 0,
            original_price DOUBLE PRECISION DEFAULT 0,
            subtotal_after_seller_discount DOUBLE PRECISION DEFAULT 0,
            platform_discount DOUBLE PRECISION DEFAULT 0,
            seller_discount DOUBLE PRECISION DEFAULT 0,
            commission DOUBLE PRECISION DEFAULT 0,
            transaction_fee_item DOUBLE PRECISION DEFAULT 0,
            settlement_amount DOUBLE PRECISION DEFAULT 0,
            raw_item_data TEXT,
            synced_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            CONSTRAINT fk_tiktok_escrow_order FOREIGN KEY (escrow_order_id) 
                REFERENCES %I.tiktok_escrow_orders(id) ON DELETE CASCADE
        )', schema_name, schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_escrow_items_tenant ON %I.tiktok_escrow_items(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_escrow_items_order_id ON %I.tiktok_escrow_items(escrow_order_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_tiktok_escrow_items_seller_sku ON %I.tiktok_escrow_items(seller_sku)', schema_name);

    RAISE NOTICE 'Analytics tables created successfully in schema: %', schema_name;
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
        RAISE NOTICE 'Creating analytics tables in schema: %', tenant_schema;
        PERFORM create_analytics_tables(tenant_schema);
    END LOOP;
    
    RAISE NOTICE '✅ Analytics tables migration completed for all tenant schemas';
END;
$$;

-- ============================================
-- Cleanup: Drop the function after use (optional)
-- ============================================
-- DROP FUNCTION IF EXISTS create_analytics_tables(TEXT);
