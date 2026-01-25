#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Comprehensive Report Generator
==============================
Generates Full and Executive reports from ComprehensiveAnalysis.
Following AGENTS.MD: Clean Code, DRY, SRP, max 300 lines.
"""

from dataclasses import dataclass
from typing import Dict, List, Optional
from datetime import datetime
import os

from ..comprehensive_engine import ComprehensiveAnalysis, ComprehensiveAnalysisEngine
from ..unified_scorer import ActionRecommendation


@dataclass 
class ReportSummary:
    """Summary statistics for report header."""
    totalProducts: int
    totalCost: float
    totalRevenue: float
    totalProfit: float
    overallRoas: float
    avgSuccessProbability: float
    
    scaleUpCount: int
    maintainCount: int
    reduceCount: int
    stopCount: int
    
    totalBudgetChange: float
    expectedRevenueChange: float


class ComprehensiveReportGenerator:
    """
    Generates comprehensive reports from analysis results.
    
    Output formats:
    - Full Report (detailed, all products)
    - Executive Report (summary, top actions)
    """
    
    def __init__(self, outputDir: str = "output"):
        self.outputDir = outputDir
        os.makedirs(outputDir, exist_ok=True)
    
    def generateFullReport(self, analyses: List[ComprehensiveAnalysis],
                            title: str = "TikTok Ads Comprehensive Report"
                            ) -> str:
        """Generate full detailed report in Markdown."""
        
        summary = self._calculateSummary(analyses)
        
        report = []
        report.append(f"# {title}")
        report.append(f"\n**Generated:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        report.append(f"\n**Analysis Type:** Comprehensive (Legacy + Intelligence Combined)\n")
        
        # Executive Summary
        report.extend(self._buildExecutiveSummary(summary, analyses))
        
        # Action Distribution
        report.extend(self._buildActionDistribution(summary))
        
        # Event Context
        report.extend(self._buildEventContext(analyses))
        
        # Products to Scale Up
        report.extend(self._buildScaleUpSection(analyses))
        
        # Products to Maintain
        report.extend(self._buildMaintainSection(analyses))
        
        # Products to Reduce/Stop
        report.extend(self._buildStopSection(analyses))
        
        # Detailed Analysis Table
        report.extend(self._buildDetailedTable(analyses))
        
        # Methodology
        report.extend(self._buildMethodology())
        
        return "\n".join(report)
    
    def generateExecutiveReport(self, analyses: List[ComprehensiveAnalysis]
                                 ) -> str:
        """Generate executive summary report."""
        
        summary = self._calculateSummary(analyses)
        
        report = []
        report.append("# 📊 EXECUTIVE SUMMARY - TikTok Ads")
        report.append(f"\n**Date:** {datetime.now().strftime('%Y-%m-%d')}")
        report.append(f"\n**Period:** Comprehensive Analysis\n")
        
        # Key Metrics Box
        report.append("---")
        report.append("## 🎯 KEY METRICS")
        report.append(f"| Metric | Value |")
        report.append(f"|--------|-------|")
        report.append(f"| Total Revenue | Rp {summary.totalRevenue:,.0f} |")
        report.append(f"| Total Profit | Rp {summary.totalProfit:,.0f} |")
        report.append(f"| Overall ROAS | {summary.overallRoas:.2f}x |")
        report.append(f"| Avg Success Probability | {summary.avgSuccessProbability:.1f}% |")
        report.append("")
        
        # Immediate Actions
        report.append("---")
        report.append("## 🚀 IMMEDIATE ACTIONS")
        
        # Scale Up
        scaleUp = [a for a in analyses if a.finalAction in [
            ActionRecommendation.SCALE_UP_AGGRESSIVE, 
            ActionRecommendation.SCALE_UP
        ]]
        scaleUp = sorted(scaleUp, key=lambda x: x.unifiedScore.compositeScore, reverse=True)[:5]
        
        if scaleUp:
            report.append("\n### ✅ NAIKKAN BUDGET (Top 5)")
            for i, a in enumerate(scaleUp, 1):
                report.append(f"\n**{i}. {a.productName}**")
                report.append(f"   - Action: {a.finalAction.value}")
                report.append(f"   - Budget: Rp {a.budgetRec.currentBudget:,.0f} → Rp {a.budgetRec.recommendedBudget:,.0f} ({a.finalBudgetChange:+.0f}%)")
                report.append(f"   - Expected Revenue: Rp {a.budgetRec.expectedRevenue:,.0f}")
                report.append(f"   - Success Probability: {a.successProbability:.1f}% ± {a.probability.uncertaintyPct:.1f}%")
                report.append(f"   - Confidence: {a.confidenceLevel}")
        
        # Stop
        stop = [a for a in analyses if a.finalAction == ActionRecommendation.STOP]
        if stop:
            totalLoss = sum(abs(a.totalProfit) for a in stop if a.totalProfit < 0)
            report.append(f"\n### 🛑 HENTIKAN SEGERA ({len(stop)} produk)")
            report.append(f"**Potensi penghematan: Rp {totalLoss:,.0f}**")
            for a in stop[:5]:
                report.append(f"   - {a.productName}: ROAS {a.roas:.2f}x, Loss Rp {abs(a.totalProfit):,.0f}")
        
        # Event Alert
        if analyses and analyses[0].currentEvents:
            report.append("\n### 📅 EVENT CONTEXT")
            report.append(f"**Current Events:** {', '.join(analyses[0].currentEvents)}")
            report.append(f"**Multiplier:** {analyses[0].currentEventMultiplier:.2f}x")
            if analyses[0].nextEventName:
                report.append(f"**Next Event:** {analyses[0].nextEventName} ({analyses[0].nextEventDate})")
        
        # Budget Summary
        report.append("\n---")
        report.append("## 💰 BUDGET RECOMMENDATION SUMMARY")
        report.append(f"| Category | Count | Budget Change |")
        report.append(f"|----------|-------|---------------|")
        report.append(f"| Scale Up | {summary.scaleUpCount} | +Rp {sum(a.budgetRec.budgetChange for a in analyses if a.budgetRec.budgetChange > 0):,.0f} |")
        report.append(f"| Maintain | {summary.maintainCount} | Rp 0 |")
        report.append(f"| Reduce | {summary.reduceCount} | -Rp {abs(sum(a.budgetRec.budgetChange for a in analyses if -0.5 < a.finalBudgetChange/100 < 0)):,.0f} |")
        report.append(f"| Stop | {summary.stopCount} | -Rp {abs(sum(a.budgetRec.budgetChange for a in analyses if a.finalAction == ActionRecommendation.STOP)):,.0f} |")
        report.append(f"| **NET** | {summary.totalProducts} | **{'+' if summary.totalBudgetChange >= 0 else ''}Rp {summary.totalBudgetChange:,.0f}** |")
        
        return "\n".join(report)
    
    def _calculateSummary(self, analyses: List[ComprehensiveAnalysis]) -> ReportSummary:
        """Calculate summary statistics."""
        if not analyses:
            return ReportSummary(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
        
        totalCost = sum(a.totalCost for a in analyses)
        totalRevenue = sum(a.totalRevenue for a in analyses)
        
        scaleUp = len([a for a in analyses if a.finalAction in [
            ActionRecommendation.SCALE_UP_AGGRESSIVE, ActionRecommendation.SCALE_UP
        ]])
        maintain = len([a for a in analyses if a.finalAction == ActionRecommendation.MAINTAIN])
        reduce = len([a for a in analyses if a.finalAction == ActionRecommendation.REDUCE])
        stop = len([a for a in analyses if a.finalAction == ActionRecommendation.STOP])
        
        return ReportSummary(
            totalProducts=len(analyses),
            totalCost=totalCost,
            totalRevenue=totalRevenue,
            totalProfit=totalRevenue - totalCost,
            overallRoas=totalRevenue / totalCost if totalCost > 0 else 0,
            avgSuccessProbability=sum(a.successProbability for a in analyses) / len(analyses),
            scaleUpCount=scaleUp,
            maintainCount=maintain,
            reduceCount=reduce,
            stopCount=stop,
            totalBudgetChange=sum(a.budgetRec.budgetChange for a in analyses),
            expectedRevenueChange=sum(a.budgetRec.revenueChange for a in analyses)
        )
    
    def _buildExecutiveSummary(self, summary: ReportSummary, 
                                analyses: List[ComprehensiveAnalysis]) -> List[str]:
        """Build executive summary section."""
        lines = [
            "---",
            "## 📊 EXECUTIVE SUMMARY",
            "",
            f"| Metric | Value |",
            f"|--------|-------|",
            f"| Total Products | {summary.totalProducts} |",
            f"| Total Cost | Rp {summary.totalCost:,.0f} |",
            f"| Total Revenue | Rp {summary.totalRevenue:,.0f} |",
            f"| Total Profit | Rp {summary.totalProfit:,.0f} |",
            f"| Overall ROAS | {summary.overallRoas:.2f}x |",
            f"| Avg Success Probability | {summary.avgSuccessProbability:.1f}% |",
            ""
        ]
        return lines
    
    def _buildActionDistribution(self, summary: ReportSummary) -> List[str]:
        """Build action distribution section."""
        return [
            "---",
            "## 📌 ACTION DISTRIBUTION",
            "",
            f"| Action | Count | % |",
            f"|--------|-------|---|",
            f"| 🚀 Scale Up | {summary.scaleUpCount} | {summary.scaleUpCount/summary.totalProducts*100:.0f}% |",
            f"| ✅ Maintain | {summary.maintainCount} | {summary.maintainCount/summary.totalProducts*100:.0f}% |",
            f"| ⬇️ Reduce | {summary.reduceCount} | {summary.reduceCount/summary.totalProducts*100:.0f}% |",
            f"| 🛑 Stop | {summary.stopCount} | {summary.stopCount/summary.totalProducts*100:.0f}% |",
            ""
        ]
    
    def _buildEventContext(self, analyses: List[ComprehensiveAnalysis]) -> List[str]:
        """Build event context section."""
        if not analyses:
            return []
        
        a = analyses[0]
        lines = [
            "---",
            "## 📅 INDONESIAN EVENT CONTEXT",
            "",
            f"**Current Date:** {datetime.now().strftime('%Y-%m-%d')}",
            f"**Active Events:** {', '.join(a.currentEvents) if a.currentEvents else 'None'}",
            f"**Event Multiplier:** {a.currentEventMultiplier:.2f}x",
        ]
        
        if a.nextEventName:
            lines.append(f"**Next Event:** {a.nextEventName} on {a.nextEventDate}")
        
        lines.append("")
        return lines
    
    def _buildScaleUpSection(self, analyses: List[ComprehensiveAnalysis]) -> List[str]:
        """Build scale up products section."""
        scaleUp = [a for a in analyses if a.finalAction in [
            ActionRecommendation.SCALE_UP_AGGRESSIVE, ActionRecommendation.SCALE_UP
        ]]
        scaleUp = sorted(scaleUp, key=lambda x: x.unifiedScore.compositeScore, reverse=True)
        
        lines = ["---", "## 🚀 PRODUCTS TO SCALE UP", ""]
        
        if not scaleUp:
            lines.append("*No products recommended for scale up*")
            return lines
        
        for i, a in enumerate(scaleUp[:10], 1):
            ci = a.budgetRec.revenueCI
            lines.extend([
                f"### {i}. {a.productName}",
                f"**Score:** {a.unifiedScore.compositeScore:.0f} ({a.unifiedScore.category})",
                f"**Action:** {a.finalAction.value}",
                f"**Current:** Rp {a.budgetRec.currentBudget:,.0f} | ROAS {a.roas:.2f}x",
                f"**Recommended:** Rp {a.budgetRec.recommendedBudget:,.0f} ({a.finalBudgetChange:+.0f}%)",
                f"**Expected Revenue:** Rp {a.budgetRec.expectedRevenue:,.0f} (CI: Rp {ci[0]:,.0f} - {ci[1]:,.0f})",
                f"**Success Probability:** {a.successProbability:.1f}% ({a.confidenceLevel})",
                f"**Rationale:** {'; '.join(a.budgetRec.reasons[:3])}",
                ""
            ])
        
        return lines
    
    def _buildMaintainSection(self, analyses: List[ComprehensiveAnalysis]) -> List[str]:
        """Build maintain products section."""
        maintain = [a for a in analyses if a.finalAction == ActionRecommendation.MAINTAIN]
        maintain = sorted(maintain, key=lambda x: x.unifiedScore.compositeScore, reverse=True)
        
        lines = ["---", "## ✅ PRODUCTS TO MAINTAIN", ""]
        
        if not maintain:
            lines.append("*No products in maintain category*")
            return lines
        
        lines.append(f"| Product | Score | ROAS | Probability | Status |")
        lines.append(f"|---------|-------|------|-------------|--------|")
        
        for a in maintain[:15]:
            lines.append(f"| {a.productName[:40]} | {a.unifiedScore.compositeScore:.0f} | {a.roas:.2f}x | {a.successProbability:.0f}% | {a.unifiedScore.category} |")
        
        lines.append("")
        return lines
    
    def _buildStopSection(self, analyses: List[ComprehensiveAnalysis]) -> List[str]:
        """Build stop products section."""
        stop = [a for a in analyses if a.finalAction in [
            ActionRecommendation.STOP, ActionRecommendation.REDUCE
        ]]
        stop = sorted(stop, key=lambda x: x.totalProfit)
        
        lines = ["---", "## 🛑 PRODUCTS TO REDUCE/STOP", ""]
        
        if not stop:
            lines.append("*No products recommended for reduction*")
            return lines
        
        totalLoss = sum(abs(a.totalProfit) for a in stop if a.totalProfit < 0)
        lines.append(f"**Total Potential Savings:** Rp {totalLoss:,.0f}")
        lines.append("")
        
        for i, a in enumerate(stop[:10], 1):
            lines.extend([
                f"### {i}. {a.productName}",
                f"**Action:** {a.finalAction.value}",
                f"**ROAS:** {a.roas:.2f}x | **Profit:** Rp {a.totalProfit:,.0f}",
                f"**Warnings:** {'; '.join(a.unifiedScore.warningFactors[:3]) if a.unifiedScore.warningFactors else 'None'}",
                ""
            ])
        
        return lines
    
    def _buildDetailedTable(self, analyses: List[ComprehensiveAnalysis]) -> List[str]:
        """Build detailed analysis table."""
        analyses = sorted(analyses, key=lambda x: x.unifiedScore.compositeScore, reverse=True)
        
        lines = ["---", "## 📋 ALL PRODUCTS DETAIL", ""]
        lines.append(f"| # | Product | Score | Action | Budget Change | Probability | ROAS |")
        lines.append(f"|---|---------|-------|--------|---------------|-------------|------|")
        
        for i, a in enumerate(analyses, 1):
            lines.append(f"| {i} | {a.productName[:35]} | {a.unifiedScore.compositeScore:.0f} | {a.finalAction.value} | {a.finalBudgetChange:+.0f}% | {a.successProbability:.0f}% | {a.roas:.2f}x |")
        
        lines.append("")
        return lines
    
    def _buildMethodology(self) -> List[str]:
        """Build methodology section."""
        return [
            "---",
            "## 📚 METHODOLOGY",
            "",
            "### Unified Scoring Components:",
            "- ROAS Score (20%): Profitability measurement",
            "- Trend Score (15%): Mann-Kendall + MACD analysis", 
            "- Momentum Score (10%): Recent vs historical performance",
            "- Elasticity Score (15%): Budget response + saturation",
            "- Volatility Score (10%): Stability measurement (CV)",
            "- Fatigue Score (10%): Creative fatigue detection",
            "- Churn Risk Score (10%): Decline risk assessment",
            "- Event Score (10%): Indonesian context (payday, twin dates, holidays)",
            "",
            "### Confidence Intervals:",
            "- 95% CI using Bootstrap method",
            "- Wilson score interval for probability estimates",
            "",
            "### Budget Recommendations:",
            "- Max budget cap: Rp 2,000,000 per product",
            "- Elasticity-adjusted scaling",
            "- Event multiplier consideration",
            "",
            f"**Report Generated:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
        ]
