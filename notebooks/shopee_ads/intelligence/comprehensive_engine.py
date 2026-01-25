#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Comprehensive Analysis Engine
====================================
Main orchestrator combining ALL analysis modules into unified output.
Following AGENTS.MD: Clean Code, DRY, SRP, max 300 lines.

This engine combines:
- Core intelligence modules (shared with TikTok)
- Unified scoring (8 components)
- Budget optimization recommendations
- Churn risk analysis
- Indonesian event context
"""

import sys
from pathlib import Path
from dataclasses import dataclass, field
from typing import Dict, List, Optional
from datetime import datetime
import pandas as pd

# Add parent to path for core imports
_parent_dir = Path(__file__).resolve().parent.parent.parent
if str(_parent_dir) not in sys.path:
    sys.path.insert(0, str(_parent_dir))

# Import from core
from core.intelligence import (
    UnifiedScorer,
    UnifiedScoreResult,
    ActionRecommendation,
    calculate_budget_elasticity,
    calculate_marginal_roi,
    detect_creative_fatigue,
    calculate_fatigue_index,
    calculate_churn_risk,
    detect_saturation_point,
)
from core.statistics import (
    mann_kendall_test,
    calculate_momentum,
    coefficient_of_variation,
    TrendDirection,
    get_trend_direction,
)
from core.calendar import IndonesianCalendar

from .unified_scorer import get_action_simple


@dataclass
class ComprehensiveAnalysis:
    """Complete analysis result for a single product."""
    productId: str
    productName: str
    biddingMode: str
    isDiscontinued: bool
    
    # Basic metrics
    totalCost: float
    totalRevenue: float
    totalProfit: float
    totalOrders: int
    roas: float
    nPeriods: int
    
    # Unified Score
    unifiedScore: UnifiedScoreResult
    
    # Indonesian Context
    currentEventMultiplier: float
    currentEvents: List[str]
    
    # Legacy metrics
    legacyMkTrend: str
    legacyMomentumPct: float
    legacyCv: float
    legacyChurnRisk: float
    
    # Summary
    finalAction: ActionRecommendation
    actionLabel: str
    successProbability: float
    confidenceLevel: str
    summary: str


class ShopeeComprehensiveEngine:
    """
    Main engine orchestrating all analysis components for Shopee.
    Produces complete analysis with all metrics and recommendations.
    """
    
    def __init__(self):
        self.calendar = IndonesianCalendar()
        self.scorer = UnifiedScorer()
    
    def analyzeAll(self, df: pd.DataFrame, verbose: bool = True) -> List[ComprehensiveAnalysis]:
        """Analyze all products comprehensively."""
        if verbose:
            print("🔬 Running Comprehensive Analysis...")
        
        if df.empty:
            print("   ⚠️ No data found!")
            return []
        
        productIds = df["product_id"].unique()
        if verbose:
            print(f"   Found {len(productIds)} products")
        
        # Get current event context
        today = datetime.now()
        eventMult, eventNames = self.calendar.getCompositeMultiplier(today)
        
        if verbose and eventNames:
            print(f"   Current events: {eventNames}")
        
        results = []
        for pid in productIds:
            productDf = df[df["product_id"] == pid]
            analysis = self._analyzeProduct(productDf, pid, eventMult, eventNames)
            if analysis:
                results.append(analysis)
        
        # Sort by unified score
        results.sort(key=lambda x: x.unifiedScore.compositeScore, reverse=True)
        
        if verbose:
            print(f"   ✅ Analyzed {len(results)} products successfully")
        
        return results
    
    def _analyzeProduct(self, df: pd.DataFrame, productId: str,
                        eventMult: float, eventNames: List[str]) -> Optional[ComprehensiveAnalysis]:
        """Analyze a single product comprehensively."""
        try:
            # Basic metrics
            productName = df["product_name"].iloc[0][:55]
            biddingMode = df["bidding_mode"].iloc[0] if "bidding_mode" in df.columns else "Unknown"
            
            totalCost = df["cost"].sum()
            totalRevenue = df["revenue"].sum()
            totalProfit = totalRevenue - totalCost
            totalOrders = df["units_sold"].sum() if "units_sold" in df.columns else 0
            roas = totalRevenue / totalCost if totalCost > 0 else 0
            nPeriods = len(df)
            
            # Time series
            revenueSeries = df["revenue"].values
            costSeries = df["cost"].values
            roiSeries = (df["revenue"] / df["cost"].replace(0, 1)).values
            
            # Detect discontinued
            latestPeriod = df["period_label"].max() if "period_label" in df.columns else None
            isDiscontinued = False  # Will set based on data recency
            
            # Statistical analysis
            mk = mann_kendall_test(roiSeries)
            momentum = calculate_momentum(revenueSeries)
            cv = coefficient_of_variation(roiSeries)
            
            # Trend direction
            trendDir = get_trend_direction(revenueSeries)
            
            # Churn risk
            churnRisk = calculate_churn_risk(revenueSeries, roiSeries)
            
            # Elasticity (if enough data)
            elasticity = 0.0
            satStatus = ""
            if nPeriods >= 4:
                elast = calculate_budget_elasticity(df, "cost", "revenue", "period_start")
                elasticity = elast.elasticity if elast.isValid else 0
            
            # Fatigue
            fatigueIndex = calculate_fatigue_index(nPeriods, 30, 60)
            fatigueStatus = "FATIGUED" if fatigueIndex > 0.7 else ""
            
            # Calculate unified score
            unifiedScore = self.scorer.calculate(
                productId=str(productId),
                productName=productName,
                roas=roas,
                mkTrend=mk.get("trend", ""),
                mkPValue=mk.get("p_value", 1.0),
                trendDirection=trendDir,
                trendRoc=0.0,
                legacyMomentumPct=momentum.get("momentum_pct", 0) if momentum.get("is_valid") else 0,
                elasticity=elasticity,
                saturationStatus=satStatus,
                marginalRoas=0.0,
                cv=cv if isinstance(cv, (int, float)) else cv.get("cv", 0),
                intelVolatilityScore=50.0,
                fatigueStatus=fatigueStatus,
                fatiguePct=fatigueIndex * 100,
                churnRiskScore=churnRisk * 100 if isinstance(churnRisk, float) else 0,
                eventMultiplier=eventMult,
                eventNames=eventNames,
                nPeriods=nPeriods,
                dataQuality="HIGH" if nPeriods >= 8 else "MEDIUM" if nPeriods >= 4 else "LOW"
            )
            
            # Get simple action label
            actionLabel = get_action_simple(unifiedScore.action)
            
            return ComprehensiveAnalysis(
                productId=str(productId),
                productName=productName,
                biddingMode=biddingMode,
                isDiscontinued=isDiscontinued,
                totalCost=totalCost,
                totalRevenue=totalRevenue,
                totalProfit=totalProfit,
                totalOrders=int(totalOrders),
                roas=roas,
                nPeriods=nPeriods,
                unifiedScore=unifiedScore,
                currentEventMultiplier=eventMult,
                currentEvents=eventNames,
                legacyMkTrend=mk.get("trend", "Stabil"),
                legacyMomentumPct=momentum.get("momentum_pct", 0) if momentum.get("is_valid") else 0,
                legacyCv=cv if isinstance(cv, (int, float)) else cv.get("cv", 0),
                legacyChurnRisk=churnRisk if isinstance(churnRisk, float) else 0,
                finalAction=unifiedScore.action,
                actionLabel=actionLabel,
                successProbability=unifiedScore.confidence,
                confidenceLevel=unifiedScore.confidenceLevel,
                summary=self._generateSummary(unifiedScore, roas, totalProfit)
            )
        except Exception as e:
            print(f"   ⚠️ Error analyzing {productId}: {e}")
            return None
    
    def _generateSummary(self, score: UnifiedScoreResult, 
                         roas: float, profit: float) -> str:
        """Generate human-readable summary."""
        if score.compositeScore >= 70:
            return f"Performa sangat baik (Score: {score.compositeScore:.0f}). ROI {roas:.1f}x. Layak scale up."
        elif score.compositeScore >= 50:
            return f"Performa cukup baik (Score: {score.compositeScore:.0f}). ROI {roas:.1f}x. Monitor & optimasi."
        elif profit < 0:
            return f"Merugi Rp {abs(profit):,.0f}. Score: {score.compositeScore:.0f}. Pertimbangkan stop."
        else:
            return f"Performa kurang baik (Score: {score.compositeScore:.0f}). ROI {roas:.1f}x. Evaluasi segera."
