#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads Reports Package
==========================
Report generation modules for Shopee Ads analysis.
"""

from .report import (
    generate_report_md,
    run_full_analysis,
    run_quarterly_analysis,
)
from .report_html import generate_html
from .report_executive import generate_executive_summary
from .report_evaluation import generate_evaluation
from .report_sections import (
    add_executive_summary,
    add_ai_evaluation_section,
    add_top_products_section,
    add_stop_products_section,
    add_profit_estimation_section,
    add_methodology_section,
    add_strategic_recommendations,
)
from .report_advanced import (
    add_efficiency_section,
    add_bidding_comparison_section,
    add_budget_optimization_section,
    add_trend_analysis_section,
    add_data_quality_section,
)
from .report_products import (
    add_top_products,
    add_stop_products,
    add_methodology,
)
from .report_projection import add_financial_projection
from .report_evaluation import generate_strategic_insights

__all__ = [
    # Main report functions
    "generate_report_md",
    "generate_html",
    "generate_executive_summary",
    "generate_evaluation",
    "run_full_analysis",
    "run_quarterly_analysis",
    # Section helpers - basic
    "add_executive_summary",
    "add_ai_evaluation_section",
    "add_top_products_section",
    "add_stop_products_section",
    "add_profit_estimation_section",
    "add_methodology_section",
    "add_strategic_recommendations",
    # Section helpers - advanced
    "add_efficiency_section",
    "add_bidding_comparison_section",
    "add_budget_optimization_section",
    "add_trend_analysis_section",
    "add_data_quality_section",
    # Products
    "add_top_products",
    "add_stop_products",
    "add_methodology",
    # Projection
    "add_financial_projection",
    # Evaluation
    "generate_strategic_insights",
]
