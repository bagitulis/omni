#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Report Sections Module - Part 2
===============================
Section builders untuk HTML report.
Recommendations dan Churn Risk sections.

v2.1: Enhanced with:
- Discontinued products filtering
- Video vs Product Card separation
- Specific budget percentage recommendations
- Data quality warning badges
- Lifecycle stage badges
- Saturation status indicators
- Confidence interval display
"""

from typing import List, Tuple
import numpy as np

from ..comprehensive_engine import ComprehensiveAnalysis
from ..unified_scorer import ActionRecommendation


# ========== BADGE HELPER FUNCTIONS ==========

def _getDataQualityBadge(nPeriods: int) -> str:
    """Get data quality badge based on number of periods."""
    if nPeriods < 4:
        return '<span class="badge badge-warning" title="Data kurang dari 4 periode - rekomendasi kurang akurat">⚠️ Data Terbatas</span>'
    elif nPeriods < 7:
        return '<span class="badge badge-info" title="Data 4-6 periode - cukup untuk analisis">📊 Data Cukup</span>'
    else:
        return '<span class="badge badge-success" title="Data 7+ periode - rekomendasi sangat akurat">✅ Data Lengkap</span>'


def _getLifecycleBadge(analysis: ComprehensiveAnalysis) -> str:
    """
    Get lifecycle stage badge - Fase Produk.
    
    - Bertumbuh: Momentum positif kuat, trend naik - potensi scale-up tinggi
    - Stabil Menguntungkan: Performa konsisten baik - pertahankan
    - Menurun: Momentum negatif atau trend turun - perlu evaluasi
    - Stagnan: Tidak ada pergerakan signifikan - monitor
    """
    momentum = analysis.legacyMomentumPct
    trend = analysis.legacyMkTrend
    roas = analysis.roas
    
    if momentum > 20 and "Naik" in trend:
        return '<span class="badge badge-growth" title="Momentum kuat +{:.0f}%, trend naik - sangat layak ditingkatkan budgetnya">🚀 Bertumbuh</span>'.format(momentum)
    elif momentum > 0 and roas >= 3:
        return '<span class="badge badge-mature" title="Performa stabil dengan ROAS {:.1f}x - pertahankan budget saat ini">📈 Stabil Untung</span>'.format(roas)
    elif momentum < -10 or "Turun" in trend:
        return '<span class="badge badge-decline" title="Momentum menurun {:.0f}% - pertimbangkan kurangi budget">📉 Menurun</span>'.format(momentum)
    else:
        return '<span class="badge badge-stable" title="Performa datar - pantau perkembangan">➡️ Stagnan</span>'


def _getSaturationBadge(analysis: ComprehensiveAnalysis) -> str:
    """
    Get saturation status badge - Respons terhadap Budget.
    
    - Responsif: Tambah budget = revenue naik proporsional
    - Cukup Responsif: Masih ada ruang untuk tambah budget
    - Hampir Jenuh: Tambah budget kurang efektif
    - Jenuh: Tambah budget tidak akan meningkatkan revenue
    """
    elasticity = analysis.unifiedScore.elasticityScore
    
    if elasticity >= 70:
        return '<span class="badge badge-success" title="Tambah budget akan meningkatkan revenue secara proporsional">🟢 Responsif</span>'
    elif elasticity >= 50:
        return '<span class="badge badge-info" title="Masih ada ruang untuk meningkatkan budget">🔵 Cukup Responsif</span>'
    elif elasticity >= 30:
        return '<span class="badge badge-warning" title="Tambah budget kurang efektif - mendekati titik jenuh">🟡 Hampir Jenuh</span>'
    else:
        return '<span class="badge badge-danger" title="Budget sudah maksimal - tambah budget tidak akan meningkatkan revenue">🔴 Jenuh</span>'


def _formatProfitWithCI(analysis: ComprehensiveAnalysis) -> str:
    """Format profit estimate with confidence interval."""
    profit = analysis.budgetRec.profitChange
    ci = analysis.budgetRec.profitCI
    
    if profit > 0:
        return f"+Rp {profit:,.0f}<br/><small>CI: [{ci[0]:,.0f}, {ci[1]:,.0f}]</small>"
    else:
        return f"Rp {profit:,.0f}"


def _filterActiveProducts(analyses: List[ComprehensiveAnalysis]) -> Tuple[List[ComprehensiveAnalysis], List[ComprehensiveAnalysis]]:
    """
    Filter discontinued products from active ones.
    Returns: (active_products, discontinued_products)
    """
    active = [a for a in analyses if not a.isDiscontinued]
    discontinued = [a for a in analyses if a.isDiscontinued]
    return active, discontinued


def _splitByCreativeType(analyses: List[ComprehensiveAnalysis]) -> Tuple[List[ComprehensiveAnalysis], List[ComprehensiveAnalysis]]:
    """
    Split products by creative type (Video vs Product Card).
    Returns: (video_products, card_products)
    """
    video = [a for a in analyses if a.creativeType and "Video" in a.creativeType]
    card = [a for a in analyses if a.creativeType and ("Kartu" in a.creativeType or "Card" in a.creativeType)]
    return video, card


def _formatProductName(name: str, maxCharsPerLine: int = 30) -> str:
    """
    Format product name to wrap nicely instead of truncating.
    Splits at ' - ' delimiter or wraps at word boundary.
    
    Example:
        "Nusseyba - Haafiz Kurta - Cotton Madina Premium"
        becomes:
        "Nusseyba - Haafiz Kurta<br/><small>Cotton Madina Premium</small>"
    """
    if len(name) <= maxCharsPerLine:
        return name
    
    # Try to split at ' - ' delimiter first
    parts = name.split(' - ')
    if len(parts) >= 2:
        # Take first two parts as main, rest as subtitle
        mainPart = f"{parts[0]} - {parts[1]}"
        if len(parts) > 2:
            subtitle = ' - '.join(parts[2:])
            return f"{mainPart}<br/><small>{subtitle}</small>"
        return mainPart
    
    # No delimiter, wrap at word boundary
    if len(name) > maxCharsPerLine:
        # Find last space before maxCharsPerLine
        breakPoint = name[:maxCharsPerLine].rfind(' ')
        if breakPoint > 10:  # Ensure reasonable first line
            firstLine = name[:breakPoint]
            secondLine = name[breakPoint+1:]
            return f"{firstLine}<br/><small>{secondLine}</small>"
    
    return name


def _getBudgetChangeLabel(budgetChangePct: float) -> str:
    """Get budget change label with percentage."""
    if budgetChangePct > 25:
        return f"📈 NAIKKAN +{budgetChangePct:.0f}%"
    elif budgetChangePct > 10:
        return f"📈 NAIKKAN +{budgetChangePct:.0f}%"
    elif budgetChangePct > 0:
        return f"📈 NAIKKAN +{budgetChangePct:.0f}%"
    elif budgetChangePct == 0:
        return "➡️ KEEP 0%"
    elif budgetChangePct > -20:
        return f"📉 TURUNKAN {budgetChangePct:.0f}%"
    elif budgetChangePct > -100:
        return f"📉 TURUNKAN {budgetChangePct:.0f}%"
    else:
        return "🛑 STOP -100%"


def buildRecommendations(analyses: List[ComprehensiveAnalysis]) -> List[str]:
    """Build recommendations sections with discontinued filter and type separation."""
    lines = ["<h2>🎯 REKOMENDASI STRATEGIS BERBASIS DATA</h2>"]
    
    # CRITICAL: Filter out discontinued products
    active, discontinued = _filterActiveProducts(analyses)
    
    # Show discontinued notice if any
    if discontinued:
        lines.extend(_buildDiscontinuedNotice(discontinued))
    
    # Hidden Gems: High ROAS but blocked (only from active products)
    hiddenGems = [a for a in active 
                  if a.roas >= 5 
                  and a.finalAction not in [ActionRecommendation.SCALE_UP, ActionRecommendation.SCALE_UP_AGGRESSIVE]]
    hiddenGems = sorted(hiddenGems, key=lambda x: x.roas, reverse=True)
    
    if hiddenGems:
        lines.extend(_buildHiddenGemsSection(hiddenGems))
    
    # Scale Up - separated by type (only from active products)
    scaleUp = [a for a in active if a.finalAction in [
        ActionRecommendation.SCALE_UP_AGGRESSIVE, ActionRecommendation.SCALE_UP
    ]]
    if scaleUp:
        lines.extend(_buildScaleUpSectionByType(scaleUp))
    
    # Maintain - separated by type (only from active products)
    maintain = [a for a in active if a.finalAction == ActionRecommendation.MAINTAIN]
    if maintain:
        lines.extend(_buildMaintainSectionByType(maintain))
    
    # Reduce (only from active products)
    reduce = [a for a in active if a.finalAction == ActionRecommendation.REDUCE]
    if reduce:
        lines.extend(_buildReduceSection(reduce))
    
    # Stop - only ACTIVE products that need stopping (only from active products)
    stop = [a for a in active if a.finalAction == ActionRecommendation.STOP]
    if stop:
        lines.extend(_buildStopSection(stop))
    
    return lines


def _buildDiscontinuedNotice(discontinued: List[ComprehensiveAnalysis]) -> List[str]:
    """Build notice for discontinued products."""
    lines = [
        "<h3>ℹ️ PRODUK DISCONTINUED (Tidak Termasuk Analisis)</h3>",
        "<p>Produk berikut sudah dihentikan dan <strong>tidak diberikan rekomendasi budget</strong>:</p>",
        "<table>",
        "<thead><tr><th>#</th><th>Produk</th><th>Tipe</th><th>Status Terakhir</th></tr></thead><tbody>",
    ]
    
    for i, a in enumerate(discontinued, 1):
        status = f"ROAS {a.roas:.1f}x" if a.roas > 0 else "Tidak ada revenue"
        lines.append(
            f"<tr><td>{i}</td><td>{a.productName}</td><td>{a.creativeType or '-'}</td>"
            f"<td>{status}</td></tr>"
        )
    
    lines.append("</tbody></table>")
    lines.append("<blockquote><strong>⚠️ Catatan:</strong> Produk discontinued "
                 "sudah diverifikasi tidak aktif dan tidak memerlukan tindakan apapun.</blockquote>")
    return lines


def _buildHiddenGemsSection(hiddenGems: List[ComprehensiveAnalysis]) -> List[str]:
    """Build hidden gems section."""
    # Separate by type
    video, card = _splitByCreativeType(hiddenGems)
    
    lines = [
        "<h3>💎 HIDDEN GEMS - Perlu Perhatian Khusus</h3>",
        "<p><strong>Produk dengan ROAS tinggi tapi memerlukan tindakan sebelum scale-up:</strong></p>",
    ]
    
    if video:
        lines.append("<h4>🎬 Video Ads</h4>")
        lines.extend(_buildHiddenGemsTable(video[:5]))
    
    if card:
        lines.append("<h4>🃏 Product Card</h4>")
        lines.extend(_buildHiddenGemsTable(card[:5]))
    
    lines.append("<blockquote><strong>💡 Insight:</strong> Produk ini memiliki potensi besar. ")
    lines.append("Setelah mengatasi blocker, produk ini berpeluang menjadi kandidat scale-up terbaik.</blockquote>")
    return lines


def _buildHiddenGemsTable(products: List[ComprehensiveAnalysis]) -> List[str]:
    """Build hidden gems table."""
    lines = [
        "<table>",
        "<thead><tr><th>#</th><th>Produk</th><th>ROAS</th><th>Profit</th><th>Blocker</th><th>Rekomendasi</th></tr></thead><tbody>",
    ]
    
    for i, a in enumerate(products, 1):
        blockers, recommendations = _identifyBlockers(a)
        blockerStr = ", ".join(blockers) if blockers else "Multiple factors"
        recStr = ", ".join(recommendations) if recommendations else "Evaluate further"
        productDisplay = _formatProductName(a.productName)
        
        lines.append(
            f"<tr><td>{i}</td><td>{productDisplay}</td><td>{a.roas:.1f}x</td>"
            f"<td>Rp {a.totalProfit:,.0f}</td><td>{blockerStr}</td><td>{recStr}</td></tr>"
        )
    
    lines.append("</tbody></table>")
    return lines


def _identifyBlockers(analysis: ComprehensiveAnalysis) -> tuple:
    """Identify blockers and recommendations for hidden gems."""
    blockers = []
    recommendations = []
    us = analysis.unifiedScore
    
    if us.fatigueScore < 30:
        blockers.append("Creative Fatigue")
        recommendations.append("Refresh creative")
    if us.volatilityScore < 30:
        blockers.append("Volatility Tinggi")
        recommendations.append("Stabilkan dulu")
    if us.elasticityScore < 30:
        blockers.append("Saturasi Budget")
        recommendations.append("Review target audience")
    
    return blockers, recommendations


def _buildScaleUpSectionByType(scaleUp: List[ComprehensiveAnalysis]) -> List[str]:
    """Build scale up section separated by Video vs Product Card."""
    video, card = _splitByCreativeType(scaleUp)
    
    lines = [
        "<h3>🚀 REKOMENDASI NAIKKAN BUDGET</h3>",
    ]
    
    if video:
        video = sorted(video, key=lambda x: x.unifiedScore.compositeScore, reverse=True)
        lines.append("<h4>🎬 Video Ads</h4>")
        lines.extend(_buildScaleUpTable(video[:8]))
    
    if card:
        card = sorted(card, key=lambda x: x.unifiedScore.compositeScore, reverse=True)
        lines.append("<h4>🃏 Product Card</h4>")
        lines.extend(_buildScaleUpTable(card[:8]))
    
    return lines


def _buildScaleUpTable(products: List[ComprehensiveAnalysis]) -> List[str]:
    """Build scale up table with badges for data quality, lifecycle, saturation."""
    totalBudgetIncrease = sum(a.budgetRec.budgetChange for a in products if a.budgetRec.budgetChange > 0)
    totalEstProfit = sum(a.budgetRec.profitChange for a in products if a.budgetRec.profitChange > 0)
    
    lines = [
        f"<p><strong>Total Tambahan Budget:</strong> Rp {totalBudgetIncrease:,.0f} | "
        f"<strong>Est. Tambahan Profit:</strong> Rp {totalEstProfit:,.0f}</p>",
        "<table>",
        "<thead><tr><th>#</th><th>Produk</th><th>ROAS</th><th>Status</th>"
        "<th>Rekomendasi</th><th>Est. Profit</th><th>Data</th></tr></thead><tbody>",
    ]
    
    for i, a in enumerate(products, 1):
        budgetPct = a.finalBudgetChange
        budgetLabel = _getBudgetChangeLabel(budgetPct)
        lifecycleBadge = _getLifecycleBadge(a)
        saturationBadge = _getSaturationBadge(a)
        dataQualityBadge = _getDataQualityBadge(a.nPeriods)
        profitWithCI = _formatProfitWithCI(a)
        productDisplay = _formatProductName(a.productName)
        
        lines.append(
            f"<tr><td>{i}</td><td>{productDisplay}</td><td>{a.roas:.1f}x</td>"
            f"<td>{lifecycleBadge} {saturationBadge}</td>"
            f"<td><strong>{budgetLabel}</strong><br/><small>+Rp {a.budgetRec.budgetChange:,.0f}</small></td>"
            f"<td>{profitWithCI}</td>"
            f"<td>{dataQualityBadge}</td></tr>"
        )
    
    lines.append("</tbody></table>")
    return lines


def _buildMaintainSectionByType(maintain: List[ComprehensiveAnalysis]) -> List[str]:
    """Build maintain section separated by type."""
    video, card = _splitByCreativeType(maintain)
    
    lines = [
        "<h3>✅ PERTAHANKAN BUDGET (Keep 0%)</h3>",
    ]
    
    if video:
        video = sorted(video, key=lambda x: x.totalProfit, reverse=True)
        lines.append("<h4>🎬 Video Ads</h4>")
        lines.extend(_buildMaintainTable(video[:5]))
    
    if card:
        card = sorted(card, key=lambda x: x.totalProfit, reverse=True)
        lines.append("<h4>🃏 Product Card</h4>")
        lines.extend(_buildMaintainTable(card[:5]))
    
    return lines


def _buildMaintainTable(products: List[ComprehensiveAnalysis]) -> List[str]:
    """Build maintain table."""
    totalProfit = sum(a.totalProfit for a in products)
    
    lines = [
        f"<p>Jumlah: {len(products)} | Total Profit: Rp {totalProfit:,.0f}</p>",
        "<table>",
        "<thead><tr><th>#</th><th>Produk</th><th>ROAS</th><th>Profit</th><th>Score</th><th>Status</th></tr></thead><tbody>",
    ]
    
    for i, a in enumerate(products, 1):
        productDisplay = _formatProductName(a.productName)
        lines.append(
            f"<tr><td>{i}</td><td>{productDisplay}</td><td>{a.roas:.2f}x</td>"
            f"<td>Rp {a.totalProfit:,.0f}</td><td>{a.unifiedScore.compositeScore:.0f}</td>"
            f"<td>➡️ KEEP 0%</td></tr>"
        )
    
    lines.append("</tbody></table>")
    return lines


def _buildReduceSection(reduce: List[ComprehensiveAnalysis]) -> List[str]:
    """Build reduce budget section with specific percentages."""
    video, card = _splitByCreativeType(reduce)
    
    lines = [
        "<h3>📉 TURUNKAN BUDGET</h3>",
        "<p>Produk dengan performa di bawah rata-rata - kurangi budget untuk optimasi.</p>",
    ]
    
    allReduce = sorted(reduce, key=lambda x: x.roas)
    lines.append("<table>")
    lines.append("<thead><tr><th>#</th><th>Produk</th><th>Tipe</th><th>ROAS</th>"
                 "<th>Rekomendasi</th><th>Alasan</th></tr></thead><tbody>")
    
    for i, a in enumerate(allReduce[:10], 1):
        budgetLabel = _getBudgetChangeLabel(a.finalBudgetChange)
        creativeType = "🎬 Video" if "Video" in (a.creativeType or "") else "🃏 Card"
        reason = "ROI rendah" if a.roas < 3 else "Trend menurun" if "Turun" in a.legacyMkTrend else "Volatility tinggi"
        productDisplay = _formatProductName(a.productName)
        
        lines.append(
            f"<tr><td>{i}</td><td>{productDisplay}</td><td>{creativeType}</td>"
            f"<td>{a.roas:.2f}x</td><td><strong>{budgetLabel}</strong></td>"
            f"<td>{reason}</td></tr>"
        )
    
    lines.append("</tbody></table>")
    return lines


def _buildScaleUpSection(scaleUp: List[ComprehensiveAnalysis]) -> List[str]:
    """Build scale up section (legacy compatibility)."""
    return _buildScaleUpSectionByType(scaleUp)


def _buildMaintainSection(maintain: List[ComprehensiveAnalysis]) -> List[str]:
    """Build maintain section (legacy compatibility)."""
    return _buildMaintainSectionByType(maintain)


def _buildStopSection(stop: List[ComprehensiveAnalysis]) -> List[str]:
    """Build stop section - only for ACTIVE products."""
    # Double check: exclude any discontinued products
    activeStop = [a for a in stop if not a.isDiscontinued]
    
    if not activeStop:
        return []
    
    totalLoss = sum(abs(a.totalProfit) for a in activeStop if a.totalProfit < 0)
    video, card = _splitByCreativeType(activeStop)
    
    lines = [
        "<h3>🛑 HENTIKAN SEGERA (IKLAN MASIH AKTIF)</h3>",
        f"<p><strong>Jumlah Produk:</strong> {len(activeStop)} | "
        f"<strong>Potensi Penghematan:</strong> Rp {totalLoss:,.0f}</p>",
        "<p><em>⚠️ Produk yang sudah discontinued tidak termasuk dalam daftar ini.</em></p>",
    ]
    
    if video:
        lines.append("<h4>🎬 Video Ads</h4>")
        lines.extend(_buildStopTable(video))
    
    if card:
        lines.append("<h4>🃏 Product Card</h4>")
        lines.extend(_buildStopTable(card))
    
    return lines


def _buildStopTable(products: List[ComprehensiveAnalysis]) -> List[str]:
    """Build stop table."""
    lines = [
        "<table>",
        "<thead><tr><th>#</th><th>Produk</th><th>ROAS</th><th>Kerugian</th><th>Rekomendasi</th><th>Alasan</th></tr></thead><tbody>",
    ]
    
    for i, a in enumerate(products, 1):
        loss = abs(a.totalProfit) if a.totalProfit < 0 else 0
        warnings = a.unifiedScore.warningFactors[:1] if a.unifiedScore.warningFactors else ["ROI terlalu rendah"]
        productDisplay = _formatProductName(a.productName)
        
        lines.append(
            f"<tr><td>{i}</td><td>{productDisplay}</td><td>{a.roas:.2f}x</td>"
            f"<td>Rp {loss:,.0f}</td><td><strong>🛑 STOP -100%</strong></td>"
            f"<td>{warnings[0][:40]}</td></tr>"
        )
    
    lines.append("</tbody></table>")
    return lines


def buildChurnRiskSection(analyses: List[ComprehensiveAnalysis]) -> List[str]:
    """Build churn risk analysis section."""
    byRisk = sorted(analyses, key=lambda x: x.unifiedScore.churnRiskScore, reverse=True)
    
    highRisk = [a for a in analyses if a.unifiedScore.churnRiskScore > 60]
    medRisk = [a for a in analyses if 30 <= a.unifiedScore.churnRiskScore <= 60]
    lowRisk = [a for a in analyses if a.unifiedScore.churnRiskScore < 30]
    
    avgRisk = np.mean([a.unifiedScore.churnRiskScore for a in analyses])
    
    lines = [
        "<h2>⚠️ SISTEM PERINGATAN DINI (Churn Risk)</h2>",
        "<p>Analisis risiko penurunan performa produk berdasarkan momentum, trend, dan volatilitas.</p>",
        
        "<h3>Gambaran Umum Risiko Portfolio</h3>",
        "<table>",
        "<thead><tr><th>Metrik</th><th>Nilai</th></tr></thead><tbody>",
        f"<tr><td>Skor Risiko Portfolio</td><td><strong>{avgRisk:.1f}/100</strong></td></tr>",
        f"<tr><td>Produk Risiko Tinggi</td><td>{len(highRisk)} ({len(highRisk)/len(analyses)*100:.1f}%)</td></tr>",
        f"<tr><td>Produk Risiko Sedang</td><td>{len(medRisk)}</td></tr>",
        f"<tr><td>Produk Risiko Rendah</td><td>{len(lowRisk)}</td></tr>",
        "</tbody></table>",
    ]
    
    if highRisk:
        lines.extend(_buildHighRiskTable(highRisk))
    
    return lines


def _buildHighRiskTable(highRisk: List[ComprehensiveAnalysis]) -> List[str]:
    """Build high risk products table."""
    lines = [
        "<h3>🔴 Produk dengan Risiko Tinggi</h3>",
        "<table>",
        "<thead><tr><th>#</th><th>Produk</th><th>Skor Risiko</th><th>Tingkat</th><th>Faktor Risiko</th></tr></thead><tbody>",
    ]
    
    for i, a in enumerate(highRisk[:10], 1):
        factors = []
        if a.legacyMomentumPct < -20:
            factors.append(f"Momentum: {a.legacyMomentumPct:.0f}%")
        if "Turun" in a.legacyMkTrend:
            factors.append("Trend menurun")
        if a.legacyCv > 50:
            factors.append(f"Volatility tinggi: CV={a.legacyCv:.0f}%")
        
        factorStr = "; ".join(factors) if factors else "Multiple factors"
        productDisplay = _formatProductName(a.productName)
        lines.append(
            f"<tr><td>{i}</td><td>{productDisplay}</td><td>{a.unifiedScore.churnRiskScore:.0f}/100</td>"
            f"<td>🔴 HIGH</td><td>{factorStr}</td></tr>"
        )
    
    lines.append("</tbody></table>")
    return lines
