#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Comprehensive Analysis Engine
=============================
Main orchestrator that combines ALL analysis modules into unified output.
Following AGENTS.MD: Clean Code, DRY, SRP, max 300 lines.

This engine combines:
- Legacy tiktok_ads analysis
- Intelligence System analysis
- Unified scoring
- Enhanced budget recommendations
- Probability with CI
- Indonesian event context
"""

from dataclasses import dataclass, field
from typing import Dict, List, Optional, Tuple
from datetime import datetime
import pandas as pd
import sqlite3
from pathlib import Path

# Import all components
from .unified_scorer import UnifiedScorer, UnifiedScore, ActionRecommendation
from .enhanced_budget import EnhancedBudgetRecommender, BudgetRecommendation
from .probability_ci import ProbabilityWithCI, ProbabilityEstimate
from .indonesian_calendar import IndonesianCalendar
from .roas_classifier import RoasClassifier
from .trend_momentum import TrendMomentum
from .volatility_analyzer import VolatilityAnalyzer
from .saturation_model import SaturationModel
from .fatigue_detector import FatigueDetector
from .product_lifecycle import ProductLifecycle
from .config_intelligence import IntelligenceConfig
from .product_resolver import ProductResolver

# Import legacy components
import sys
sys.path.insert(0, str(Path(__file__).parent.parent))
from ..stats import mann_kendall_test, calculate_momentum, coefficient_of_variation
from ..correlation import calculate_churn_risk
from ..elasticity import calculate_budget_elasticity, calculate_marginal_roi


@dataclass
class ComprehensiveAnalysis:
    """Complete analysis result for a single product."""
    # Identity
    productId: str
    productName: str
    creativeType: str
    isDiscontinued: bool
    
    # Basic metrics
    totalCost: float
    totalRevenue: float
    totalProfit: float
    totalOrders: int
    roas: float
    nPeriods: int
    
    # Unified Score
    unifiedScore: UnifiedScore
    
    # Budget Recommendation
    budgetRec: BudgetRecommendation
    
    # Probability Estimate
    probability: ProbabilityEstimate
    
    # Indonesian Context
    currentEventMultiplier: float
    currentEvents: List[str]
    nextEventDate: Optional[str]
    nextEventName: Optional[str]
    
    # Legacy metrics (for comparison)
    legacyMkTrend: str
    legacyMomentumPct: float
    legacyCv: float
    legacyChurnRisk: float
    
    # Summary
    finalAction: ActionRecommendation
    finalBudgetChange: float
    finalBudgetNominal: float
    successProbability: float
    confidenceLevel: str
    summary: str


class ComprehensiveAnalysisEngine:
    """
    Main engine that orchestrates all analysis components.
    Produces complete analysis with all metrics, recommendations, and probabilities.
    """
    
    def __init__(self, dbPath: str, tenantId: str = "yumna_bertigamart"):
        self.dbPath = dbPath
        self.tenantId = tenantId
        self.config = IntelligenceConfig()
        
        # Initialize all analyzers
        self.calendar = IndonesianCalendar()
        self.scorer = UnifiedScorer()
        self.budgetRecommender = EnhancedBudgetRecommender()
        self.probabilityEngine = ProbabilityWithCI()
        self.roasClassifier = RoasClassifier(self.config)
        self.trendAnalyzer = TrendMomentum()
        self.volatilityAnalyzer = VolatilityAnalyzer(self.config)
        self.saturationModel = SaturationModel(self.config.budget.maxBudgetPerProduct)
        self.fatigueDetector = FatigueDetector(self.config)
        self.lifecycleAnalyzer = ProductLifecycle(self.config)
        
        # Initialize ProductResolver with Excel reference
        excelPath = str(Path(__file__).parent.parent / "ads_data" / "Tiktoksellercenter_batchedit_20260114_basic_information_template.xlsx")
        self.productResolver = ProductResolver(excelPath=excelPath)
        self.productResolver.load()  # Pre-load product names
    
    def analyzeAll(self, minCost: float = 50000,
                   verbose: bool = True) -> List[ComprehensiveAnalysis]:
        """Analyze all products comprehensively."""
        if verbose:
            print("🔬 Running Comprehensive Analysis...")
            print("   Loading data...")
        
        # Load raw data
        rawDf = self._loadData(minCost)
        
        if rawDf.empty:
            print("   ⚠️ No data found!")
            return []
        
        productIds = rawDf['productId'].unique()
        if verbose:
            print(f"   Found {len(productIds)} products")
        
        # Get current event context
        today = datetime.now()
        eventMult, eventNames = self.calendar.getCompositeMultiplier(today)
        nextEvent = self._getNextBigEvent(today)
        
        if verbose:
            print(f"   Current events: {eventNames if eventNames else 'None'}")
        
        results = []
        for i, pid in enumerate(productIds):
            if verbose and (i + 1) % 20 == 0:
                print(f"   Processing {i + 1}/{len(productIds)}...")
            
            analysis = self._analyzeProduct(pid, rawDf, eventMult, eventNames, nextEvent)
            if analysis:
                results.append(analysis)
        
        if verbose:
            print(f"✅ Analyzed {len(results)} products")
        
        return results
    
    def _getNextBigEvent(self, fromDate: datetime) -> Optional[Tuple[str, str]]:
        """Get next big event (twin date or holiday) from calendar."""
        try:
            # Check next 30 days for twin dates
            for dayOffset in range(1, 31):
                checkDate = fromDate + pd.Timedelta(days=dayOffset)
                day = checkDate.day
                month = checkDate.month
                dateObj = checkDate.date()
                
                # Twin date
                if day == month:
                    return (checkDate.strftime('%Y-%m-%d'), f"Tanggal Kembar {month}.{day}")
                
                # Holiday
                if dateObj in self.calendar.NATIONAL_HOLIDAYS:
                    return (checkDate.strftime('%Y-%m-%d'), self.calendar.NATIONAL_HOLIDAYS[dateObj])
            
            return None
        except Exception:
            return None
    
    def _loadData(self, minCost: float) -> pd.DataFrame:
        """Load data from database."""
        conn = sqlite3.connect(self.dbPath)
        
        query = """
            SELECT 
                productId, campaignName, creativeType, videoTitle,
                periodStart, periodEnd, periodLabel,
                cost, grossRevenue, ordersSku, roi,
                impressions, clicks, ctr, conversionRate
            FROM TiktokAdsCreativeData 
            WHERE tenantId = ? AND cost > 0
            ORDER BY periodStart
        """
        
        df = pd.read_sql_query(query, conn, params=(self.tenantId,))
        conn.close()
        
        # Filter by min cost
        productCosts = df.groupby('productId')['cost'].sum()
        validProducts = productCosts[productCosts >= minCost].index
        
        return df[df['productId'].isin(validProducts)]
    
    def _analyzeProduct(self, productId: str, rawDf: pd.DataFrame,
                        eventMult: float, eventNames: List[str],
                        nextEvent: Optional[Dict]) -> Optional[ComprehensiveAnalysis]:
        """Analyze a single product comprehensively."""
        productDf = rawDf[rawDf['productId'] == productId].copy()
        
        if productDf.empty:
            return None
        
        # Basic metrics
        totalCost = productDf['cost'].sum()
        totalRevenue = productDf['grossRevenue'].sum()
        totalProfit = totalRevenue - totalCost
        totalOrders = productDf['ordersSku'].sum()
        roas = totalRevenue / totalCost if totalCost > 0 else 0
        nPeriods = len(productDf)
        
        productName = self._getProductName(productId, productDf)
        creativeType = productDf['creativeType'].iloc[0]
        
        # Time series for analysis
        productDf = productDf.sort_values('periodStart')
        roiSeries = pd.Series(productDf['roi'].values)
        costSeries = pd.Series(productDf['cost'].values)
        revenueSeries = pd.Series(productDf['grossRevenue'].values)
        
        # === Legacy Analysis ===
        mkResult = mann_kendall_test(roiSeries)
        momentumResult = calculate_momentum(roiSeries) if len(roiSeries) >= 3 else {'change_pct': 0}
        momentumPct = momentumResult.get('change_pct', 0) if isinstance(momentumResult, dict) else (momentumResult or 0)
        cvResult = coefficient_of_variation(roiSeries) if len(roiSeries) >= 2 else {'cv': 50}
        cv = cvResult.get('cv', 50) if isinstance(cvResult, dict) else (cvResult or 50)
        
        # Ensure numeric values
        momentumPct = float(momentumPct) if momentumPct is not None else 0.0
        cv = float(cv) if cv is not None else 50.0
        
        # Churn risk (simplified)
        churnRisk = self._calculateSimpleChurnRisk(mkResult, momentumPct, cv)
        
        # Elasticity
        elasticityResult = calculate_budget_elasticity(productDf)
        elasticity = elasticityResult.get('elasticity', 1.0) if isinstance(elasticityResult, dict) else 1.0
        mRoi = calculate_marginal_roi(productDf)
        
        # === Intelligence Analysis ===
        trendResult = self.trendAnalyzer.analyze(roiSeries)
        volatilityResult = self.volatilityAnalyzer.analyze(roiSeries)
        saturationResult = self.saturationModel.analyze(costSeries, revenueSeries)
        fatigueResult = self.fatigueDetector.analyzeFromDf(productDf)
        
        # === Unified Scoring ===
        unifiedScore = self.scorer.calculate(
            productId=productId,
            productName=productName,
            roas=roas,
            mkTrend=mkResult['trend'],
            mkPValue=mkResult['p_value'],
            trendDirection=trendResult.direction,
            trendRoc=trendResult.roc,
            legacyMomentumPct=momentumPct,
            elasticity=elasticity,
            saturationStatus=saturationResult.status.value,
            marginalRoas=mRoi,
            cv=cv,
            intelVolatilityScore=volatilityResult.score,
            fatigueStatus=fatigueResult.status.value,
            fatiguePct=fatigueResult.ctrDecay,
            churnRiskScore=churnRisk,
            eventMultiplier=eventMult,
            eventNames=eventNames,
            nPeriods=nPeriods,
            dataQuality=mkResult['data_quality']
        )
        
        # === Budget Recommendation ===
        currentBudget = costSeries[-4:].mean() if len(costSeries) >= 4 else totalCost / max(nPeriods, 1)
        
        budgetRec = self.budgetRecommender.recommend(
            productId=productId,
            productName=productName,
            currentBudget=currentBudget,
            currentRevenue=totalRevenue / max(nPeriods, 1),
            roas=roas,
            action=unifiedScore.action,
            compositeScore=unifiedScore.compositeScore,
            elasticity=elasticity,
            saturationStatus=saturationResult.status.value,
            marginalRoas=mRoi,
            eventMultiplier=eventMult,
            eventNames=eventNames,
            confidence=unifiedScore.confidence,
            trendDirection=trendResult.direction,
            fatigueStatus=fatigueResult.status.value,
            churnRiskScore=churnRisk
        )
        
        # === Probability Estimate ===
        historicalRoas = [r for r in roiSeries if r > 0]
        historicalProfits = list(revenueSeries - costSeries)
        
        probability = self.probabilityEngine.estimate(
            compositeScore=unifiedScore.compositeScore,
            roasScore=unifiedScore.roasScore,
            trendScore=unifiedScore.trendScore,
            elasticityScore=unifiedScore.elasticityScore,
            historicalRoas=historicalRoas,
            historicalProfits=historicalProfits,
            eventMultiplier=eventMult,
            churnRiskScore=churnRisk,
            fatigueStatus=fatigueResult.status.value,
            nPeriods=nPeriods,
            dataQuality=mkResult['data_quality']
        )
        
        # Generate summary
        summary = self._generateSummary(unifiedScore, budgetRec, probability)
        
        # Check if product is discontinued (not in Excel reference)
        isDiscontinued = self.productResolver.isDiscontinued(productId)
        
        return ComprehensiveAnalysis(
            productId=productId,
            productName=productName,
            creativeType=creativeType,
            isDiscontinued=isDiscontinued,
            totalCost=totalCost,
            totalRevenue=totalRevenue,
            totalProfit=totalProfit,
            totalOrders=int(totalOrders),
            roas=roas,
            nPeriods=nPeriods,
            unifiedScore=unifiedScore,
            budgetRec=budgetRec,
            probability=probability,
            currentEventMultiplier=eventMult,
            currentEvents=eventNames,
            nextEventDate=nextEvent[0] if nextEvent else None,
            nextEventName=nextEvent[1] if nextEvent else None,
            legacyMkTrend=mkResult['trend'],
            legacyMomentumPct=momentumPct,
            legacyCv=cv,
            legacyChurnRisk=churnRisk,
            finalAction=unifiedScore.action,
            finalBudgetChange=budgetRec.budgetChangePct,
            finalBudgetNominal=budgetRec.budgetChange,
            successProbability=probability.successProbability,
            confidenceLevel=probability.confidenceLevel,
            summary=summary
        )
    
    def _getProductName(self, productId: str, df: pd.DataFrame) -> str:
        """
        Get product name using ProductResolver (Excel mapping).
        
        Priority:
        1. Excel mapping (most accurate)
        2. Video title (if contains 'Nusseyba')
        3. Campaign name (fallback)
        """
        videoTitle = str(df['videoTitle'].iloc[0]) if 'videoTitle' in df else ''
        campaignName = str(df['campaignName'].iloc[0]) if 'campaignName' in df else ''
        
        # Use ProductResolver with Excel mapping
        return self.productResolver.resolve(
            productId=str(productId),
            campaignName=campaignName,
            videoTitle=videoTitle
        )
    
    def _calculateSimpleChurnRisk(self, mkResult: Dict, momentumPct: float, cv: float) -> float:
        """Calculate simplified churn risk score."""
        risk = 0
        
        # Handle None values
        momentumPct = momentumPct or 0
        cv = cv or 50
        
        if "Turun" in mkResult.get('trend', ''):
            risk += 30
        if momentumPct < -20:
            risk += 30
        if cv > 80:
            risk += 20
        if mkResult.get('p_value', 1.0) < 0.05 and "Turun" in mkResult.get('trend', ''):
            risk += 20
        
        return min(100, risk)
    
    def _generateSummary(self, score: UnifiedScore, budget: BudgetRecommendation,
                          prob: ProbabilityEstimate) -> str:
        """Generate human-readable summary."""
        parts = [
            f"{score.category}",
            f"Score: {score.compositeScore:.0f}",
            f"Action: {score.action.value}",
            f"Budget: {budget.budgetChangePct:+.0f}%",
            f"Success: {prob.successProbability:.0f}% ({prob.confidenceLevel})"
        ]
        return " | ".join(parts)
