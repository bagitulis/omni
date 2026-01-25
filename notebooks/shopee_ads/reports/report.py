#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads Analysis - Report Generation (Main Entry Point)
==========================================================
Main functions for generating comprehensive analysis reports.
HTML-only output matching TikTok Ads report quality.
"""

from datetime import datetime
from typing import List, Dict

from ..config import OUTPUT_DIR
from ..data import load_data, get_available_periods, filter_data_by_window
from ..analysis import analyze_products

from .report_html import generate_html
from .report_executive import generate_executive_summary
from .report_evaluation import generate_evaluation
from .executive_report_generator import generate_executive_html_report
from .report_advanced import (
    add_efficiency_section, add_bidding_comparison_section,
    add_budget_optimization_section, add_trend_analysis_section,
    add_data_quality_section, add_churn_risk_section, add_correlation_section
)
from .report_sections import add_strategic_recommendations
from .report_products import add_top_products, add_reduce_products, add_stop_products, add_methodology
from .report_projection import add_financial_projection


def generate_report_md(products: List[Dict], title: str = "Shopee Ads Report",
                       period_info: Dict = None, verbose: bool = True) -> tuple:
    """Generate comprehensive professional Markdown report."""
    if verbose:
        print("📝 Generating Markdown report...")
    
    evaluation = generate_evaluation(products, period_info)
    
    report = []
    report.append(f"# {title}")
    report.append(f"\n**Generated:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    report.append("**System:** Shopee Ads Intelligence System v1.1.0\n")
    
    # Core sections
    _add_executive_summary(report, evaluation, period_info, products)
    _add_ai_evaluation(report, evaluation, products)
    add_financial_projection(report, products, evaluation)
    add_strategic_recommendations(report, products, evaluation)
    add_top_products(report, products)
    add_reduce_products(report, products)
    add_stop_products(report, products)
    
    # Advanced sections (matching TikTok)
    add_correlation_section(report, products)
    add_churn_risk_section(report, products)
    add_efficiency_section(report, products)
    add_bidding_comparison_section(report, products)
    add_budget_optimization_section(report, products)
    add_trend_analysis_section(report, products)
    add_data_quality_section(report, products)
    
    # Methodology
    add_methodology(report)
    
    return "\n".join(report), evaluation


def _add_executive_summary(report: List[str], evaluation: Dict, period_info: Dict, products: List[Dict]):
    """Add executive summary with health score - matching TikTok structure."""
    report.append("\n## 🎯 RINGKASAN EKSEKUTIF\n")
    
    def fmt(n): return f"Rp {n:,.0f}".replace(",", ".")
    
    if period_info:
        report.append(f"**Periode Analisis:** {period_info.get('start', 'N/A')} sampai {period_info.get('end', 'N/A')}")
        if period_info.get("rolling_window"):
            report.append(f" | **Metode:** Rolling Window {period_info.get('window_months', 3)} Bulan\n")
        else:
            report.append("\n")
    
    # Health Score
    health = evaluation.get("health", {})
    risk = evaluation.get("risk", {})
    report.append("### 📊 Skor Kesehatan Portfolio\n")
    report.append("| Metrik | Nilai | Kategori |")
    report.append("|--------|-------|----------|")
    report.append(f"| **Health Score** | **{health.get('score', 0):.1f}/100** | **{health.get('grade', 'N/A')}** |")
    report.append(f"| Tingkat Risiko | {risk.get('level_emoji', '')} {risk.get('level', 'N/A')} | Skor: {risk.get('score', 0)}/100 |")
    report.append(f"| Status Keseluruhan | {health.get('interpretation', 'N/A')} | - |")
    report.append("")
    
    # Financial Overview
    report.append("### 💰 Metrik Utama Performa\n")
    report.append("| Metrik | Nilai |")
    report.append("|--------|-------|")
    report.append(f"| Total Investasi Iklan | {fmt(evaluation['total_cost'])} |")
    report.append(f"| Total Pendapatan | {fmt(evaluation['total_revenue'])} |")
    report.append(f"| **Total Keuntungan Bersih** | **{fmt(evaluation['total_profit'])}** |")
    report.append(f"| **ROI Keseluruhan** | **{evaluation['roi']:.2f}x** |")
    report.append(f"| Jumlah Produk Dianalisis | {evaluation['total_products']} |")
    report.append("")
    
    # Distribution
    dist = evaluation["distribution"]
    total = max(evaluation["total_products"], 1)
    report.append("### 📈 Distribusi Rekomendasi Produk\n")
    report.append("| Rekomendasi | Jumlah Produk | Persentase |")
    report.append("|-------------|---------------|------------|")
    report.append(f"| 🟢 **LANJUTKAN** (Scale Up) | {dist['LANJUTKAN']} | {dist['LANJUTKAN']/total*100:.1f}% |")
    report.append(f"| 🟡 **PANTAU** (Monitor) | {dist['PANTAU']} | {dist['PANTAU']/total*100:.1f}% |")
    report.append(f"| 🔴 **HENTIKAN** (Stop) | {dist['HENTIKAN']} | {dist['HENTIKAN']/total*100:.1f}% |")
    report.append("")


def _add_ai_evaluation(report: List[str], evaluation: Dict, products: List[Dict]):
    """Add AI/ML evaluation section with health score components."""
    report.append("## 🤖 EVALUASI BERBASIS AI & MACHINE LEARNING\n")
    
    health = evaluation.get("health", {})
    components = health.get("components", {})
    
    report.append("### Komponen Health Score\n")
    report.append("Tabel berikut menunjukkan rincian perhitungan Health Score portfolio iklan:\n")
    report.append("| Komponen | Skor | Bobot | Nilai Aktual | Rumus Perhitungan |")
    report.append("|----------|------|-------|--------------|-------------------|")
    for name, data in components.items():
        value = data.get('value', '-')
        report.append(f"| {name} | {data.get('score', 0):.1f} | {data.get('weight', 0)}% | {value} | {data.get('formula', '-')} |")
    report.append("")
    
    # Risk factors
    risk = evaluation.get("risk", {})
    report.append(f"### ⚠️ Penilaian Risiko Portfolio\n")
    report.append(f"**Tingkat Risiko:** {risk.get('level_emoji', '')} {risk.get('level', 'N/A')} (Skor: {risk.get('score', 0)}/100)\n")
    
    factors = risk.get("factors", [])
    if factors:
        report.append("**Faktor-Faktor Risiko yang Teridentifikasi:**\n")
        for factor in factors:
            report.append(f"- {factor}")
    else:
        report.append("✅ Tidak ada faktor risiko signifikan teridentifikasi.\n")
    report.append("")


def run_full_analysis(verbose: bool = True) -> Dict:
    """Run full analysis with all available data - HTML only."""
    df = load_data(verbose=verbose)
    periods = get_available_periods(df)
    period_info = {
        "start": periods[0] if periods else None,
        "end": periods[-1] if periods else None,
        "count": len(periods),
        "rolling_window": False
    }
    
    products = analyze_products(df, verbose=verbose)
    report_md, evaluation = generate_report_md(products, "Shopee Ads - Full Analysis", period_info, verbose)
    report_html = generate_html(report_md, "Shopee Ads Intelligence Report")
    
    # Use new executive HTML generator (aligned with TikTok format)
    exec_html = generate_executive_html_report(products, evaluation, period_info)
    
    return _save_reports(report_html, products, period_info, verbose, evaluation, exec_html)


def run_quarterly_analysis(verbose: bool = True) -> Dict:
    """Run quarterly rolling window analysis - HTML only."""
    df = load_data(verbose=verbose)
    df_filtered = filter_data_by_window(df, n_months=3, verbose=verbose)
    periods = get_available_periods(df_filtered)
    period_info = {
        "start": periods[0] if periods else None,
        "end": periods[-1] if periods else None,
        "count": len(periods),
        "rolling_window": True,
        "window_months": 3
    }
    
    products = analyze_products(df_filtered, verbose=verbose)
    
    if periods:
        year = periods[-1].split("-")[0]
        week = int(periods[-1].split("-W")[1]) if "-W" in periods[-1] else 1
        quarter = f"Q{(week-1)//13 + 1}"
        title = f"Shopee Ads Analysis - {quarter} {year}"
    else:
        title = "Shopee Ads Analysis - Quarterly"
    
    report_md, evaluation = generate_report_md(products, title, period_info, verbose)
    report_html = generate_html(report_md, f"{title} - Intelligence Report")
    
    # Use new executive HTML generator (aligned with TikTok format)
    exec_html = generate_executive_html_report(products, evaluation, period_info)
    
    return _save_reports(report_html, products, period_info, verbose, evaluation, exec_html)


def _save_reports(report_html: str, products: List[Dict], period_info: Dict,
                  verbose: bool, evaluation: Dict, exec_html: str = None) -> Dict:
    """Save HTML reports only to output directory."""
    timestamp = datetime.now().strftime("%Y%m%d_%H%M")
    
    if period_info and period_info.get("end"):
        end = period_info["end"]
        if "-W" in str(end):
            year, week = end.split("-")[0], int(end.split("-W")[1])
            period_str = f"Q{(week-1)//13 + 1}_{year}"
        else:
            period_str = str(end).replace("-", "")[:6]
    else:
        period_str = "FULL"
    
    base_name = f"SHOPEE_ADS_REPORT_{period_str}_{timestamp}"
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    
    html_path = OUTPUT_DIR / f"{base_name}.html"
    html_path.write_text(report_html, encoding="utf-8")
    
    result = {
        "success": True,
        "full_report": str(html_path),
        "products_analyzed": len(products),
        "evaluation": evaluation,
    }
    
    if exec_html:
        exec_html_path = OUTPUT_DIR / f"{base_name}_EXECUTIVE.html"
        exec_html_path.write_text(exec_html, encoding="utf-8")
        result["executive_summary"] = str(exec_html_path)
    
    if verbose:
        print(f"   📄 Reports saved to {OUTPUT_DIR}")
    
    return result
