-- ============================================================
-- SYSTEM.DB SCHEMA - GLOBAL CONFIG ONLY
-- ============================================================
-- 
-- STRICT RULES - DO NOT VIOLATE:
-- 1. Only GlobalConfig table is allowed
-- 2. NO tenant-specific data (tokens, users, orders)  
-- 3. NO OAuthState/OAuthLog (these go to tenant DBs)
-- 4. Access via GlobalConfigService ONLY
--
-- Location: backend/config/databases/system.db
-- ============================================================

-- Global configuration for all tenants
-- Used for: Partner IDs, Partner Keys, App Keys, App Secrets
CREATE TABLE IF NOT EXISTS GlobalConfig (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    platform    TEXT NOT NULL,           -- shopee, tiktok, lazada
    configKey   TEXT NOT NULL,           -- partnerId, partnerKey, appKey, appSecret
    configValue TEXT NOT NULL DEFAULT '',
    isEncrypted INTEGER DEFAULT 0,       -- 1 = encrypted, 0 = plain
    description TEXT,                    -- Human readable description
    createdAt   TEXT DEFAULT CURRENT_TIMESTAMP,
    updatedAt   TEXT DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(platform, configKey)
);

-- Indices for fast lookup
CREATE INDEX IF NOT EXISTS idx_globalconfig_platform ON GlobalConfig(platform);
CREATE INDEX IF NOT EXISTS idx_globalconfig_key ON GlobalConfig(platform, configKey);

-- ============================================================
-- EXAMPLE DATA (for reference only)
-- ============================================================
-- INSERT INTO GlobalConfig (platform, configKey, configValue, isEncrypted, description)
-- VALUES 
--   ('shopee', 'partnerId', '123456', 0, 'Shopee Partner ID'),
--   ('shopee', 'partnerKey', 'enc:xxxxx', 1, 'Shopee Partner Key (encrypted)'),
--   ('tiktok', 'appKey', 'abcdef', 0, 'TikTok App Key'),
--   ('tiktok', 'appSecret', 'enc:xxxxx', 1, 'TikTok App Secret (encrypted)'),
--   ('lazada', 'appKey', 'ghijkl', 0, 'Lazada App Key'),
--   ('lazada', 'appSecret', 'enc:xxxxx', 1, 'Lazada App Secret (encrypted)');

-- ============================================================
-- FORBIDDEN TABLES (DO NOT ADD TO SYSTEM.DB!)
-- ============================================================
-- These belong in TENANT databases ({tenant}_bertigamart.db):
-- - User (authentication per tenant)
-- - PlatformConfig (access/refresh tokens per tenant)
-- - OAuthState (OAuth flow state per tenant)
-- - OAuthLog (OAuth logs per tenant)
-- - ShopeeOrder, TiktokOrder, LazadaOrder
-- - Products, Inventory, etc.
