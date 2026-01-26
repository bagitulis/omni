--
-- PostgreSQL database dump
--

\restrict h0o8YNX50bfY2YAbrYnHBPfoUlcMyEiVi1o1UiSGiD90v7BeUy1SknOMS4eEWtx

-- Dumped from database version 16.11
-- Dumped by pg_dump version 16.11

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: tenant_yumna_bertigamart; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA tenant_yumna_bertigamart;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: analytics_settings; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.analytics_settings (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    platform character varying(100) DEFAULT 'shopee'::character varying,
    price_column character varying(255) DEFAULT 'HARGA'::character varying,
    sku_column character varying(255),
    formula_deduction double precision DEFAULT 1500,
    formula_multiplier double precision DEFAULT 0.84,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: auto_functions_config; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.auto_functions_config (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    enabled boolean DEFAULT false,
    interval_minutes integer DEFAULT 30,
    start_time character varying(20),
    end_time character varying(20),
    last_executed timestamp with time zone,
    next_scheduled_execution timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: auto_functions_config_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.auto_functions_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: auto_functions_config_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.auto_functions_config_id_seq OWNED BY tenant_yumna_bertigamart.auto_functions_config.id;


--
-- Name: auto_functions_history; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.auto_functions_history (
    id integer NOT NULL,
    function_name character varying(255) NOT NULL,
    executed_at timestamp with time zone DEFAULT now(),
    status character varying(50) NOT NULL,
    error_message text,
    duration_ms integer
);


--
-- Name: auto_functions_history_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.auto_functions_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: auto_functions_history_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.auto_functions_history_id_seq OWNED BY tenant_yumna_bertigamart.auto_functions_history.id;


--
-- Name: filter_preferences; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.filter_preferences (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    platform character varying(100) NOT NULL,
    tab character varying(100) DEFAULT 'master'::character varying,
    visible_columns text,
    column_filters text,
    search_query text DEFAULT ''::text,
    locked_columns text DEFAULT '[]'::text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: google_sheets_settings; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.google_sheets_settings (
    id character varying(255) NOT NULL,
    spreadsheet_id character varying(500),
    selected_sheet character varying(500),
    manual_mode boolean DEFAULT false,
    wallet_spreadsheet_id character varying(500),
    shipping_spreadsheet_id character varying(500),
    inventory_spreadsheet_id character varying(500),
    inventory_sheet_name character varying(500),
    inventory_selected_columns text,
    inventory_available_worksheets text,
    wallet_available_worksheets text,
    shipping_available_worksheets text,
    order_available_worksheets text,
    order_spreadsheet_id character varying(500),
    available_spreadsheets text,
    available_worksheets text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: inventory_records; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.inventory_records (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    data jsonb NOT NULL,
    key_value character varying(500) NOT NULL,
    key_column_name character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: inventory_settings; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.inventory_settings (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    spreadsheet_id character varying(500),
    sheet_name character varying(500),
    selected_columns text,
    header_row integer DEFAULT 1,
    data_start_row integer DEFAULT 2,
    key_column character varying(255),
    auto_sync boolean DEFAULT false,
    sync_interval_seconds integer DEFAULT 300,
    last_sync_timestamp timestamp with time zone,
    all_columns text,
    last_headers_hash character varying(255),
    last_sync_status character varying(100),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: inventory_sku_platform_status; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.inventory_sku_platform_status (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    sku character varying(500) NOT NULL,
    lazada boolean DEFAULT false,
    shopee boolean DEFAULT false,
    tiktok boolean DEFAULT false,
    checked_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: inventory_sync_history; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.inventory_sync_history (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    sync_direction character varying(100),
    status character varying(100),
    total_records integer DEFAULT 0,
    synced_records integer DEFAULT 0,
    failed_records integer DEFAULT 0,
    duration double precision DEFAULT 0,
    message text,
    created_at timestamp with time zone DEFAULT now(),
    new_records integer DEFAULT 0,
    updated_records integer DEFAULT 0,
    unchanged_records integer DEFAULT 0,
    error_message text,
    headers_changed boolean DEFAULT false,
    synced_at timestamp with time zone DEFAULT now()
);


--
-- Name: job_history; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.job_history (
    id integer NOT NULL,
    job_id character varying(255) NOT NULL,
    job_type character varying(255),
    status character varying(50) NOT NULL,
    error_message text,
    duration_ms integer,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: job_history_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.job_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: job_history_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.job_history_id_seq OWNED BY tenant_yumna_bertigamart.job_history.id;


--
-- Name: jobs; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.jobs (
    id character varying(255) NOT NULL,
    type character varying(255) NOT NULL,
    status character varying(50) NOT NULL,
    priority character varying(50) DEFAULT 'normal'::character varying,
    data jsonb NOT NULL,
    error_message text,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: lazada_order_items; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.lazada_order_items (
    id integer NOT NULL,
    order_sn character varying(255) NOT NULL,
    item_id bigint NOT NULL,
    sku_id character varying(500),
    seller_sku character varying(500),
    product_name text,
    variation_name text,
    quantity integer,
    price double precision,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    tenant_id character varying(255)
);


--
-- Name: lazada_order_items_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.lazada_order_items_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: lazada_order_items_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.lazada_order_items_id_seq OWNED BY tenant_yumna_bertigamart.lazada_order_items.id;


--
-- Name: lazada_orders; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.lazada_orders (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    order_sn character varying(255) NOT NULL,
    shop_id bigint,
    order_status character varying(100),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: lazada_orders_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.lazada_orders_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: lazada_orders_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.lazada_orders_id_seq OWNED BY tenant_yumna_bertigamart.lazada_orders.id;


--
-- Name: lazada_products; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.lazada_products (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    item_id character varying(255) NOT NULL,
    name text NOT NULL,
    description text,
    status character varying(100),
    price double precision DEFAULT 0,
    quantity integer DEFAULT 0,
    image text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    brand character varying(255)
);


--
-- Name: lazada_products_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.lazada_products_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: lazada_products_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.lazada_products_id_seq OWNED BY tenant_yumna_bertigamart.lazada_products.id;


--
-- Name: lazada_skus; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.lazada_skus (
    id integer NOT NULL,
    product_id integer NOT NULL,
    sku_id character varying(500) NOT NULL,
    seller_sku character varying(500),
    variant_name text,
    variant_data jsonb,
    price double precision DEFAULT 0,
    quantity integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    tenant_id character varying(255) DEFAULT 'yumna_bertigamart'::character varying NOT NULL,
    item_id character varying(255),
    shop_sku character varying(255),
    name character varying(500),
    special_price double precision DEFAULT 0,
    available integer DEFAULT 0
);


--
-- Name: lazada_skus_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.lazada_skus_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: lazada_skus_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.lazada_skus_id_seq OWNED BY tenant_yumna_bertigamart.lazada_skus.id;


--
-- Name: locked_orders; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.locked_orders (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    sku character varying(500) NOT NULL,
    product_name text NOT NULL,
    variation_name text,
    qty integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: oauth_logs; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.oauth_logs (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    platform character varying(100) NOT NULL,
    event_type character varying(100) NOT NULL,
    shop_id character varying(255),
    code text,
    state character varying(500),
    status character varying(100) DEFAULT 'received'::character varying,
    error_msg text,
    metadata jsonb,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: oauth_states; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.oauth_states (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    platform character varying(100) NOT NULL,
    state character varying(500) NOT NULL,
    redirect_url text,
    metadata jsonb,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: order_today_items; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.order_today_items (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    platform character varying(100) NOT NULL,
    order_sn character varying(255) NOT NULL,
    tracking_no character varying(255),
    courier character varying(255),
    seller_sku character varying(500),
    product_name text,
    variation_name text,
    quantity integer DEFAULT 1,
    synced_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: order_today_items_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.order_today_items_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: order_today_items_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.order_today_items_id_seq OWNED BY tenant_yumna_bertigamart.order_today_items.id;


--
-- Name: platform_configs; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.platform_configs (
    id character varying(255) NOT NULL,
    platform character varying(100) NOT NULL,
    config_key character varying(255) NOT NULL,
    config_value text,
    data_type character varying(100) DEFAULT 'string'::character varying,
    is_encrypted boolean DEFAULT true,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: route_configs; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.route_configs (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    route_path character varying(500) NOT NULL,
    route_method character varying(20) DEFAULT 'GET'::character varying,
    route_name character varying(255),
    description text,
    category character varying(100),
    enabled boolean DEFAULT true,
    caching_enabled boolean DEFAULT true,
    cache_ttl integer DEFAULT 300,
    cache_strategy character varying(100) DEFAULT 'standard'::character varying,
    cache_key_template text,
    queue_enabled boolean DEFAULT true,
    queue_max_size integer DEFAULT 100,
    queue_priority character varying(50) DEFAULT 'normal'::character varying,
    max_concurrent integer DEFAULT 5,
    rate_limit_enabled boolean DEFAULT false,
    rate_limit_window integer DEFAULT 60,
    rate_limit_max integer DEFAULT 100,
    min_interval_ms integer DEFAULT 0,
    timeout integer DEFAULT 30000,
    retry_enabled boolean DEFAULT true,
    max_retries integer DEFAULT 3,
    retry_delay_ms integer DEFAULT 1000,
    last_executed timestamp with time zone,
    execution_count integer DEFAULT 0,
    failure_count integer DEFAULT 0,
    avg_execution_time_ms double precision DEFAULT 0,
    custom_config jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: route_execution_config; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.route_execution_config (
    id integer NOT NULL,
    route_key character varying(255) NOT NULL,
    route_name character varying(255) NOT NULL,
    description text,
    execution_mode character varying(50) DEFAULT 'direct'::character varying,
    priority character varying(50) DEFAULT 'normal'::character varying,
    enabled boolean DEFAULT true,
    icon character varying(50) DEFAULT '????'::character varying,
    category character varying(100) DEFAULT 'general'::character varying,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: route_execution_config_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.route_execution_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: route_execution_config_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.route_execution_config_id_seq OWNED BY tenant_yumna_bertigamart.route_execution_config.id;


--
-- Name: sheet_snapshots; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.sheet_snapshots (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    sku character varying(500) NOT NULL,
    snapshot_data jsonb NOT NULL,
    row_index integer NOT NULL,
    data_hash character varying(255),
    synced_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: shopee_ads_product_data; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.shopee_ads_product_data (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    upload_batch_id character varying(255),
    period_start timestamp with time zone,
    period_end timestamp with time zone,
    period_label character varying(100),
    product_id character varying(255),
    product_name text,
    status character varying(100),
    bidding_mode character varying(255),
    placement character varying(255),
    start_date timestamp with time zone,
    end_date character varying(100),
    impressions integer DEFAULT 0,
    clicks integer DEFAULT 0,
    ctr double precision DEFAULT 0,
    conversions integer DEFAULT 0,
    direct_conversions integer DEFAULT 0,
    conversion_rate double precision DEFAULT 0,
    direct_conversion_rate double precision DEFAULT 0,
    cost_per_conversion double precision DEFAULT 0,
    cost_per_direct_conversion double precision DEFAULT 0,
    units_sold integer DEFAULT 0,
    direct_units_sold integer DEFAULT 0,
    revenue double precision DEFAULT 0,
    direct_revenue double precision DEFAULT 0,
    cost double precision DEFAULT 0,
    roas double precision DEFAULT 0,
    direct_roas double precision DEFAULT 0,
    acos double precision DEFAULT 0,
    direct_acos double precision DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: shopee_ads_product_data_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.shopee_ads_product_data_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: shopee_ads_product_data_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.shopee_ads_product_data_id_seq OWNED BY tenant_yumna_bertigamart.shopee_ads_product_data.id;


--
-- Name: shopee_ads_upload_batches; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.shopee_ads_upload_batches (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    file_name character varying(500),
    period_start timestamp with time zone,
    period_end timestamp with time zone,
    total_rows integer DEFAULT 0,
    inserted_rows integer DEFAULT 0,
    skipped_rows integer DEFAULT 0,
    updated_rows integer DEFAULT 0,
    status character varying(100) DEFAULT 'pending'::character varying,
    error_message text,
    uploaded_by character varying(255),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    period_label character varying(20) DEFAULT ''::character varying
);


--
-- Name: shopee_escrow_items; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.shopee_escrow_items (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    escrow_order_id character varying(255) NOT NULL,
    order_id character varying(255),
    order_sn character varying(255),
    month integer,
    year integer,
    item_id bigint,
    model_id bigint,
    sku character varying(500),
    model_sku character varying(500),
    item_name text,
    model_name text,
    quantity integer DEFAULT 0,
    original_price double precision DEFAULT 0,
    selling_price double precision DEFAULT 0,
    discounted_price double precision DEFAULT 0,
    seller_discount double precision DEFAULT 0,
    shopee_discount double precision DEFAULT 0,
    discount_from_coin double precision DEFAULT 0,
    discount_from_voucher_seller double precision DEFAULT 0,
    discount_from_voucher_shopee double precision DEFAULT 0,
    ams_commission_fee double precision DEFAULT 0,
    seller_order_processing_fee double precision DEFAULT 0,
    raw_item_data jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: shopee_escrow_orders; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.shopee_escrow_orders (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    order_sn character varying(255) NOT NULL,
    month integer,
    year integer,
    order_date timestamp with time zone,
    buyer_user_name character varying(500),
    escrow_amount double precision DEFAULT 0,
    commission_fee double precision DEFAULT 0,
    service_fee double precision DEFAULT 0,
    seller_processing_fee double precision DEFAULT 0,
    buyer_paid_shipping_fee double precision DEFAULT 0,
    actual_shipping_fee double precision DEFAULT 0,
    shopee_shipping_rebate double precision DEFAULT 0,
    estimated_shipping_fee double precision DEFAULT 0,
    buyer_total_amount double precision DEFAULT 0,
    buyer_payment_method character varying(255),
    raw_order_income jsonb,
    raw_buyer_payment_info jsonb,
    synced_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: shopee_escrow_sync; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.shopee_escrow_sync (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    month integer NOT NULL,
    year integer NOT NULL,
    total_orders integer DEFAULT 0,
    synced_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: shopee_order_items; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.shopee_order_items (
    id integer NOT NULL,
    order_sn character varying(255) NOT NULL,
    item_id bigint NOT NULL,
    model_id bigint,
    item_name text,
    model_name text,
    item_sku character varying(500),
    model_sku character varying(500),
    quantity integer,
    price double precision,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    tenant_id character varying(255)
);


--
-- Name: shopee_order_items_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.shopee_order_items_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: shopee_order_items_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.shopee_order_items_id_seq OWNED BY tenant_yumna_bertigamart.shopee_order_items.id;


--
-- Name: shopee_orders; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.shopee_orders (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    order_sn character varying(255) NOT NULL,
    shop_id bigint,
    order_status character varying(100),
    order_timestamp integer,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: shopee_orders_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.shopee_orders_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: shopee_orders_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.shopee_orders_id_seq OWNED BY tenant_yumna_bertigamart.shopee_orders.id;


--
-- Name: shopee_products; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.shopee_products (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    item_id bigint NOT NULL,
    name text NOT NULL,
    description text,
    status character varying(100),
    price double precision DEFAULT 0,
    quantity integer DEFAULT 0,
    image text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: shopee_products_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.shopee_products_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: shopee_products_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.shopee_products_id_seq OWNED BY tenant_yumna_bertigamart.shopee_products.id;


--
-- Name: shopee_skus; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.shopee_skus (
    id integer NOT NULL,
    product_id integer NOT NULL,
    item_id bigint NOT NULL,
    model_id bigint,
    seller_sku character varying(500),
    variant_name text,
    variant_data jsonb,
    price double precision DEFAULT 0,
    quantity integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    tenant_id character varying(255) DEFAULT 'yumna_bertigamart'::character varying NOT NULL
);


--
-- Name: shopee_skus_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.shopee_skus_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: shopee_skus_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.shopee_skus_id_seq OWNED BY tenant_yumna_bertigamart.shopee_skus.id;


--
-- Name: sku_platform_check_results; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.sku_platform_check_results (
    id bigint NOT NULL,
    tenant_id text NOT NULL,
    sku text NOT NULL,
    lazada boolean DEFAULT false,
    shopee boolean DEFAULT false,
    tiktok boolean DEFAULT false,
    checked_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: sku_platform_check_results_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.sku_platform_check_results_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sku_platform_check_results_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.sku_platform_check_results_id_seq OWNED BY tenant_yumna_bertigamart.sku_platform_check_results.id;


--
-- Name: spreadsheets; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.spreadsheets (
    id character varying(255) NOT NULL,
    spreadsheet_id character varying(500) NOT NULL,
    spreadsheet_url text,
    spreadsheet_name character varying(500),
    purpose character varying(100) DEFAULT 'other'::character varying,
    sheets text,
    tenant_id character varying(255) NOT NULL,
    registered_by character varying(255),
    last_used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: tiktok_ads_creative_data; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.tiktok_ads_creative_data (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    upload_batch_id character varying(255),
    period_start timestamp with time zone,
    period_end timestamp with time zone,
    campaign_id character varying(255),
    campaign_name text,
    product_id character varying(255),
    creative_type character varying(255),
    video_title text,
    video_id character varying(255),
    tiktok_account character varying(500),
    posting_time timestamp with time zone,
    status character varying(100),
    authorization_type character varying(255),
    cost double precision DEFAULT 0,
    orders_sku integer DEFAULT 0,
    cost_per_order double precision DEFAULT 0,
    gross_revenue double precision DEFAULT 0,
    roi double precision DEFAULT 0,
    impressions integer DEFAULT 0,
    clicks integer DEFAULT 0,
    ctr double precision DEFAULT 0,
    conversion_rate double precision DEFAULT 0,
    watch_rate_2s double precision,
    watch_rate_6s double precision,
    watch_rate_25pct double precision,
    watch_rate_50pct double precision,
    watch_rate_75pct double precision,
    watch_rate_100pct double precision,
    currency character varying(20) DEFAULT 'IDR'::character varying,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    period_label character varying(20)
);


--
-- Name: tiktok_ads_creative_data_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.tiktok_ads_creative_data_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tiktok_ads_creative_data_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.tiktok_ads_creative_data_id_seq OWNED BY tenant_yumna_bertigamart.tiktok_ads_creative_data.id;


--
-- Name: tiktok_ads_ml_predictions; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.tiktok_ads_ml_predictions (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    product_id character varying(255) NOT NULL,
    prediction_date timestamp with time zone,
    target_month integer,
    target_year integer,
    predicted_roi double precision,
    roi_confidence double precision,
    recommended_budget double precision,
    budget_reason text,
    performance_score character varying(10),
    performance_reason text,
    creative_score double precision,
    best_creative_type character varying(255),
    insights jsonb,
    has_anomaly boolean DEFAULT false,
    anomaly_type character varying(255),
    anomaly_description text,
    model_version character varying(50) DEFAULT 'v1.0'::character varying,
    model_type character varying(100),
    training_data_points integer,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: tiktok_ads_product_summaries; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.tiktok_ads_product_summaries (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    product_id character varying(255) NOT NULL,
    period_type character varying(50),
    period_date timestamp with time zone,
    total_cost double precision DEFAULT 0,
    total_orders integer DEFAULT 0,
    total_revenue double precision DEFAULT 0,
    avg_roi double precision DEFAULT 0,
    avg_ctr double precision DEFAULT 0,
    avg_conversion_rate double precision DEFAULT 0,
    total_impressions integer DEFAULT 0,
    total_clicks integer DEFAULT 0,
    best_creative_type character varying(255),
    top_video_id character varying(255),
    top_video_title text,
    roi_trend character varying(50),
    cost_trend character varying(50),
    orders_trend character varying(50),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: tiktok_ads_upload_batches; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.tiktok_ads_upload_batches (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    file_name character varying(500),
    period_start timestamp with time zone,
    period_end timestamp with time zone,
    total_rows integer DEFAULT 0,
    inserted_rows integer DEFAULT 0,
    skipped_rows integer DEFAULT 0,
    updated_rows integer DEFAULT 0,
    status character varying(100) DEFAULT 'pending'::character varying,
    error_message text,
    uploaded_by character varying(255),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    period_label character varying(20) DEFAULT ''::character varying
);


--
-- Name: tiktok_escrow_items; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.tiktok_escrow_items (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    escrow_order_id character varying(255) NOT NULL,
    order_id character varying(255),
    product_id character varying(255),
    product_name text,
    sku_id character varying(500),
    seller_sku character varying(500),
    quantity integer DEFAULT 0,
    sale_price double precision DEFAULT 0,
    original_price double precision DEFAULT 0,
    subtotal_after_seller_discount double precision DEFAULT 0,
    platform_discount double precision DEFAULT 0,
    seller_discount double precision DEFAULT 0,
    commission double precision DEFAULT 0,
    transaction_fee_item double precision DEFAULT 0,
    settlement_amount double precision DEFAULT 0,
    raw_item_data jsonb,
    synced_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: tiktok_escrow_orders; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.tiktok_escrow_orders (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    order_id character varying(255) NOT NULL,
    month integer,
    year integer,
    order_status character varying(100),
    order_date timestamp with time zone,
    buyer_name character varying(500),
    transaction_id character varying(255),
    transaction_type character varying(100),
    statement_time timestamp with time zone,
    total_settlement_amount double precision DEFAULT 0,
    product_revenue double precision DEFAULT 0,
    platform_commission double precision DEFAULT 0,
    transaction_fee double precision DEFAULT 0,
    shipping_fee_customer_paid double precision DEFAULT 0,
    shipping_fee_actual double precision DEFAULT 0,
    shipping_fee_platform_discount double precision DEFAULT 0,
    seller_shipping_discount double precision DEFAULT 0,
    refund_amount double precision DEFAULT 0,
    adjustment double precision DEFAULT 0,
    buyer_total_amount double precision DEFAULT 0,
    currency character varying(20) DEFAULT 'IDR'::character varying,
    raw_transaction_data jsonb,
    raw_order_data jsonb,
    synced_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: tiktok_escrow_sync; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.tiktok_escrow_sync (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    month integer NOT NULL,
    year integer NOT NULL,
    total_orders integer DEFAULT 0,
    synced_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: tiktok_order_items; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.tiktok_order_items (
    id integer NOT NULL,
    order_sn character varying(255) NOT NULL,
    line_item_id character varying(255),
    product_id bigint NOT NULL,
    sku_id character varying(500),
    seller_sku character varying(500),
    product_name text,
    variation_name text,
    quantity integer,
    price double precision,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    tenant_id character varying(255)
);


--
-- Name: tiktok_order_items_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.tiktok_order_items_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tiktok_order_items_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.tiktok_order_items_id_seq OWNED BY tenant_yumna_bertigamart.tiktok_order_items.id;


--
-- Name: tiktok_orders; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.tiktok_orders (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    order_sn character varying(255) NOT NULL,
    shop_id bigint,
    order_status character varying(100),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: tiktok_orders_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.tiktok_orders_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tiktok_orders_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.tiktok_orders_id_seq OWNED BY tenant_yumna_bertigamart.tiktok_orders.id;


--
-- Name: tiktok_products; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.tiktok_products (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    product_id character varying(255) NOT NULL,
    name text NOT NULL,
    description text,
    status character varying(100),
    price double precision DEFAULT 0,
    quantity integer DEFAULT 0,
    image text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: tiktok_products_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.tiktok_products_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tiktok_products_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.tiktok_products_id_seq OWNED BY tenant_yumna_bertigamart.tiktok_products.id;


--
-- Name: tiktok_skus; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.tiktok_skus (
    id integer NOT NULL,
    product_id integer NOT NULL,
    sku_id character varying(500) NOT NULL,
    seller_sku character varying(500),
    variant_name text,
    variant_data jsonb,
    price double precision DEFAULT 0,
    quantity integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    tenant_id character varying(255) DEFAULT 'yumna_bertigamart'::character varying NOT NULL
);


--
-- Name: tiktok_skus_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.tiktok_skus_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tiktok_skus_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.tiktok_skus_id_seq OWNED BY tenant_yumna_bertigamart.tiktok_skus.id;


--
-- Name: users; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.users (
    id character varying(255) NOT NULL,
    username character varying(255) NOT NULL,
    email character varying(500) NOT NULL,
    password character varying(255) NOT NULL,
    role character varying(100) DEFAULT 'owner'::character varying,
    failed_login_attempts integer DEFAULT 0,
    account_locked_until timestamp with time zone,
    last_failed_login timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: webhook_fbs_events; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.webhook_fbs_events (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    event_type character varying(255) NOT NULL,
    shop_id character varying(255),
    item_id character varying(255),
    sku_id character varying(255),
    stock_change integer,
    invoice_number character varying(255),
    payload jsonb,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: webhook_fbs_events_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.webhook_fbs_events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: webhook_fbs_events_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.webhook_fbs_events_id_seq OWNED BY tenant_yumna_bertigamart.webhook_fbs_events.id;


--
-- Name: webhook_logs; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.webhook_logs (
    id character varying(255) NOT NULL,
    tenant_id character varying(255),
    platform character varying(100) NOT NULL,
    event_type character varying(255) NOT NULL,
    payload jsonb,
    headers jsonb,
    status character varying(100) DEFAULT 'received'::character varying,
    error_msg text,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: webhook_marketing_events; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.webhook_marketing_events (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    event_type character varying(255) NOT NULL,
    shop_id character varying(255),
    item_id character varying(255),
    promotion_id character varying(255),
    promotion_type character varying(255),
    action character varying(100),
    payload jsonb,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: webhook_marketing_events_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.webhook_marketing_events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: webhook_marketing_events_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.webhook_marketing_events_id_seq OWNED BY tenant_yumna_bertigamart.webhook_marketing_events.id;


--
-- Name: webhook_order_events; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.webhook_order_events (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    platform character varying(100) NOT NULL,
    event_type character varying(255) NOT NULL,
    order_sn character varying(255) NOT NULL,
    shop_id character varying(255),
    old_status character varying(100),
    new_status character varying(100),
    fulfillment_status character varying(100),
    package_number character varying(255),
    payload jsonb,
    webhook_log_id character varying(255),
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: webhook_order_events_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.webhook_order_events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: webhook_order_events_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.webhook_order_events_id_seq OWNED BY tenant_yumna_bertigamart.webhook_order_events.id;


--
-- Name: webhook_product_events; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.webhook_product_events (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    event_type character varying(255) NOT NULL,
    shop_id character varying(255),
    item_id character varying(255),
    variation_id character varying(255),
    action character varying(100),
    payload jsonb,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: webhook_product_events_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.webhook_product_events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: webhook_product_events_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.webhook_product_events_id_seq OWNED BY tenant_yumna_bertigamart.webhook_product_events.id;


--
-- Name: webhook_return_events; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.webhook_return_events (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    event_type character varying(255) NOT NULL,
    shop_id character varying(255),
    order_sn character varying(255),
    return_sn character varying(255),
    status character varying(100),
    reason text,
    payload jsonb,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: webhook_return_events_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.webhook_return_events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: webhook_return_events_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.webhook_return_events_id_seq OWNED BY tenant_yumna_bertigamart.webhook_return_events.id;


--
-- Name: webhook_shopee_events; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.webhook_shopee_events (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    event_type character varying(255) NOT NULL,
    shop_id character varying(255),
    action character varying(100),
    expiry_time timestamp with time zone,
    penalty_points integer,
    payload jsonb,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: webhook_shopee_events_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.webhook_shopee_events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: webhook_shopee_events_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.webhook_shopee_events_id_seq OWNED BY tenant_yumna_bertigamart.webhook_shopee_events.id;


--
-- Name: webhook_webchat_events; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.webhook_webchat_events (
    id integer NOT NULL,
    tenant_id character varying(255) NOT NULL,
    event_type character varying(255) NOT NULL,
    shop_id character varying(255),
    conversation_id character varying(255),
    message_type character varying(100),
    sender_id character varying(255),
    payload jsonb,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: webhook_webchat_events_id_seq; Type: SEQUENCE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE SEQUENCE tenant_yumna_bertigamart.webhook_webchat_events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: webhook_webchat_events_id_seq; Type: SEQUENCE OWNED BY; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER SEQUENCE tenant_yumna_bertigamart.webhook_webchat_events_id_seq OWNED BY tenant_yumna_bertigamart.webhook_webchat_events.id;


--
-- Name: wholesale_settings; Type: TABLE; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE TABLE tenant_yumna_bertigamart.wholesale_settings (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    platform character varying(100) DEFAULT 'shopee'::character varying,
    admin_fee integer DEFAULT 1500,
    max_order_tier3 integer DEFAULT 1000,
    min_order1 integer DEFAULT 2,
    max_order1 integer DEFAULT 3,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: auto_functions_config id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.auto_functions_config ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.auto_functions_config_id_seq'::regclass);


--
-- Name: auto_functions_history id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.auto_functions_history ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.auto_functions_history_id_seq'::regclass);


--
-- Name: job_history id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.job_history ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.job_history_id_seq'::regclass);


--
-- Name: lazada_order_items id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.lazada_order_items ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.lazada_order_items_id_seq'::regclass);


--
-- Name: lazada_orders id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.lazada_orders ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.lazada_orders_id_seq'::regclass);


--
-- Name: lazada_products id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.lazada_products ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.lazada_products_id_seq'::regclass);


--
-- Name: lazada_skus id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.lazada_skus ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.lazada_skus_id_seq'::regclass);


--
-- Name: order_today_items id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.order_today_items ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.order_today_items_id_seq'::regclass);


--
-- Name: route_execution_config id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.route_execution_config ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.route_execution_config_id_seq'::regclass);


--
-- Name: shopee_ads_product_data id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_ads_product_data ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.shopee_ads_product_data_id_seq'::regclass);


--
-- Name: shopee_order_items id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_order_items ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.shopee_order_items_id_seq'::regclass);


--
-- Name: shopee_orders id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_orders ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.shopee_orders_id_seq'::regclass);


--
-- Name: shopee_products id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_products ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.shopee_products_id_seq'::regclass);


--
-- Name: shopee_skus id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_skus ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.shopee_skus_id_seq'::regclass);


--
-- Name: sku_platform_check_results id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.sku_platform_check_results ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.sku_platform_check_results_id_seq'::regclass);


--
-- Name: tiktok_ads_creative_data id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_ads_creative_data ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.tiktok_ads_creative_data_id_seq'::regclass);


--
-- Name: tiktok_order_items id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_order_items ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.tiktok_order_items_id_seq'::regclass);


--
-- Name: tiktok_orders id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_orders ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.tiktok_orders_id_seq'::regclass);


--
-- Name: tiktok_products id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_products ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.tiktok_products_id_seq'::regclass);


--
-- Name: tiktok_skus id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_skus ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.tiktok_skus_id_seq'::regclass);


--
-- Name: webhook_fbs_events id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_fbs_events ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.webhook_fbs_events_id_seq'::regclass);


--
-- Name: webhook_marketing_events id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_marketing_events ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.webhook_marketing_events_id_seq'::regclass);


--
-- Name: webhook_order_events id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_order_events ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.webhook_order_events_id_seq'::regclass);


--
-- Name: webhook_product_events id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_product_events ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.webhook_product_events_id_seq'::regclass);


--
-- Name: webhook_return_events id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_return_events ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.webhook_return_events_id_seq'::regclass);


--
-- Name: webhook_shopee_events id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_shopee_events ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.webhook_shopee_events_id_seq'::regclass);


--
-- Name: webhook_webchat_events id; Type: DEFAULT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_webchat_events ALTER COLUMN id SET DEFAULT nextval('tenant_yumna_bertigamart.webhook_webchat_events_id_seq'::regclass);


--
-- Name: analytics_settings analytics_settings_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.analytics_settings
    ADD CONSTRAINT analytics_settings_pkey PRIMARY KEY (id);


--
-- Name: analytics_settings analytics_settings_tenant_id_platform_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.analytics_settings
    ADD CONSTRAINT analytics_settings_tenant_id_platform_key UNIQUE (tenant_id, platform);


--
-- Name: auto_functions_config auto_functions_config_name_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.auto_functions_config
    ADD CONSTRAINT auto_functions_config_name_key UNIQUE (name);


--
-- Name: auto_functions_config auto_functions_config_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.auto_functions_config
    ADD CONSTRAINT auto_functions_config_pkey PRIMARY KEY (id);


--
-- Name: auto_functions_history auto_functions_history_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.auto_functions_history
    ADD CONSTRAINT auto_functions_history_pkey PRIMARY KEY (id);


--
-- Name: filter_preferences filter_preferences_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.filter_preferences
    ADD CONSTRAINT filter_preferences_pkey PRIMARY KEY (id);


--
-- Name: filter_preferences filter_preferences_tenant_id_platform_tab_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.filter_preferences
    ADD CONSTRAINT filter_preferences_tenant_id_platform_tab_key UNIQUE (tenant_id, platform, tab);


--
-- Name: google_sheets_settings google_sheets_settings_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.google_sheets_settings
    ADD CONSTRAINT google_sheets_settings_pkey PRIMARY KEY (id);


--
-- Name: inventory_records inventory_records_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.inventory_records
    ADD CONSTRAINT inventory_records_pkey PRIMARY KEY (id);


--
-- Name: inventory_records inventory_records_tenant_id_key_column_name_key_value_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.inventory_records
    ADD CONSTRAINT inventory_records_tenant_id_key_column_name_key_value_key UNIQUE (tenant_id, key_column_name, key_value);


--
-- Name: inventory_settings inventory_settings_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.inventory_settings
    ADD CONSTRAINT inventory_settings_pkey PRIMARY KEY (id);


--
-- Name: inventory_settings inventory_settings_tenant_id_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.inventory_settings
    ADD CONSTRAINT inventory_settings_tenant_id_key UNIQUE (tenant_id);


--
-- Name: inventory_sku_platform_status inventory_sku_platform_status_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.inventory_sku_platform_status
    ADD CONSTRAINT inventory_sku_platform_status_pkey PRIMARY KEY (id);


--
-- Name: inventory_sku_platform_status inventory_sku_platform_status_tenant_id_sku_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.inventory_sku_platform_status
    ADD CONSTRAINT inventory_sku_platform_status_tenant_id_sku_key UNIQUE (tenant_id, sku);


--
-- Name: inventory_sync_history inventory_sync_history_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.inventory_sync_history
    ADD CONSTRAINT inventory_sync_history_pkey PRIMARY KEY (id);


--
-- Name: job_history job_history_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.job_history
    ADD CONSTRAINT job_history_pkey PRIMARY KEY (id);


--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (id);


--
-- Name: lazada_order_items lazada_order_items_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.lazada_order_items
    ADD CONSTRAINT lazada_order_items_pkey PRIMARY KEY (id);


--
-- Name: lazada_orders lazada_orders_order_sn_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.lazada_orders
    ADD CONSTRAINT lazada_orders_order_sn_key UNIQUE (order_sn);


--
-- Name: lazada_orders lazada_orders_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.lazada_orders
    ADD CONSTRAINT lazada_orders_pkey PRIMARY KEY (id);


--
-- Name: lazada_products lazada_products_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.lazada_products
    ADD CONSTRAINT lazada_products_pkey PRIMARY KEY (id);


--
-- Name: lazada_products lazada_products_tenant_id_item_id_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.lazada_products
    ADD CONSTRAINT lazada_products_tenant_id_item_id_key UNIQUE (tenant_id, item_id);


--
-- Name: lazada_skus lazada_skus_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.lazada_skus
    ADD CONSTRAINT lazada_skus_pkey PRIMARY KEY (id);


--
-- Name: lazada_skus lazada_skus_sku_id_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.lazada_skus
    ADD CONSTRAINT lazada_skus_sku_id_key UNIQUE (sku_id);


--
-- Name: locked_orders locked_orders_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.locked_orders
    ADD CONSTRAINT locked_orders_pkey PRIMARY KEY (id);


--
-- Name: locked_orders locked_orders_tenant_id_sku_product_name_variation_name_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.locked_orders
    ADD CONSTRAINT locked_orders_tenant_id_sku_product_name_variation_name_key UNIQUE (tenant_id, sku, product_name, variation_name);


--
-- Name: oauth_logs oauth_logs_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.oauth_logs
    ADD CONSTRAINT oauth_logs_pkey PRIMARY KEY (id);


--
-- Name: oauth_states oauth_states_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.oauth_states
    ADD CONSTRAINT oauth_states_pkey PRIMARY KEY (id);


--
-- Name: oauth_states oauth_states_state_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.oauth_states
    ADD CONSTRAINT oauth_states_state_key UNIQUE (state);


--
-- Name: order_today_items order_today_items_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.order_today_items
    ADD CONSTRAINT order_today_items_pkey PRIMARY KEY (id);


--
-- Name: order_today_items order_today_items_tenant_id_platform_order_sn_seller_sku_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.order_today_items
    ADD CONSTRAINT order_today_items_tenant_id_platform_order_sn_seller_sku_key UNIQUE (tenant_id, platform, order_sn, seller_sku);


--
-- Name: platform_configs platform_configs_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.platform_configs
    ADD CONSTRAINT platform_configs_pkey PRIMARY KEY (id);


--
-- Name: platform_configs platform_configs_platform_config_key_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.platform_configs
    ADD CONSTRAINT platform_configs_platform_config_key_key UNIQUE (platform, config_key);


--
-- Name: route_configs route_configs_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.route_configs
    ADD CONSTRAINT route_configs_pkey PRIMARY KEY (id);


--
-- Name: route_configs route_configs_tenant_id_route_path_route_method_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.route_configs
    ADD CONSTRAINT route_configs_tenant_id_route_path_route_method_key UNIQUE (tenant_id, route_path, route_method);


--
-- Name: route_execution_config route_execution_config_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.route_execution_config
    ADD CONSTRAINT route_execution_config_pkey PRIMARY KEY (id);


--
-- Name: route_execution_config route_execution_config_route_key_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.route_execution_config
    ADD CONSTRAINT route_execution_config_route_key_key UNIQUE (route_key);


--
-- Name: sheet_snapshots sheet_snapshots_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.sheet_snapshots
    ADD CONSTRAINT sheet_snapshots_pkey PRIMARY KEY (id);


--
-- Name: sheet_snapshots sheet_snapshots_tenant_id_sku_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.sheet_snapshots
    ADD CONSTRAINT sheet_snapshots_tenant_id_sku_key UNIQUE (tenant_id, sku);


--
-- Name: shopee_ads_product_data shopee_ads_product_data_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_ads_product_data
    ADD CONSTRAINT shopee_ads_product_data_pkey PRIMARY KEY (id);


--
-- Name: shopee_ads_upload_batches shopee_ads_upload_batches_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_ads_upload_batches
    ADD CONSTRAINT shopee_ads_upload_batches_pkey PRIMARY KEY (id);


--
-- Name: shopee_escrow_items shopee_escrow_items_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_escrow_items
    ADD CONSTRAINT shopee_escrow_items_pkey PRIMARY KEY (id);


--
-- Name: shopee_escrow_orders shopee_escrow_orders_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_escrow_orders
    ADD CONSTRAINT shopee_escrow_orders_pkey PRIMARY KEY (id);


--
-- Name: shopee_escrow_orders shopee_escrow_orders_tenant_id_order_sn_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_escrow_orders
    ADD CONSTRAINT shopee_escrow_orders_tenant_id_order_sn_key UNIQUE (tenant_id, order_sn);


--
-- Name: shopee_escrow_sync shopee_escrow_sync_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_escrow_sync
    ADD CONSTRAINT shopee_escrow_sync_pkey PRIMARY KEY (id);


--
-- Name: shopee_escrow_sync shopee_escrow_sync_tenant_id_month_year_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_escrow_sync
    ADD CONSTRAINT shopee_escrow_sync_tenant_id_month_year_key UNIQUE (tenant_id, month, year);


--
-- Name: shopee_order_items shopee_order_items_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_order_items
    ADD CONSTRAINT shopee_order_items_pkey PRIMARY KEY (id);


--
-- Name: shopee_orders shopee_orders_order_sn_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_orders
    ADD CONSTRAINT shopee_orders_order_sn_key UNIQUE (order_sn);


--
-- Name: shopee_orders shopee_orders_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_orders
    ADD CONSTRAINT shopee_orders_pkey PRIMARY KEY (id);


--
-- Name: shopee_products shopee_products_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_products
    ADD CONSTRAINT shopee_products_pkey PRIMARY KEY (id);


--
-- Name: shopee_products shopee_products_tenant_id_item_id_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_products
    ADD CONSTRAINT shopee_products_tenant_id_item_id_key UNIQUE (tenant_id, item_id);


--
-- Name: shopee_skus shopee_skus_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.shopee_skus
    ADD CONSTRAINT shopee_skus_pkey PRIMARY KEY (id);


--
-- Name: sku_platform_check_results sku_platform_check_results_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.sku_platform_check_results
    ADD CONSTRAINT sku_platform_check_results_pkey PRIMARY KEY (id);


--
-- Name: spreadsheets spreadsheets_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.spreadsheets
    ADD CONSTRAINT spreadsheets_pkey PRIMARY KEY (id);


--
-- Name: spreadsheets spreadsheets_spreadsheet_id_tenant_id_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.spreadsheets
    ADD CONSTRAINT spreadsheets_spreadsheet_id_tenant_id_key UNIQUE (spreadsheet_id, tenant_id);


--
-- Name: tiktok_ads_creative_data tiktok_ads_creative_data_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_ads_creative_data
    ADD CONSTRAINT tiktok_ads_creative_data_pkey PRIMARY KEY (id);


--
-- Name: tiktok_ads_ml_predictions tiktok_ads_ml_predictions_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_ads_ml_predictions
    ADD CONSTRAINT tiktok_ads_ml_predictions_pkey PRIMARY KEY (id);


--
-- Name: tiktok_ads_ml_predictions tiktok_ads_ml_predictions_tenant_id_product_id_target_month_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_ads_ml_predictions
    ADD CONSTRAINT tiktok_ads_ml_predictions_tenant_id_product_id_target_month_key UNIQUE (tenant_id, product_id, target_month, target_year);


--
-- Name: tiktok_ads_product_summaries tiktok_ads_product_summaries_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_ads_product_summaries
    ADD CONSTRAINT tiktok_ads_product_summaries_pkey PRIMARY KEY (id);


--
-- Name: tiktok_ads_product_summaries tiktok_ads_product_summaries_tenant_id_product_id_period_ty_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_ads_product_summaries
    ADD CONSTRAINT tiktok_ads_product_summaries_tenant_id_product_id_period_ty_key UNIQUE (tenant_id, product_id, period_type, period_date);


--
-- Name: tiktok_ads_upload_batches tiktok_ads_upload_batches_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_ads_upload_batches
    ADD CONSTRAINT tiktok_ads_upload_batches_pkey PRIMARY KEY (id);


--
-- Name: tiktok_escrow_items tiktok_escrow_items_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_escrow_items
    ADD CONSTRAINT tiktok_escrow_items_pkey PRIMARY KEY (id);


--
-- Name: tiktok_escrow_orders tiktok_escrow_orders_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_escrow_orders
    ADD CONSTRAINT tiktok_escrow_orders_pkey PRIMARY KEY (id);


--
-- Name: tiktok_escrow_orders tiktok_escrow_orders_tenant_id_order_id_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_escrow_orders
    ADD CONSTRAINT tiktok_escrow_orders_tenant_id_order_id_key UNIQUE (tenant_id, order_id);


--
-- Name: tiktok_escrow_sync tiktok_escrow_sync_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_escrow_sync
    ADD CONSTRAINT tiktok_escrow_sync_pkey PRIMARY KEY (id);


--
-- Name: tiktok_escrow_sync tiktok_escrow_sync_tenant_id_month_year_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_escrow_sync
    ADD CONSTRAINT tiktok_escrow_sync_tenant_id_month_year_key UNIQUE (tenant_id, month, year);


--
-- Name: tiktok_order_items tiktok_order_items_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_order_items
    ADD CONSTRAINT tiktok_order_items_pkey PRIMARY KEY (id);


--
-- Name: tiktok_orders tiktok_orders_order_sn_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_orders
    ADD CONSTRAINT tiktok_orders_order_sn_key UNIQUE (order_sn);


--
-- Name: tiktok_orders tiktok_orders_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_orders
    ADD CONSTRAINT tiktok_orders_pkey PRIMARY KEY (id);


--
-- Name: tiktok_products tiktok_products_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_products
    ADD CONSTRAINT tiktok_products_pkey PRIMARY KEY (id);


--
-- Name: tiktok_products tiktok_products_tenant_id_product_id_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_products
    ADD CONSTRAINT tiktok_products_tenant_id_product_id_key UNIQUE (tenant_id, product_id);


--
-- Name: tiktok_skus tiktok_skus_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_skus
    ADD CONSTRAINT tiktok_skus_pkey PRIMARY KEY (id);


--
-- Name: tiktok_skus tiktok_skus_sku_id_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.tiktok_skus
    ADD CONSTRAINT tiktok_skus_sku_id_key UNIQUE (sku_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: webhook_fbs_events webhook_fbs_events_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_fbs_events
    ADD CONSTRAINT webhook_fbs_events_pkey PRIMARY KEY (id);


--
-- Name: webhook_logs webhook_logs_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_logs
    ADD CONSTRAINT webhook_logs_pkey PRIMARY KEY (id);


--
-- Name: webhook_marketing_events webhook_marketing_events_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_marketing_events
    ADD CONSTRAINT webhook_marketing_events_pkey PRIMARY KEY (id);


--
-- Name: webhook_order_events webhook_order_events_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_order_events
    ADD CONSTRAINT webhook_order_events_pkey PRIMARY KEY (id);


--
-- Name: webhook_product_events webhook_product_events_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_product_events
    ADD CONSTRAINT webhook_product_events_pkey PRIMARY KEY (id);


--
-- Name: webhook_return_events webhook_return_events_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_return_events
    ADD CONSTRAINT webhook_return_events_pkey PRIMARY KEY (id);


--
-- Name: webhook_shopee_events webhook_shopee_events_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_shopee_events
    ADD CONSTRAINT webhook_shopee_events_pkey PRIMARY KEY (id);


--
-- Name: webhook_webchat_events webhook_webchat_events_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.webhook_webchat_events
    ADD CONSTRAINT webhook_webchat_events_pkey PRIMARY KEY (id);


--
-- Name: wholesale_settings wholesale_settings_pkey; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.wholesale_settings
    ADD CONSTRAINT wholesale_settings_pkey PRIMARY KEY (id);


--
-- Name: wholesale_settings wholesale_settings_tenant_id_key; Type: CONSTRAINT; Schema: tenant_yumna_bertigamart; Owner: -
--

ALTER TABLE ONLY tenant_yumna_bertigamart.wholesale_settings
    ADD CONSTRAINT wholesale_settings_tenant_id_key UNIQUE (tenant_id);


--
-- Name: idx_analytics_settings_tenant; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_analytics_settings_tenant ON tenant_yumna_bertigamart.analytics_settings USING btree (tenant_id);


--
-- Name: idx_inventory_records_tenant; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_inventory_records_tenant ON tenant_yumna_bertigamart.inventory_records USING btree (tenant_id);


--
-- Name: idx_inventory_sync_history_synced_at; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_inventory_sync_history_synced_at ON tenant_yumna_bertigamart.inventory_sync_history USING btree (synced_at DESC);


--
-- Name: idx_inventory_sync_history_tenant; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_inventory_sync_history_tenant ON tenant_yumna_bertigamart.inventory_sync_history USING btree (tenant_id);


--
-- Name: idx_lazada_order_items_ordersn; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_lazada_order_items_ordersn ON tenant_yumna_bertigamart.lazada_order_items USING btree (order_sn);


--
-- Name: idx_lazada_orders_tenant; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_lazada_orders_tenant ON tenant_yumna_bertigamart.lazada_orders USING btree (tenant_id);


--
-- Name: idx_lazada_skus_item_id; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_lazada_skus_item_id ON tenant_yumna_bertigamart.lazada_skus USING btree (item_id);


--
-- Name: idx_platform_configs_platform; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_platform_configs_platform ON tenant_yumna_bertigamart.platform_configs USING btree (platform);


--
-- Name: idx_shopee_ads_batch_period; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_ads_batch_period ON tenant_yumna_bertigamart.shopee_ads_upload_batches USING btree (period_label);


--
-- Name: idx_shopee_ads_data_period; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_ads_data_period ON tenant_yumna_bertigamart.shopee_ads_product_data USING btree (period_label);


--
-- Name: idx_shopee_ads_period; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_ads_period ON tenant_yumna_bertigamart.shopee_ads_product_data USING btree (period_start, period_end);


--
-- Name: idx_shopee_ads_product; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_ads_product ON tenant_yumna_bertigamart.shopee_ads_product_data USING btree (product_id);


--
-- Name: idx_shopee_escrow_items_model_sku; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_escrow_items_model_sku ON tenant_yumna_bertigamart.shopee_escrow_items USING btree (model_sku);


--
-- Name: idx_shopee_escrow_items_order; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_escrow_items_order ON tenant_yumna_bertigamart.shopee_escrow_items USING btree (escrow_order_id);


--
-- Name: idx_shopee_escrow_items_order_id; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_escrow_items_order_id ON tenant_yumna_bertigamart.shopee_escrow_items USING btree (escrow_order_id);


--
-- Name: idx_shopee_escrow_items_sku; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_escrow_items_sku ON tenant_yumna_bertigamart.shopee_escrow_items USING btree (sku);


--
-- Name: idx_shopee_escrow_items_tenant; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_escrow_items_tenant ON tenant_yumna_bertigamart.shopee_escrow_items USING btree (tenant_id);


--
-- Name: idx_shopee_escrow_orders_month_year; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_escrow_orders_month_year ON tenant_yumna_bertigamart.shopee_escrow_orders USING btree (month, year);


--
-- Name: idx_shopee_escrow_orders_order_sn; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_escrow_orders_order_sn ON tenant_yumna_bertigamart.shopee_escrow_orders USING btree (order_sn);


--
-- Name: idx_shopee_escrow_orders_tenant; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_escrow_orders_tenant ON tenant_yumna_bertigamart.shopee_escrow_orders USING btree (tenant_id);


--
-- Name: idx_shopee_escrow_sync_month_year; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_escrow_sync_month_year ON tenant_yumna_bertigamart.shopee_escrow_sync USING btree (month, year);


--
-- Name: idx_shopee_escrow_sync_tenant; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_escrow_sync_tenant ON tenant_yumna_bertigamart.shopee_escrow_sync USING btree (tenant_id);


--
-- Name: idx_shopee_order_items_ordersn; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_order_items_ordersn ON tenant_yumna_bertigamart.shopee_order_items USING btree (order_sn);


--
-- Name: idx_shopee_orders_status; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_orders_status ON tenant_yumna_bertigamart.shopee_orders USING btree (order_status);


--
-- Name: idx_shopee_orders_tenant; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_orders_tenant ON tenant_yumna_bertigamart.shopee_orders USING btree (tenant_id);


--
-- Name: idx_shopee_products_tenant; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_products_tenant ON tenant_yumna_bertigamart.shopee_products USING btree (tenant_id);


--
-- Name: idx_shopee_skus_model; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_skus_model ON tenant_yumna_bertigamart.shopee_skus USING btree (model_id);


--
-- Name: idx_shopee_skus_product; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_shopee_skus_product ON tenant_yumna_bertigamart.shopee_skus USING btree (product_id);


--
-- Name: idx_shopee_tenant_ordersn; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE UNIQUE INDEX idx_shopee_tenant_ordersn ON tenant_yumna_bertigamart.shopee_orders USING btree (tenant_id, order_sn);


--
-- Name: idx_sku_platform_check_results_sku; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_sku_platform_check_results_sku ON tenant_yumna_bertigamart.sku_platform_check_results USING btree (sku);


--
-- Name: idx_sku_platform_check_results_tenant_id; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_sku_platform_check_results_tenant_id ON tenant_yumna_bertigamart.sku_platform_check_results USING btree (tenant_id);


--
-- Name: idx_tiktok_ads_batch_period; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_tiktok_ads_batch_period ON tenant_yumna_bertigamart.tiktok_ads_upload_batches USING btree (period_label);


--
-- Name: idx_tiktok_ads_campaign; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_tiktok_ads_campaign ON tenant_yumna_bertigamart.tiktok_ads_creative_data USING btree (campaign_id);


--
-- Name: idx_tiktok_ads_data_period; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_tiktok_ads_data_period ON tenant_yumna_bertigamart.tiktok_ads_creative_data USING btree (period_label);


--
-- Name: idx_tiktok_ads_product; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_tiktok_ads_product ON tenant_yumna_bertigamart.tiktok_ads_creative_data USING btree (product_id);


--
-- Name: idx_tiktok_escrow_items_order_id; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_tiktok_escrow_items_order_id ON tenant_yumna_bertigamart.tiktok_escrow_items USING btree (escrow_order_id);


--
-- Name: idx_tiktok_escrow_items_seller_sku; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_tiktok_escrow_items_seller_sku ON tenant_yumna_bertigamart.tiktok_escrow_items USING btree (seller_sku);


--
-- Name: idx_tiktok_escrow_items_tenant; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_tiktok_escrow_items_tenant ON tenant_yumna_bertigamart.tiktok_escrow_items USING btree (tenant_id);


--
-- Name: idx_tiktok_escrow_orders_month_year; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_tiktok_escrow_orders_month_year ON tenant_yumna_bertigamart.tiktok_escrow_orders USING btree (month, year);


--
-- Name: idx_tiktok_escrow_orders_order_id; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_tiktok_escrow_orders_order_id ON tenant_yumna_bertigamart.tiktok_escrow_orders USING btree (order_id);


--
-- Name: idx_tiktok_escrow_orders_tenant; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_tiktok_escrow_orders_tenant ON tenant_yumna_bertigamart.tiktok_escrow_orders USING btree (tenant_id);


--
-- Name: idx_tiktok_escrow_sync_month_year; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_tiktok_escrow_sync_month_year ON tenant_yumna_bertigamart.tiktok_escrow_sync USING btree (month, year);


--
-- Name: idx_tiktok_escrow_sync_tenant; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_tiktok_escrow_sync_tenant ON tenant_yumna_bertigamart.tiktok_escrow_sync USING btree (tenant_id);


--
-- Name: idx_tiktok_orders_tenant; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_tiktok_orders_tenant ON tenant_yumna_bertigamart.tiktok_orders USING btree (tenant_id);


--
-- Name: idx_users_email; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_users_email ON tenant_yumna_bertigamart.users USING btree (email);


--
-- Name: idx_users_username; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_users_username ON tenant_yumna_bertigamart.users USING btree (username);


--
-- Name: idx_webhook_order_events_ordersn; Type: INDEX; Schema: tenant_yumna_bertigamart; Owner: -
--

CREATE INDEX idx_webhook_order_events_ordersn ON tenant_yumna_bertigamart.webhook_order_events USING btree (order_sn);


--
-- PostgreSQL database dump complete
--

\unrestrict h0o8YNX50bfY2YAbrYnHBPfoUlcMyEiVi1o1UiSGiD90v7BeUy1SknOMS4eEWtx

