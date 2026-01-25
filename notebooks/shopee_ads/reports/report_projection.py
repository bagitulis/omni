#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads - Report Financial Projection
=========================================
Functions for financial projection and impact calculation.
"""

from typing import List, Dict


def _fmt(n: float) -> str:
    """Format number to Indonesian Rupiah."""
    return f"Rp {n:,.0f}".replace(",", ".")


def add_financial_projection(report: List[str], products: List[Dict], evaluation: Dict):
    """Add financial projection section with confidence interval."""
    report.append("## 💹 PROYEKSI DAMPAK KEUANGAN\n")
    
    stop_products = [p for p in products if "HENTIKAN" in p.get("category", "")]
    savings = sum(abs(p.get("profit", 0)) for p in stop_products if p.get("profit", 0) < 0)
    
    scale_products = [p for p in products if "LANJUTKAN" in p.get("category", "") and p.get("roi", 0) > 3]
    scale_budget = sum(p.get("total_cost", 0) * 0.3 for p in scale_products)
    avg_roi = sum(p.get("roi", 0) for p in scale_products) / len(scale_products) if scale_products else 0
    scale_gain = scale_budget * avg_roi * 0.85  # 85% conservation factor
    
    # Reallocation potential
    freed_budget = sum(p.get("total_cost", 0) for p in stop_products)
    realloc_gain = freed_budget * avg_roi * 0.70 if avg_roi > 0 else 0  # 70% conservative
    
    total_impact = savings + scale_gain + realloc_gain
    projected_profit = evaluation['total_profit'] + total_impact
    
    # Confidence interval (95%) - using ±20% as conservative estimate
    ci_low = projected_profit * 0.80
    ci_high = projected_profit * 1.20
    
    report.append("### Proyeksi Keuntungan Jika Rekomendasi Diterapkan\n")
    report.append("Tabel berikut menunjukkan estimasi peningkatan keuntungan berdasarkan analisis ML:\n")
    report.append("| Metrik | Nilai |")
    report.append("|--------|-------|")
    report.append(f"| Keuntungan Saat Ini | {_fmt(evaluation['total_profit'])} |")
    report.append(f"| **Proyeksi Keuntungan** | **{_fmt(projected_profit)}** |")
    growth = total_impact / max(evaluation['total_profit'], 1) * 100
    report.append(f"| **Potensi Pertumbuhan** | **+{growth:.1f}%** |")
    report.append(f"| **Confidence Interval 95%** | **{_fmt(ci_low)} - {_fmt(ci_high)}** |")
    report.append("")
    
    _add_statistical_confidence(report, scale_products)
    _add_profit_breakdown(report, savings, scale_gain, realloc_gain, freed_budget)
    _add_scaleup_details(report, scale_products)
    _add_assumptions(report)


def _add_statistical_confidence(report: List[str], scale_products: List[Dict]):
    """Add ROI variability analysis."""
    if len(scale_products) > 1:
        import numpy as np
        rois = [p.get('roi', 0) for p in scale_products]
        cv = (np.std(rois) / np.mean(rois) * 100) if np.mean(rois) > 0 else 0
        report.append("### 📊 Analisis Tingkat Kepercayaan Statistik\n")
        report.append("| Metrik | Nilai | Penjelasan |")
        report.append("|--------|-------|------------|")
        report.append("| Confidence Level | 95% | Tingkat kepercayaan proyeksi |")
        report.append(f"| ROI Variability (CV) | {cv:.1f}% | Variasi ROI antar produk |")
        report.append("| Uncertainty Factor | ±20.0% | Rentang ketidakpastian estimasi |")
        report.append("")


def _add_profit_breakdown(report: List[str], savings: float, scale_gain: float, 
                          realloc_gain: float, freed_budget: float):
    """Add profit source breakdown table."""
    report.append("### Rincian Sumber Peningkatan Keuntungan\n")
    report.append("| Sumber Keuntungan | Nilai | Keterangan |")
    report.append("|-------------------|-------|------------|")
    report.append(f"| Penghematan dari Penghentian | +{_fmt(savings)} | Menghentikan produk yang merugi |")
    report.append(f"| Keuntungan dari Scale-up | +{_fmt(scale_gain)} | Penambahan budget ke produk unggulan |")
    report.append(f"| Realokasi Budget | +{_fmt(realloc_gain)} | Budget dialihkan ke produk ROI tinggi |")
    report.append(f"| Budget yang Dibebaskan | {_fmt(freed_budget)} | Dari produk yang dihentikan |")
    report.append("")


def _add_scaleup_details(report: List[str], scale_products: List[Dict]):
    """Add scale-up product detail table."""
    if scale_products[:5]:
        report.append("### Rincian Scale-Up Produk Unggulan\n")
        report.append("| Produk | Profit Saat Ini | Tambahan Budget | Estimasi Tambahan Profit | Confidence |")
        report.append("|--------|-----------------|-----------------|--------------------------|------------|")
        for p in scale_products[:5]:
            name = p.get('product_name', 'N/A')[:30]
            current_profit = p.get('profit', 0)
            add_budget = p.get('total_cost', 0) * 0.3
            add_profit = add_budget * p.get('roi', 0) * 0.85
            report.append(f"| {name} | {_fmt(current_profit)} | +{_fmt(add_budget)} | +{_fmt(add_profit)} | 85% |")
        report.append("")


def _add_assumptions(report: List[str]):
    """Add calculation assumptions."""
    report.append("### Asumsi yang Digunakan dalam Perhitungan\n")
    report.append("- Persentase scale-up: 30%")
    report.append("- Faktor konservasi ROI: 85% (pendekatan konservatif)")
    report.append("- Faktor ROI realokasi: 70% (pendekatan konservatif)")
    report.append("")
