-- ============================================
-- Create Tenant Schemas
-- Run after init-multi-tenant.sql
-- ============================================

-- Create tenant schemas
CREATE SCHEMA IF NOT EXISTS tenant_yumna_bertigamart;
CREATE SCHEMA IF NOT EXISTS tenant_tika_nusseyba;

-- Insert tenant records
INSERT INTO system.tenants (tenant_id, shop_name, is_active, db_schema) VALUES
('yumna_bertigamart', 'Bertigamart', true, 'tenant_yumna_bertigamart'),
('tika_nusseyba', 'Nusseyba Shop', true, 'tenant_tika_nusseyba')
ON CONFLICT (tenant_id) DO NOTHING;

COMMIT;
