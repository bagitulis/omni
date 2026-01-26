--
-- PostgreSQL database dump
--

\restrict d5RNBj5rudQEkmJVJiH8yvvjJ9TCVdWrgSjLr7ZvCXF0Ya0txhfDLSXg3GYc9oO

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
-- Name: system; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA system;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: audit_log; Type: TABLE; Schema: system; Owner: -
--

CREATE TABLE system.audit_log (
    id bigint NOT NULL,
    tenant_id character varying(255),
    user_id character varying(255),
    action character varying(255) NOT NULL,
    entity_type character varying(255),
    entity_id character varying(255),
    old_value jsonb,
    new_value jsonb,
    ip_address inet,
    user_agent text,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: audit_log_id_seq; Type: SEQUENCE; Schema: system; Owner: -
--

CREATE SEQUENCE system.audit_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: audit_log_id_seq; Type: SEQUENCE OWNED BY; Schema: system; Owner: -
--

ALTER SEQUENCE system.audit_log_id_seq OWNED BY system.audit_log.id;


--
-- Name: global_config; Type: TABLE; Schema: system; Owner: -
--

CREATE TABLE system.global_config (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    platform character varying(100) NOT NULL,
    config_key character varying(255) NOT NULL,
    config_value text,
    is_encrypted boolean DEFAULT false,
    description text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: oauth_logs; Type: TABLE; Schema: system; Owner: -
--

CREATE TABLE system.oauth_logs (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    platform character varying(50) NOT NULL,
    event_type character varying(50) NOT NULL,
    shop_id character varying(255),
    code text,
    state character varying(255),
    status character varying(50) DEFAULT 'received'::character varying,
    error_msg text,
    metadata text,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: oauth_states; Type: TABLE; Schema: system; Owner: -
--

CREATE TABLE system.oauth_states (
    id character varying(255) NOT NULL,
    tenant_id character varying(255) NOT NULL,
    platform character varying(50) NOT NULL,
    state character varying(255) NOT NULL,
    redirect_url text,
    metadata text,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: tenants; Type: TABLE; Schema: system; Owner: -
--

CREATE TABLE system.tenants (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id character varying(255) NOT NULL,
    shop_name character varying(255),
    is_active boolean DEFAULT true,
    db_schema character varying(255),
    settings jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: users; Type: TABLE; Schema: system; Owner: -
--

CREATE TABLE system.users (
    id character varying(255) DEFAULT (gen_random_uuid())::text NOT NULL,
    username character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    password character varying(255) NOT NULL,
    role character varying(50) DEFAULT 'developer'::character varying,
    failed_login_attempts integer DEFAULT 0,
    account_locked_until timestamp with time zone,
    last_failed_login timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: audit_log id; Type: DEFAULT; Schema: system; Owner: -
--

ALTER TABLE ONLY system.audit_log ALTER COLUMN id SET DEFAULT nextval('system.audit_log_id_seq'::regclass);


--
-- Name: audit_log audit_log_pkey; Type: CONSTRAINT; Schema: system; Owner: -
--

ALTER TABLE ONLY system.audit_log
    ADD CONSTRAINT audit_log_pkey PRIMARY KEY (id);


--
-- Name: global_config global_config_pkey; Type: CONSTRAINT; Schema: system; Owner: -
--

ALTER TABLE ONLY system.global_config
    ADD CONSTRAINT global_config_pkey PRIMARY KEY (id);


--
-- Name: oauth_logs oauth_logs_pkey; Type: CONSTRAINT; Schema: system; Owner: -
--

ALTER TABLE ONLY system.oauth_logs
    ADD CONSTRAINT oauth_logs_pkey PRIMARY KEY (id);


--
-- Name: oauth_states oauth_states_pkey; Type: CONSTRAINT; Schema: system; Owner: -
--

ALTER TABLE ONLY system.oauth_states
    ADD CONSTRAINT oauth_states_pkey PRIMARY KEY (id);


--
-- Name: oauth_states oauth_states_state_key; Type: CONSTRAINT; Schema: system; Owner: -
--

ALTER TABLE ONLY system.oauth_states
    ADD CONSTRAINT oauth_states_state_key UNIQUE (state);


--
-- Name: tenants tenants_pkey; Type: CONSTRAINT; Schema: system; Owner: -
--

ALTER TABLE ONLY system.tenants
    ADD CONSTRAINT tenants_pkey PRIMARY KEY (id);


--
-- Name: tenants tenants_tenant_id_key; Type: CONSTRAINT; Schema: system; Owner: -
--

ALTER TABLE ONLY system.tenants
    ADD CONSTRAINT tenants_tenant_id_key UNIQUE (tenant_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: system; Owner: -
--

ALTER TABLE ONLY system.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: system; Owner: -
--

ALTER TABLE ONLY system.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: system; Owner: -
--

ALTER TABLE ONLY system.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: idx_audit_log_created; Type: INDEX; Schema: system; Owner: -
--

CREATE INDEX idx_audit_log_created ON system.audit_log USING btree (created_at);


--
-- Name: idx_audit_log_tenant; Type: INDEX; Schema: system; Owner: -
--

CREATE INDEX idx_audit_log_tenant ON system.audit_log USING btree (tenant_id);


--
-- Name: idx_oauth_logs_created_at; Type: INDEX; Schema: system; Owner: -
--

CREATE INDEX idx_oauth_logs_created_at ON system.oauth_logs USING btree (created_at);


--
-- Name: idx_oauth_logs_event_type; Type: INDEX; Schema: system; Owner: -
--

CREATE INDEX idx_oauth_logs_event_type ON system.oauth_logs USING btree (event_type);


--
-- Name: idx_oauth_logs_platform; Type: INDEX; Schema: system; Owner: -
--

CREATE INDEX idx_oauth_logs_platform ON system.oauth_logs USING btree (platform);


--
-- Name: idx_oauth_logs_status; Type: INDEX; Schema: system; Owner: -
--

CREATE INDEX idx_oauth_logs_status ON system.oauth_logs USING btree (status);


--
-- Name: idx_oauth_logs_tenant_id; Type: INDEX; Schema: system; Owner: -
--

CREATE INDEX idx_oauth_logs_tenant_id ON system.oauth_logs USING btree (tenant_id);


--
-- Name: idx_oauth_states_expires_at; Type: INDEX; Schema: system; Owner: -
--

CREATE INDEX idx_oauth_states_expires_at ON system.oauth_states USING btree (expires_at);


--
-- Name: idx_oauth_states_platform; Type: INDEX; Schema: system; Owner: -
--

CREATE INDEX idx_oauth_states_platform ON system.oauth_states USING btree (platform);


--
-- Name: idx_oauth_states_tenant_id; Type: INDEX; Schema: system; Owner: -
--

CREATE INDEX idx_oauth_states_tenant_id ON system.oauth_states USING btree (tenant_id);


--
-- Name: idx_system_users_email; Type: INDEX; Schema: system; Owner: -
--

CREATE INDEX idx_system_users_email ON system.users USING btree (email);


--
-- Name: idx_system_users_username; Type: INDEX; Schema: system; Owner: -
--

CREATE INDEX idx_system_users_username ON system.users USING btree (username);


--
-- PostgreSQL database dump complete
--

\unrestrict d5RNBj5rudQEkmJVJiH8yvvjJ9TCVdWrgSjLr7ZvCXF0Ya0txhfDLSXg3GYc9oO

