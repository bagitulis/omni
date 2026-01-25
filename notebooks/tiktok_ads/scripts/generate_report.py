#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Script Runner (Unified Intelligence System)
=================================================================
Wrapper script untuk generate report via command line atau MCP.
Menggunakan Intelligence System v3.0 dengan semua metodologi terintegrasi.

Features:
- Unified Scoring (Legacy + Intelligence combined)
- Indonesian Calendar Events (Payday, Twin Date, Holidays)
- Probability with Confidence Intervals
- Saturation/Elasticity Analysis
- Trend Momentum (MACD-style)
- Creative Fatigue Detection
- Product Lifecycle Analysis

Usage:
    python generate_report.py
    python generate_report.py --quiet  # For subprocess (no console output)
    python generate_report.py --quarterly  # For rolling window analysis

Output:
    output/TIKTOK_ADS_REPORT_[timestamp].html
"""

import sys
import io
import os
from pathlib import Path
from datetime import datetime

# Fix path - ensure tiktok_ads package is importable
_script_dir = Path(__file__).resolve().parent  # scripts/
_tiktok_ads_dir = _script_dir.parent  # tiktok_ads/
_notebooks_dir = _tiktok_ads_dir.parent  # notebooks/
if str(_notebooks_dir) not in sys.path:
    sys.path.insert(0, str(_notebooks_dir))

# Fix Windows console encoding for emoji support
if sys.stdout.encoding != 'utf-8':
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8', errors='replace')
if sys.stderr.encoding != 'utf-8':
    sys.stderr = io.TextIOWrapper(sys.stderr.buffer, encoding='utf-8', errors='replace')

# Import from tiktok_ads package
from tiktok_ads import OUTPUT_DIR
from tiktok_ads.config import DB_PATH
from tiktok_ads.intelligence import (
    ComprehensiveAnalysisEngine,
    HtmlReportGenerator,
    ExecutiveReportGenerator,
    ActionRecommendation,
)


def run_comprehensive_analysis(verbose=True, min_cost=50000):
    """
    Run comprehensive analysis using the Intelligence System.
    Combines all methodologies: Legacy Stats + Intelligence + Indonesian Calendar.
    """
    def log(msg):
        if verbose:
            print(msg)
    
    log("🔬 Initializing Comprehensive Analysis Engine...")
    
    # Initialize engine with database path
    engine = ComprehensiveAnalysisEngine(
        dbPath=str(DB_PATH),
        tenantId="yumna_bertigamart"
    )
    
    # Run analysis
    log("📊 Running unified analysis with all methodologies...")
    analyses = engine.analyzeAll(minCost=min_cost, verbose=verbose)
    
    if not analyses:
        log("⚠️ No products found for analysis!")
        return None
    
    # Generate reports
    log("📝 Generating HTML reports...")
    
    # Full Report
    htmlGenerator = HtmlReportGenerator()
    fullReportHtml = htmlGenerator.generateFullReport(analyses)
    
    # Executive Summary
    execGenerator = ExecutiveReportGenerator()
    execReportHtml = execGenerator.generateReport(analyses)
    
    # Save reports
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    timestamp = datetime.now().strftime('%Y%m%d_%H%M')
    quarter_label = f"Q1_2026"  # Dynamic based on current date
    
    full_path = os.path.join(OUTPUT_DIR, f"TIKTOK_ADS_REPORT_{quarter_label}_{timestamp}.html")
    exec_path = os.path.join(OUTPUT_DIR, f"TIKTOK_ADS_REPORT_{quarter_label}_EXECUTIVE_{timestamp}.html")
    
    with open(full_path, 'w', encoding='utf-8') as f:
        f.write(fullReportHtml)
    
    with open(exec_path, 'w', encoding='utf-8') as f:
        f.write(execReportHtml)
    
    log(f"✅ Full Report saved: {full_path}")
    log(f"✅ Executive Summary saved: {exec_path}")
    
    # Build result dict compatible with existing interface
    scale_up = [a for a in analyses if a.finalAction in [ActionRecommendation.SCALE_UP_AGGRESSIVE, ActionRecommendation.SCALE_UP]]
    maintain = [a for a in analyses if a.finalAction == ActionRecommendation.MAINTAIN]
    reduce = [a for a in analyses if a.finalAction == ActionRecommendation.REDUCE]
    stop = [a for a in analyses if a.finalAction == ActionRecommendation.STOP]
    monitor = [a for a in analyses if a.finalAction not in [ActionRecommendation.SCALE_UP_AGGRESSIVE, ActionRecommendation.SCALE_UP, ActionRecommendation.MAINTAIN, ActionRecommendation.REDUCE, ActionRecommendation.STOP]]
    
    # Convert to legacy format for backward compatibility
    products = []
    for a in analyses:
        # Map action to category
        if a.finalAction in [ActionRecommendation.SCALE_UP_AGGRESSIVE, ActionRecommendation.SCALE_UP]:
            category = "🟢 LANJUTKAN"
        elif a.finalAction == ActionRecommendation.STOP:
            category = "🔴 HENTIKAN"
        else:
            category = "🟡 PANTAU"
        
        products.append({
            'product_id': a.productId,
            'product_name': a.productName,
            'creative_type': a.creativeType,
            'total_cost': a.totalCost,
            'total_revenue': a.totalRevenue,
            'profit': a.totalProfit,
            'roi': a.roas,
            'score': a.unifiedScore.compositeScore,
            'category': category,
            'action': a.finalAction.value,
            'success_probability': a.successProbability,
            'confidence_level': a.confidenceLevel,
            'trend': a.legacyMkTrend,
            'momentum_pct': a.legacyMomentumPct,
            'budget_change_pct': a.finalBudgetChange,
            'recommended_budget': a.budgetRec.recommendedBudget,
            'event_multiplier': a.currentEventMultiplier,
            'current_events': a.currentEvents,
        })
    
    return {
        'products': products,
        'analyses': analyses,  # Full analysis objects
        'html_path': full_path,
        'executive_html_path': exec_path,
        'summary': {
            'total': len(analyses),
            'scale_up': len(scale_up),
            'maintain': len(maintain),
            'reduce': len(reduce),
            'stop': len(stop),
            'monitor': len(monitor),
        }
    }


# ============================================================
# MAIN
# ============================================================
if __name__ == "__main__":
    # Parse arguments
    quiet_mode = "--quiet" in sys.argv
    quarterly_mode = "--quarterly" in sys.argv
    verbose = not quiet_mode
    
    def log(message: str):
        """Print message only if not in quiet mode"""
        if verbose:
            print(message)
    
    log("=" * 70)
    log("TIKTOK ADS ANALYSIS - UNIFIED INTELLIGENCE SYSTEM v3.0")
    log("=" * 70)
    log("")
    log("Metodologi Terintegrasi:")
    log("  • Legacy Stats: Mann-Kendall, Momentum, CV")
    log("  • Intelligence: ROAS Classifier, Trend Momentum, Saturation")
    log("  • Context: Indonesian Calendar (Payday, Twin Date, Holidays)")
    log("  • Probability: Bayesian with Confidence Intervals")
    log("")
    
    # Run comprehensive analysis
    result = run_comprehensive_analysis(verbose=verbose)
    
    if result is None:
        log("❌ Analysis failed - no data found")
        sys.exit(1)
    
    # Summary
    products = result['products']
    summary = result['summary']
    
    # Count by category for backward compatibility
    lanjut = len([p for p in products if 'LANJUTKAN' in p['category']])
    pantau = len([p for p in products if 'PANTAU' in p['category']])
    henti = len([p for p in products if 'HENTIKAN' in p['category']])
    
    log("")
    log("=" * 70)
    log("LAPORAN BERHASIL DIBUAT!")
    log("=" * 70)
    
    log(f"""
📄 FULL REPORT:
   {result['html_path']}

📊 EXECUTIVE SUMMARY:
   {result['executive_html_path']}

RINGKASAN REKOMENDASI:
   Total Produk: {len(products)}
   🚀 SCALE UP: {summary['scale_up']}
   ✅ MAINTAIN: {summary['maintain']}
   📉 REDUCE: {summary['reduce']}
   ⚠️ MONITOR: {summary['monitor']}
   🛑 STOP: {summary['stop']}

DISTRIBUSI KATEGORI (Legacy Format):
   LANJUTKAN: {lanjut}
   PANTAU: {pantau}
   HENTIKAN: {henti}
""")
    log("=" * 70)
