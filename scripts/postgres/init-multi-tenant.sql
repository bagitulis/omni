-- ============================================
-- PostgreSQL Multi-Tenant Initialization Script
-- Database: omni_main
-- Version: 2.0
-- Date: 2026-01-26
-- ============================================
-- This script initializes a fresh PostgreSQL database with:
-- 1. Required extensions (uuid-ossp)
-- 2. System schema with global tables
-- 3. Function to create tenant schemas
-- 4. Default tenants
-- ============================================

-- ============================================
-- 0. Setup & Extensions
-- ============================================
\set ON_ERROR_STOP off

-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- 1. System Schema
-- ============================================
CREATE SCHEMA IF NOT EXISTS system;

-- ============================================
-- 1.1 System Tables
-- ============================================

-- Tenants Registry
CREATE TABLE IF NOT EXISTS system.tenants (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL UNIQUE,
    shop_name VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    db_schema VARCHAR(255),
    settings JSONB,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Global Configuration
CREATE TABLE IF NOT EXISTS system.global_config (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    platform VARCHAR(100) NOT NULL,
    config_key VARCHAR(255) NOT NULL,
    config_value TEXT,
    is_encrypted BOOLEAN DEFAULT false,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- System Users
CREATE TABLE IF NOT EXISTS system.users (
    id VARCHAR(255) DEFAULT gen_random_uuid()::text NOT NULL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'developer',
    failed_login_attempts INTEGER DEFAULT 0,
    account_locked_until TIMESTAMPTZ,
    last_failed_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Audit Log
CREATE TABLE IF NOT EXISTS system.audit_log (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(255),
    user_id VARCHAR(255),
    action VARCHAR(255) NOT NULL,
    entity_type VARCHAR(255),
    entity_id VARCHAR(255),
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- OAuth Logs
CREATE TABLE IF NOT EXISTS system.oauth_logs (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    platform VARCHAR(50) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    shop_id VARCHAR(255),
    code TEXT,
    state VARCHAR(255),
    status VARCHAR(50) DEFAULT 'received',
    error_msg TEXT,
    metadata TEXT,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- OAuth States
CREATE TABLE IF NOT EXISTS system.oauth_states (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    platform VARCHAR(50) NOT NULL,
    state VARCHAR(255) NOT NULL UNIQUE,
    redirect_url TEXT,
    metadata TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 1.2 System Indexes
-- ============================================
CREATE INDEX IF NOT EXISTS idx_audit_log_tenant ON system.audit_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_created ON system.audit_log(created_at);
CREATE INDEX IF NOT EXISTS idx_oauth_logs_tenant_id ON system.oauth_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_oauth_logs_platform ON system.oauth_logs(platform);
CREATE INDEX IF NOT EXISTS idx_oauth_logs_event_type ON system.oauth_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_oauth_logs_status ON system.oauth_logs(status);
CREATE INDEX IF NOT EXISTS idx_oauth_logs_created_at ON system.oauth_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_oauth_states_tenant_id ON system.oauth_states(tenant_id);
CREATE INDEX IF NOT EXISTS idx_oauth_states_platform ON system.oauth_states(platform);
CREATE INDEX IF NOT EXISTS idx_oauth_states_expires_at ON system.oauth_states(expires_at);
CREATE INDEX IF NOT EXISTS idx_system_users_email ON system.users(email);
CREATE INDEX IF NOT EXISTS idx_system_users_username ON system.users(username);

-- ============================================
-- 2. Tenant Schema Creation Function
-- ============================================
CREATE OR REPLACE FUNCTION create_tenant_schema(schema_name TEXT)
RETURNS VOID AS $$
BEGIN
    -- Create schema
    EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', schema_name);
    
    -- Set search path
    EXECUTE format('SET search_path TO %I, public', schema_name);
    
    -- ========================================
    -- Users & Settings
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.users (
            id VARCHAR(255) DEFAULT gen_random_uuid()::text NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            username VARCHAR(255) NOT NULL UNIQUE,
            email VARCHAR(255) NOT NULL UNIQUE,
            password VARCHAR(255) NOT NULL,
            role VARCHAR(50) DEFAULT ''user'',
            failed_login_attempts INTEGER DEFAULT 0,
            account_locked_until TIMESTAMPTZ,
            last_failed_login TIMESTAMPTZ,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.platform_configs (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            platform VARCHAR(100) NOT NULL,
            shop_id VARCHAR(255),
            shop_name VARCHAR(500),
            access_token TEXT,
            refresh_token TEXT,
            token_expires_at TIMESTAMPTZ,
            shop_cipher TEXT,
            is_connected BOOLEAN DEFAULT false,
            auth_status VARCHAR(50) DEFAULT ''disconnected'',
            last_sync_at TIMESTAMPTZ,
            settings JSONB,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.google_sheets_settings (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            spreadsheet_id VARCHAR(500),
            selected_sheet VARCHAR(500),
            manual_mode BOOLEAN DEFAULT false,
            wallet_spreadsheet_id VARCHAR(500),
            shipping_spreadsheet_id VARCHAR(500),
            inventory_spreadsheet_id VARCHAR(500),
            inventory_sheet_name VARCHAR(500),
            inventory_selected_columns TEXT,
            inventory_available_worksheets TEXT,
            wallet_available_worksheets TEXT,
            shipping_available_worksheets TEXT,
            order_available_worksheets TEXT,
            order_spreadsheet_id VARCHAR(500),
            available_spreadsheets TEXT,
            available_worksheets TEXT,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.filter_preferences (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            platform VARCHAR(100) NOT NULL,
            tab VARCHAR(100) DEFAULT ''master'',
            visible_columns TEXT,
            column_filters TEXT,
            search_query TEXT DEFAULT '''',
            locked_columns TEXT DEFAULT ''[]'',
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.analytics_settings (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            platform VARCHAR(100) DEFAULT ''shopee'',
            price_column VARCHAR(255) DEFAULT ''HARGA'',
            sku_column VARCHAR(255),
            formula_deduction DOUBLE PRECISION DEFAULT 1500,
            formula_multiplier DOUBLE PRECISION DEFAULT 0.84,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.wholesale_settings (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            platform VARCHAR(100) DEFAULT ''shopee'',
            min_qty INTEGER DEFAULT 2,
            discount_pct DOUBLE PRECISION DEFAULT 5,
            enabled BOOLEAN DEFAULT false,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    -- ========================================
    -- Shopee Tables
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_orders (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            order_sn VARCHAR(255) NOT NULL UNIQUE,
            shop_id BIGINT,
            region VARCHAR(50),
            currency VARCHAR(10),
            cod BOOLEAN DEFAULT false,
            total_amount DOUBLE PRECISION,
            shipping_carrier VARCHAR(255),
            payment_method VARCHAR(255),
            estimated_shipping_fee DOUBLE PRECISION,
            message_to_seller TEXT,
            create_time BIGINT,
            update_time BIGINT,
            days_to_ship INTEGER,
            ship_by_date BIGINT,
            buyer_user_id BIGINT,
            buyer_username VARCHAR(255),
            recipient_address JSONB,
            actual_shipping_fee DOUBLE PRECISION,
            goods_to_declare BOOLEAN DEFAULT false,
            note TEXT,
            note_update_time BIGINT,
            item_list JSONB,
            invoice_data JSONB,
            checkout_shipping_carrier VARCHAR(255),
            reverse_shipping_fee DOUBLE PRECISION,
            order_chargeable_weight_gram INTEGER,
            edt JSONB,
            package_list JSONB,
            dropshipper VARCHAR(255),
            dropshipper_phone VARCHAR(50),
            split_up BOOLEAN DEFAULT false,
            buyer_cancel_reason TEXT,
            cancel_by VARCHAR(50),
            cancel_reason TEXT,
            actual_shipping_fee_confirmed BOOLEAN DEFAULT false,
            buyer_cpf_id VARCHAR(255),
            fulfillment_flag VARCHAR(50),
            pickup_done_time BIGINT,
            escrow_amount DOUBLE PRECISION,
            escrow_tax DOUBLE PRECISION,
            escrow_release_date TIMESTAMPTZ,
            order_status VARCHAR(50),
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_order_items (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            order_sn VARCHAR(255) NOT NULL,
            item_id BIGINT,
            item_name TEXT,
            item_sku VARCHAR(255),
            model_id BIGINT,
            model_name VARCHAR(500),
            model_sku VARCHAR(255),
            model_quantity_purchased INTEGER,
            model_original_price DOUBLE PRECISION,
            model_discounted_price DOUBLE PRECISION,
            wholesale BOOLEAN DEFAULT false,
            weight DOUBLE PRECISION,
            add_on_deal BOOLEAN DEFAULT false,
            main_item BOOLEAN DEFAULT false,
            add_on_deal_id BIGINT,
            promotion_type VARCHAR(100),
            promotion_id BIGINT,
            order_item_id BIGINT,
            promotion_group_id BIGINT,
            image_info JSONB,
            product_location_id TEXT,
            is_prescription_item BOOLEAN DEFAULT false,
            is_b2c_owned_item BOOLEAN DEFAULT false,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_products (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            item_id BIGINT NOT NULL UNIQUE,
            shop_id BIGINT,
            category_id BIGINT,
            item_name TEXT,
            description TEXT,
            item_sku VARCHAR(255),
            create_time BIGINT,
            update_time BIGINT,
            image JSONB,
            weight VARCHAR(50),
            dimension JSONB,
            pre_order JSONB,
            condition VARCHAR(50),
            has_model BOOLEAN DEFAULT false,
            item_status VARCHAR(50),
            brand JSONB,
            item_dangerous INTEGER,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_skus (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            item_id BIGINT NOT NULL,
            model_id BIGINT NOT NULL,
            model_sku VARCHAR(255),
            model_name VARCHAR(500),
            price_info JSONB,
            stock_info JSONB,
            stock_info_v2 JSONB,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now(),
            UNIQUE(item_id, model_id)
        )', schema_name);

    -- ========================================
    -- Shopee Escrow Tables
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_escrow_sync (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            sync_date DATE NOT NULL,
            status VARCHAR(50) DEFAULT ''pending'',
            total_orders INTEGER DEFAULT 0,
            synced_orders INTEGER DEFAULT 0,
            failed_orders INTEGER DEFAULT 0,
            error_message TEXT,
            started_at TIMESTAMPTZ,
            completed_at TIMESTAMPTZ,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_escrow_orders (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            order_sn VARCHAR(255) NOT NULL UNIQUE,
            buyer_user_id BIGINT,
            buyer_user_name VARCHAR(255),
            buyer_total_amount DOUBLE PRECISION,
            escrow_amount DOUBLE PRECISION,
            escrow_tax DOUBLE PRECISION,
            actual_shipping_fee DOUBLE PRECISION,
            order_income DOUBLE PRECISION,
            order_income_after_adjustment DOUBLE PRECISION,
            commission_fee DOUBLE PRECISION,
            service_fee DOUBLE PRECISION,
            drc_adjustable_refund DOUBLE PRECISION,
            final_shipping_fee DOUBLE PRECISION,
            final_escrow_product_gst DOUBLE PRECISION,
            final_escrow_shipping_gst DOUBLE PRECISION,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_escrow_items (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            escrow_order_id VARCHAR(255) NOT NULL,
            order_sn VARCHAR(255) NOT NULL,
            item_id BIGINT,
            item_name TEXT,
            item_sku VARCHAR(255),
            model_id BIGINT,
            model_name VARCHAR(500),
            model_sku VARCHAR(255),
            quantity INTEGER DEFAULT 1,
            original_price DOUBLE PRECISION,
            discounted_price DOUBLE PRECISION,
            seller_discount DOUBLE PRECISION,
            is_main_item BOOLEAN DEFAULT false,
            is_add_on_deal BOOLEAN DEFAULT false,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    -- ========================================
    -- Shopee Ads Tables
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_ads_upload_batches (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            file_name VARCHAR(500),
            status VARCHAR(50) DEFAULT ''pending'',
            total_rows INTEGER DEFAULT 0,
            processed_rows INTEGER DEFAULT 0,
            success_rows INTEGER DEFAULT 0,
            failed_rows INTEGER DEFAULT 0,
            errors JSONB,
            started_at TIMESTAMPTZ,
            completed_at TIMESTAMPTZ,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.shopee_ads_product_data (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            batch_id VARCHAR(255),
            date DATE NOT NULL,
            campaign_id BIGINT,
            campaign_name VARCHAR(500),
            campaign_type VARCHAR(100),
            product_id BIGINT,
            product_name TEXT,
            sku VARCHAR(255),
            impression BIGINT DEFAULT 0,
            clicks BIGINT DEFAULT 0,
            ctr DOUBLE PRECISION DEFAULT 0,
            direct_conversions INTEGER DEFAULT 0,
            broad_conversions INTEGER DEFAULT 0,
            direct_gmv DOUBLE PRECISION DEFAULT 0,
            broad_gmv DOUBLE PRECISION DEFAULT 0,
            expense DOUBLE PRECISION DEFAULT 0,
            direct_roas DOUBLE PRECISION DEFAULT 0,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    -- ========================================
    -- Lazada Tables
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.lazada_orders (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            order_id BIGINT NOT NULL UNIQUE,
            order_number VARCHAR(255),
            customer_first_name VARCHAR(255),
            customer_last_name VARCHAR(255),
            created_at TIMESTAMPTZ,
            updated_at TIMESTAMPTZ,
            address_billing JSONB,
            address_shipping JSONB,
            items_count INTEGER,
            payment_method VARCHAR(255),
            price DOUBLE PRECISION,
            delivery_info TEXT,
            statuses TEXT,
            voucher DOUBLE PRECISION,
            voucher_seller DOUBLE PRECISION,
            voucher_code VARCHAR(255),
            voucher_code_seller VARCHAR(255),
            national_registration_number VARCHAR(255),
            branch_number VARCHAR(255),
            tax_code VARCHAR(255),
            extra_attributes TEXT,
            promised_shipping_times TEXT,
            warehouse_code VARCHAR(255),
            shipping_fee DOUBLE PRECISION,
            shipping_fee_original DOUBLE PRECISION,
            shipping_fee_discount_seller DOUBLE PRECISION,
            shipping_fee_discount_platform DOUBLE PRECISION,
            gift_option BOOLEAN DEFAULT false,
            gift_message TEXT,
            remarks TEXT
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.lazada_order_items (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            order_id BIGINT NOT NULL,
            order_item_id BIGINT,
            shop_id VARCHAR(255),
            shop_sku VARCHAR(255),
            sku_id BIGINT,
            sku VARCHAR(255),
            name TEXT,
            variation VARCHAR(500),
            item_price DOUBLE PRECISION,
            paid_price DOUBLE PRECISION,
            wallet_credits DOUBLE PRECISION,
            tax_amount DOUBLE PRECISION,
            shipping_fee_original DOUBLE PRECISION,
            shipping_fee_discount_seller DOUBLE PRECISION,
            shipping_fee_discount_platform DOUBLE PRECISION,
            shipping_amount DOUBLE PRECISION,
            voucher_seller DOUBLE PRECISION,
            voucher_code_seller VARCHAR(255),
            voucher_platform DOUBLE PRECISION,
            voucher_code VARCHAR(255),
            status VARCHAR(100),
            tracking_code VARCHAR(255),
            tracking_code_pre VARCHAR(255),
            shipment_provider VARCHAR(255),
            shipping_provider_type VARCHAR(100),
            is_digital INTEGER DEFAULT 0,
            is_fbl INTEGER DEFAULT 0,
            digital_delivery_info TEXT,
            reason TEXT,
            reason_detail TEXT,
            product_main_image TEXT,
            product_detail_url TEXT,
            promised_shipping_time TIMESTAMPTZ,
            cancel_return_initiator VARCHAR(100),
            sla_time_stamp TIMESTAMPTZ,
            warehouse_code VARCHAR(255),
            buyer_id BIGINT,
            invoice_number VARCHAR(255),
            extra_attributes TEXT,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.lazada_products (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            item_id BIGINT NOT NULL UNIQUE,
            primary_category BIGINT,
            attributes JSONB,
            skus JSONB,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.lazada_skus (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            item_id BIGINT NOT NULL,
            sku_id BIGINT NOT NULL UNIQUE,
            seller_sku VARCHAR(255),
            shop_sku VARCHAR(255),
            url TEXT,
            package_height VARCHAR(50),
            package_length VARCHAR(50),
            package_width VARCHAR(50),
            package_weight VARCHAR(50),
            package_content TEXT,
            images JSONB,
            special_from_time TIMESTAMPTZ,
            special_to_time TIMESTAMPTZ,
            special_from_date DATE,
            special_to_date DATE,
            special_time_format VARCHAR(100),
            special_price DOUBLE PRECISION,
            price DOUBLE PRECISION,
            quantity INTEGER,
            saleable_stock INTEGER,
            non_saleable_stock INTEGER,
            fulfillment_stock JSONB,
            status VARCHAR(50),
            compatible_vehicles TEXT,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    -- ========================================
    -- TikTok Tables
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_orders (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            order_id VARCHAR(255) NOT NULL UNIQUE,
            shop_id VARCHAR(255),
            buyer_uid VARCHAR(255),
            buyer_message TEXT,
            seller_note TEXT,
            payment_info JSONB,
            recipient_address JSONB,
            shipping_provider VARCHAR(255),
            shipping_provider_id VARCHAR(255),
            tracking_number VARCHAR(255),
            delivery_option VARCHAR(100),
            delivery_type VARCHAR(100),
            delivery_sla BIGINT,
            shipping_type VARCHAR(100),
            shipping_due_time BIGINT,
            fulfillment_type VARCHAR(100),
            split_or_combine_tag VARCHAR(50),
            line_items JSONB,
            packages JSONB,
            payment_method VARCHAR(255),
            collection_time BIGINT,
            order_status VARCHAR(50),
            paid_time BIGINT,
            rts_time BIGINT,
            delivery_time BIGINT,
            cancel_reason TEXT,
            cancel_user VARCHAR(100),
            cancel_time BIGINT,
            update_time BIGINT,
            create_time BIGINT,
            warehouse_id VARCHAR(255),
            is_cod BOOLEAN DEFAULT false,
            is_on_hold_order BOOLEAN DEFAULT false,
            is_replacement_order BOOLEAN DEFAULT false,
            replaced_order_id VARCHAR(255),
            cpf VARCHAR(255),
            request_cancel_time BIGINT,
            is_buyer_request_cancel BOOLEAN DEFAULT false,
            rtc_time BIGINT,
            is_sample_order BOOLEAN DEFAULT false,
            has_updated_recipient_address BOOLEAN DEFAULT false,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_order_items (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            order_id VARCHAR(255) NOT NULL,
            sku_id VARCHAR(255),
            product_id VARCHAR(255),
            product_name TEXT,
            sku_name VARCHAR(500),
            seller_sku VARCHAR(255),
            sku_image TEXT,
            quantity INTEGER DEFAULT 1,
            sale_price DOUBLE PRECISION,
            original_price DOUBLE PRECISION,
            platform_discount DOUBLE PRECISION DEFAULT 0,
            seller_discount DOUBLE PRECISION DEFAULT 0,
            sku_type VARCHAR(50),
            is_gift BOOLEAN DEFAULT false,
            display_status VARCHAR(50),
            cancel_reason TEXT,
            cancel_user VARCHAR(100),
            rts_time BIGINT,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_products (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            product_id VARCHAR(255) NOT NULL UNIQUE,
            shop_id VARCHAR(255),
            title TEXT,
            description TEXT,
            category_chains JSONB,
            brand JSONB,
            main_images JSONB,
            skus JSONB,
            package_dimensions JSONB,
            package_weight JSONB,
            product_status VARCHAR(50),
            create_time BIGINT,
            update_time BIGINT,
            is_cod_open BOOLEAN DEFAULT false,
            certifications JSONB,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_skus (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            product_id VARCHAR(255) NOT NULL,
            sku_id VARCHAR(255) NOT NULL UNIQUE,
            seller_sku VARCHAR(255),
            sales_attributes JSONB,
            price JSONB,
            stock_infos JSONB,
            external_sku_id VARCHAR(255),
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    -- ========================================
    -- TikTok Escrow Tables
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_escrow_sync (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            sync_date DATE NOT NULL,
            status VARCHAR(50) DEFAULT ''pending'',
            total_orders INTEGER DEFAULT 0,
            synced_orders INTEGER DEFAULT 0,
            failed_orders INTEGER DEFAULT 0,
            error_message TEXT,
            started_at TIMESTAMPTZ,
            completed_at TIMESTAMPTZ,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_escrow_orders (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            order_id VARCHAR(255) NOT NULL UNIQUE,
            statement_id VARCHAR(255),
            statement_time BIGINT,
            settlement_time BIGINT,
            sku_settlement_price DOUBLE PRECISION,
            sku_platform_discount DOUBLE PRECISION,
            sku_seller_discount DOUBLE PRECISION,
            subtotal_after_seller_discounts DOUBLE PRECISION,
            subtotal_before_discounts DOUBLE PRECISION,
            customer_paid_shipping_fee DOUBLE PRECISION,
            shipping_fee_seller_discount DOUBLE PRECISION,
            shipping_fee_platform_discount DOUBLE PRECISION,
            small_order_fee DOUBLE PRECISION,
            transaction_fee DOUBLE PRECISION,
            referral_fee DOUBLE PRECISION,
            refund_subtotal DOUBLE PRECISION,
            refund_shipping_fee DOUBLE PRECISION,
            final_shipping_fee DOUBLE PRECISION,
            settlement_amount DOUBLE PRECISION,
            adjustment_amount DOUBLE PRECISION,
            adjustment_reason TEXT,
            payment_method VARCHAR(255),
            currency VARCHAR(10),
            revenue DOUBLE PRECISION,
            cost DOUBLE PRECISION,
            fees_total DOUBLE PRECISION,
            net_sales DOUBLE PRECISION,
            acos DOUBLE PRECISION,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_escrow_items (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            escrow_order_id VARCHAR(255) NOT NULL,
            order_id VARCHAR(255) NOT NULL,
            sku_id VARCHAR(255),
            product_id VARCHAR(255),
            product_name TEXT,
            sku_name VARCHAR(500),
            seller_sku VARCHAR(255),
            quantity INTEGER DEFAULT 1,
            sale_price DOUBLE PRECISION,
            original_price DOUBLE PRECISION,
            platform_discount DOUBLE PRECISION DEFAULT 0,
            seller_discount DOUBLE PRECISION DEFAULT 0,
            settlement_price DOUBLE PRECISION,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    -- ========================================
    -- TikTok Ads Tables
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_ads_upload_batches (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            file_name VARCHAR(500),
            status VARCHAR(50) DEFAULT ''pending'',
            total_rows INTEGER DEFAULT 0,
            processed_rows INTEGER DEFAULT 0,
            success_rows INTEGER DEFAULT 0,
            failed_rows INTEGER DEFAULT 0,
            errors JSONB,
            started_at TIMESTAMPTZ,
            completed_at TIMESTAMPTZ,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_ads_creative_data (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            batch_id VARCHAR(255),
            product_id VARCHAR(255),
            sku_id VARCHAR(255),
            seller_sku VARCHAR(255),
            product_name TEXT,
            sku_name VARCHAR(500),
            date DATE NOT NULL,
            impression BIGINT DEFAULT 0,
            video_views BIGINT DEFAULT 0,
            clicks BIGINT DEFAULT 0,
            ctr DOUBLE PRECISION DEFAULT 0,
            cpc DOUBLE PRECISION DEFAULT 0,
            spend DOUBLE PRECISION DEFAULT 0,
            orders INTEGER DEFAULT 0,
            cvr DOUBLE PRECISION DEFAULT 0,
            cpa DOUBLE PRECISION DEFAULT 0,
            gmv DOUBLE PRECISION DEFAULT 0,
            roas DOUBLE PRECISION DEFAULT 0,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_ads_product_summaries (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            product_id VARCHAR(255) NOT NULL,
            seller_sku VARCHAR(255),
            product_name TEXT,
            period_start DATE NOT NULL,
            period_end DATE NOT NULL,
            total_impression BIGINT DEFAULT 0,
            total_clicks BIGINT DEFAULT 0,
            total_spend DOUBLE PRECISION DEFAULT 0,
            total_orders INTEGER DEFAULT 0,
            total_gmv DOUBLE PRECISION DEFAULT 0,
            avg_ctr DOUBLE PRECISION DEFAULT 0,
            avg_cvr DOUBLE PRECISION DEFAULT 0,
            avg_roas DOUBLE PRECISION DEFAULT 0,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.tiktok_ads_ml_predictions (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            product_id VARCHAR(255) NOT NULL,
            prediction_date DATE NOT NULL,
            predicted_gmv DOUBLE PRECISION,
            predicted_roas DOUBLE PRECISION,
            confidence_score DOUBLE PRECISION,
            model_version VARCHAR(50),
            features JSONB,
            created_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    -- ========================================
    -- Inventory Tables
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.inventory_settings (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            spreadsheet_id VARCHAR(500),
            sheet_name VARCHAR(500),
            selected_columns TEXT,
            header_row INTEGER DEFAULT 1,
            data_start_row INTEGER DEFAULT 2,
            key_column VARCHAR(255),
            auto_sync BOOLEAN DEFAULT false,
            sync_interval_seconds INTEGER DEFAULT 300,
            last_sync_timestamp TIMESTAMPTZ,
            all_columns TEXT,
            last_headers_hash VARCHAR(255),
            last_sync_status VARCHAR(100),
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.inventory_records (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            data JSONB NOT NULL,
            key_value VARCHAR(500) NOT NULL,
            key_column_name VARCHAR(255) NOT NULL,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.inventory_sync_history (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            sync_direction VARCHAR(100),
            status VARCHAR(100),
            total_records INTEGER DEFAULT 0,
            synced_records INTEGER DEFAULT 0,
            failed_records INTEGER DEFAULT 0,
            duration DOUBLE PRECISION DEFAULT 0,
            message TEXT,
            created_at TIMESTAMPTZ DEFAULT now(),
            new_records INTEGER DEFAULT 0,
            updated_records INTEGER DEFAULT 0,
            unchanged_records INTEGER DEFAULT 0,
            error_message TEXT,
            headers_changed BOOLEAN DEFAULT false,
            synced_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.inventory_sku_platform_status (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            sku VARCHAR(500) NOT NULL,
            lazada BOOLEAN DEFAULT false,
            shopee BOOLEAN DEFAULT false,
            tiktok BOOLEAN DEFAULT false,
            checked_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.sku_platform_check_results (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            sku VARCHAR(500) NOT NULL,
            platform VARCHAR(50) NOT NULL,
            exists BOOLEAN DEFAULT false,
            product_id VARCHAR(255),
            sku_id VARCHAR(255),
            product_name TEXT,
            checked_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.sheet_snapshots (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            snapshot_type VARCHAR(50) NOT NULL,
            data JSONB NOT NULL,
            row_count INTEGER DEFAULT 0,
            created_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    -- ========================================
    -- Orders Today & Jobs
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.order_today_items (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            platform VARCHAR(50) NOT NULL,
            order_sn VARCHAR(255) NOT NULL,
            order_status VARCHAR(100),
            sku VARCHAR(255),
            product_name TEXT,
            quantity INTEGER DEFAULT 1,
            price DOUBLE PRECISION,
            order_date DATE NOT NULL,
            created_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.locked_orders (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            platform VARCHAR(50) NOT NULL,
            order_sn VARCHAR(255) NOT NULL UNIQUE,
            locked_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.jobs (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            type VARCHAR(255) NOT NULL,
            status VARCHAR(50) NOT NULL,
            priority VARCHAR(50) DEFAULT ''normal'',
            data JSONB NOT NULL,
            error_message TEXT,
            progress_percent INTEGER DEFAULT 0,
            progress_message VARCHAR(255),
            total_items INTEGER DEFAULT 0,
            processed_items INTEGER DEFAULT 0,
            result_data JSONB,
            started_at TIMESTAMPTZ,
            completed_at TIMESTAMPTZ,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.job_history (
            id SERIAL PRIMARY KEY,
            job_id VARCHAR(255) NOT NULL,
            job_type VARCHAR(255),
            status VARCHAR(50) NOT NULL,
            error_message TEXT,
            duration_ms INTEGER,
            started_at TIMESTAMPTZ,
            completed_at TIMESTAMPTZ,
            created_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    -- ========================================
    -- Auto Functions Tables
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.auto_functions_config (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            enabled BOOLEAN DEFAULT false,
            interval_minutes INTEGER DEFAULT 30,
            start_time VARCHAR(20),
            end_time VARCHAR(20),
            last_executed TIMESTAMPTZ,
            next_scheduled_execution TIMESTAMPTZ,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.auto_functions_history (
            id SERIAL PRIMARY KEY,
            function_name VARCHAR(255) NOT NULL,
            executed_at TIMESTAMPTZ DEFAULT now(),
            status VARCHAR(50) NOT NULL,
            error_message TEXT,
            duration_ms INTEGER
        )', schema_name);

    -- ========================================
    -- OAuth Tables (Per Tenant)
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.oauth_states (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            platform VARCHAR(50) NOT NULL,
            state VARCHAR(255) NOT NULL UNIQUE,
            redirect_url TEXT,
            metadata TEXT,
            expires_at TIMESTAMPTZ NOT NULL,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.oauth_logs (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            platform VARCHAR(50) NOT NULL,
            event_type VARCHAR(50) NOT NULL,
            shop_id VARCHAR(255),
            code TEXT,
            state VARCHAR(255),
            status VARCHAR(50) DEFAULT ''received'',
            error_msg TEXT,
            metadata TEXT,
            processed_at TIMESTAMPTZ,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
        )', schema_name);

    -- ========================================
    -- Webhook Tables
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.webhook_logs (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            platform VARCHAR(50) NOT NULL,
            event_type VARCHAR(100) NOT NULL,
            payload JSONB,
            status VARCHAR(50) DEFAULT ''received'',
            processed_at TIMESTAMPTZ,
            error_message TEXT,
            created_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.webhook_order_events (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            platform VARCHAR(50) NOT NULL,
            order_sn VARCHAR(255) NOT NULL,
            event_type VARCHAR(100) NOT NULL,
            old_status VARCHAR(100),
            new_status VARCHAR(100),
            payload JSONB,
            processed BOOLEAN DEFAULT false,
            created_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.webhook_product_events (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            platform VARCHAR(50) NOT NULL,
            product_id VARCHAR(255) NOT NULL,
            event_type VARCHAR(100) NOT NULL,
            payload JSONB,
            processed BOOLEAN DEFAULT false,
            created_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.webhook_return_events (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            platform VARCHAR(50) NOT NULL,
            return_id VARCHAR(255) NOT NULL,
            order_sn VARCHAR(255),
            event_type VARCHAR(100) NOT NULL,
            status VARCHAR(100),
            payload JSONB,
            processed BOOLEAN DEFAULT false,
            created_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.webhook_marketing_events (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            platform VARCHAR(50) NOT NULL,
            promotion_id VARCHAR(255),
            event_type VARCHAR(100) NOT NULL,
            payload JSONB,
            processed BOOLEAN DEFAULT false,
            created_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.webhook_shopee_events (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            shop_id VARCHAR(255),
            code INTEGER,
            event_type VARCHAR(100) NOT NULL,
            payload JSONB,
            processed BOOLEAN DEFAULT false,
            created_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.webhook_webchat_events (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            platform VARCHAR(50) NOT NULL,
            conversation_id VARCHAR(255),
            event_type VARCHAR(100) NOT NULL,
            payload JSONB,
            processed BOOLEAN DEFAULT false,
            created_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.webhook_fbs_events (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255),
            platform VARCHAR(50) NOT NULL,
            event_type VARCHAR(100) NOT NULL,
            payload JSONB,
            processed BOOLEAN DEFAULT false,
            created_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    -- ========================================
    -- Route Config Tables
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.route_configs (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            route_name VARCHAR(255) NOT NULL,
            config JSONB,
            is_active BOOLEAN DEFAULT true,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.route_execution_config (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            route_name VARCHAR(255) NOT NULL UNIQUE,
            last_execution TIMESTAMPTZ,
            execution_count INTEGER DEFAULT 0,
            last_status VARCHAR(50),
            last_error TEXT,
            config JSONB,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    -- ========================================
    -- Spreadsheets Table
    -- ========================================
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.spreadsheets (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            tenant_id VARCHAR(255) NOT NULL,
            name VARCHAR(500) NOT NULL,
            spreadsheet_id VARCHAR(500) NOT NULL,
            type VARCHAR(100),
            last_sync TIMESTAMPTZ,
            created_at TIMESTAMPTZ DEFAULT now(),
            updated_at TIMESTAMPTZ DEFAULT now()
        )', schema_name);

    -- ========================================
    -- Create Indexes for Performance
    -- ========================================
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_shopee_orders_order_sn ON %I.shopee_orders(order_sn)', replace(schema_name, 'tenant_', ''), schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_shopee_orders_status ON %I.shopee_orders(order_status)', replace(schema_name, 'tenant_', ''), schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_shopee_orders_created ON %I.shopee_orders(created_at)', replace(schema_name, 'tenant_', ''), schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_shopee_items_order ON %I.shopee_order_items(order_sn)', replace(schema_name, 'tenant_', ''), schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_lazada_orders_id ON %I.lazada_orders(order_id)', replace(schema_name, 'tenant_', ''), schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_lazada_items_order ON %I.lazada_order_items(order_id)', replace(schema_name, 'tenant_', ''), schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_tiktok_orders_id ON %I.tiktok_orders(order_id)', replace(schema_name, 'tenant_', ''), schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_tiktok_items_order ON %I.tiktok_order_items(order_id)', replace(schema_name, 'tenant_', ''), schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_platform_configs_tenant ON %I.platform_configs(tenant_id)', replace(schema_name, 'tenant_', ''), schema_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_inventory_records_key ON %I.inventory_records(key_value)', replace(schema_name, 'tenant_', ''), schema_name);

    -- Reset search path
    SET search_path TO public;
    
    RAISE NOTICE 'Created tenant schema: %', schema_name;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 3. Create Default Tenants
-- ============================================
DO $$
BEGIN
    -- Insert default tenants if not exists
    INSERT INTO system.tenants (tenant_id, shop_name, is_active, db_schema)
    VALUES 
        ('yumna_bertigamart', 'Bertigamart', true, 'tenant_yumna_bertigamart'),
        ('tika_nusseyba', 'Nusseyba Shop', true, 'tenant_tika_nusseyba')
    ON CONFLICT (tenant_id) DO NOTHING;
    
    -- Create tenant schemas
    PERFORM create_tenant_schema('tenant_yumna_bertigamart');
    PERFORM create_tenant_schema('tenant_tika_nusseyba');
END $$;

-- ============================================
-- 4. Verification
-- ============================================
DO $$
DECLARE
    schema_count INT;
    table_count INT;
BEGIN
    SELECT COUNT(*) INTO schema_count FROM pg_namespace WHERE nspname LIKE 'tenant_%' OR nspname = 'system';
    SELECT COUNT(*) INTO table_count FROM pg_tables WHERE schemaname LIKE 'tenant_%' OR schemaname = 'system';
    
    RAISE NOTICE '===========================================';
    RAISE NOTICE 'Database initialization complete!';
    RAISE NOTICE 'Schemas created: %', schema_count;
    RAISE NOTICE 'Tables created: %', table_count;
    RAISE NOTICE '===========================================';
END $$;
