#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads Analysis - Configuration
===================================
Configuration constants dan thresholds.
"""

from pathlib import Path

# ============================================================
# PATHS
# ============================================================
PACKAGE_DIR = Path(__file__).parent  # shopee_ads/
BASE_DIR = PACKAGE_DIR.parent        # notebooks/
DATA_DIR = PACKAGE_DIR / "data"
OUTPUT_DIR = PACKAGE_DIR / "output"

# Create output dir if not exists
OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

# ============================================================
# TENANT
# ============================================================
TENANT_ID = "yumna_bertigamart"

# ============================================================
# COLUMN MAPPINGS (Indonesian CSV → Internal)
# ============================================================
COLUMN_MAP = {
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

# Columns to keep after renaming
REQUIRED_COLUMNS = [
    "product_id", "product_name", "status", "bidding_mode",
    "impressions", "clicks", "conversions", "units_sold",
    "revenue", "cost", "roas",
]

# ============================================================
# DATA QUALITY THRESHOLDS
# ============================================================
MIN_PERIODS_FOR_TREND = 4      # Minimum periods untuk analisis trend yang valid
MIN_PERIODS_FOR_MOMENTUM = 3   # Minimum periods untuk momentum analysis
MIN_COST_FOR_VALID = 10000     # Minimum cost (Rp) untuk dianggap valid
MIN_SAMPLES_HIGH_CONFIDENCE = 8  # Minimum samples untuk high confidence

# ============================================================
# SCORING WEIGHTS
# ============================================================
WEIGHT_ROI = 0.30
WEIGHT_PROFIT = 0.25
WEIGHT_MOMENTUM = 0.20
WEIGHT_CONSISTENCY = 0.15
WEIGHT_TREND = 0.10

# ============================================================
# CATEGORY THRESHOLDS
# ============================================================
SCORE_THRESHOLD_GOOD = 60      # Score >= ini = LANJUTKAN
SCORE_THRESHOLD_BAD = 40       # Score < ini = HENTIKAN
ROI_THRESHOLD_GOOD = 2.0       # ROI >= ini untuk LANJUTKAN
ROI_THRESHOLD_BAD = 1.0        # ROI < ini = HENTIKAN
LOSS_THRESHOLD = -100000       # Rugi > ini = HENTIKAN

# ============================================================
# CSV PARSING OPTIONS
# ============================================================
CSV_SKIP_ROWS = 8              # Skip header rows in Shopee CSV
CSV_ENCODING = "utf-8"
