#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Methodology Section
==========================================
Comprehensive methodology documentation.
"""


def add_methodology_section(report):
    """Add comprehensive methodology section"""
    report.append("## 📚 METODOLOGI ANALISIS\n")
    report.append("Bagian ini menjelaskan metode statistik dan machine learning yang digunakan dalam analisis.\n")
    
    report.append("### 1. Mann-Kendall Trend Test\n")
    report.append("**Tujuan:** Mendeteksi trend monotonic (konsisten naik/turun) dalam data time series\n")
    report.append("**Formula:** S = Σ sign(xⱼ - xᵢ) untuk semua pasangan i < j\n")
    report.append("**Cara Membaca Hasil:**")
    report.append("- p-value < 0.05: Terdapat trend yang signifikan secara statistik")
    report.append("- τ (tau) > 0: Trend cenderung naik")
    report.append("- τ (tau) < 0: Trend cenderung turun\n")
    
    report.append("### 2. Linear Regression\n")
    report.append("**Tujuan:** Mengukur kemiringan (slope) trend dan tingkat kecocokan model (R²)\n")
    report.append("**Formula:** y = mx + b, dimana R² = 1 - (SS_res / SS_tot)\n")
    report.append("**Cara Membaca Hasil:**")
    report.append("- R² > 0.7: Model sangat cocok dengan data")
    report.append("- p-value < 0.1: Kemiringan trend signifikan\n")
    
    report.append("### 3. Momentum Analysis\n")
    report.append("**Tujuan:** Membandingkan performa terkini dengan performa historis\n")
    report.append("**Formula:** Data dibagi menjadi 70% historis dan 30% terkini, lalu dihitung persentase perubahan\n")
    report.append("**Cara Membaca Hasil:**")
    report.append("- > +20%: Momentum Sangat Positif 🚀")
    report.append("- +5% hingga +20%: Momentum Positif 📈")
    report.append("- -5% hingga +5%: Momentum Stabil ➡️")
    report.append("- < -5%: Momentum Negatif ⚠️\n")
    
    report.append("### 4. Coefficient of Variation (CV)\n")
    report.append("**Tujuan:** Mengukur tingkat konsistensi performa produk\n")
    report.append("**Formula:** CV = (Standar Deviasi / Rata-rata) × 100%\n")
    report.append("**Cara Membaca Hasil:**")
    report.append("- CV < 20%: Performa Sangat Konsisten ✅")
    report.append("- CV 20-50%: Performa Cukup Konsisten 🔶")
    report.append("- CV > 50%: Performa Tidak Konsisten ⚠️\n")
    
    report.append("### 5. Composite Scoring\n")
    report.append("**Tujuan:** Menghitung skor keseluruhan produk berdasarkan berbagai faktor\n")
    report.append("**Formula:**")
    report.append("```")
    report.append("Skor = ROI_score × 0.30 +")
    report.append("       Profit_score × 0.25 +")
    report.append("       Momentum_score × 0.20 +")
    report.append("       Consistency_score × 0.15 +")
    report.append("       Trend_score × 0.10")
    report.append("```\n")
    
    report.append("### 6. Health Score Portfolio\n")
    report.append("**Tujuan:** Menilai kesehatan keseluruhan portfolio iklan\n")
    report.append("**Formula:**")
    report.append("```")
    report.append("Health = Tingkat_Profitabilitas × 0.40 +")
    report.append("         Performa_ROI × 0.30 +")
    report.append("         Kesehatan_Trend × 0.15 +")
    report.append("         Kualitas_Data × 0.15")
    report.append("```\n")
    
    report.append("### 7. T-Test & Mann-Whitney U Test\n")
    report.append("**Tujuan:** Memvalidasi secara statistik apakah perbedaan performa antar creative type signifikan\n")
    report.append("**Formula:** t = (x̄₁ - x̄₂) / √(s₁²/n₁ + s₂²/n₂)\n")
    report.append("**Cara Membaca Hasil:**")
    report.append("- p-value < 0.05: Perbedaan signifikan secara statistik")
    report.append("- Cohen's d > 0.8: Ukuran efek besar (perbedaan praktis yang bermakna)\n")
    
    report.append("### 8. Confidence Interval\n")
    report.append("**Tujuan:** Mengestimasi rentang proyeksi keuntungan dengan tingkat kepercayaan tertentu\n")
    report.append("**Formula:** CI = x̄ ± (z × σ/√n) atau menggunakan metode Bootstrap percentile\n")
    report.append("**Cara Membaca Hasil:**")
    report.append("- Confidence Interval 95%: Terdapat 95% kepercayaan bahwa nilai sebenarnya berada dalam rentang tersebut\n")
    
    report.append("### 9. Churn Risk Score\n")
    report.append("**Tujuan:** Mengidentifikasi produk yang berisiko mengalami penurunan performa\n")
    report.append("**Formula:**")
    report.append("```")
    report.append("Skor Risiko = Risiko_Momentum (0-35 poin) +")
    report.append("              Risiko_Trend (0-30 poin) +")
    report.append("              Risiko_Volatilitas (0-20 poin) +")
    report.append("              Risiko_Kualitas_Data (0-15 poin)")
    report.append("```\n")
    
    report.append("---\n")
    report.append("*Laporan dibuat oleh TikTok Ads ML Analysis System v2.2.0*\n")
    report.append("*Menggunakan: Mann-Kendall Test, Linear Regression, T-Test, Bootstrap CI, Churn Risk Model*\n")
