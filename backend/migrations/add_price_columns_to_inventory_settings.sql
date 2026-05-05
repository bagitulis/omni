-- Add price column configuration fields to inventory_settings
-- These columns allow per-platform price column customization for inventory sync

ALTER TABLE inventory_settings ADD COLUMN IF NOT EXISTS price_column VARCHAR(100) DEFAULT 'HARGA';
ALTER TABLE inventory_settings ADD COLUMN IF NOT EXISTS price_column_shopee VARCHAR(100) DEFAULT 'HARGA_SHOPEE';
ALTER TABLE inventory_settings ADD COLUMN IF NOT EXISTS price_column_tiktok VARCHAR(100) DEFAULT 'HARGA_TIKTOK';
ALTER TABLE inventory_settings ADD COLUMN IF NOT EXISTS price_column_lazada VARCHAR(100) DEFAULT 'HARGA_LAZADA';
