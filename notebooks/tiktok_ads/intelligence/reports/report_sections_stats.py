#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Report Sections Module - Part 3
===============================
Section builders untuk HTML report.
Statistical Analysis, Correlation, Budget sections.
"""

from typing import List
from datetime import datetime
import numpy as np
from scipy import stats

from ..comprehensive_engine import ComprehensiveAnalysis
from ..unified_scorer import ActionRecommendation


def buildCreativeComparison(analyses: List[ComprehensiveAnalysis]) -> List[str]:
    """Build creative type comparison with statistical tests."""
    video = [a for a in analyses if a.creativeType and "Video" in a.creativeType]
    card = [a for a in analyses if a.creativeType and ("Kartu" in a.creativeType or "Card" in a.creativeType)]
    
    if len(video) < 2 or len(card) < 2:
        return ["<h2>📱 PERBANDINGAN CREATIVE TYPE</h2>", "<p>Data tidak cukup untuk perbandingan statistik.</p>"]
    
    videoRoas = [a.roas for a in video]
    cardRoas = [a.roas for a in card]
    
    # Statistical tests
    tStat, tPval = stats.ttest_ind(videoRoas, cardRoas)
    uStat, uPval = stats.mannwhitneyu(videoRoas, cardRoas, alternative='two-sided')
    
    # Cohen's d
    pooledStd = np.sqrt(((len(videoRoas)-1)*np.var(videoRoas) + (len(cardRoas)-1)*np.var(cardRoas)) / (len(videoRoas)+len(cardRoas)-2))
    cohensD = (np.mean(cardRoas) - np.mean(videoRoas)) / pooledStd if pooledStd > 0 else 0
    
    videoTotalCost = sum(a.totalCost for a in video)
    videoTotalRev = sum(a.totalRevenue for a in video)
    cardTotalCost = sum(a.totalCost for a in card)
    cardTotalRev = sum(a.totalRevenue for a in card)
    
    return [
        "<h2>📱 PERBANDINGAN PERFORMA CREATIVE TYPE</h2>",
        
        "<h3>🎬 Performa Video</h3>",
        "<table>",
        "<thead><tr><th>Metrik</th><th>Nilai</th></tr></thead><tbody>",
        f"<tr><td>Jumlah Produk</td><td>{len(video)}</td></tr>",
        f"<tr><td>Total Biaya Iklan</td><td>Rp {videoTotalCost:,.0f}</td></tr>",
        f"<tr><td>Total Pendapatan</td><td>Rp {videoTotalRev:,.0f}</td></tr>",
        f"<tr><td><strong>ROAS Keseluruhan</strong></td><td><strong>{videoTotalRev/videoTotalCost:.2f}x</strong></td></tr>",
        "</tbody></table>",
        
        "<h3>🃏 Performa Kartu</h3>",
        "<table>",
        "<thead><tr><th>Metrik</th><th>Nilai</th></tr></thead><tbody>",
        f"<tr><td>Jumlah Produk</td><td>{len(card)}</td></tr>",
        f"<tr><td>Total Biaya Iklan</td><td>Rp {cardTotalCost:,.0f}</td></tr>",
        f"<tr><td>Total Pendapatan</td><td>Rp {cardTotalRev:,.0f}</td></tr>",
        f"<tr><td><strong>ROAS Keseluruhan</strong></td><td><strong>{cardTotalRev/cardTotalCost:.2f}x</strong></td></tr>",
        "</tbody></table>",
        
        "<h3>📊 Validasi Statistik</h3>",
        "<table>",
        "<thead><tr><th>Metrik Statistik</th><th>Video</th><th>Kartu</th></tr></thead><tbody>",
        f"<tr><td>Ukuran Sampel (n)</td><td>{len(video)}</td><td>{len(card)}</td></tr>",
        f"<tr><td>Rata-rata ROAS</td><td>{np.mean(videoRoas):.2f}x</td><td>{np.mean(cardRoas):.2f}x</td></tr>",
        f"<tr><td>Standar Deviasi</td><td>{np.std(videoRoas):.2f}</td><td>{np.std(cardRoas):.2f}</td></tr>",
        f"<tr><td>Median ROAS</td><td>{np.median(videoRoas):.2f}x</td><td>{np.median(cardRoas):.2f}x</td></tr>",
        "</tbody></table>",
        
        "<table>",
        "<thead><tr><th>Uji Statistik</th><th>Nilai Statistik</th><th>p-value</th><th>Signifikan?</th></tr></thead><tbody>",
        f"<tr><td>T-Test</td><td>{tStat:.3f}</td><td>{tPval:.4f}</td><td>{'✅ Ya' if tPval < 0.05 else '❌ Tidak'}</td></tr>",
        f"<tr><td>Mann-Whitney U</td><td>{uStat:.1f}</td><td>{uPval:.4f}</td><td>{'✅ Ya' if uPval < 0.05 else '❌ Tidak'}</td></tr>",
        "</tbody></table>",
        
        f"<p><strong>Cohen's d:</strong> {cohensD:.3f} ({'Large' if abs(cohensD) > 0.8 else 'Medium' if abs(cohensD) > 0.5 else 'Small'})</p>",
    ]


def buildCorrelationSection(analyses: List[ComprehensiveAnalysis]) -> List[str]:
    """Build correlation and efficiency analysis."""
    costs = [a.totalCost for a in analyses]
    revenues = [a.totalRevenue for a in analyses]
    roas_vals = [a.roas for a in analyses]
    
    pearsonR, pearsonP = stats.pearsonr(costs, revenues)
    spearmanR, spearmanP = stats.spearmanr(costs, revenues)
    costRoiCorr, _ = stats.pearsonr(costs, roas_vals)
    
    # Efficiency grouping
    medianCost = np.median(costs)
    avgRoas = np.mean(roas_vals)
    highEfficiency = [a for a in analyses if a.totalCost < medianCost and a.roas > avgRoas]
    lowEfficiency = [a for a in analyses if a.totalCost > medianCost and a.roas < avgRoas]
    
    lines = [
        "<h2>📈 ANALISIS KORELASI & EFISIENSI PENGELUARAN</h2>",
        
        "<h3>Korelasi antara Biaya dan Pendapatan</h3>",
        "<table>",
        "<thead><tr><th>Metrik</th><th>Nilai</th><th>Interpretasi</th></tr></thead><tbody>",
        f"<tr><td>Pearson r</td><td>{pearsonR:.3f}</td><td>{'Strong' if abs(pearsonR) > 0.7 else 'Moderate' if abs(pearsonR) > 0.4 else 'Weak'}</td></tr>",
        f"<tr><td>p-value</td><td>{pearsonP:.4f}</td><td>{'✅ Signifikan' if pearsonP < 0.05 else '❌ Tidak signifikan'}</td></tr>",
        f"<tr><td>Spearman r</td><td>{spearmanR:.3f}</td><td>Validasi non-parametrik</td></tr>",
        "</tbody></table>",
        
        "<h3>Analisis Diminishing Returns</h3>",
        "<table>",
        "<thead><tr><th>Metrik</th><th>Nilai</th><th>Interpretasi</th></tr></thead><tbody>",
        f"<tr><td>Korelasi Biaya-ROI</td><td>{costRoiCorr:.3f}</td><td>{'Diminishing returns terdeteksi' if costRoiCorr < -0.3 else 'Tidak ada diminishing returns signifikan'}</td></tr>",
        "</tbody></table>",
    ]
    
    if highEfficiency:
        lines.extend(_buildHighEfficiencyTable(highEfficiency))
    
    if lowEfficiency:
        lines.extend(_buildLowEfficiencyTable(lowEfficiency))
    
    return lines


def _buildHighEfficiencyTable(highEfficiency: List[ComprehensiveAnalysis]) -> List[str]:
    """Build high efficiency products table."""
    lines = [
        "<h3>🎯 Produk dengan Efisiensi Tinggi (Kandidat Scale Up)</h3>",
        "<p><em>Pengeluaran rendah dengan ROI tinggi</em></p>",
        "<table>",
        "<thead><tr><th>Produk</th><th>Pengeluaran</th><th>ROAS</th><th>Peluang</th></tr></thead><tbody>",
    ]
    
    for a in sorted(highEfficiency, key=lambda x: x.roas, reverse=True)[:5]:
        lines.append(
            f"<tr><td>{a.productName[:35]}</td><td>Rp {a.totalCost:,.0f}</td>"
            f"<td>{a.roas:.2f}x</td><td>Scale up prioritas</td></tr>"
        )
    
    lines.append("</tbody></table>")
    return lines


def _buildLowEfficiencyTable(lowEfficiency: List[ComprehensiveAnalysis]) -> List[str]:
    """Build low efficiency products table."""
    lines = [
        "<h3>⚠️ Produk dengan Efisiensi Rendah (Perlu Evaluasi)</h3>",
        "<p><em>Pengeluaran tinggi dengan ROI di bawah rata-rata</em></p>",
        "<table>",
        "<thead><tr><th>Produk</th><th>Pengeluaran</th><th>ROAS</th><th>Masalah</th></tr></thead><tbody>",
    ]
    
    for a in sorted(lowEfficiency, key=lambda x: x.totalCost, reverse=True)[:5]:
        lines.append(
            f"<tr><td>{a.productName[:35]}</td><td>Rp {a.totalCost:,.0f}</td>"
            f"<td>{a.roas:.2f}x</td><td>Review budget</td></tr>"
        )
    
    lines.append("</tbody></table>")
    return lines


def buildBudgetOptimization(analyses: List[ComprehensiveAnalysis]) -> List[str]:
    """Build budget optimization recommendations."""
    totalCurrentBudget = sum(a.budgetRec.currentBudget for a in analyses)
    totalRecommended = sum(a.budgetRec.recommendedBudget for a in analyses)
    totalImpact = sum(a.budgetRec.revenueChange for a in analyses)
    
    stopCount = len([a for a in analyses if a.finalAction == ActionRecommendation.STOP])
    reduceCount = len([a for a in analyses if a.finalAction == ActionRecommendation.REDUCE])
    maintainCount = len([a for a in analyses if a.finalAction == ActionRecommendation.MAINTAIN])
    increaseCount = len([a for a in analyses if a.finalAction in [
        ActionRecommendation.SCALE_UP, ActionRecommendation.SCALE_UP_AGGRESSIVE
    ]])
    
    freedBudget = sum(a.budgetRec.currentBudget for a in analyses if a.finalAction == ActionRecommendation.STOP)
    
    lines = [
        "<h2>💰 REKOMENDASI OPTIMASI BUDGET</h2>",
        "<p>Alokasi budget optimal berdasarkan historical ROI, momentum, dan composite score.</p>",
        
        "<h3>Ringkasan Rekomendasi Budget</h3>",
        "<table>",
        "<thead><tr><th>Metrik</th><th>Nilai</th></tr></thead><tbody>",
        f"<tr><td>Total Budget Saat Ini</td><td>Rp {totalCurrentBudget:,.0f}</td></tr>",
        f"<tr><td>Total Budget yang Direkomendasikan</td><td>Rp {totalRecommended:,.0f}</td></tr>",
        f"<tr><td>Budget yang Dapat Dibebaskan (dari STOP)</td><td>Rp {freedBudget:,.0f}</td></tr>",
        f"<tr><td>Estimasi Dampak Pendapatan</td><td><strong>+Rp {totalImpact:,.0f}</strong></td></tr>",
        "</tbody></table>",
        
        "<h3>Ringkasan Tindakan</h3>",
        "<table>",
        "<thead><tr><th>Tindakan</th><th>Jumlah Produk</th></tr></thead><tbody>",
        f"<tr><td>🛑 HENTIKAN</td><td>{stopCount}</td></tr>",
        f"<tr><td>📉 KURANGI</td><td>{reduceCount}</td></tr>",
        f"<tr><td>➡️ PERTAHANKAN</td><td>{maintainCount}</td></tr>",
        f"<tr><td>📈 TINGKATKAN</td><td>{increaseCount}</td></tr>",
        "</tbody></table>",
    ]
    
    scaleUp = [a for a in analyses if a.finalAction in [
        ActionRecommendation.SCALE_UP, ActionRecommendation.SCALE_UP_AGGRESSIVE
    ]]
    
    if scaleUp:
        lines.extend(_buildScaleUpDetailTable(scaleUp))
    
    return lines


def _buildScaleUpDetailTable(scaleUp: List[ComprehensiveAnalysis]) -> List[str]:
    """Build scale up detail table."""
    scaleUp = sorted(scaleUp, key=lambda x: x.budgetRec.revenueChange, reverse=True)
    
    lines = [
        "<h3>📈 Produk untuk Scale-Up (Detail)</h3>",
        "<table>",
        "<thead><tr><th>Produk</th><th>Saat Ini</th><th>Direkomendasikan</th><th>Perubahan</th><th>Est. Dampak</th></tr></thead><tbody>",
    ]
    
    for a in scaleUp[:10]:
        lines.append(
            f"<tr><td>{a.productName[:30]}</td><td>Rp {a.budgetRec.currentBudget:,.0f}</td>"
            f"<td>Rp {a.budgetRec.recommendedBudget:,.0f}</td><td>+Rp {a.budgetRec.budgetChange:,.0f}</td>"
            f"<td>+Rp {a.budgetRec.revenueChange:,.0f}</td></tr>"
        )
    
    lines.append("</tbody></table>")
    return lines


def buildMethodology() -> List[str]:
    """Build methodology section."""
    timestamp = datetime.now()
    return [
        "<h2>📚 METODOLOGI ANALISIS</h2>",
        "<p>Bagian ini menjelaskan metode statistik dan machine learning yang digunakan.</p>",
        
        "<h3>1. Unified Scoring System (v3.0)</h3>",
        "<p><strong>Tujuan:</strong> Menggabungkan 8 komponen analisis untuk skor komprehensif</p>",
        "<pre>",
        "Composite Score = ",
        "  ROAS Score × 0.20 +",
        "  Trend Score × 0.15 +",
        "  Momentum Score × 0.10 +",
        "  Elasticity Score × 0.15 +",
        "  Volatility Score × 0.10 +",
        "  Fatigue Score × 0.10 +",
        "  Churn Risk Score × 0.10 +",
        "  Event Score × 0.10",
        "</pre>",
        
        "<h3>2. Mann-Kendall Trend Test</h3>",
        "<p><strong>Tujuan:</strong> Mendeteksi trend monotonic dalam time series</p>",
        "<ul>",
        "<li>p-value < 0.05: Trend signifikan secara statistik</li>",
        "<li>τ (tau) > 0: Trend naik | τ < 0: Trend turun</li>",
        "</ul>",
        
        "<h3>3. MACD-Style Trend Analysis</h3>",
        "<p><strong>Tujuan:</strong> Mendeteksi momentum dan perubahan trend</p>",
        "<ul>",
        "<li>EMA Short (3 period) vs EMA Long (8 period)</li>",
        "<li>Signal = EMA_short - EMA_long</li>",
        "</ul>",
        
        "<h3>4. Momentum Analysis (70/30)</h3>",
        "<p><strong>Tujuan:</strong> Membandingkan performa terkini dengan historis</p>",
        "<ul>",
        "<li>> +20%: Sangat Positif 🚀</li>",
        "<li>+5% ~ +20%: Positif 📈</li>",
        "<li>-5% ~ +5%: Stabil ➡️</li>",
        "<li>< -5%: Negatif ⚠️</li>",
        "</ul>",
        
        "<h3>5. Bootstrap Confidence Interval</h3>",
        "<p><strong>Tujuan:</strong> Estimasi rentang proyeksi dengan 95% confidence</p>",
        "<ul>",
        "<li>1000 bootstrap samples</li>",
        "<li>Percentile method untuk CI</li>",
        "</ul>",
        
        "<h3>6. Saturation Model (Logistic Growth)</h3>",
        "<p><strong>Tujuan:</strong> Mendeteksi diminishing returns</p>",
        "<ul>",
        "<li>Marginal ROAS calculation</li>",
        "<li>Saturation point estimation</li>",
        "</ul>",
        
        "<h3>7. Fatigue Detection (CTR Decay)</h3>",
        "<p><strong>Tujuan:</strong> Mendeteksi creative fatigue</p>",
        "<ul>",
        "<li>CTR decay < 10%: Fresh</li>",
        "<li>CTR decay 10-25%: Aging</li>",
        "<li>CTR decay > 25%: Fatigued</li>",
        "</ul>",
        
        "<h3>8. Indonesian Calendar Events</h3>",
        "<p><strong>Tujuan:</strong> Memperhitungkan event lokal Indonesia</p>",
        "<ul>",
        "<li>Payday Prime (25-3): 1.20x multiplier</li>",
        "<li>Twin Dates (1.1, 2.2, ...): 1.35x multiplier</li>",
        "<li>National Holidays: 1.15x multiplier</li>",
        "</ul>",
        
        "<hr>",
        f"<p><em>Report generated by Unified Intelligence System v3.0.0</em></p>",
        f"<p><em>{timestamp.strftime('%Y-%m-%d %H:%M:%S')}</em></p>",
    ]
