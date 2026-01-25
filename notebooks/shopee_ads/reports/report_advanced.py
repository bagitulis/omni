#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads - Report Sections (Advanced)
=======================================
Correlation, Efficiency, Budget Optimization sections.
Matches TikTok Ads report structure.
"""

from typing import List, Dict, Any
import numpy as np


def add_efficiency_section(report: List[str], products: List[Dict]):
    """Add spending efficiency analysis section."""
    report.append("## 📈 ANALISIS EFISIENSI PENGELUARAN\n")
    
    if not products:
        report.append("*Data tidak mencukupi untuk analisis efisiensi.*\n")
        return
    
    # Calculate averages
    avg_cost = np.mean([p.get("total_cost", 0) for p in products])
    avg_roi = np.mean([p.get("roi", 0) for p in products])
    median_cost = np.median([p.get("total_cost", 0) for p in products])
    
    # High efficiency: low cost, high ROI
    efficient = [p for p in products if p.get("total_cost", 0) < median_cost and p.get("roi", 0) > avg_roi]
    # Low efficiency: high cost, low ROI
    inefficient = [p for p in products if p.get("total_cost", 0) > avg_cost and p.get("roi", 0) < avg_roi]
    
    if efficient:
        report.append("### 🎯 Produk dengan Efisiensi Tinggi (Kandidat Scale Up)\n")
        report.append("*Pengeluaran rendah dengan ROI tinggi - potensi yang belum dimanfaatkan*\n")
        report.append("| Produk | Pengeluaran | ROI | Peluang |")
        report.append("|--------|-------------|-----|---------|")
        for p in sorted(efficient, key=lambda x: x.get("roi", 0), reverse=True)[:5]:
            cost_str = f"Rp {p.get('total_cost', 0):,.0f}".replace(",", ".")
            report.append(f"| {p.get('product_name', 'N/A')[:30]} | {cost_str} | {p.get('roi', 0):.1f}x | Scale up prioritas |")
        report.append("")
    
    if inefficient:
        report.append("### ⚠️ Produk dengan Efisiensi Rendah (Perlu Evaluasi)\n")
        report.append("*Pengeluaran tinggi dengan ROI di bawah rata-rata*\n")
        report.append("| Produk | Pengeluaran | ROI | Masalah |")
        report.append("|--------|-------------|-----|---------|")
        for p in sorted(inefficient, key=lambda x: x.get("total_cost", 0), reverse=True)[:5]:
            cost_str = f"Rp {p.get('total_cost', 0):,.0f}".replace(",", ".")
            report.append(f"| {p.get('product_name', 'N/A')[:30]} | {cost_str} | {p.get('roi', 0):.1f}x | Review budget |")
        report.append("")


def add_bidding_comparison_section(report: List[str], products: List[Dict]):
    """Add bidding mode comparison section (similar to TikTok creative type)."""
    report.append("## 📱 PERBANDINGAN MODE BIDDING\n")
    
    # Group by bidding mode
    modes = {}
    for p in products:
        mode = p.get("bidding_mode", "Unknown")
        if mode not in modes:
            modes[mode] = {"products": [], "cost": 0, "revenue": 0}
        modes[mode]["products"].append(p)
        modes[mode]["cost"] += p.get("total_cost", 0)
        modes[mode]["revenue"] += p.get("total_revenue", 0)
    
    if len(modes) < 2:
        report.append("*Hanya ada satu mode bidding - tidak ada perbandingan.*\n")
        return
    
    report.append("### Performa per Mode Bidding\n")
    report.append("| Mode Bidding | Produk | Total Biaya | Total Revenue | ROI |")
    report.append("|--------------|--------|-------------|---------------|-----|")
    
    for mode, data in modes.items():
        roi = data["revenue"] / data["cost"] if data["cost"] > 0 else 0
        cost_str = f"Rp {data['cost']:,.0f}".replace(",", ".")
        rev_str = f"Rp {data['revenue']:,.0f}".replace(",", ".")
        report.append(f"| {mode} | {len(data['products'])} | {cost_str} | {rev_str} | **{roi:.2f}x** |")
    report.append("")
    
    # Find best mode
    best_mode = max(modes.items(), key=lambda x: x[1]["revenue"]/x[1]["cost"] if x[1]["cost"] > 0 else 0)
    report.append(f"> 💡 **Insight:** Mode **{best_mode[0]}** memiliki ROI terbaik ")
    report.append(f"({best_mode[1]['revenue']/best_mode[1]['cost']:.2f}x). Pertimbangkan untuk mengalokasikan lebih banyak budget ke mode ini.\n")


def add_budget_optimization_section(report: List[str], products: List[Dict]):
    """Add budget optimization recommendations."""
    report.append("## 💰 REKOMENDASI OPTIMASI BUDGET\n")
    report.append("Alokasi budget optimal berdasarkan historical ROI dan composite score.\n")
    
    # Categorize products
    scale_up = [p for p in products if "LANJUTKAN" in p.get("category", "") and p.get("roi", 0) > 3]
    maintain = [p for p in products if "LANJUTKAN" in p.get("category", "") and p.get("roi", 0) <= 3]
    reduce = [p for p in products if "PANTAU" in p.get("category", "")]
    stop = [p for p in products if "HENTIKAN" in p.get("category", "")]
    
    # Calculate metrics
    total_budget = sum(p.get("total_cost", 0) for p in products)
    budget_freed = sum(p.get("total_cost", 0) for p in stop)
    
    report.append("### Ringkasan Tindakan\n")
    report.append("| Tindakan | Jumlah Produk | Rekomendasi Budget |")
    report.append("|----------|---------------|--------------------|")
    report.append(f"| 🚀 SCALE UP | {len(scale_up)} | +30-50% budget |")
    report.append(f"| ➡️ PERTAHANKAN | {len(maintain)} | Jaga budget saat ini |")
    report.append(f"| 📉 KURANGI | {len(reduce)} | -20-30% budget |")
    report.append(f"| 🛑 HENTIKAN | {len(stop)} | Stop 100% |")
    report.append("")
    
    if budget_freed > 0:
        budget_str = f"Rp {budget_freed:,.0f}".replace(",", ".")
        report.append(f"### Budget yang Dapat Dibebaskan: {budget_str}\n")
        report.append("Alokasikan budget ini ke produk dengan rekomendasi SCALE UP:\n")
    
    if scale_up:
        report.append("### 📈 Detail Scale-Up\n")
        report.append("| Produk | ROI Saat Ini | Score | Tambah Budget | Est. Impact |")
        report.append("|--------|--------------|-------|---------------|-------------|")
        for p in sorted(scale_up, key=lambda x: x.get("score", 0), reverse=True)[:8]:
            current_cost = p.get("total_cost", 0)
            additional = current_cost * 0.3  # 30% increase
            expected_roi = p.get("roi", 0) * 0.85  # Conservative estimate
            impact = additional * expected_roi
            add_str = f"+Rp {additional:,.0f}".replace(",", ".")
            impact_str = f"+Rp {impact:,.0f}".replace(",", ".")
            report.append(f"| {p.get('product_name', 'N/A')[:30]} | {p.get('roi', 0):.1f}x | {p.get('score', 0):.0f} | {add_str} | {impact_str} |")
        report.append("")


def add_trend_analysis_section(report: List[str], products: List[Dict]):
    """Add trend analysis section."""
    report.append("## 📊 ANALISIS TREND\n")
    
    # Group by trend
    trends = {"increasing": [], "stable": [], "decreasing": []}
    for p in products:
        trend = p.get("trend", "stable")
        if trend in trends:
            trends[trend].append(p)
    
    total = len(products)
    
    report.append("### Distribusi Trend Produk\n")
    report.append("| Trend | Jumlah | Persentase | Interpretasi |")
    report.append("|-------|--------|------------|--------------|")
    report.append(f"| 📈 Meningkat | {len(trends['increasing'])} | {len(trends['increasing'])/total*100:.1f}% | Performa membaik |")
    report.append(f"| ➡️ Stabil | {len(trends['stable'])} | {len(trends['stable'])/total*100:.1f}% | Performa konsisten |")
    report.append(f"| 📉 Menurun | {len(trends['decreasing'])} | {len(trends['decreasing'])/total*100:.1f}% | Perlu perhatian |")
    report.append("")
    
    # Alert for declining products with high cost
    declining_high_cost = sorted(
        [p for p in trends["decreasing"] if p.get("total_cost", 0) > 1000000],
        key=lambda x: x.get("total_cost", 0), reverse=True
    )
    
    if declining_high_cost:
        report.append("### ⚠️ Produk Menurun dengan Budget Tinggi\n")
        report.append("*Produk ini memerlukan evaluasi segera*\n")
        report.append("| Produk | Budget | ROI | Aksi Disarankan |")
        report.append("|--------|--------|-----|-----------------|")
        for p in declining_high_cost[:5]:
            cost_str = f"Rp {p.get('total_cost', 0):,.0f}".replace(",", ".")
            report.append(f"| {p.get('product_name', 'N/A')[:30]} | {cost_str} | {p.get('roi', 0):.1f}x | Kurangi budget 30% |")
        report.append("")


def add_data_quality_section(report: List[str], products: List[Dict]):
    """Add data quality assessment section."""
    report.append("## 🔍 KUALITAS DATA\n")
    
    # Group by data quality
    quality = {"HIGH": [], "MEDIUM": [], "LOW": []}
    for p in products:
        q = p.get("data_quality", "MEDIUM")
        if q in quality:
            quality[q].append(p)
    
    total = len(products)
    
    report.append("### Distribusi Kualitas Data\n")
    report.append("| Kualitas | Jumlah | Persentase | Confidence |")
    report.append("|----------|--------|------------|------------|")
    report.append(f"| ✅ Tinggi | {len(quality['HIGH'])} | {len(quality['HIGH'])/total*100:.1f}% | Rekomendasi sangat akurat |")
    report.append(f"| 🟡 Sedang | {len(quality['MEDIUM'])} | {len(quality['MEDIUM'])/total*100:.1f}% | Rekomendasi cukup akurat |")
    report.append(f"| ⚠️ Rendah | {len(quality['LOW'])} | {len(quality['LOW'])/total*100:.1f}% | Perlu data lebih lanjut |")
    report.append("")
    
    # Overall confidence
    high_pct = len(quality['HIGH']) / total * 100
    if high_pct >= 70:
        confidence = "TINGGI ✅"
        msg = "Mayoritas produk memiliki data yang cukup untuk analisis akurat."
    elif high_pct >= 40:
        confidence = "SEDANG 🟡"
        msg = "Data cukup untuk sebagian besar analisis."
    else:
        confidence = "RENDAH ⚠️"
        msg = "Perlu mengumpulkan lebih banyak data historis."
    
    report.append(f"> 📊 **Confidence Level:** {confidence}")
    report.append(f"> {msg}\n")


def add_churn_risk_section(report: List[str], products: List[Dict]):
    """Add churn risk / early warning section - similar to TikTok."""
    report.append("## ⚠️ SISTEM PERINGATAN DINI (Churn Risk)\n")
    report.append("Analisis risiko penurunan performa produk berdasarkan momentum, trend, dan volatilitas.\n")
    
    # Analyze churn risk
    high_risk = []
    medium_risk = []
    low_risk = []
    
    for p in products:
        # Calculate risk score from multiple factors
        risk_score = 0
        risk_factors = []
        
        # ROI factor
        roi = p.get("roi", 0)
        if roi < 1.0:
            risk_score += 30
            risk_factors.append({"factor": "ROI Negatif", "value": f"{roi:.2f}x", "points": 30, "severity": "HIGH"})
        elif roi < 2.0:
            risk_score += 15
            risk_factors.append({"factor": "ROI Rendah", "value": f"{roi:.2f}x", "points": 15, "severity": "MEDIUM"})
        
        # Trend factor
        trend = p.get("trend", "stable")
        if "decreasing" in trend.lower() or "turun" in trend.lower():
            risk_score += 25
            risk_factors.append({"factor": "Trend Menurun", "value": trend, "points": 25, "severity": "HIGH"})
        
        # Momentum factor  
        momentum = p.get("momentum_pct", 0)
        if momentum < -20:
            risk_score += 20
            risk_factors.append({"factor": "Momentum Negatif", "value": f"{momentum:.1f}%", "points": 20, "severity": "HIGH"})
        elif momentum < 0:
            risk_score += 10
            risk_factors.append({"factor": "Momentum Turun", "value": f"{momentum:.1f}%", "points": 10, "severity": "MEDIUM"})
        
        # Consistency factor
        cv = p.get("cv", 0)
        if cv > 80:
            risk_score += 15
            risk_factors.append({"factor": "Volatilitas Tinggi", "value": f"CV={cv:.0f}%", "points": 15, "severity": "MEDIUM"})
        
        p["risk_score"] = risk_score
        p["risk_factors"] = risk_factors
        
        if risk_score >= 50:
            high_risk.append(p)
        elif risk_score >= 25:
            medium_risk.append(p)
        else:
            low_risk.append(p)
    
    total = len(products)
    portfolio_risk = sum(p.get("risk_score", 0) for p in products) / len(products) if products else 0
    
    report.append("### Gambaran Umum Risiko Portfolio\n")
    report.append("| Metrik | Nilai |")
    report.append("|--------|-------|")
    report.append(f"| Skor Risiko Portfolio | **{portfolio_risk:.1f}/100** |")
    report.append(f"| Produk Risiko Tinggi | {len(high_risk)} ({len(high_risk)/total*100:.1f}%) |")
    report.append(f"| Produk Risiko Sedang | {len(medium_risk)} |")
    report.append(f"| Produk Risiko Rendah | {len(low_risk)} |")
    report.append("")
    
    if high_risk:
        report.append("### 🔴 Produk dengan Risiko Tinggi\n")
        report.append("| # | Produk | Skor Risiko | Tingkat | Prediksi |")
        report.append("|---|--------|-------------|---------|----------|")
        for i, p in enumerate(sorted(high_risk, key=lambda x: x.get("risk_score", 0), reverse=True)[:10], 1):
            name = p.get("product_name", "N/A")[:35]
            report.append(f"| {i} | {name} | {p['risk_score']}/100 | HIGH | Kemungkinan tinggi mengalami penurunan |")
        report.append("")
        
        # Detail risk factors for top 3
        report.append("### Rincian Faktor Risiko (3 Produk Teratas)\n")
        for p in sorted(high_risk, key=lambda x: x.get("risk_score", 0), reverse=True)[:3]:
            name = p.get("product_name", "N/A")[:40]
            report.append(f"**{name}** (Skor: {p['risk_score']})\n")
            for f in p.get("risk_factors", []):
                report.append(f"- [{f['severity']}] {f['factor']}: {f['value']} (+{f['points']} poin)")
            report.append("")
    else:
        report.append("✅ Tidak ada produk dengan risiko tinggi saat ini. Portfolio dalam kondisi baik.\n")


def add_correlation_section(report: List[str], products: List[Dict]):
    """Add correlation analysis section - similar to TikTok."""
    report.append("## 📈 ANALISIS KORELASI & EFISIENSI PENGELUARAN\n")
    
    if len(products) < 5:
        report.append("*Data tidak mencukupi untuk melakukan analisis korelasi (min 5 produk).*\n")
        return
    
    # Simple correlation analysis
    costs = [p.get("total_cost", 0) for p in products]
    revenues = [p.get("total_revenue", 0) for p in products]
    rois = [p.get("roi", 0) for p in products]
    
    # Calculate correlation
    import numpy as np
    if len(costs) > 2:
        cost_rev_corr = np.corrcoef(costs, revenues)[0, 1]
        cost_roi_corr = np.corrcoef(costs, rois)[0, 1]
    else:
        cost_rev_corr = 0
        cost_roi_corr = 0
    
    # Interpret correlation
    def interpret_corr(r):
        if r > 0.7:
            return "Sangat Kuat Positif"
        elif r > 0.4:
            return "Positif Moderat"
        elif r > 0:
            return "Positif Lemah"
        elif r > -0.4:
            return "Negatif Lemah"
        elif r > -0.7:
            return "Negatif Moderat"
        else:
            return "Sangat Kuat Negatif"
    
    report.append("### Korelasi antara Biaya dan Pendapatan\n")
    report.append("| Metrik | Nilai | Interpretasi |")
    report.append("|--------|-------|--------------|")
    report.append(f"| Pearson r | {cost_rev_corr:.3f} | {interpret_corr(cost_rev_corr)} |")
    report.append(f"| Signifikansi | {'✅ Signifikan' if abs(cost_rev_corr) > 0.3 else '❌ Tidak signifikan'} | - |")
    report.append("")
    
    # Diminishing returns analysis
    report.append("### Analisis Diminishing Returns\n")
    report.append("| Metrik | Nilai | Interpretasi |")
    report.append("|--------|-------|--------------|")
    report.append(f"| Korelasi Biaya-ROI | {cost_roi_corr:.3f} | {interpret_corr(cost_roi_corr)} |")
    
    diminishing = cost_roi_corr < -0.2
    report.append(f"| Diminishing Returns? | {'⚠️ Ya - ROI menurun saat pengeluaran naik' if diminishing else '✅ Tidak - Pengeluaran masih efektif'} | {'Perlu batasi budget' if diminishing else 'Aman untuk scale up'} |")
    report.append("")
    
    # Efficient products
    avg_cost = np.mean(costs)
    avg_roi = np.mean(rois)
    median_cost = np.median(costs)
    
    efficient = [p for p in products if p.get("total_cost", 0) < median_cost and p.get("roi", 0) > avg_roi]
    inefficient = [p for p in products if p.get("total_cost", 0) > avg_cost and p.get("roi", 0) < avg_roi]
    
    if efficient:
        report.append("### 🎯 Produk dengan Efisiensi Tinggi (Kandidat Scale Up)\n")
        report.append("*Pengeluaran rendah dengan ROI tinggi - potensi yang belum dimanfaatkan*\n")
        report.append("| Produk | Pengeluaran | ROI | Peluang |")
        report.append("|--------|-------------|-----|---------|")
        for p in sorted(efficient, key=lambda x: x.get("roi", 0), reverse=True)[:5]:
            cost_str = f"Rp {p.get('total_cost', 0):,.0f}".replace(",", ".")
            report.append(f"| {p.get('product_name', 'N/A')[:30]} | {cost_str} | {p.get('roi', 0):.1f}x | Scale up prioritas |")
        report.append("")
    
    if inefficient:
        report.append("### ⚠️ Produk dengan Efisiensi Rendah (Perlu Evaluasi)\n")
        report.append("*Pengeluaran tinggi dengan ROI di bawah rata-rata*\n")
        report.append("| Produk | Pengeluaran | ROI | Masalah |")
        report.append("|--------|-------------|-----|---------|")
        for p in sorted(inefficient, key=lambda x: x.get("total_cost", 0), reverse=True)[:5]:
            cost_str = f"Rp {p.get('total_cost', 0):,.0f}".replace(",", ".")
            report.append(f"| {p.get('product_name', 'N/A')[:30]} | {cost_str} | {p.get('roi', 0):.1f}x | Review budget |")
        report.append("")

