#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Report Sections Module - Part 1
===============================
Section builders untuk HTML report.
Executive Summary, Health, Financial sections.
"""

from typing import List
import numpy as np

from ..comprehensive_engine import ComprehensiveAnalysis
from ..unified_scorer import ActionRecommendation
from ..portfolio_metrics import PortfolioHealth, FinancialProjection


def buildExecutiveSummary(analyses: List[ComprehensiveAnalysis],
                          health: PortfolioHealth) -> List[str]:
    """Build executive summary section."""
    totalCost = sum(a.totalCost for a in analyses)
    totalRevenue = sum(a.totalRevenue for a in analyses)
    totalProfit = totalRevenue - totalCost
    overallRoas = totalRevenue / totalCost if totalCost > 0 else 0
    avgProb = np.mean([a.successProbability for a in analyses])
    
    scaleUp = len([a for a in analyses if a.finalAction in [
        ActionRecommendation.SCALE_UP_AGGRESSIVE, ActionRecommendation.SCALE_UP
    ]])
    maintain = len([a for a in analyses if a.finalAction == ActionRecommendation.MAINTAIN])
    reduce = len([a for a in analyses if a.finalAction == ActionRecommendation.REDUCE])
    stop = len([a for a in analyses if a.finalAction == ActionRecommendation.STOP])
    
    return [
        "<h2>🎯 RINGKASAN EKSEKUTIF</h2>",
        "<h3>📊 Skor Kesehatan Portfolio</h3>",
        "<table>",
        "<thead><tr><th>Metrik</th><th>Nilai</th><th>Kategori</th></tr></thead><tbody>",
        f"<tr><td><strong>Health Score</strong></td><td><strong>{health.healthScore}/100</strong></td><td><strong>Grade {health.grade}</strong></td></tr>",
        f"<tr><td>Tingkat Risiko</td><td>{health.riskLevel}</td><td>Skor: {health.riskScore}/100</td></tr>",
        f"<tr><td>Avg Success Probability</td><td>{avgProb:.1f}%</td><td>-</td></tr>",
        "</tbody></table>",
        
        "<h3>💰 Metrik Utama Performa</h3>",
        "<table>",
        "<thead><tr><th>Metrik</th><th>Nilai</th></tr></thead><tbody>",
        f"<tr><td>Total Investasi Iklan</td><td>Rp {totalCost:,.0f}</td></tr>",
        f"<tr><td>Total Pendapatan</td><td>Rp {totalRevenue:,.0f}</td></tr>",
        f"<tr><td><strong>Total Keuntungan Bersih</strong></td><td><strong>Rp {totalProfit:,.0f}</strong></td></tr>",
        f"<tr><td><strong>ROAS Keseluruhan</strong></td><td><strong>{overallRoas:.2f}x</strong></td></tr>",
        f"<tr><td>Jumlah Produk Dianalisis</td><td>{len(analyses)}</td></tr>",
        "</tbody></table>",
        
        "<h3>📈 Distribusi Rekomendasi Produk</h3>",
        "<table>",
        "<thead><tr><th>Rekomendasi</th><th>Jumlah Produk</th><th>Persentase</th></tr></thead><tbody>",
        f"<tr><td>🚀 <strong>SCALE UP</strong></td><td>{scaleUp}</td><td>{scaleUp/len(analyses)*100:.1f}%</td></tr>",
        f"<tr><td>✅ <strong>MAINTAIN</strong></td><td>{maintain}</td><td>{maintain/len(analyses)*100:.1f}%</td></tr>",
        f"<tr><td>⬇️ <strong>REDUCE</strong></td><td>{reduce}</td><td>{reduce/len(analyses)*100:.1f}%</td></tr>",
        f"<tr><td>🛑 <strong>STOP</strong></td><td>{stop}</td><td>{stop/len(analyses)*100:.1f}%</td></tr>",
        "</tbody></table>",
    ]


def buildHealthScoreSection(health: PortfolioHealth,
                            analyses: List[ComprehensiveAnalysis]) -> List[str]:
    """Build health score breakdown section."""
    return [
        "<h2>🤖 EVALUASI BERBASIS AI & MACHINE LEARNING</h2>",
        "<h3>Komponen Health Score</h3>",
        "<p>Tabel berikut menunjukkan rincian perhitungan Health Score portfolio iklan:</p>",
        "<table>",
        "<thead><tr><th>Komponen</th><th>Skor</th><th>Bobot</th><th>Nilai Aktual</th><th>Rumus Perhitungan</th></tr></thead><tbody>",
        f"<tr><td>Tingkat Profitabilitas</td><td>{health.profitabilityScore:.1f}</td><td>40%</td><td>{health.profitabilityScore:.1f}%</td><td>(Produk Untung / Total) × 100</td></tr>",
        f"<tr><td>Performa ROI</td><td>{health.roiScore:.1f}</td><td>30%</td><td>-</td><td>50 + log₁₀(ROI) × 30</td></tr>",
        f"<tr><td>Kesehatan Trend</td><td>{health.trendScore:.1f}</td><td>15%</td><td>-</td><td>(Naik + Stabil×0.7) / Total</td></tr>",
        f"<tr><td>Kualitas Data</td><td>{health.dataQualityScore:.1f}</td><td>15%</td><td>-</td><td>(Tinggi + Sedang×0.6) / Total</td></tr>",
        "</tbody></table>",
        
        "<h3>⚠️ Penilaian Risiko Portfolio</h3>",
        f"<p><strong>Tingkat Risiko:</strong> {health.riskLevel} (Skor: {health.riskScore}/100)</p>",
        "<p><strong>Faktor-Faktor Risiko yang Teridentifikasi:</strong></p>",
        "<ul>",
        *[f"<li>{factor}</li>" for factor in health.riskFactors],
        "</ul>",
    ]


def buildFinancialProjection(projection: FinancialProjection,
                             analyses: List[ComprehensiveAnalysis]) -> List[str]:
    """Build financial projection section."""
    roas_values = [a.roas for a in analyses if a.roas > 0]
    cv = (np.std(roas_values) / np.mean(roas_values) * 100) if roas_values else 0
    
    return [
        "<h2>💹 PROYEKSI DAMPAK KEUANGAN</h2>",
        "<h3>Proyeksi Keuntungan Jika Rekomendasi Diterapkan</h3>",
        "<table>",
        "<thead><tr><th>Metrik</th><th>Nilai</th></tr></thead><tbody>",
        f"<tr><td>Keuntungan Saat Ini</td><td>Rp {projection.currentProfit:,.0f}</td></tr>",
        f"<tr><td><strong>Proyeksi Keuntungan</strong></td><td><strong>Rp {projection.projectedProfit:,.0f}</strong></td></tr>",
        f"<tr><td><strong>Potensi Pertumbuhan</strong></td><td><strong>+{projection.growthPct:.1f}%</strong></td></tr>",
        f"<tr><td><strong>Confidence Interval 95%</strong></td><td><strong>Rp {projection.ciLower:,.0f} - Rp {projection.ciUpper:,.0f}</strong></td></tr>",
        "</tbody></table>",
        
        "<h3>📊 Analisis Tingkat Kepercayaan Statistik</h3>",
        "<table>",
        "<thead><tr><th>Metrik</th><th>Nilai</th><th>Penjelasan</th></tr></thead><tbody>",
        "<tr><td>Confidence Level</td><td>95%</td><td>Tingkat kepercayaan proyeksi</td></tr>",
        f"<tr><td>ROI Variability (CV)</td><td>{cv:.1f}%</td><td>Variasi ROI antar produk</td></tr>",
        "<tr><td>Methodology</td><td>Bootstrap + Bayesian</td><td>Kombinasi metode statistik</td></tr>",
        "</tbody></table>",
        
        "<h3>Rincian Sumber Peningkatan Keuntungan</h3>",
        "<table>",
        "<thead><tr><th>Sumber Keuntungan</th><th>Nilai</th><th>Keterangan</th></tr></thead><tbody>",
        f"<tr><td>Penghematan dari Penghentian</td><td>+Rp {projection.savingsFromStop:,.0f}</td><td>Menghentikan produk yang merugi</td></tr>",
        f"<tr><td>Keuntungan dari Scale-up</td><td>+Rp {projection.gainsFromScaleUp:,.0f}</td><td>Penambahan budget ke produk unggulan</td></tr>",
        f"<tr><td>Realokasi Budget</td><td>+Rp {projection.reallocationGains:,.0f}</td><td>Budget dialihkan ke produk dengan ROI tinggi</td></tr>",
        "</tbody></table>",
    ]
