#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - CORE MODULE (Thin Wrapper)
=================================================
DEPRECATED: File ini sekarang menjadi thin wrapper.
Semua fungsi sudah dipindahkan ke package tiktok_ads/

Usage tetap sama:
    from tiktok_ads_core import (
        load_product_names, load_data, analyze_products,
        generate_report_md, generate_html
    )

Atau langsung dari package:
    from tiktok_ads import run_full_analysis, run_quarterly_analysis
"""

# Re-export semua dari package tiktok_ads
from tiktok_ads import *

# Backward compatibility - tetap expose semua fungsi dan konstanta
__all__ = [
    # Config
    "TENANT_ID", "DB_PATH", "EXCEL_PRODUCT_REF", "OUTPUT_DIR",
    "MIN_PERIODS_FOR_TREND", "MIN_PERIODS_FOR_MOMENTUM", "MIN_COST_FOR_VALID",
    "MIN_SAMPLES_HIGH_CONFIDENCE",
    "WEIGHT_ROI", "WEIGHT_PROFIT", "WEIGHT_MOMENTUM", "WEIGHT_CONSISTENCY", "WEIGHT_TREND",
    "SCORE_THRESHOLD_GOOD", "SCORE_THRESHOLD_BAD",
    "ROI_THRESHOLD_GOOD", "ROI_THRESHOLD_BAD", "LOSS_THRESHOLD",
    
    # Period
    "get_period_bin", "get_period_label", "get_available_periods",
    "get_rolling_window_periods", "filter_data_by_window", "analyze_period_performance",
    
    # Stats
    "mann_kendall_test", "linear_trend_analysis", "calculate_momentum", "coefficient_of_variation",
    
    # Data
    "load_product_names", "load_data", "get_product_display_name", "summary_stats",
    
    # Analysis
    "assess_data_quality", "analyze_products",
    "get_top_products", "get_stop_products", "get_monitor_products",
    
    # Report
    "generate_report_md", "generate_html", "run_full_analysis", "run_quarterly_analysis",
]
