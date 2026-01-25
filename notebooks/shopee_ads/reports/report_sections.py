#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads - Report Sections
============================
Helper functions for generating individual report sections.
Follows TikTok Ads report structure for consistency.
"""

from typing import List, Dict, Any


def add_executive_summary(report: List[str], products: List[Dict], 
                          evaluation: Dict, period_info: Dict = None):
    """Add executive summary with health score."""
    report.append("## 🎯 RINGKASAN EKSEKUTIF\n")
    
    health = evaluation.get("health", {})
    risk = evaluation.get("risk", {})
    
    total_cost = sum(p["total_cost"] for p in products)
    total_revenue = sum(p["total_revenue"] for p in products)
    total_profit = sum(p["profit"] for p in products)
    avg_roi = total_revenue / total_cost if total_cost > 0 else 0
    
    if period_info:
        report.append(f"**Periode Analisis:** {period_info.get('start', 'N/A')} sampai {period_info.get('end', 'N/A')}")
        if period_info.get("rolling_window"):
            report.append(f" | **Metode:** Rolling Window {period_info.get('window_months', 3)} Bulan\n")
        else:
            report.append("\n")
    
    # Health Score section
    if health:
        report.append("### 📊 Skor Kesehatan Portfolio\n")
        report.append("| Metrik | Nilai | Kategori |")
        report.append("|--------|-------|----------|")
        report.append(f"| **Health Score** | **{health.get('score', 0)}/100** | **{health.get('grade', 'N/A')}** |")
        report.append(f"| Tingkat Risiko | {risk.get('level', 'N/A')} | Skor: {risk.get('score', 0)}/100 |")
        report.append("")
    
    # Main metrics
    report.append("### 💰 Metrik Utama Performa\n")
    report.append("| Metrik | Nilai |")
    report.append("|--------|-------|")
    report.append(f"| Total Investasi Iklan | Rp {total_cost:,.0f} |".replace(",", "."))
    report.append(f"| Total Pendapatan | Rp {total_revenue:,.0f} |".replace(",", "."))
    report.append(f"| **Total Keuntungan Bersih** | **Rp {total_profit:,.0f}** |".replace(",", "."))
    report.append(f"| **ROI Keseluruhan** | **{avg_roi:.2f}x** |")
    report.append(f"| Jumlah Produk Dianalisis | {len(products)} |")
    report.append("")
    
    # Distribution
    dist = evaluation.get("distribution", {})
    if dist:
        report.append("### 📈 Distribusi Rekomendasi Produk\n")
        report.append("| Rekomendasi | Jumlah Produk | Persentase |")
        report.append("|-------------|---------------|------------|")
        total = len(products)
        report.append(f"| 🟢 **LANJUTKAN** (Scale Up) | {dist.get('LANJUTKAN', 0)} | {dist.get('LANJUTKAN', 0)/total*100:.1f}% |")
        report.append(f"| 🟡 **PANTAU** (Monitor) | {dist.get('PANTAU', 0)} | {dist.get('PANTAU', 0)/total*100:.1f}% |")
        report.append(f"| 🔴 **HENTIKAN** (Stop) | {dist.get('HENTIKAN', 0)} | {dist.get('HENTIKAN', 0)/total*100:.1f}% |")
        report.append("")


def add_ai_evaluation_section(report: List[str], evaluation: Dict):
    """Add comprehensive AI evaluation section."""
    report.append("## 🤖 EVALUASI BERBASIS AI & MACHINE LEARNING\n")
    
    health = evaluation.get("health", {})
    components = health.get("components", {})
    
    if components:
        report.append("### Komponen Health Score\n")
        report.append("| Komponen | Skor | Bobot | Rumus |")
        report.append("|----------|------|-------|-------|")
        
        for name, data in components.items():
            report.append(f"| {name} | {data.get('score', 0)} | {data.get('weight', 0)} | {data.get('formula', '-')} |")
        report.append("")
    
    # Risk factors
    risk = evaluation.get("risk", {})
    factors = risk.get("factors", [])
    if factors:
        report.append("### ⚠️ Faktor Risiko Teridentifikasi\n")
        for factor in factors:
            report.append(f"- {factor}")
        report.append("")


def add_top_products_section(report: List[str], products: List[Dict], n: int = 10):
    """Add top products section."""
    report.append(f"## ⭐ TOP {n} PRODUK TERBAIK\n")
    
    # Sort by score descending
    sorted_products = sorted(products, key=lambda x: x.get("score", 0), reverse=True)[:n]
    
    if not sorted_products:
        report.append("*Tidak ada produk yang memenuhi kriteria*\n")
        return
    
    report.append("| # | Produk | ROI | Score | Trend | Aksi |")
    report.append("|---|--------|-----|-------|-------|------|")
    
    for i, p in enumerate(sorted_products, 1):
        name = p.get("product_name", "Unknown")[:40]
        roi = p.get("roi", 0)
        score = p.get("score", 0)
        trend = p.get("trend", "stable")
        action = p.get("action", "Monitor")
        report.append(f"| {i} | {name} | {roi:.2f}x | {score:.0f} | {trend} | {action} |")
    report.append("")


def add_stop_products_section(report: List[str], products: List[Dict]):
    """Add products to stop section."""
    report.append("## 🛑 PRODUK YANG HARUS DIHENTIKAN\n")
    
    # Filter products with HENTIKAN category
    stop_products = [p for p in products if "HENTIKAN" in p.get("category", "")]
    
    if not stop_products:
        report.append("*Tidak ada produk yang perlu dihentikan - semua profitable!* ✅\n")
        return
    
    report.append("| # | Produk | Kerugian | ROI | Alasan |")
    report.append("|---|--------|----------|-----|--------|")
    
    for i, p in enumerate(stop_products[:10], 1):
        name = p.get("product_name", "Unknown")[:40]
        loss = abs(p.get("profit", 0)) if p.get("profit", 0) < 0 else 0
        roi = p.get("roi", 0)
        action = p.get("action", "Stop")
        loss_str = f"Rp {loss:,.0f}".replace(",", ".")
        report.append(f"| {i} | {name} | {loss_str} | {roi:.2f}x | {action} |")
    report.append("")


def add_profit_estimation_section(report: List[str], evaluation: Dict, products: List[Dict]):
    """Add profit estimation if recommendations are followed."""
    report.append("## 💹 PROYEKSI DAMPAK KEUANGAN\n")
    
    # Calculate potential savings
    stop_products = [p for p in products if "HENTIKAN" in p.get("category", "")]
    potential_savings = sum(abs(p.get("profit", 0)) for p in stop_products if p.get("profit", 0) < 0)
    
    # Calculate potential gains from scaling
    scale_products = [p for p in products if "LANJUTKAN" in p.get("category", "")]
    avg_scale_roi = sum(p.get("roi", 0) for p in scale_products) / len(scale_products) if scale_products else 0
    
    report.append("### Proyeksi Jika Rekomendasi Diterapkan\n")
    report.append("| Aksi | Dampak Estimasi |")
    report.append("|------|-----------------|")
    report.append(f"| Hentikan produk rugi | Hemat Rp {potential_savings:,.0f} |".replace(",", "."))
    report.append(f"| Scale up produk terbaik | +20-50% revenue (avg ROI {avg_scale_roi:.1f}x) |")
    report.append("")
    
    if potential_savings > 0:
        report.append(f"> 💡 **Potensi Penghematan:** Dengan menghentikan {len(stop_products)} produk rugi, ")
        report.append(f"Anda dapat menghemat **Rp {potential_savings:,.0f}** per periode.\n".replace(",", "."))


def add_methodology_section(report: List[str]):
    """Add methodology explanation section with unified scoring formula."""
    report.append("## 📐 METODOLOGI ANALISIS\n")
    
    report.append("### 🧠 Unified Composite Score (8 Komponen)\n")
    report.append("Formula scoring yang sama dengan TikTok Ads untuk konsistensi lintas platform:\n")
    report.append("```")
    report.append("Composite Score = ")
    report.append("  ROAS Score × 0.20 +")
    report.append("  Trend Score × 0.15 +")
    report.append("  Momentum Score × 0.10 +")
    report.append("  Elasticity Score × 0.15 +")
    report.append("  Volatility Score × 0.10 +")
    report.append("  Fatigue Score × 0.10 +")
    report.append("  Churn Risk Score × 0.10 +")
    report.append("  Event Score × 0.10")
    report.append("```\n")
    
    report.append("### Rincian Komponen Scoring\n")
    report.append("| Komponen | Bobot | Deskripsi | Metode |")
    report.append("|----------|-------|-----------|--------|")
    report.append("| ROAS | 20% | Return on Ad Spend | Revenue/Cost ratio, scaled 0-100 |")
    report.append("| Trend | 15% | Arah performa | Mann-Kendall non-parametric test |")
    report.append("| Momentum | 10% | Performa terkini vs historis | 70/30 split comparison |")
    report.append("| Elasticity | 15% | Respons revenue terhadap budget | % change analysis |")
    report.append("| Volatility | 10% | Stabilitas performa | Coefficient of Variation |")
    report.append("| Fatigue | 10% | Kelelahan iklan/creative | Days active decay model |")
    report.append("| Churn Risk | 10% | Risiko penurunan | Multi-factor risk assessment |")
    report.append("| Event | 10% | Konteks event Indonesia | Calendar-based multiplier |")
    report.append("")
    
    report.append("### Category Thresholds\n")
    report.append("| Score Range | Kategori | Rekomendasi |")
    report.append("|-------------|----------|-------------|")
    report.append("| ≥ 80 | ⭐ STAR | Scale up agresif +30% |")
    report.append("| 65-79 | 🚀 GROWTH | Scale up +10-30% |")
    report.append("| 50-64 | ✅ STABLE | Maintain budget |")
    report.append("| 35-49 | 👀 WATCH | Monitor & optimize |")
    report.append("| < 35 | ⚠️ PROBLEM | Reduce atau Stop |")
    report.append("")
    
    report.append("### Statistical Methods\n")
    report.append("- **Mann-Kendall Test:** Non-parametric trend detection dengan p-value validation")
    report.append("- **Linear Regression:** Trend slope, R², dan direction analysis")
    report.append("- **Momentum Analysis:** 70/30 historical vs recent split")
    report.append("- **CV Analysis:** Consistency via coefficient of variation")
    report.append("- **Elasticity:** Budget response analysis per period")
    report.append("- **Churn Risk:** Multi-factor decline probability")
    report.append("")


def add_strategic_recommendations(report: List[str], products: List[Dict], evaluation: Dict):
    """Add strategic recommendations based on data analysis."""
    from .report_evaluation import generate_strategic_insights
    
    insights = generate_strategic_insights(products)
    if not insights:
        return
    
    report.append("## 🎯 REKOMENDASI STRATEGIS BERBASIS DATA\n")
    
    # Type emoji mapping
    type_emoji = {
        "SAVINGS": "💰",
        "GROWTH": "📈",
        "OPPORTUNITY": "🌟",
        "ALERT": "⚠️",
        "WARNING": "⚠️"
    }
    
    for i, insight in enumerate(insights, 1):
        itype = insight["type"]
        emoji = type_emoji.get(itype, "📌")
        report.append(f"### {i}. {emoji} {itype}\n")
        report.append(f"**Temuan Analisis:** {insight['insight']}\n")
        report.append(f"**Tindakan yang Disarankan:** {insight['action']}\n")
        report.append(f"**Metodologi:** {insight['methodology']}\n")
        report.append("")
