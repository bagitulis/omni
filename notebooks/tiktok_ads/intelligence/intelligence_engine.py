#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Intelligence Engine
===================
Main orchestrator that combines all analysis modules.
Following AGENTS.MD: Clean Code, DRY, SRP.
"""

from dataclasses import dataclass, field
from typing import Dict, List, Optional
from datetime import datetime
import pandas as pd
from pathlib import Path

from .config_intelligence import IntelligenceConfig, RecommendationAction
from .indonesian_calendar import IndonesianCalendar
from .funnel_analyzer import FunnelAnalyzer
from .roas_classifier import RoasClassifier
from .trend_momentum import TrendMomentum
from .volatility_analyzer import VolatilityAnalyzer
from .saturation_model import SaturationModel
from .budget_optimizer import BudgetOptimizer
from .fatigue_detector import FatigueDetector
from .product_lifecycle import ProductLifecycle
from .composite_scorer import CompositeScorer
from .probability_engine import ProbabilityEngine
from .data_aggregator import DataAggregator
from .product_resolver import ProductResolver


@dataclass
class ProductAnalysis:
    """Complete analysis result for a single product."""
    productId: str
    productName: str
    isDiscontinued: bool
    
    # Basic metrics
    totalCost: float
    totalRevenue: float
    totalProfit: float
    totalOrders: int
    periodsCount: int
    
    # Analysis results
    roasScore: float
    roasTier: str
    trendScore: float
    trendDirection: str
    volatilityScore: float
    riskLevel: str
    saturationScore: float
    saturationStatus: str
    fatigueScore: float
    fatigueStatus: str
    lifecycleStage: str
    
    # Final recommendation
    compositeScore: float
    category: str
    action: RecommendationAction
    successProbability: float
    confidenceLevel: str
    
    # Budget recommendation
    currentBudget: float
    recommendedBudget: float
    budgetChange: float
    budgetReasons: List[str]
    
    # Indonesian context
    indoMultiplier: float
    indoEvents: List[str]
    
    # Summary
    summary: str


class IntelligenceEngine:
    """
    Main orchestrator for the Ad Intelligence System.
    Coordinates all analyzers and produces comprehensive analysis.
    """
    
    def __init__(self, dbPath: str, excelPath: str = None,
                 tenantId: str = "yumna_bertigamart"):
        self.config = IntelligenceConfig()
        
        # Initialize all analyzers
        self.calendar = IndonesianCalendar()
        self.funnelAnalyzer = FunnelAnalyzer()
        self.roasClassifier = RoasClassifier(self.config)
        self.trendMomentum = TrendMomentum()
        self.volatilityAnalyzer = VolatilityAnalyzer(self.config)
        self.saturationModel = SaturationModel(self.config.budget.maxBudgetPerProduct)
        self.budgetOptimizer = BudgetOptimizer(self.config)
        self.fatigueDetector = FatigueDetector(self.config)
        self.productLifecycle = ProductLifecycle(self.config)
        self.compositeScorer = CompositeScorer(self.config)
        self.probabilityEngine = ProbabilityEngine(self.config)
        
        # Data layer
        self.dataAggregator = DataAggregator(dbPath, tenantId)
        self.productResolver = ProductResolver(excelPath)
    
    def analyzeProduct(self, productId: str, 
                       rawDf: pd.DataFrame = None) -> ProductAnalysis:
        """Analyze a single product comprehensively."""
        # Get time series data
        timeSeries = self.dataAggregator.getProductTimeSeries(productId, rawDf)
        
        if timeSeries.empty:
            return self._emptyAnalysis(productId)
        
        # Get aggregated metrics
        aggDf = self.dataAggregator.aggregateByProduct(rawDf)
        productAgg = aggDf[aggDf['productId'] == productId]
        
        if productAgg.empty:
            return self._emptyAnalysis(productId)
        
        row = productAgg.iloc[0]
        
        # Resolve product name
        productName = self.productResolver.resolve(
            productId, 
            row.get('campaignName', ''),
            row.get('videoTitle', '')
        )
        isDiscontinued = self.productResolver.isDiscontinued(productId)
        
        # Run all analyzers
        roasResult = self.roasClassifier.classify(
            row['totalRevenue'], row['totalCost']
        )
        
        trendResult = self.trendMomentum.analyzeFromDf(timeSeries, 'roi')
        volatilityResult = self.volatilityAnalyzer.analyzeFromDf(timeSeries, 'roi')
        saturationResult = self.saturationModel.analyzeFromDf(timeSeries)
        fatigueResult = self.fatigueDetector.analyzeFromDf(timeSeries)
        lifecycleResult = self.productLifecycle.analyze(timeSeries)
        
        # Get Indonesian context
        today = datetime.now()
        indoMult, indoEvents = self.calendar.getCompositeMultiplier(today)
        
        # Calculate composite score
        compositeResult = self.compositeScorer.calculate(
            roasScore=roasResult.score,
            trendScore=trendResult.score,
            elasticityScore=saturationResult.score,
            volatilityScore=volatilityResult.score,
            fatigueScore=fatigueResult.score,
            indoContextMultiplier=indoMult
        )
        
        # Budget recommendation
        avgBudgetPerPeriod = row['totalCost'] / max(row['periodsCount'], 1)
        budgetRec = self.budgetOptimizer.recommend(
            currentBudget=avgBudgetPerPeriod,
            compositeScore=compositeResult.finalScore,
            marginalRoas=saturationResult.marginalRoas,
            volatilityScore=volatilityResult.score,
            trendScore=trendResult.score
        )
        
        # Success probability
        probResult = self.probabilityEngine.predict(
            compositeScore=compositeResult.finalScore,
            historicalRoas=row['avgRoas'],
            trendDirection=trendResult.direction.name.replace('_', ' '),
            volatilityCv=volatilityResult.cv,
            indoMultiplier=indoMult,
            dataPoints=len(timeSeries)
        )
        
        # Generate summary
        summary = self._generateSummary(
            productName, compositeResult, budgetRec, probResult
        )
        
        return ProductAnalysis(
            productId=productId,
            productName=productName,
            isDiscontinued=isDiscontinued,
            totalCost=row['totalCost'],
            totalRevenue=row['totalRevenue'],
            totalProfit=row['totalProfit'],
            totalOrders=int(row['totalOrders']),
            periodsCount=int(row['periodsCount']),
            roasScore=roasResult.score,
            roasTier=roasResult.tier.value,
            trendScore=trendResult.score,
            trendDirection=trendResult.direction.value,
            volatilityScore=volatilityResult.score,
            riskLevel=volatilityResult.riskLevel.value,
            saturationScore=saturationResult.score,
            saturationStatus=saturationResult.status.value,
            fatigueScore=fatigueResult.score,
            fatigueStatus=fatigueResult.status.value,
            lifecycleStage=lifecycleResult.stage.value,
            compositeScore=compositeResult.finalScore,
            category=compositeResult.category.value,
            action=compositeResult.action,
            successProbability=probResult.successProbability,
            confidenceLevel=probResult.confidenceLevel.value,
            currentBudget=budgetRec.currentBudget,
            recommendedBudget=budgetRec.recommendedBudget,
            budgetChange=budgetRec.changePercent,
            budgetReasons=budgetRec.reasons,
            indoMultiplier=indoMult,
            indoEvents=indoEvents,
            summary=summary
        )
    
    def analyzeAll(self, minCost: float = 10000,
                   verbose: bool = True) -> List[ProductAnalysis]:
        """Analyze all products in the database."""
        if verbose:
            print("📊 Loading data from database...")
        
        rawDf = self.dataAggregator.loadRawData(minCost)
        
        if rawDf.empty:
            print("⚠️ No data found!")
            return []
        
        productIds = rawDf['productId'].unique()
        
        if verbose:
            print(f"   Found {len(productIds)} products to analyze")
        
        results = []
        for i, pid in enumerate(productIds):
            if verbose and (i + 1) % 50 == 0:
                print(f"   Processing {i + 1}/{len(productIds)}...")
            
            analysis = self.analyzeProduct(pid, rawDf)
            results.append(analysis)
        
        if verbose:
            print(f"✅ Analyzed {len(results)} products")
        
        return results
    
    def toDataFrame(self, analyses: List[ProductAnalysis]) -> pd.DataFrame:
        """Convert list of ProductAnalysis to DataFrame."""
        records = []
        for a in analyses:
            records.append({
                'productId': a.productId,
                'productName': a.productName,
                'isDiscontinued': a.isDiscontinued,
                'totalCost': a.totalCost,
                'totalRevenue': a.totalRevenue,
                'totalProfit': a.totalProfit,
                'totalOrders': a.totalOrders,
                'periodsCount': a.periodsCount,
                'roasScore': a.roasScore,
                'roasTier': a.roasTier,
                'trendScore': a.trendScore,
                'trendDirection': a.trendDirection,
                'volatilityScore': a.volatilityScore,
                'riskLevel': a.riskLevel,
                'saturationScore': a.saturationScore,
                'saturationStatus': a.saturationStatus,
                'fatigueScore': a.fatigueScore,
                'fatigueStatus': a.fatigueStatus,
                'lifecycleStage': a.lifecycleStage,
                'compositeScore': a.compositeScore,
                'category': a.category,
                'action': a.action.value,
                'successProbability': a.successProbability,
                'confidenceLevel': a.confidenceLevel,
                'currentBudget': a.currentBudget,
                'recommendedBudget': a.recommendedBudget,
                'budgetChange': a.budgetChange,
                'indoMultiplier': a.indoMultiplier,
                'summary': a.summary,
            })
        
        return pd.DataFrame(records)
    
    def _generateSummary(self, name: str, composite, budget, prob) -> str:
        """Generate summary text for a product."""
        actionMap = {
            RecommendationAction.SCALE_UP_AGGRESSIVE: "Scale Up Agresif",
            RecommendationAction.SCALE_UP_MODERATE: "Scale Up Moderat",
            RecommendationAction.MAINTAIN: "Pertahankan",
            RecommendationAction.REDUCE_BUDGET: "Kurangi Budget",
            RecommendationAction.STOP_IMMEDIATELY: "Stop Segera",
            RecommendationAction.MONITOR_CLOSELY: "Monitor Ketat",
        }
        
        action = actionMap.get(composite.action, "Monitor")
        
        return (
            f"{name[:40]}: {composite.category.value} "
            f"(Score {composite.finalScore:.0f}). "
            f"Action: {action}. "
            f"Budget: {budget.changePercent:+.0f}%. "
            f"Success Prob: {prob.successProbability:.0f}%"
        )
    
    def _emptyAnalysis(self, productId: str) -> ProductAnalysis:
        """Return empty analysis for products with no data."""
        return ProductAnalysis(
            productId=productId,
            productName=f"Unknown-{productId[-6:]}",
            isDiscontinued=True,
            totalCost=0, totalRevenue=0, totalProfit=0,
            totalOrders=0, periodsCount=0,
            roasScore=0, roasTier="UNKNOWN",
            trendScore=50, trendDirection="INSUFFICIENT_DATA",
            volatilityScore=50, riskLevel="MEDIUM",
            saturationScore=50, saturationStatus="INSUFFICIENT_DATA",
            fatigueScore=50, fatigueStatus="INSUFFICIENT_DATA",
            lifecycleStage="LAUNCH",
            compositeScore=0, category="STOP",
            action=RecommendationAction.STOP_IMMEDIATELY,
            successProbability=0, confidenceLevel="VERY_LOW",
            currentBudget=0, recommendedBudget=0, budgetChange=0,
            budgetReasons=["No data available"],
            indoMultiplier=1.0, indoEvents=[],
            summary="No data available for analysis"
        )
