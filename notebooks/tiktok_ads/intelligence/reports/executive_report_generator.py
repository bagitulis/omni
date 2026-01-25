#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Executive Report Generator v2.1
================================
Compact executive summary for management with:
- KPI cards with visual highlights
- Video vs Card separation
- Specific budget percentages + nominal
- Indonesian event context
- ALL stop products (no limit)
- Discontinued products filtered out
- Data quality warning badges
- Lifecycle stage indicators
- Saturation status badges
- Confidence interval display
"""

from typing import List, Tuple
from datetime import datetime, timedelta
import numpy as np

from ..comprehensive_engine import ComprehensiveAnalysis
from ..unified_scorer import ActionRecommendation
from .html_styles import CSS_STYLE
from ..portfolio_metrics import (
    PortfolioHealth, FinancialProjection,
    calculateHealthScore, calculateProjection
)
from ..indonesian_calendar import IndonesianCalendar


def _filterActive(analyses: List[ComprehensiveAnalysis]) -> List[ComprehensiveAnalysis]:
    """Filter out discontinued products."""
    return [a for a in analyses if not a.isDiscontinued]


def _splitByType(analyses: List[ComprehensiveAnalysis]) -> Tuple[List, List]:
    """Split by Video vs Product Card."""
    video = [a for a in analyses if a.creativeType and "Video" in a.creativeType]
    card = [a for a in analyses if a.creativeType and ("Kartu" in a.creativeType or "Card" in a.creativeType)]
    return video, card


def _getBudgetLabel(pct: float) -> str:
    """Get budget change label."""
    if pct > 0:
        return f"📈 +{pct:.0f}%"
    elif pct == 0:
        return "➡️ 0%"
    elif pct > -100:
        return f"📉 {pct:.0f}%"
    else:
        return "🛑 STOP"


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


def _getDataBadge(nPeriods: int) -> str:
    """Get data quality badge - Bahasa Indonesia."""
    if nPeriods < 4:
        return '<span class="badge badge-warning" title="Data kurang dari 4 periode">⚠️ Terbatas</span>'
    elif nPeriods < 7:
        return '<span class="badge badge-info" title="Data 4-6 periode">📊 Cukup</span>'
    return '<span class="badge badge-success" title="Data 7+ periode">✅ Lengkap</span>'


def _getLifecycleBadge(a: ComprehensiveAnalysis) -> str:
    """
    Get lifecycle badge - Fase Produk (Bahasa Indonesia).
    - Bertumbuh: Momentum kuat, trend naik
    - Menurun: Momentum negatif atau trend turun
    - Stabil Untung: ROAS bagus, performa konsisten
    - Stagnan: Tidak ada pergerakan signifikan
    """
    if a.legacyMomentumPct > 20 and "Naik" in a.legacyMkTrend:
        return '<span class="badge badge-growth" title="Pertumbuhan kuat - layak naikkan budget">🚀 Bertumbuh</span>'
    elif a.legacyMomentumPct < -10 or "Turun" in a.legacyMkTrend:
        return '<span class="badge badge-decline" title="Performa menurun - evaluasi budget">📉 Menurun</span>'
    elif a.roas >= 3:
        return '<span class="badge badge-mature" title="Stabil menguntungkan - pertahankan">📈 Stabil Untung</span>'
    return '<span class="badge badge-stable" title="Performa datar - pantau">➡️ Stagnan</span>'


def _getSatBadge(a: ComprehensiveAnalysis) -> str:
    """
    Get saturation badge - Respons Budget (Bahasa Indonesia).
    - Responsif: Tambah budget = revenue naik
    - Cukup Responsif: Masih ada ruang
    - Hampir Jenuh: Tambah budget kurang efektif
    """
    elasticity = a.unifiedScore.elasticityScore
    if elasticity >= 60:
        return '<span class="badge badge-success" title="Tambah budget akan meningkatkan revenue">🟢 Responsif</span>'
    elif elasticity >= 40:
        return '<span class="badge badge-info" title="Masih ada ruang untuk tambah budget">🔵 Cukup Responsif</span>'
    return '<span class="badge badge-warning" title="Tambah budget kurang efektif">🟡 Hampir Jenuh</span>'


class ExecutiveReportGenerator:
    """
    Generates compact executive HTML report v2.0.
    Designed for quick management decisions with clear actions.
    """
    
    def __init__(self):
        self.timestamp = datetime.now()
        self.CSS_STYLE = CSS_STYLE
        self.calendar = IndonesianCalendar()
    
    def generateReport(self, analyses: List[ComprehensiveAnalysis],
                        title: str = "TikTok Ads - Executive Report"
                        ) -> str:
        """Generate executive HTML report."""
        
        # Filter active products for recommendations
        active = _filterActive(analyses)
        discontinued = [a for a in analyses if a.isDiscontinued]
        
        health = calculateHealthScore(analyses)
        projection = calculateProjection(analyses)
        
        html = [
            "<!DOCTYPE html>",
            '<html lang="id">',
            "<head>",
            '<meta charset="UTF-8">',
            '<meta name="viewport" content="width=device-width, initial-scale=1.0">',
            f"<title>{title}</title>",
            self.CSS_STYLE,
            "</head>",
            "<body>",
            '<div class="container">',
        ]
        
        # Header with period
        html.extend(self._buildHeader(title, analyses))
        
        # KPI Cards
        html.extend(self._buildKpiCards(analyses, health, projection))
        
        # Event Context
        html.extend(self._buildEventContext())
        
        # Priority Actions
        html.extend(self._buildPriorityActions(active))
        
        # STOP Products (ALL - no limit)
        html.extend(self._buildStopSection(active))
        
        # Scale Up by Type
        html.extend(self._buildScaleUpSection(active))
        
        # Reduce Budget
        html.extend(self._buildReduceSection(active))
        
        # Risk Monitoring
        html.extend(self._buildRiskSection(active))
        
        # Legend/Catatan Istilah
        html.extend(self._buildLegend())
        
        # Discontinued Notice
        if discontinued:
            html.extend(self._buildDiscontinuedNotice(discontinued))
        
        # Footer
        html.extend(self._buildFooter())
        
        return "\n".join(html)
    
    def _buildHeader(self, title: str, analyses: List[ComprehensiveAnalysis]) -> List[str]:
        """Build header with period info."""
        period_start = "Jul 2025"
        period_end = self.timestamp.strftime("%d %b %Y")
        
        return [
            f"<h1>{title}</h1>",
            f"<p><strong>Periode Data:</strong> {period_start} - {period_end}</p>",
            f"<p><strong>Generated:</strong> {self.timestamp.strftime('%Y-%m-%d %H:%M')}</p>",
        ]
    
    def _buildKpiCards(self, analyses: List[ComprehensiveAnalysis],
                        health: PortfolioHealth,
                        projection: FinancialProjection) -> List[str]:
        """Build KPI highlight cards."""
        totalCost = sum(a.totalCost for a in analyses)
        totalRevenue = sum(a.totalRevenue for a in analyses)
        totalProfit = totalRevenue - totalCost
        overallRoas = totalRevenue / totalCost if totalCost > 0 else 0
        
        profitClass = "kpi-success" if totalProfit > 0 else "kpi-danger"
        roasClass = "kpi-success" if overallRoas >= 3 else "kpi-warning" if overallRoas >= 1 else "kpi-danger"
        healthClass = "kpi-success" if health.healthScore >= 70 else "kpi-warning" if health.healthScore >= 50 else "kpi-danger"
        
        highRisk = len([a for a in analyses if a.unifiedScore.churnRiskScore > 60])
        riskClass = "kpi-danger" if highRisk > 20 else "kpi-warning" if highRisk > 10 else "kpi-success"
        
        if totalProfit >= 1_000_000_000:
            profitStr = f"Rp {totalProfit/1_000_000_000:.2f} M"
        else:
            profitStr = f"Rp {totalProfit/1_000_000:.0f} Jt"
        
        return [
            "<h2>📊 SNAPSHOT BISNIS</h2>",
            '<div class="kpi-grid">',
            f'<div class="kpi-card {profitClass}">',
            f'<div class="kpi-value">{profitStr}</div>',
            '<div class="kpi-label">Total Profit</div>',
            f'<div class="kpi-status">{"✅ SANGAT BAIK" if totalProfit > 1_000_000_000 else "✅ BAIK" if totalProfit > 0 else "❌ RUGI"}</div>',
            '</div>',
            f'<div class="kpi-card {roasClass}">',
            f'<div class="kpi-value">{overallRoas:.2f}x</div>',
            '<div class="kpi-label">ROAS Overall</div>',
            f'<div class="kpi-status">{"✅ EXCELLENT" if overallRoas >= 5 else "✅ BAIK" if overallRoas >= 2 else "⚠️ PERHATIAN"}</div>',
            '</div>',
            f'<div class="kpi-card {healthClass}">',
            f'<div class="kpi-value">{health.healthScore:.0f}/100</div>',
            '<div class="kpi-label">Health Score</div>',
            f'<div class="kpi-status">Grade {health.grade}</div>',
            '</div>',
            f'<div class="kpi-card {riskClass}">',
            f'<div class="kpi-value">{highRisk}</div>',
            '<div class="kpi-label">High Risk Products</div>',
            f'<div class="kpi-status">{health.riskLevel}</div>',
            '</div>',
            '</div>',
        ]
    
    def _buildEventContext(self) -> List[str]:
        """Build Indonesian event context section."""
        today = self.timestamp.date()
        day = today.day
        
        if day in list(range(25, 32)) + list(range(1, 4)):
            phase = "Payday Prime 💰"
            phaseNote = "Tingkatkan budget - daya beli tinggi"
        elif day in range(4, 11):
            phase = "Post Payday"
            phaseNote = "Budget normal"
        elif day in range(11, 18):
            phase = "Mid Month"
            phaseNote = "Fokus high performer saja"
        else:
            phase = "Dry Season"
            phaseNote = "Turunkan budget - daya beli rendah"
        
        lines = [
            "<h2>📅 KONTEKS EVENT INDONESIA</h2>",
            '<div class="event-timeline">',
            f'<div class="event-item event-today">📍 {today.strftime("%d %b")} - {phase}</div>',
        ]
        
        # Check upcoming twin dates
        for offset in range(1, 31):
            checkDate = today + timedelta(days=offset)
            if checkDate.day == checkDate.month:
                lines.append(f'<div class="event-item event-{"upcoming" if offset <= 7 else "future"}">'
                           f'🎯 {checkDate.strftime("%d %b")} - Tanggal Kembar {checkDate.month}.{checkDate.day} ({offset} hari)</div>')
                break
        
        # Check holidays
        for holidayDate, holidayName in self.calendar.NATIONAL_HOLIDAYS.items():
            if today < holidayDate <= today + timedelta(days=14):
                daysUntil = (holidayDate - today).days
                lines.append(f'<div class="event-item event-upcoming">'
                           f'⭐ {holidayDate.strftime("%d %b")} - {holidayName} ({daysUntil} hari)</div>')
        
        lines.append('</div>')
        lines.append(f'<p><em>💡 {phaseNote}</em></p>')
        
        return lines
    
    def _buildPriorityActions(self, active: List[ComprehensiveAnalysis]) -> List[str]:
        """Build priority action boxes."""
        stop = [a for a in active if a.finalAction == ActionRecommendation.STOP]
        reduce = [a for a in active if a.finalAction == ActionRecommendation.REDUCE]
        scaleUp = [a for a in active if a.finalAction in [
            ActionRecommendation.SCALE_UP_AGGRESSIVE, ActionRecommendation.SCALE_UP]]
        
        totalStopSavings = sum(abs(a.totalProfit) for a in stop if a.totalProfit < 0)
        totalScaleProfit = sum(a.budgetRec.profitChange for a in scaleUp if a.budgetRec.profitChange > 0)
        
        lines = ["<h2>⚡ PRIORITAS AKSI</h2>"]
        
        if stop:
            lines.extend([
                '<div class="priority-box priority-critical">',
                '<div class="priority-badge">🔴 P0 - SEGERA (Hari Ini)</div>',
                '<div class="priority-content">',
                f'<strong>STOP {len(stop)} iklan merugi</strong> - Hemat Rp {totalStopSavings:,.0f}',
                '</div></div>',
            ])
        
        if reduce:
            lines.extend([
                '<div class="priority-box priority-high">',
                '<div class="priority-badge">🟠 P1 - Minggu Ini</div>',
                '<div class="priority-content">',
                f'<strong>TURUNKAN budget {len(reduce)} produk</strong> dengan ROI rendah',
                '</div></div>',
            ])
        
        if scaleUp:
            lines.extend([
                '<div class="priority-box priority-normal">',
                '<div class="priority-badge">🟢 P2 - Scale Winners</div>',
                '<div class="priority-content">',
                f'<strong>NAIKKAN budget {len(scaleUp)} produk</strong> - Est. profit +Rp {totalScaleProfit:,.0f}',
                '</div></div>',
            ])
        
        return lines
    
    def _buildStopSection(self, active: List[ComprehensiveAnalysis]) -> List[str]:
        """Build STOP section - ALL products, no limit."""
        stop = [a for a in active if a.finalAction == ActionRecommendation.STOP]
        if not stop:
            return []
        
        video, card = _splitByType(stop)
        totalLoss = sum(abs(a.totalProfit) for a in stop if a.totalProfit < 0)
        
        lines = [
            "<h2>🛑 HENTIKAN SEGERA - SEMUA PRODUK</h2>",
            f"<p><strong>Total: {len(stop)} produk</strong> | Potensi Hemat: <strong>Rp {totalLoss:,.0f}</strong></p>",
        ]
        
        if video:
            lines.append("<h4 class='type-video'>🎬 Video Ads</h4>")
            lines.extend(self._buildStopTable(video))
        
        if card:
            lines.append("<h4 class='type-card'>🃏 Product Card</h4>")
            lines.extend(self._buildStopTable(card))
        
        return lines
    
    def _buildStopTable(self, products: List[ComprehensiveAnalysis]) -> List[str]:
        """Build stop products table - ALL items."""
        lines = [
            "<table>",
            "<thead><tr><th>#</th><th>Produk</th><th>ROAS</th><th>Kerugian</th><th>Budget</th><th>Aksi</th></tr></thead><tbody>",
        ]
        for i, a in enumerate(products, 1):
            loss = abs(a.totalProfit) if a.totalProfit < 0 else 0
            productDisplay = _formatProductName(a.productName)
            lines.append(
                f"<tr><td>{i}</td><td>{productDisplay}</td><td>{a.roas:.2f}x</td>"
                f"<td>Rp {loss:,.0f}</td><td>Rp {a.budgetRec.currentBudget:,.0f}</td>"
                f"<td><strong>🛑 STOP -100%</strong></td></tr>"
            )
        lines.append("</tbody></table>")
        return lines
    
    def _buildScaleUpSection(self, active: List[ComprehensiveAnalysis]) -> List[str]:
        """Build scale up section by type."""
        scaleUp = [a for a in active if a.finalAction in [
            ActionRecommendation.SCALE_UP_AGGRESSIVE, ActionRecommendation.SCALE_UP]]
        if not scaleUp:
            return []
        
        video, card = _splitByType(scaleUp)
        lines = ["<h2>🚀 NAIKKAN BUDGET</h2>"]
        
        if video:
            video = sorted(video, key=lambda x: x.budgetRec.profitChange, reverse=True)
            lines.append("<h4 class='type-video'>🎬 Video Ads (Top 8)</h4>")
            lines.extend(self._buildScaleTable(video[:8]))
        
        if card:
            card = sorted(card, key=lambda x: x.budgetRec.profitChange, reverse=True)
            lines.append("<h4 class='type-card'>🃏 Product Card (Top 8)</h4>")
            lines.extend(self._buildScaleTable(card[:8]))
        
        return lines
    
    def _buildScaleTable(self, products: List[ComprehensiveAnalysis]) -> List[str]:
        """Build scale up table with badges for status, saturation, data quality."""
        lines = [
            "<table>",
            "<thead><tr><th>#</th><th>Produk</th><th>ROAS</th><th>Status</th>"
            "<th>Rekomendasi</th><th>Est. Profit (CI)</th><th>Data</th></tr></thead><tbody>",
        ]
        for i, a in enumerate(products, 1):
            pctLabel = _getBudgetLabel(a.finalBudgetChange)
            nominalChange = a.budgetRec.budgetChange
            lifecycleBadge = _getLifecycleBadge(a)
            satBadge = _getSatBadge(a)
            dataBadge = _getDataBadge(a.nPeriods)
            productDisplay = _formatProductName(a.productName)
            
            # Format profit with CI
            ci = a.budgetRec.profitCI
            profitStr = f"+Rp {a.budgetRec.profitChange:,.0f}<br/><small>CI: [{ci[0]:,.0f}, {ci[1]:,.0f}]</small>"
            
            lines.append(
                f"<tr><td>{i}</td><td>{productDisplay}</td><td>{a.roas:.1f}x</td>"
                f"<td>{lifecycleBadge}<br/>{satBadge}</td>"
                f"<td><strong>{pctLabel}</strong><br/><small>+Rp {nominalChange:,.0f}</small></td>"
                f"<td>{profitStr}</td><td>{dataBadge}</td></tr>"
            )
        lines.append("</tbody></table>")
        return lines
    
    def _buildReduceSection(self, active: List[ComprehensiveAnalysis]) -> List[str]:
        """Build reduce budget section."""
        reduce = [a for a in active if a.finalAction == ActionRecommendation.REDUCE]
        if not reduce:
            return []
        
        lines = [
            "<h2>📉 TURUNKAN BUDGET</h2>",
            "<table>",
            "<thead><tr><th>#</th><th>Produk</th><th>Tipe</th><th>ROAS</th>"
            "<th>Budget</th><th>Rekomendasi</th><th>Alasan</th></tr></thead><tbody>",
        ]
        
        for i, a in enumerate(sorted(reduce, key=lambda x: x.roas), 1):
            pctLabel = _getBudgetLabel(a.finalBudgetChange)
            creativeType = "🎬 Video" if "Video" in (a.creativeType or "") else "🃏 Card"
            reason = "ROI rendah" if a.roas < 3 else "Trend menurun" if "Turun" in a.legacyMkTrend else "Volatility"
            productDisplay = _formatProductName(a.productName)
            
            lines.append(
                f"<tr><td>{i}</td><td>{productDisplay}</td><td>{creativeType}</td>"
                f"<td>{a.roas:.2f}x</td><td>Rp {a.budgetRec.currentBudget:,.0f}</td>"
                f"<td><strong>{pctLabel}</strong><br/><small>Rp {a.budgetRec.budgetChange:,.0f}</small></td>"
                f"<td>{reason}</td></tr>"
            )
        
        lines.append("</tbody></table>")
        return lines
    
    def _buildRiskSection(self, active: List[ComprehensiveAnalysis]) -> List[str]:
        """Build risk monitoring with data quality badges."""
        highRisk = [a for a in active if a.unifiedScore.churnRiskScore > 60 
                    and a.finalAction not in [ActionRecommendation.STOP]]
        
        if not highRisk:
            return [
                "<h2>⚠️ RISK MONITORING</h2>",
                "<p>✅ Tidak ada produk aktif dengan risiko tinggi.</p>"
            ]
        
        lines = [
            "<h2>⚠️ RISK MONITORING</h2>",
            f"<p><strong>{len(highRisk)} produk</strong> risiko tinggi (exclude yang harus STOP)</p>",
            "<table>",
            "<thead><tr><th>#</th><th>Produk</th><th>Status</th><th>Risk</th><th>Faktor</th><th>Data</th></tr></thead><tbody>",
        ]
        
        for i, a in enumerate(sorted(highRisk, key=lambda x: x.unifiedScore.churnRiskScore, reverse=True)[:10], 1):
            creativeType = "🎬 Video" if "Video" in (a.creativeType or "") else "🃏 Card"
            lifecycleBadge = _getLifecycleBadge(a)
            dataBadge = _getDataBadge(a.nPeriods)
            productDisplay = _formatProductName(a.productName)
            issue = a.unifiedScore.warningFactors[0][:35] if a.unifiedScore.warningFactors else "Multiple"
            lines.append(
                f"<tr><td>{i}</td><td>{productDisplay}</td><td>{lifecycleBadge}</td>"
                f"<td><span class='badge badge-danger'>{a.unifiedScore.churnRiskScore:.0f}/100</span></td>"
                f"<td>{issue}</td><td>{dataBadge}</td></tr>"
            )
        
        lines.append("</tbody></table>")
        return lines
    
    def _buildDiscontinuedNotice(self, discontinued: List[ComprehensiveAnalysis]) -> List[str]:
        """Build discontinued notice."""
        return [
            "<h2>ℹ️ PRODUK DISCONTINUED</h2>",
            '<div class="discontinued-notice">',
            f"<strong>{len(discontinued)} produk</strong> sudah dihentikan, tidak perlu aksi.",
            '</div>',
        ]
    
    def _buildLegend(self) -> List[str]:
        """
        Build legend section - Catatan Istilah dalam Bahasa Indonesia.
        Menjelaskan semua badge dan indikator yang digunakan di laporan.
        """
        return [
            "<h2>📖 CATATAN ISTILAH</h2>",
            '<div class="legend-section">',
            
            # Fase Produk
            '<div class="legend-group">',
            '<h4>Fase Produk:</h4>',
            '<ul>',
            '<li><span class="badge badge-growth">🚀 Bertumbuh</span> = Momentum kuat & trend naik. <strong>Layak naikkan budget.</strong></li>',
            '<li><span class="badge badge-mature">📈 Stabil Untung</span> = Performa konsisten dengan ROAS bagus. <strong>Pertahankan budget.</strong></li>',
            '<li><span class="badge badge-decline">📉 Menurun</span> = Performa menurun. <strong>Evaluasi atau kurangi budget.</strong></li>',
            '<li><span class="badge badge-stable">➡️ Stagnan</span> = Tidak ada pergerakan signifikan. <strong>Pantau perkembangan.</strong></li>',
            '</ul>',
            '</div>',
            
            # Respons Budget
            '<div class="legend-group">',
            '<h4>Respons terhadap Budget:</h4>',
            '<ul>',
            '<li><span class="badge badge-success">🟢 Responsif</span> = Tambah budget akan meningkatkan revenue secara proporsional.</li>',
            '<li><span class="badge badge-info">🔵 Cukup Responsif</span> = Masih ada ruang untuk meningkatkan budget.</li>',
            '<li><span class="badge badge-warning">🟡 Hampir Jenuh</span> = Tambah budget kurang efektif, mendekati titik jenuh.</li>',
            '<li><span class="badge badge-danger">🔴 Jenuh</span> = Budget sudah maksimal, tambah budget tidak akan meningkatkan revenue.</li>',
            '</ul>',
            '</div>',
            
            # Kualitas Data
            '<div class="legend-group">',
            '<h4>Kualitas Data:</h4>',
            '<ul>',
            '<li><span class="badge badge-success">✅ Lengkap</span> = Data 7+ periode - rekomendasi sangat akurat.</li>',
            '<li><span class="badge badge-info">📊 Cukup</span> = Data 4-6 periode - cukup untuk analisis.</li>',
            '<li><span class="badge badge-warning">⚠️ Terbatas</span> = Data kurang dari 4 periode - rekomendasi kurang akurat.</li>',
            '</ul>',
            '</div>',
            
            # Rekomendasi Budget
            '<div class="legend-group">',
            '<h4>Rekomendasi Budget:</h4>',
            '<ul>',
            '<li><strong>📈 +50%</strong> = Naikkan budget 50% dari budget saat ini.</li>',
            '<li><strong>📈 +25%</strong> = Naikkan budget 25% dari budget saat ini.</li>',
            '<li><strong>➡️ 0%</strong> = Pertahankan budget saat ini.</li>',
            '<li><strong>📉 -30%</strong> = Kurangi budget 30% dari budget saat ini.</li>',
            '<li><strong>🛑 STOP</strong> = Hentikan iklan sepenuhnya.</li>',
            '</ul>',
            '</div>',
            
            # Catatan Penting
            '<div class="legend-group">',
            '<h4>Catatan Penting:</h4>',
            '<ul>',
            '<li><strong>ROAS (Return on Ad Spend)</strong> = Revenue dibagi Cost. ROAS 3x artinya setiap Rp 1 yang dikeluarkan menghasilkan Rp 3.</li>',
            '<li><strong>Confidence Interval (CI)</strong> = Rentang estimasi dengan 95% kepercayaan. Contoh: CI [500k, 800k] artinya profit diperkirakan antara Rp 500k - Rp 800k.</li>',
            '<li><strong>Score</strong> = Skor gabungan 0-100 dari 8 faktor: ROAS, Trend, Momentum, Elastisitas, Volatilitas, Kelelahan Kreatif, Risiko Churn, dan Konteks Event.</li>',
            '</ul>',
            '</div>',
            
            '</div>',  # End legend-section
        ]
    
    def _buildFooter(self) -> List[str]:
        """Build footer."""
        nextReview = self.timestamp + timedelta(days=7)
        return [
            '<div class="footer">',
            "<p><strong>Unified Intelligence System v3.0.0</strong></p>",
            f"<p>Generated: {self.timestamp.strftime('%Y-%m-%d %H:%M')}</p>",
            f"<p>Next Review: {nextReview.strftime('%d %b %Y')}</p>",
            "</div></div></body></html>"
        ]
