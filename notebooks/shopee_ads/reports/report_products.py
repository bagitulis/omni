#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads - Report Products Section
====================================
Functions for generating top and stop products sections.
"""

from typing import List, Dict


def _fmt(n: float) -> str:
    """Format number to Indonesian Rupiah."""
    return f"Rp {n:,.0f}".replace(",", ".")


def add_top_products(report: List[str], products: List[Dict], n: int = 10):
    """Add top products section with detailed info - matching TikTok format."""
    from ..analysis import get_top_products
    
    report.append("## 🚀 PRODUK REKOMENDASI SCALE UP\n")
    top = get_top_products(products, n=n)
    if not top:
        report.append("*Tidak ada produk yang memenuhi kriteria scale up*\n")
        return
    
    # TikTok-style columns: #, Produk, ROAS, Profit, Score, Rekomendasi, Alasan
    report.append("| # | Produk | ROAS | Profit | Score | Rekomendasi | Alasan |")
    report.append("|---|--------|------|--------|-------|-------------|--------|")
    for i, p in enumerate(top, 1):
        name = p.get('product_name', 'N/A')[:35]
        roi = p.get('roi', 0)
        profit = _fmt(p.get('profit', 0))
        score = p.get('score', 0)
        trend = p.get('trend', 'stable')
        # Determine recommendation
        if roi >= 5:
            reco = "🚀 +30%"
            reason = "ROI sangat tinggi"
        elif roi >= 3:
            reco = "📈 +20%"
            reason = "ROI bagus"
        else:
            reco = "📈 +10%"
            reason = "Performa stabil"
        report.append(f"| {i} | {name} | {roi:.1f}x | {profit} | {score:.0f} | {reco} | {reason} |")
    report.append("")


def add_reduce_products(report: List[str], products: List[Dict]):
    """Add reduce budget products section - matching TikTok TURUNKAN BUDGET."""
    from ..analysis import get_monitor_products
    
    report.append("## 📉 TURUNKAN BUDGET\n")
    report.append("Produk dengan performa di bawah rata-rata - kurangi budget untuk optimasi.\n")
    
    # Get products that should be monitored (potential reduce)
    monitor = get_monitor_products(products)
    # Filter to those with lower ROI
    reduce = [p for p in monitor if p.get('roi', 0) < 3.0]
    
    if not reduce:
        report.append("*Tidak ada produk yang perlu dikurangi budget-nya*\n")
        return
    
    # TikTok-style columns: #, Produk, Tipe, ROAS, Rekomendasi, Alasan
    report.append("| # | Produk | Tipe | ROAS | Budget | Rekomendasi | Alasan |")
    report.append("|---|--------|------|------|--------|-------------|--------|")
    for i, p in enumerate(sorted(reduce, key=lambda x: x.get('roi', 0))[:10], 1):
        name = p.get('product_name', 'N/A')[:30]
        bidding_mode = p.get('bidding_mode', 'Unknown')[:10]
        roi = p.get('roi', 0)
        budget = _fmt(p.get('total_cost', 0))
        # Determine recommendation based on ROI
        if roi < 1.5:
            reco = "📉 -30%"
            reason = "ROI rendah"
        elif roi < 2.0:
            reco = "📉 -20%"
            reason = "ROI di bawah target"
        else:
            reco = "➡️ Monitor"
            reason = "Perlu evaluasi"
        report.append(f"| {i} | {name} | {bidding_mode} | {roi:.2f}x | {budget} | {reco} | {reason} |")
    report.append("")


def add_stop_products(report: List[str], products: List[Dict]):
    """Add stop products section - matching TikTok format."""
    from ..analysis import get_stop_products
    
    report.append("## 🛑 HENTIKAN SEGERA\n")
    stop = get_stop_products(products)
    if not stop:
        report.append("✅ **Tidak ada produk yang perlu dihentikan** - Semua produk profitable!\n")
        return
    
    potential_savings = sum(abs(p.get("profit", 0)) for p in stop if p.get("profit", 0) < 0)
    
    report.append(f"**Jumlah Produk:** {len(stop)} | **Potensi Penghematan:** {_fmt(potential_savings)}\n")
    # TikTok-style columns: #, Produk, ROAS, Kerugian, Rekomendasi, Alasan
    report.append("| # | Produk | ROAS | Kerugian | Rekomendasi | Alasan |")
    report.append("|---|--------|------|----------|-------------|--------|")
    for i, p in enumerate(stop[:10], 1):
        loss = abs(p.get("profit", 0)) if p.get("profit", 0) < 0 else 0
        name = p.get('product_name', 'N/A')[:30]
        roi = p.get('roi', 0)
        # Determine reason
        if roi < 1.0:
            reason = "ROI negatif (rugi)"
        elif loss > 500000:
            reason = "Kerugian signifikan"
        else:
            reason = "Skor rendah"
        report.append(f"| {i} | {name} | {roi:.2f}x | {_fmt(loss)} | ⛔ STOP | {reason} |")
    report.append("")


def add_methodology(report: List[str]):
    """Add comprehensive methodology section - matching TikTok unified scoring."""
    report.append("## 📚 METODOLOGI ANALISIS\n")
    
    report.append("### 1. Unified Composite Scoring System (8 Komponen)\n")
    report.append("Sistem scoring yang sama dengan TikTok untuk konsistensi lintas platform:\n")
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
    
    report.append("### 2. Mann-Kendall Trend Test\n")
    report.append("- **Tujuan:** Mendeteksi trend monotonic dalam time series")
    report.append("- p-value < 0.05: Trend signifikan secara statistik")
    report.append("- τ (tau) > 0: Trend naik | τ < 0: Trend turun\n")
    
    report.append("### 3. Budget Elasticity Analysis\n")
    report.append("- **Tujuan:** Mengukur respons revenue terhadap perubahan budget")
    report.append("- E > 1.0: Elastic - revenue grows faster than cost")
    report.append("- 0 < E < 1.0: Inelastic - diminishing returns")
    report.append("- E < 0: Negative - perlu evaluasi\n")
    
    report.append("### 4. Churn Risk Assessment\n")
    report.append("- **Tujuan:** Memprediksi produk yang berisiko mengalami penurunan")
    report.append("- Faktor: ROI, Trend, Momentum, Volatility")
    report.append("- Score > 50: High Risk")
    report.append("- Score 25-50: Medium Risk")
    report.append("- Score < 25: Low Risk\n")
    
    report.append("### 5. Indonesian Event Calendar\n")
    report.append("- **Tujuan:** Menyesuaikan scoring berdasarkan konteks event")
    report.append("- Twin dates (11.11, 12.12): Boost multiplier")
    report.append("- Payday (25-5): Boost multiplier")
    report.append("- Hari besar: Lebaran, Natal, dll\n")
