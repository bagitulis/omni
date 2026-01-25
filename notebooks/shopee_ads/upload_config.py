#!/usr/bin/env python
"""
Shopee Ads - Upload Configuration
=================================
Configuration constants for CSV-to-database upload operations.
"""

import os
from pathlib import Path

# ============================================================
# PATHS
# ============================================================
_SCRIPT_DIR = Path(__file__).parent.resolve()
ADS_DATA_DIR = _SCRIPT_DIR / "data"

# ============================================================
# TENANT
# ============================================================
TENANT_ID = "yumna_bertigamart"

# ============================================================
# POSTGRES CONNECTION (fallback to env vars)
# ============================================================
PG_HOST = os.getenv("PG_HOST", "localhost")
PG_PORT = int(os.getenv("PG_PORT", "5433"))
PG_USER = os.getenv("PG_USER", "omni")
PG_PASSWORD = os.getenv("PG_PASSWORD", "omni_secure_2026")
PG_DATABASE = os.getenv("PG_DATABASE", "omni_main")
PG_SCHEMA = os.getenv("PG_SCHEMA", f"tenant_{TENANT_ID}")

# ============================================================
# CSV PARSING
# ============================================================
CSV_SKIP_ROWS = 8
CSV_ENCODING = "utf-8"

# ============================================================
# COLUMN MAPPING (Indonesian CSV → Database snake_case)
# ============================================================
CSV_TO_DB_COLUMN_MAP = {
    "Urutan": "order",
    "Nama Iklan": "product_name",
    "Status": "status",
    "Kode Produk": "product_id",
    "Mode Bidding": "bidding_mode",
    "Penempatan Iklan": "placement",
    "Tanggal Mulai": "start_date",
    "Tanggal Selesai": "end_date",
    "Dilihat": "impressions",
    "Jumlah Klik": "clicks",
    "Persentase Klik": "ctr",
    "Konversi": "conversions",
    "Konversi Langsung": "direct_conversions",
    "Tingkat konversi": "conversion_rate",
    "Tingkat Konversi Langsung": "direct_conversion_rate",
    "Biaya per Konversi": "cost_per_conversion",
    "Biaya per Konversi Langsung": "cost_per_direct_conversion",
    "Produk Terjual": "units_sold",
    "Terjual Langsung": "direct_units_sold",
    "Omzet Penjualan": "revenue",
    "Penjualan Langsung (GMV Langsung)": "direct_revenue",
    "Biaya": "cost",
    "Efektifitas Iklan": "roas",
    "Efektivitas Langsung": "direct_roas",
    "Persentase Biaya Iklan terhadap Penjualan dari Iklan (ACOS)": "acos",
    "Persentase Biaya Iklan terhadap Penjualan dari Iklan Langsung (ACOS Langsung)": "direct_acos",
}

# ============================================================
# DATABASE COLUMNS (ordered for INSERT)
# ============================================================
DB_COLUMNS = [
    'tenant_id', 'upload_batch_id', 'period_start', 'period_end', 'period_label',
    'product_id', 'product_name', 'status', 'bidding_mode', 'placement',
    'start_date', 'end_date', 'impressions', 'clicks', 'ctr',
    'conversions', 'direct_conversions', 'conversion_rate', 'direct_conversion_rate',
    'cost_per_conversion', 'cost_per_direct_conversion',
    'units_sold', 'direct_units_sold', 'revenue', 'direct_revenue',
    'cost', 'roas', 'direct_roas', 'acos', 'direct_acos',
    'created_at', 'updated_at'
]

# Numeric columns that need conversion
NUMERIC_COLUMNS = [
    'impressions', 'clicks', 'ctr', 'conversions', 'direct_conversions',
    'conversion_rate', 'direct_conversion_rate', 'cost_per_conversion', 'cost_per_direct_conversion',
    'units_sold', 'direct_units_sold', 'revenue', 'direct_revenue',
    'cost', 'roas', 'direct_roas', 'acos', 'direct_acos'
]

# Percentage columns (need % removal)
PERCENTAGE_COLUMNS = ['ctr', 'conversion_rate', 'direct_conversion_rate', 'acos', 'direct_acos']
