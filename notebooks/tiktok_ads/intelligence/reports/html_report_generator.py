#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
HTML Report Generator
=====================
Generates professional HTML reports matching legacy format.
Combines ALL methodologies for comprehensive analysis.

Refactored for AGENTS.MD compliance - split into modules:
- html_styles.py: CSS styles
- portfolio_metrics.py: Health score & projection calculations
- report_sections_core.py: Executive summary & core sections
- report_sections_reco.py: Recommendations & churn sections
- report_sections_stats.py: Statistical analysis sections
"""

from typing import List
from datetime import datetime

from ..comprehensive_engine import ComprehensiveAnalysis
from .html_styles import CSS_STYLE
from ..portfolio_metrics import (
    PortfolioHealth, FinancialProjection,
    calculateHealthScore, calculateProjection
)
from .report_sections_core import (
    buildExecutiveSummary, buildHealthScoreSection, buildFinancialProjection
)
from .report_sections_reco import (
    buildRecommendations, buildChurnRiskSection
)
from .report_sections_stats import (
    buildCreativeComparison, buildCorrelationSection,
    buildBudgetOptimization, buildMethodology
)

# Re-export for backwards compatibility
__all__ = [
    'HtmlReportGenerator',
    'PortfolioHealth',
    'FinancialProjection'
]


class HtmlReportGenerator:
    """
    Generates comprehensive HTML reports combining all methodologies.
    Matches the detail level of legacy reports.
    
    Usage:
        generator = HtmlReportGenerator()
        html = generator.generateFullReport(analyses, "My Report Title")
    """
    
    def __init__(self):
        self.timestamp = datetime.now()
    
    def generateFullReport(self, analyses: List[ComprehensiveAnalysis],
                           title: str = "TikTok Ads Unified Intelligence Report"
                           ) -> str:
        """
        Generate comprehensive HTML report.
        
        Args:
            analyses: List of ComprehensiveAnalysis results
            title: Report title
            
        Returns:
            Complete HTML string
        """
        # Calculate portfolio metrics
        health = calculateHealthScore(analyses)
        projection = calculateProjection(analyses)
        
        html = self._buildHeader(title)
        
        # Executive Summary
        html.extend(buildExecutiveSummary(analyses, health))
        
        # Health Score Details
        html.extend(buildHealthScoreSection(health, analyses))
        
        # Financial Projection
        html.extend(buildFinancialProjection(projection, analyses))
        
        # Recommendations by Action
        html.extend(buildRecommendations(analyses))
        
        # Churn Risk Analysis
        html.extend(buildChurnRiskSection(analyses))
        
        # Creative Type Comparison
        html.extend(buildCreativeComparison(analyses))
        
        # Correlation & Efficiency
        html.extend(buildCorrelationSection(analyses))
        
        # Budget Optimization
        html.extend(buildBudgetOptimization(analyses))
        
        # Methodology
        html.extend(buildMethodology())
        
        # Footer
        html.extend(self._buildFooter())
        
        return "\n".join(html)
    
    def _buildHeader(self, title: str) -> List[str]:
        """Build HTML header and opening tags."""
        return [
            "<!DOCTYPE html>",
            '<html lang="id">',
            "<head>",
            '<meta charset="UTF-8">',
            '<meta name="viewport" content="width=device-width, initial-scale=1.0">',
            f"<title>{title}</title>",
            CSS_STYLE,
            "</head>",
            "<body>",
            '<div class="container">',
            f"<h1>{title}</h1>",
            f"<p><strong>Generated:</strong> {self.timestamp.strftime('%Y-%m-%d %H:%M:%S')}</p>",
            f"<p><strong>System:</strong> Unified Intelligence System v3.0.0</p>",
        ]
    
    def _buildFooter(self) -> List[str]:
        """Build HTML footer and closing tags."""
        return [
            '<div class="footer">',
            "<p><strong>Unified Intelligence System v3.0.0</strong></p>",
            f"<p>Generated: {self.timestamp.strftime('%Y-%m-%d %H:%M:%S')}</p>",
            "<p>Methodology: Mann-Kendall • MACD Trend • Momentum 70/30 • Bootstrap CI • Saturation Model • Fatigue Detection • Indonesian Calendar</p>",
            "</div>",
            "</div>",
            "</body>",
            "</html>"
        ]
