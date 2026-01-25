#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads Analysis Package v1.1.0
==================================
Modular ML-based Shopee Ads performance analysis.
Reuses core modules for platform-agnostic analytics.

Structure:
- config: Configuration and thresholds
- data: Data loading from DB/CSV
- analysis: Product analysis with ML scoring
- intelligence: Advanced ML (fatigue, saturation, lifecycle)
- reports: Report generation (HTML/MD)
"""

# Configuration
from .config import (
    TENANT_ID, DATA_DIR, OUTPUT_DIR,
    MIN_PERIODS_FOR_TREND, MIN_PERIODS_FOR_MOMENTUM, MIN_COST_FOR_VALID,
    MIN_SAMPLES_HIGH_CONFIDENCE,
    WEIGHT_ROI, WEIGHT_PROFIT, WEIGHT_MOMENTUM, WEIGHT_CONSISTENCY, WEIGHT_TREND,
    SCORE_THRESHOLD_GOOD, SCORE_THRESHOLD_BAD,
    ROI_THRESHOLD_GOOD, ROI_THRESHOLD_BAD, LOSS_THRESHOLD,
)

# Data functions
from .data import (
    load_all_csv_files, load_data, get_available_periods,
    summary_stats, get_period_from_filename,
)

# Analysis functions
from .analysis import (
    assess_data_quality, analyze_products,
    get_top_products, get_stop_products, get_monitor_products,
)

# Report functions
from .reports import (
    generate_report_md, generate_html, generate_executive_summary,
    run_full_analysis, run_quarterly_analysis,
)

__version__ = "1.1.0"
__author__ = "Omni Analytics"

__all__ = [
    # Config
    "TENANT_ID", "DATA_DIR", "OUTPUT_DIR",
    "MIN_PERIODS_FOR_TREND", "MIN_PERIODS_FOR_MOMENTUM", "MIN_COST_FOR_VALID",
    # Data
    "load_all_csv_files", "load_data", "get_available_periods", "summary_stats",
    # Analysis
    "assess_data_quality", "analyze_products",
    "get_top_products", "get_stop_products", "get_monitor_products",
    # Report
    "generate_report_md", "generate_html", "generate_executive_summary",
    "run_full_analysis", "run_quarterly_analysis",
]
