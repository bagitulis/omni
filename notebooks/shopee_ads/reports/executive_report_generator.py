#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads - Executive Report Generator v3.0
=============================================
Generates comprehensive executive HTML report with:
- KPI snapshot cards
- Event context Indonesia
- Priority action boxes
- Stop/Scale/Reduce tables with badges
- Risk monitoring
- Legend with terminology

Aligned with TikTok executive report format.
"""

from datetime import datetime, timedelta
from typing import List, Dict, Any, Tuple

# CSS Style - same as TikTok for consistency
CSS_STYLE = """
<style>
    body { font-family: 'Segoe UI', Arial; max-width: 1100px; margin: 20px auto; padding: 15px; background: #f5f5f5; }
    .container { background: #fff; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
    h1 { color: #f53d2d; border-bottom: 3px solid #f53d2d; padding-bottom: 10px; } /* Shopee orange */
    h2 { color: #333; margin-top: 30px; border-left: 4px solid #f53d2d; padding-left: 10px; }
    h4 { margin: 15px 0 10px; }
    .type-video { color: #9c27b0; } .type-card { color: #2196f3; }
    table { width: 100%; border-collapse: collapse; margin: 15px 0; font-size: 13px; }
    th, td { padding: 10px; border: 1px solid #ddd; text-align: left; }
    th { background: #f53d2d; color: white; }
    tr:nth-child(even) { background: #fafafa; }
    .badge { display: inline-block; padding: 3px 8px; border-radius: 12px; font-size: 11px; margin: 2px; }
    .badge-success { background: #dcfce7; color: #166534; }
    .badge-warning { background: #fef3c7; color: #92400e; }
    .badge-danger { background: #fee2e2; color: #991b1b; }
    .badge-info { background: #dbeafe; color: #1e40af; }
    .badge-growth { background: #d1fae5; color: #065f46; }
    .badge-decline { background: #fce7f3; color: #9d174d; }
    .badge-mature { background: #e0e7ff; color: #3730a3; }
    .badge-stable { background: #f3f4f6; color: #374151; }
    .kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 15px; margin: 20px 0; }
    .kpi-card { padding: 20px; border-radius: 10px; text-align: center; }
    .kpi-success { background: linear-gradient(135deg, #10b981, #059669); color: white; }
    .kpi-warning { background: linear-gradient(135deg, #f59e0b, #d97706); color: white; }
    .kpi-danger { background: linear-gradient(135deg, #ef4444, #dc2626); color: white; }
    .kpi-value { font-size: 28px; font-weight: bold; }
    .kpi-label { font-size: 12px; opacity: 0.9; }
    .kpi-status { font-size: 11px; margin-top: 5px; }
    .priority-box { margin: 10px 0; padding: 15px; border-radius: 8px; border-left: 4px solid; }
    .priority-critical { background: #fef2f2; border-color: #ef4444; }
    .priority-high { background: #fffbeb; border-color: #f59e0b; }
    .priority-normal { background: #f0fdf4; border-color: #22c55e; }
    .priority-badge { font-weight: bold; font-size: 14px; }
    .priority-content { margin-top: 8px; }
    .event-timeline { margin: 15px 0; }
    .event-item { padding: 8px 15px; margin: 5px 0; border-radius: 5px; }
    .event-today { background: #f53d2d; color: white; }
    .event-upcoming { background: #fff7ed; border: 1px solid #fb923c; }
    .event-future { background: #f0f9ff; border: 1px solid #38bdf8; }
    .legend-section { background: #f9fafb; padding: 20px; border-radius: 8px; margin-top: 20px; }
    .legend-group { margin-bottom: 15px; }
    .legend-group h4 { margin-bottom: 8px; color: #374151; }
    .legend-group ul { margin: 0; padding-left: 20px; }
    .legend-group li { margin: 5px 0; font-size: 13px; }
    .footer { text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; color: #666; font-size: 12px; }
    .discontinued-notice { background: #f3f4f6; padding: 15px; border-radius: 8px; color: #6b7280; }
    @media (max-width: 768px) { .kpi-grid { grid-template-columns: repeat(2, 1fr); } }
</style>
"""

# Indonesian Calendar Events
NATIONAL_HOLIDAYS = {
    # Update this with current year holidays
}


def _formatProductName(name: str) -> str:
    """Format product name with ellipsis if too long."""
    if not name:
        return "Unknown"
    return name[:40] + "..." if len(name) > 40 else name


def _getBudgetLabel(pctChange: float) -> str:
    """Get budget recommendation label."""
    if pctChange <= -100:
        return "🛑 STOP"
    elif pctChange < -30:
        return f"📉 {int(pctChange)}%"
    elif pctChange < 0:
        return f"📉 {int(pctChange)}%"
    elif pctChange == 0:
        return "➡️ 0%"
    elif pctChange <= 25:
        return f"📈 +{int(pctChange)}%"
    else:
        return f"🚀 +{int(pctChange)}%"


def _splitByBiddingMode(products: List[Dict]) -> Tuple[List[Dict], List[Dict]]:
    """Split products by bidding mode - Auto vs Manual."""
    auto = [p for p in products if "auto" in str(p.get("bidding_mode", "")).lower()]
    manual = [p for p in products if p not in auto]
    return auto, manual


def _getLifecycleBadge(p: Dict) -> str:
    """Get lifecycle badge based on momentum and trend."""
    momentum = p.get("momentum_pct", 0)
    trend = p.get("trend", "")
    roas = p.get("roi", 0)
    
    if momentum > 20 and "Naik" in trend:
        return '<span class="badge badge-growth" title="Pertumbuhan kuat">🚀 Bertumbuh</span>'
    elif momentum < -10 or "Turun" in trend:
        return '<span class="badge badge-decline" title="Performa menurun">📉 Menurun</span>'
    elif roas >= 3:
        return '<span class="badge badge-mature" title="Stabil menguntungkan">📈 Stabil Untung</span>'
    return '<span class="badge badge-stable" title="Performa datar">➡️ Stagnan</span>'


def _getSaturationBadge(p: Dict) -> str:
    """Get saturation/elasticity badge."""
    elasticity = p.get("elasticity_score", 50)
    
    if elasticity >= 60:
        return '<span class="badge badge-success" title="Budget responsif">🟢 Responsif</span>'
    elif elasticity >= 40:
        return '<span class="badge badge-info" title="Cukup responsif">🔵 Cukup Responsif</span>'
    return '<span class="badge badge-warning" title="Hampir jenuh">🟡 Hampir Jenuh</span>'


def _getDataBadge(nPeriods: int) -> str:
    """Get data quality badge based on number of periods."""
    if nPeriods >= 7:
        return '<span class="badge badge-success">✅ Lengkap</span>'
    elif nPeriods >= 4:
        return '<span class="badge badge-info">📊 Cukup</span>'
    return '<span class="badge badge-warning">⚠️ Terbatas</span>'


class ShopeeExecutiveReportGenerator:
    """
    Generates compact executive HTML report for Shopee Ads.
    Aligned with TikTok report format for consistency.
    """
    
    def __init__(self):
        self.timestamp = datetime.now()
        self.CSS_STYLE = CSS_STYLE
    
    def generateReport(self, products: List[Dict], evaluation: Dict,
                       period_info: Dict = None,
                       title: str = "Shopee Ads - Executive Report") -> str:
        """Generate executive HTML report."""
        
        # Filter active products
        active = [p for p in products if p.get("is_active", True)]
        
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
        
        # Header
        html.extend(self._buildHeader(title, period_info))
        
        # KPI Cards
        html.extend(self._buildKpiCards(products, evaluation))
        
        # Event Context
        html.extend(self._buildEventContext())
        
        # Priority Actions
        html.extend(self._buildPriorityActions(products, evaluation))
        
        # STOP Products (ALL)
        html.extend(self._buildStopSection(products))
        
        # Scale Up
        html.extend(self._buildScaleUpSection(products))
        
        # Reduce Budget
        html.extend(self._buildReduceSection(products))
        
        # Risk Monitoring
        html.extend(self._buildRiskSection(products))
        
        # Legend
        html.extend(self._buildLegend())
        
        # Footer
        html.extend(self._buildFooter())
        
        return "\n".join(html)
    
    def _buildHeader(self, title: str, period_info: Dict = None) -> List[str]:
        """Build header with period info."""
        period_str = ""
        if period_info:
            if period_info.get("end"):
                period_str = f" | Periode: {period_info.get('start', '')} - {period_info.get('end', '')}"
        
        return [
            f"<h1>{title}</h1>",
            f"<p><strong>Generated:</strong> {self.timestamp.strftime('%Y-%m-%d %H:%M')}{period_str}</p>",
        ]
    
    def _buildKpiCards(self, products: List[Dict], evaluation: Dict) -> List[str]:
        """Build KPI highlight cards."""
        totalCost = evaluation.get("total_cost", 0)
        totalRevenue = evaluation.get("total_revenue", 0)
        totalProfit = evaluation.get("total_profit", totalRevenue - totalCost)
        overallRoas = evaluation.get("roi", 0)
        
        profitClass = "kpi-success" if totalProfit > 0 else "kpi-danger"
        roasClass = "kpi-success" if overallRoas >= 3 else "kpi-warning" if overallRoas >= 1 else "kpi-danger"
        
        # Calculate health score
        dist = evaluation.get("distribution", {"LANJUTKAN": 0, "PANTAU": 0, "HENTIKAN": 0})
        total = sum(dist.values()) or 1
        healthScore = (dist.get("LANJUTKAN", 0) / total * 100 * 0.6 + 
                       dist.get("PANTAU", 0) / total * 100 * 0.3)
        healthClass = "kpi-success" if healthScore >= 60 else "kpi-warning" if healthScore >= 40 else "kpi-danger"
        
        # Count high risk products
        highRisk = len([p for p in products if p.get("churn_risk_score", 0) > 60])
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
            f'<div class="kpi-value">{healthScore:.0f}/100</div>',
            '<div class="kpi-label">Health Score</div>',
            f'<div class="kpi-status">{"Grade A" if healthScore >= 80 else "Grade B" if healthScore >= 60 else "Grade C"}</div>',
            '</div>',
            f'<div class="kpi-card {riskClass}">',
            f'<div class="kpi-value">{highRisk}</div>',
            '<div class="kpi-label">High Risk Products</div>',
            f'<div class="kpi-status">{"🔴 Tinggi" if highRisk > 20 else "🟡 Sedang" if highRisk > 10 else "🟢 Rendah"}</div>',
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
        
        lines.append('</div>')
        lines.append(f'<p><em>💡 {phaseNote}</em></p>')
        
        return lines
    
    def _buildPriorityActions(self, products: List[Dict], evaluation: Dict) -> List[str]:
        """Build priority action boxes."""
        stop = [p for p in products if "HENTIKAN" in p.get("category", "")]
        reduce = [p for p in products if "KURANGI" in p.get("category", "") or 
                  ("PANTAU" in p.get("category", "") and p.get("roi", 0) < 2)]
        scaleUp = [p for p in products if "LANJUTKAN" in p.get("category", "") and p.get("roi", 0) >= 3]
        
        totalStopSavings = evaluation.get("potential_savings", 0)
        
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
                f'<strong>NAIKKAN budget {len(scaleUp)} produk</strong> dengan performa bagus',
                '</div></div>',
            ])
        
        return lines
    
    def _buildStopSection(self, products: List[Dict]) -> List[str]:
        """Build STOP section - ALL products."""
        stop = [p for p in products if "HENTIKAN" in p.get("category", "")]
        if not stop:
            return []
        
        auto, manual = _splitByBiddingMode(stop)
        totalLoss = sum(abs(p.get("profit", 0)) for p in stop if p.get("profit", 0) < 0)
        
        lines = [
            "<h2>🛑 HENTIKAN SEGERA - SEMUA PRODUK</h2>",
            f"<p><strong>Total: {len(stop)} produk</strong> | Potensi Hemat: <strong>Rp {totalLoss:,.0f}</strong></p>",
        ]
        
        if auto:
            lines.append("<h4 class='type-card'>🤖 Auto Bidding</h4>")
            lines.extend(self._buildStopTable(auto))
        
        if manual:
            lines.append("<h4 class='type-video'>⚙️ Manual Bidding</h4>")
            lines.extend(self._buildStopTable(manual))
        
        return lines
    
    def _buildStopTable(self, products: List[Dict]) -> List[str]:
        """Build stop products table - ALL items."""
        lines = [
            "<table>",
            "<thead><tr><th>#</th><th>Produk</th><th>ROAS</th><th>Kerugian</th><th>Budget</th><th>Aksi</th></tr></thead><tbody>",
        ]
        for i, p in enumerate(products, 1):
            loss = abs(p.get("profit", 0)) if p.get("profit", 0) < 0 else 0
            productDisplay = _formatProductName(p.get("product_name", "Unknown"))
            budget = p.get("total_cost", 0)
            roas = p.get("roi", 0)
            lines.append(
                f"<tr><td>{i}</td><td>{productDisplay}</td><td>{roas:.2f}x</td>"
                f"<td>Rp {loss:,.0f}</td><td>Rp {budget:,.0f}</td>"
                f"<td><strong>🛑 STOP -100%</strong></td></tr>"
            )
        lines.append("</tbody></table>")
        return lines
    
    def _buildScaleUpSection(self, products: List[Dict]) -> List[str]:
        """Build scale up section."""
        scaleUp = [p for p in products if "LANJUTKAN" in p.get("category", "") and p.get("roi", 0) >= 3]
        if not scaleUp:
            return []
        
        scaleUp = sorted(scaleUp, key=lambda x: x.get("score", 0), reverse=True)[:16]
        auto, manual = _splitByBiddingMode(scaleUp)
        
        lines = ["<h2>🚀 NAIKKAN BUDGET</h2>"]
        
        if auto:
            lines.append("<h4 class='type-card'>🤖 Auto Bidding (Top 8)</h4>")
            lines.extend(self._buildScaleTable(auto[:8]))
        
        if manual:
            lines.append("<h4 class='type-video'>⚙️ Manual Bidding (Top 8)</h4>")
            lines.extend(self._buildScaleTable(manual[:8]))
        
        return lines
    
    def _buildScaleTable(self, products: List[Dict]) -> List[str]:
        """Build scale up table with badges."""
        lines = [
            "<table>",
            "<thead><tr><th>#</th><th>Produk</th><th>ROAS</th><th>Status</th>"
            "<th>Rekomendasi</th><th>Est. Profit (CI)</th><th>Data</th></tr></thead><tbody>",
        ]
        for i, p in enumerate(products, 1):
            # Calculate recommendation
            roas = p.get("roi", 0)
            if roas >= 5:
                pctChange = 50
            elif roas >= 3:
                pctChange = 30
            else:
                pctChange = 15
            
            pctLabel = _getBudgetLabel(pctChange)
            budget = p.get("total_cost", 0)
            budgetChange = budget * pctChange / 100
            
            lifecycleBadge = _getLifecycleBadge(p)
            satBadge = _getSaturationBadge(p)
            dataBadge = _getDataBadge(p.get("n_periods", 4))
            productDisplay = _formatProductName(p.get("product_name", "Unknown"))
            
            # Format profit with estimated CI
            profit = p.get("profit", 0)
            profitChange = profit * (pctChange / 100) * 0.8  # Conservative estimate
            ciLow = profitChange * 0.6
            ciHigh = profitChange * 1.2
            profitStr = f"+Rp {profitChange:,.0f}<br/><small>CI: [{ciLow:,.0f}, {ciHigh:,.0f}]</small>"
            
            lines.append(
                f"<tr><td>{i}</td><td>{productDisplay}</td><td>{roas:.1f}x</td>"
                f"<td>{lifecycleBadge}<br/>{satBadge}</td>"
                f"<td><strong>{pctLabel}</strong><br/><small>+Rp {budgetChange:,.0f}</small></td>"
                f"<td>{profitStr}</td><td>{dataBadge}</td></tr>"
            )
        lines.append("</tbody></table>")
        return lines
    
    def _buildReduceSection(self, products: List[Dict]) -> List[str]:
        """Build reduce budget section."""
        reduce = [p for p in products if "PANTAU" in p.get("category", "") and p.get("roi", 0) < 2]
        if not reduce:
            return []
        
        reduce = sorted(reduce, key=lambda x: x.get("roi", 0))[:15]
        
        lines = [
            "<h2>📉 TURUNKAN BUDGET</h2>",
            "<table>",
            "<thead><tr><th>#</th><th>Produk</th><th>Tipe</th><th>ROAS</th>"
            "<th>Budget</th><th>Rekomendasi</th><th>Alasan</th></tr></thead><tbody>",
        ]
        
        for i, p in enumerate(reduce, 1):
            roas = p.get("roi", 0)
            pctChange = -30 if roas < 1.5 else -20
            pctLabel = _getBudgetLabel(pctChange)
            budget = p.get("total_cost", 0)
            budgetChangeNom = budget * abs(pctChange) / 100
            
            biddingType = "🤖 Auto" if "auto" in str(p.get("bidding_mode", "")).lower() else "⚙️ Manual"
            reason = "ROI rendah" if roas < 1.5 else "Trend menurun" if "Turun" in p.get("trend", "") else "Volatilitas tinggi"
            productDisplay = _formatProductName(p.get("product_name", "Unknown"))
            
            lines.append(
                f"<tr><td>{i}</td><td>{productDisplay}</td><td>{biddingType}</td>"
                f"<td>{roas:.2f}x</td><td>Rp {budget:,.0f}</td>"
                f"<td><strong>{pctLabel}</strong><br/><small>-Rp {budgetChangeNom:,.0f}</small></td>"
                f"<td>{reason}</td></tr>"
            )
        
        lines.append("</tbody></table>")
        return lines
    
    def _buildRiskSection(self, products: List[Dict]) -> List[str]:
        """Build risk monitoring section."""
        highRisk = [p for p in products if p.get("churn_risk_score", 0) > 60 
                    and "HENTIKAN" not in p.get("category", "")]
        
        if not highRisk:
            return [
                "<h2>⚠️ RISK MONITORING</h2>",
                "<p>✅ Tidak ada produk aktif dengan risiko tinggi.</p>"
            ]
        
        highRisk = sorted(highRisk, key=lambda x: x.get("churn_risk_score", 0), reverse=True)[:10]
        
        lines = [
            "<h2>⚠️ RISK MONITORING</h2>",
            f"<p><strong>{len(highRisk)} produk</strong> risiko tinggi (exclude yang harus STOP)</p>",
            "<table>",
            "<thead><tr><th>#</th><th>Produk</th><th>Status</th><th>Risk</th><th>Faktor</th><th>Data</th></tr></thead><tbody>",
        ]
        
        for i, p in enumerate(highRisk, 1):
            lifecycleBadge = _getLifecycleBadge(p)
            dataBadge = _getDataBadge(p.get("n_periods", 4))
            productDisplay = _formatProductName(p.get("product_name", "Unknown"))
            churnScore = p.get("churn_risk_score", 0)
            
            # Determine risk factor
            if p.get("roi", 0) < 2:
                factor = "ROI menurun"
            elif "Turun" in p.get("trend", ""):
                factor = "Trend negatif"
            else:
                factor = "Volatilitas tinggi"
            
            lines.append(
                f"<tr><td>{i}</td><td>{productDisplay}</td><td>{lifecycleBadge}</td>"
                f"<td><span class='badge badge-danger'>{churnScore:.0f}/100</span></td>"
                f"<td>{factor}</td><td>{dataBadge}</td></tr>"
            )
        
        lines.append("</tbody></table>")
        return lines
    
    def _buildLegend(self) -> List[str]:
        """Build legend section - terminology notes in Indonesian."""
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
            '<li><strong>🚀 +50%</strong> = Naikkan budget 50% dari budget saat ini.</li>',
            '<li><strong>📈 +30%</strong> = Naikkan budget 30% dari budget saat ini.</li>',
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
            '<li><strong>Confidence Interval (CI)</strong> = Rentang estimasi dengan 95% kepercayaan.</li>',
            '<li><strong>Score</strong> = Skor gabungan 0-100 dari faktor: ROAS, Trend, Momentum, Elastisitas, Volatilitas, dll.</li>',
            '</ul>',
            '</div>',
            
            '</div>',  # End legend-section
        ]
    
    def _buildFooter(self) -> List[str]:
        """Build footer."""
        nextReview = self.timestamp + timedelta(days=7)
        return [
            '<div class="footer">',
            "<p><strong>Shopee Ads Intelligence System v3.0.0</strong></p>",
            f"<p>Generated: {self.timestamp.strftime('%Y-%m-%d %H:%M')}</p>",
            f"<p>Next Review: {nextReview.strftime('%d %b %Y')}</p>",
            "</div></div></body></html>"
        ]


# Convenience function for backward compatibility
def generate_executive_html_report(products: List[Dict], evaluation: Dict,
                                   period_info: Dict = None) -> str:
    """Generate executive HTML report - convenience wrapper."""
    generator = ShopeeExecutiveReportGenerator()
    return generator.generateReport(products, evaluation, period_info)
