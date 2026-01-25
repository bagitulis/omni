#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Executive Summary Report
===============================================
Ringkasan eksekutif untuk atasan dan stakeholder.
Semua rekomendasi menggunakan metodologi statistik yang ketat.

Structure (refactored per AGENTS.MD max 300 lines):
- scripts/exec_helpers.py: Helper functions (confidence, estimation)
- scripts/exec_categorizers.py: Product categorization functions
- scripts/exec_generator.py: Report generation functions
"""

from .report_html import generate_html
from ..scripts.exec_helpers import calculate_confidence_label, estimate_with_methodology
from ..scripts.exec_categorizers import (
    get_high_impact_products,
    get_potential_restart_products,
    get_maintain_budget_products,
    get_all_stop_products_categorized,
)
from ..scripts.exec_generator import generate_executive_summary_md


def generate_executive_summary(products, evaluation, period_info=None):
    """
    Generate both MD and HTML executive summary.
    
    Args:
        products: List of analyzed products
        evaluation: AI evaluation results
        period_info: Period information dict
        
    Returns:
        dict: {'md': markdown_content, 'html': html_content}
    """
    md_content = generate_executive_summary_md(products, evaluation, period_info)
    html_content = generate_html(md_content, "TikTok Ads - Executive Summary")
    
    return {
        'md': md_content,
        'html': html_content
    }


# Re-export for backward compatibility
__all__ = [
    'generate_executive_summary',
    'generate_executive_summary_md',
    'get_high_impact_products',
    'get_potential_restart_products',
    'get_maintain_budget_products',
    'get_all_stop_products_categorized',
    'calculate_confidence_label',
    'estimate_with_methodology',
]
