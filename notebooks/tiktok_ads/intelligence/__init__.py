#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Intelligence System
==============================
Version 3.0.0 - Unified Intelligence System

Comprehensive Ad Analysis with Multi-Layer Methodology.
Combines legacy statistical methods with modern ML techniques.
Generic logic is in `notebooks/core/`, TikTok-specific wrappers here.

Modules:
- config: Thresholds, weights, and constants
- calendar: Indonesian event detection
- analyzers: Various analysis methods
- scorers: Composite scoring system
- engine: Main orchestrator
- unified: Combined scoring from ALL methodologies
"""

__version__ = "3.0.0"
__author__ = "YN Digital"

# Core Config & Types
from .config_intelligence import IntelligenceConfig, RecommendationAction
from .indonesian_calendar import IndonesianCalendar, EventType
from .funnel_analyzer import FunnelAnalyzer
from .roas_classifier import RoasClassifier, RoasTier, TikTokRoasClassifier
from .trend_momentum import TrendMomentum, TrendDirection
from .volatility_analyzer import VolatilityAnalyzer
from .saturation_model import SaturationModel, SaturationStatus
from .budget_optimizer import BudgetOptimizer
from .fatigue_detector import FatigueDetector, FatigueStatus
from .product_lifecycle import ProductLifecycle
from .composite_scorer import CompositeScorer
from .probability_engine import ProbabilityEngine
from .data_aggregator import DataAggregator
from .product_resolver import ProductResolver
from .intelligence_engine import IntelligenceEngine, ProductAnalysis

# ============================================
# UNIFIED INTELLIGENCE SYSTEM (v3.0.0)
# ============================================

# Unified Scorer - combines ALL scoring components
from .unified_scorer import (
    UnifiedScorer,
    UnifiedScore,
    ActionRecommendation
)

# Enhanced Budget - precise recommendations with CI
from .enhanced_budget import (
    EnhancedBudgetRecommender,
    BudgetRecommendation
)

# Probability with CI - statistical probability estimates
from .probability_ci import (
    ProbabilityWithCI,
    ProbabilityEstimate
)

# Comprehensive Engine - main orchestrator
from .comprehensive_engine import (
    ComprehensiveAnalysisEngine,
    ComprehensiveAnalysis
)

# Report Generator (Markdown) - now in reports/
from .reports.comprehensive_report import (
    ComprehensiveReportGenerator,
    ReportSummary
)

# HTML Report Generator (Full Detail) - now in reports/
from .reports.html_report_generator import HtmlReportGenerator

# Executive HTML Report - now in reports/
from .reports.executive_report_generator import ExecutiveReportGenerator


# ============================================
# QUICK START FUNCTIONS
# ============================================

def runComprehensiveAnalysis(dbPath: str, verbose: bool = True, outputFormat: str = "html"):
    """
    Quick start function for comprehensive analysis.
    
    Args:
        dbPath: Path to SQLite database
        verbose: Print progress info
        outputFormat: 'html' for HTML reports, 'md' for Markdown
        
    Returns:
        tuple: (analyses, fullReport, executiveReport)
    
    Example:
        analyses, full, exec = runComprehensiveAnalysis('path/to/db.db')
        with open('report.html', 'w', encoding='utf-8') as f:
            f.write(full)
    """
    engine = ComprehensiveAnalysisEngine(dbPath, verbose=verbose)
    analyses = engine.analyzeAll()
    
    if not analyses:
        return [], "", ""
    
    if outputFormat == "html":
        reporter = HtmlReportGenerator()
        fullReport = reporter.generateFullReport(analyses)
        execReporter = ExecutiveReportGenerator()
        executiveReport = execReporter.generateReport(analyses)
    else:
        reporter = ComprehensiveReportGenerator()
        fullReport = reporter.generateFullReport(analyses)
        executiveReport = reporter.generateExecutiveReport(analyses)
    
    return analyses, fullReport, executiveReport


def getTopProducts(dbPath: str, limit: int = 10):
    """Get top products by unified score."""
    engine = ComprehensiveAnalysisEngine(dbPath, verbose=False)
    analyses = engine.analyzeAll()
    analyses = sorted(analyses, key=lambda x: x.unifiedScore.compositeScore, reverse=True)
    return analyses[:limit]


def getStopProducts(dbPath: str):
    """Get products recommended to stop."""
    engine = ComprehensiveAnalysisEngine(dbPath, verbose=False)
    analyses = engine.analyzeAll()
    return [a for a in analyses if a.finalAction == ActionRecommendation.STOP]


# ============================================
# ALL EXPORTS
# ============================================

__all__ = [
    # Version
    '__version__',
    
    # Config
    'IntelligenceConfig',
    'RecommendationAction',
    
    # Analyzers
    'IndonesianCalendar',
    'EventType',
    'FunnelAnalyzer',
    'RoasClassifier',
    'RoasTier',
    'TikTokRoasClassifier',
    'TrendMomentum',
    'TrendDirection',
    'VolatilityAnalyzer',
    'SaturationModel',
    'SaturationStatus',
    'BudgetOptimizer',
    'FatigueDetector',
    'FatigueStatus',
    'ProductLifecycle',
    'CompositeScorer',
    'ProbabilityEngine',
    
    # Data
    'DataAggregator',
    'ProductResolver',
    
    # Engine
    'IntelligenceEngine',
    'ProductAnalysis',
    
    # ===== UNIFIED SYSTEM v3.0.0 =====
    'UnifiedScorer',
    'UnifiedScore',
    'ActionRecommendation',
    
    'EnhancedBudgetRecommender',
    'BudgetRecommendation',
    
    'ProbabilityWithCI',
    'ProbabilityEstimate',
    
    'ComprehensiveAnalysisEngine',
    'ComprehensiveAnalysis',
    
    'ComprehensiveReportGenerator',
    'ReportSummary',
    
    # HTML Reports
    'HtmlReportGenerator',
    'PortfolioHealth',
    'FinancialProjection',
    'ExecutiveReportGenerator',
    
    # Quick Functions
    'runComprehensiveAnalysis',
    'getTopProducts',
    'getStopProducts',
]
