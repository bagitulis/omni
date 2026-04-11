#!/bin/bash
# Script to import TikTok Ads product names into the database
# Run this AFTER build.py smart completes and containers are up
#
# Usage: 
#   docker exec -i omni-backend sh < scripts/import_product_names.sh
#   OR manually via curl after the app is running

TENANT="yumna_bertigamart"
SCHEMA="tenant_${TENANT}"

# Read the product_names.json and generate SQL
cat <<'EOSQL' | PGPASSWORD=$PG_PASSWORD psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DATABASE -q
SET search_path TO tenant_yumna_bertigamart, public;

-- Create table if not exists (GORM should have done this, but just in case)
CREATE TABLE IF NOT EXISTS tiktok_ads_product_names (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    product_id VARCHAR(255) NOT NULL,
    name TEXT,
    category VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Clear existing data for this tenant (idempotent)
DELETE FROM tiktok_ads_product_names WHERE tenant_id = 'yumna_bertigamart';

-- Insert all product names
INSERT INTO tiktok_ads_product_names (tenant_id, product_id, name, category, created_at, updated_at) VALUES
EOSQL

echo "Product names import script generated. Use the API endpoint instead."
