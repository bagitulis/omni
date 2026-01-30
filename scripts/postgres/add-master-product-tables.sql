-- ============================================
-- Migration: Add Master Product Tables
-- Database: omni_main
-- Target Schema: All tenant schemas (tenant_*)
-- Date: 2026-01-30
-- Description: Creates unified master product tables for cross-platform product management
-- ============================================

-- Function to create master product tables in a specific tenant schema
CREATE OR REPLACE FUNCTION create_master_product_tables(schema_name TEXT)
RETURNS VOID AS $$
BEGIN
    -- Set search path to target schema
    EXECUTE format('SET search_path TO %I, public', schema_name);

    -- ============================================
    -- 1. Master Products Table
    -- Unified product data across all platforms
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.master_products (
            id SERIAL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            title VARCHAR(120) NOT NULL,
            description TEXT,
            images JSONB DEFAULT ''[]'',
            status VARCHAR(50) DEFAULT ''draft'' CHECK (status IN (''draft'', ''active'', ''archived'')),
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
        )', schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_master_products_tenant ON %I.master_products(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_master_products_status ON %I.master_products(status)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_master_products_created ON %I.master_products(created_at DESC)', schema_name);

    -- ============================================
    -- 2. Master Product SKUs Table
    -- SKU/variant data for master products
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.master_product_skus (
            id SERIAL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            master_product_id INTEGER NOT NULL,
            seller_sku VARCHAR(255) NOT NULL,
            variant_name VARCHAR(255),
            variant_data JSONB DEFAULT ''{}'',
            price DECIMAL(15,2) DEFAULT 0,
            stock INTEGER DEFAULT 0,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            CONSTRAINT fk_master_product_sku_product 
                FOREIGN KEY (master_product_id) 
                REFERENCES %I.master_products(id) ON DELETE CASCADE
        )', schema_name, schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_master_product_skus_tenant ON %I.master_product_skus(tenant_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_master_product_skus_product ON %I.master_product_skus(master_product_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_master_product_skus_seller_sku ON %I.master_product_skus(seller_sku)', schema_name);
    EXECUTE format('CREATE UNIQUE INDEX IF NOT EXISTS idx_master_product_skus_unique ON %I.master_product_skus(tenant_id, master_product_id, seller_sku)', schema_name);

    -- ============================================
    -- 3. Master Product Platform Links Table
    -- Maps master products/SKUs to platform-specific products
    -- ============================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.master_product_platform_links (
            id SERIAL PRIMARY KEY,
            master_product_id INTEGER NOT NULL,
            master_sku_id INTEGER,
            platform VARCHAR(50) NOT NULL CHECK (platform IN (''shopee'', ''tiktok'', ''lazada'')),
            platform_product_id VARCHAR(255),
            platform_item_id VARCHAR(255),
            platform_sku_id VARCHAR(255),
            sync_status VARCHAR(50) DEFAULT ''pending'' CHECK (sync_status IN (''pending'', ''synced'', ''error'', ''outdated'')),
            last_synced_at TIMESTAMPTZ,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            CONSTRAINT fk_platform_link_product 
                FOREIGN KEY (master_product_id) 
                REFERENCES %I.master_products(id) ON DELETE CASCADE,
            CONSTRAINT fk_platform_link_sku 
                FOREIGN KEY (master_sku_id) 
                REFERENCES %I.master_product_skus(id) ON DELETE CASCADE
        )', schema_name, schema_name, schema_name);

    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_platform_links_product ON %I.master_product_platform_links(master_product_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_platform_links_sku ON %I.master_product_platform_links(master_sku_id)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_platform_links_platform ON %I.master_product_platform_links(platform)', schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_platform_links_sync_status ON %I.master_product_platform_links(sync_status)', schema_name);
    EXECUTE format('CREATE UNIQUE INDEX IF NOT EXISTS idx_platform_links_unique ON %I.master_product_platform_links(platform, platform_product_id) WHERE platform_product_id IS NOT NULL', schema_name);

    RAISE NOTICE 'Master product tables created successfully in schema: %', schema_name;
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
        RAISE NOTICE 'Creating master product tables in schema: %', tenant_schema;
        PERFORM create_master_product_tables(tenant_schema);
    END LOOP;
END;
$$;

-- ============================================
-- Verification query (run manually to verify)
-- ============================================
-- SELECT table_name FROM information_schema.tables 
-- WHERE table_schema LIKE 'tenant_%' AND table_name LIKE 'master_product%';
