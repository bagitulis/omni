#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Configuration
===================================
Configuration constants dan thresholds.
"""

import os
from pathlib import Path

# ============================================================
# PATHS
# ============================================================
PACKAGE_DIR = Path(__file__).parent  # tiktok_ads/
BASE_DIR = PACKAGE_DIR.parent        # notebooks/
EXCEL_PRODUCT_REF = PACKAGE_DIR / "data/Tiktoksellercenter_batchedit_20260114_basic_information_template.xlsx"
OUTPUT_DIR = PACKAGE_DIR / "output"

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
