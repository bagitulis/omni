#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis Package v2.2.0
===================================
Modular ML-based TikTok Ads performance analysis.

Structure:
- config: Configuration and thresholds
- period: Period/time functions
- stats: Statistical tests (Mann-Kendall, regression)
- data: Data loading functions
- analysis: Product analysis
- confidence: CI and T-Test
- correlation: Correlation and churn risk
- optimization: Budget optimization
- health: Health score and risk assessment
- evaluation: Main AI evaluation
- report*: Report generation modules
"""

# Configuration
from .config import (
    TENANT_ID, DB_PATH, EXCEL_PRODUCT_REF, OUTPUT_DIR,
    MIN_PERIODS_FOR_TREND, MIN_PERIODS_FOR_MOMENTUM, MIN_COST_FOR_VALID,
    MIN_SAMPLES_HIGH_CONFIDENCE,
    WEIGHT_ROI, WEIGHT_PROFIT, WEIGHT_MOMENTUM, WEIGHT_CONSISTENCY, WEIGHT_TREND,
    SCORE_THRESHOLD_GOOD, SCORE_THRESHOLD_BAD,
    ROI_THRESHOLD_GOOD, ROI_THRESHOLD_BAD, LOSS_THRESHOLD,
)

# Period functions
from .period import (
    get_period_bin, get_period_label, get_available_periods,
    get_rolling_window_periods, filter_data_by_window, analyze_period_performance,
)

# Statistical functions
from .stats import (
    mann_kendall_test, linear_trend_analysis,
    calculate_momentum, coefficient_of_variation,
)

# Data functions
from .data import (
    load_product_names, load_data, get_product_display_name, summary_stats,
)

# Analysis functions
from .analysis import (
    assess_data_quality, analyze_products,
    get_top_products, get_stop_products, get_monitor_products,
)

# Confidence Interval & T-Test
from .confidence import (
    calculate_confidence_interval, calculate_profit_ci, compare_creative_types,
)

# Correlation & Churn Risk
from .correlation import (
    analyze_cost_revenue_correlation, calculate_churn_risk, analyze_portfolio_churn_risk,
)

# Budget Optimization
from .optimization import (
    calculate_optimal_budget, optimize_portfolio_budget,
)

# Health Score & Risk
from .health import (
    calculate_health_score, calculate_risk_assessment,
)

# Evaluation (main entry)
from .evaluation import (
    estimate_profit_impact, generate_ai_evaluation,
)

# Report functions - now in reports/ subfolder
from .reports.report import (
    generate_report_md, run_full_analysis, run_quarterly_analysis,
)
from .reports.report_html import generate_html
from .reports.report_executive import (
    generate_executive_summary, generate_executive_summary_md,
    get_high_impact_products, get_all_stop_products_categorized,
    get_potential_restart_products, get_maintain_budget_products,
    calculate_confidence_label, estimate_with_methodology,
)

# Backward compatibility - re-export from advanced_stats
from .advanced_stats import *

# NEW: Intelligence System v3.0
from .intelligence import (
    IntelligenceEngine, IntelligenceConfig, ProductAnalysis,
    RecommendationAction,
)

__version__ = "3.0.0"
__author__ = "TikTok Ads ML Analysis"

__all__ = [
    # Config
    "TENANT_ID", "DB_PATH", "EXCEL_PRODUCT_REF", "OUTPUT_DIR",
    "MIN_PERIODS_FOR_TREND", "MIN_PERIODS_FOR_MOMENTUM", "MIN_COST_FOR_VALID",
    # Period
    "get_period_bin", "get_period_label", "get_available_periods",
    "get_rolling_window_periods", "filter_data_by_window",
    # Stats
    "mann_kendall_test", "linear_trend_analysis", "calculate_momentum", "coefficient_of_variation",
    # Data
    "load_product_names", "load_data", "get_product_display_name", "summary_stats",
    # Analysis
    "assess_data_quality", "analyze_products", "get_top_products", "get_stop_products",
    # Advanced Stats
    "calculate_confidence_interval", "calculate_profit_ci", "compare_creative_types",
    "analyze_cost_revenue_correlation", "calculate_churn_risk", "analyze_portfolio_churn_risk",
    "calculate_optimal_budget", "optimize_portfolio_budget",
    # Evaluation
    "calculate_health_score", "calculate_risk_assessment", "estimate_profit_impact", "generate_ai_evaluation",
    # Report
    "generate_report_md", "generate_html", "run_full_analysis", "run_quarterly_analysis",
    "generate_executive_summary", "get_high_impact_products", "get_all_stop_products_categorized",
    "get_potential_restart_products", "get_maintain_budget_products",
    # NEW: Intelligence System
    "IntelligenceEngine", "IntelligenceConfig", "IntelligenceIntegration",
    "run_enhanced_analysis", "ProductAnalysis", "EnhancedProductAnalysis",
    "RecommendationAction",
]
