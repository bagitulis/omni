#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Report Generation (Main Entry Point)
===========================================================
Main functions for generating comprehensive analysis reports.
"""

import os
from datetime import datetime

from ..config import OUTPUT_DIR
from ..data import load_product_names, load_data
from ..period import get_available_periods, get_rolling_window_periods
from ..analysis import analyze_products
from ..evaluation import generate_ai_evaluation

# Import section builders
from .report_sections import (
    add_executive_summary,
    add_ai_evaluation_section,
    add_profit_estimation_section,
    add_strategic_recommendations,
)
from .report_advanced import (
    add_correlation_section,
    add_churn_risk_section,
    add_budget_optimization_section,
    add_data_quality_section,
    add_top_products_section,
    add_stop_products_section,
    add_creative_type_section,
)
from .report_methodology import add_methodology_section
from .report_html import generate_html
from .report_executive import generate_executive_summary


def generate_report_md(products, title="TikTok Ads Performance Report", period_info=None, verbose=True):
    """Generate comprehensive professional Markdown report with AI evaluation."""
    if verbose:
        print("📝 Generating Markdown report...")
    
    evaluation = generate_ai_evaluation(products, period_info)
    
    report = []
    report.append(f"# {title}")
    report.append(f"\n**Generated:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
    
    # Core sections
    add_executive_summary(report, products, evaluation, period_info)
    add_ai_evaluation_section(report, evaluation)
    add_profit_estimation_section(report, evaluation)
    add_strategic_recommendations(report, products, evaluation)
    
    # Advanced statistical sections
    add_correlation_section(report, evaluation)
    add_churn_risk_section(report, evaluation)
    add_budget_optimization_section(report, evaluation)
    
    # Product lists
    add_data_quality_section(report, products)
    add_top_products_section(report, products)
    add_stop_products_section(report, products)
    add_creative_type_section(report, products)
    
    # Methodology
    add_methodology_section(report)
    
    return "\n".join(report), evaluation


def run_full_analysis(verbose=True):
    """Run full analysis with AI evaluation."""
    product_map = load_product_names(verbose=verbose)
    df = load_data(verbose=verbose)
    
    periods = get_available_periods(df)
    period_info = {
        'start': periods[0] if periods else None,
        'end': periods[-1] if periods else None,
        'count': len(periods),
        'rolling_window': False
    }
    
    products = analyze_products(df, product_map, verbose=verbose)
    
    report_md, evaluation = generate_report_md(products, "TikTok Ads - Full Analysis Report", period_info, verbose=verbose)
    report_html = generate_html(report_md, "TikTok Ads Analysis Report")
    
    return _save_reports(report_md, report_html, products, period_info, "full_analysis", verbose, evaluation=evaluation)


def run_quarterly_analysis(verbose=True):
    """Run quarterly rolling window analysis with AI evaluation."""
    product_map = load_product_names(verbose=verbose)
    df = load_data(verbose=verbose)
    
    window = get_rolling_window_periods(df, n_months=3)
    period_info = {
        'start': window['periods'][0] if window['periods'] else None,
        'end': window['periods'][-1] if window['periods'] else None,
        'count': window['n_periods'],
        'rolling_window': True,
        'window_months': 3,
        'quarter_label': window['quarter_label']
    }
    
    products = analyze_products(df, product_map, verbose=verbose, use_rolling_window=True, window_months=3)
    
    title = f"TikTok Ads - {window['quarter_label']} Analysis (Rolling Window 3 Bulan)"
    report_md, evaluation = generate_report_md(products, title, period_info, verbose=verbose)
    report_html = generate_html(report_md, f"TikTok Ads {window['quarter_label']}")
    
    quarter = window['quarter_label'].replace(' ', '_')
    return _save_reports(report_md, report_html, products, period_info, f"TIKTOK_ADS_REPORT_{quarter}", verbose, evaluation=evaluation)


def _save_reports(report_md, report_html, products, period_info, prefix, verbose, evaluation=None):
    """Save reports to disk and return result dict. Only HTML files are saved."""
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    
    timestamp = datetime.now().strftime('%Y%m%d_%H%M')
    html_path = os.path.join(OUTPUT_DIR, f"{prefix}_{timestamp}.html")
    
    with open(html_path, 'w', encoding='utf-8') as f:
        f.write(report_html)
    
    if verbose:
        print(f"✅ Full Report saved to {OUTPUT_DIR}")
    
    # Generate Executive Summary if evaluation is provided
    exec_html_path = None
    
    if evaluation is not None:
        exec_summary = generate_executive_summary(products, evaluation, period_info)
        exec_html_path = os.path.join(OUTPUT_DIR, f"{prefix}_EXECUTIVE_{timestamp}.html")
        
        with open(exec_html_path, 'w', encoding='utf-8') as f:
            f.write(exec_summary['html'])
        
        if verbose:
            print(f"✅ Executive Summary saved to {OUTPUT_DIR}")
    
    return {
        'products': products,
        'report_md': report_md,
        'report_html': report_html,
        'period_info': period_info,
        'output_files': {
            'html': html_path,
            'executive_html': exec_html_path
        },
        'html_path': html_path,
        'executive_html_path': exec_html_path
    }
