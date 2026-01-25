#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Portfolio Metrics Module
========================
Health score dan financial projection calculations.
Dipisah dari main generator untuk SRP compliance.
"""

from dataclasses import dataclass
from typing import List
import numpy as np

from .comprehensive_engine import ComprehensiveAnalysis
from .unified_scorer import ActionRecommendation


@dataclass
class PortfolioHealth:
    """Portfolio health metrics."""
    healthScore: float
    grade: str
    profitabilityScore: float
    roiScore: float
    trendScore: float
    dataQualityScore: float
    riskLevel: str
    riskScore: float
    riskFactors: List[str]


@dataclass
class FinancialProjection:
    """Financial projection with CI."""
    currentProfit: float
    projectedProfit: float
    growthPct: float
    ciLower: float
    ciUpper: float
    savingsFromStop: float
    gainsFromScaleUp: float
    reallocationGains: float
    methodology: str


def calculateHealthScore(analyses: List[ComprehensiveAnalysis]) -> PortfolioHealth:
    """
    Calculate portfolio health score with components.
    
    Components:
    - Profitability (40%): Persentase produk yang menguntungkan
    - ROI Performance (30%): Logarithmic scaling dari ROAS keseluruhan
    - Trend Health (15%): Produk dengan trend naik/stabil
    - Data Quality (15%): Produk dengan confidence level tinggi/medium
    """
    if not analyses:
        return PortfolioHealth(0, "F", 0, 0, 0, 0, "UNKNOWN", 0, [])
    
    # Profitability (40%)
    profitable = sum(1 for a in analyses if a.totalProfit > 0)
    profitabilityScore = (profitable / len(analyses)) * 100
    
    # ROI Performance (30%)
    totalCost = sum(a.totalCost for a in analyses)
    totalRevenue = sum(a.totalRevenue for a in analyses)
    overallRoas = totalRevenue / totalCost if totalCost > 0 else 0
    roiScore = min(100, 50 + np.log10(max(overallRoas, 0.1)) * 30)
    
    # Trend Health (15%)
    uptrend = sum(1 for a in analyses if "Naik" in a.legacyMkTrend or a.unifiedScore.trendScore > 60)
    stable = sum(1 for a in analyses if "Stabil" in a.legacyMkTrend or 40 <= a.unifiedScore.trendScore <= 60)
    trendScore = ((uptrend + stable * 0.7) / len(analyses)) * 100
    
    # Data Quality (15%)
    highConf = sum(1 for a in analyses if a.confidenceLevel == "HIGH")
    medConf = sum(1 for a in analyses if a.confidenceLevel == "MEDIUM")
    dataQualityScore = ((highConf + medConf * 0.6) / len(analyses)) * 100
    
    # Health Score
    healthScore = (
        profitabilityScore * 0.40 +
        roiScore * 0.30 +
        trendScore * 0.15 +
        dataQualityScore * 0.15
    )
    
    # Grade
    if healthScore >= 80: grade = "A"
    elif healthScore >= 70: grade = "B"
    elif healthScore >= 60: grade = "C"
    elif healthScore >= 50: grade = "D"
    else: grade = "F"
    
    # Risk Assessment
    riskFactors = []
    riskScore = 0
    
    stopProducts = [a for a in analyses if a.finalAction == ActionRecommendation.STOP]
    if len(stopProducts) > 0:
        riskFactors.append(f"Products to STOP: {len(stopProducts)}")
        riskScore += min(30, len(stopProducts) * 10)
    
    highChurn = [a for a in analyses if a.unifiedScore.churnRiskScore > 60]
    if highChurn:
        riskFactors.append(f"High churn risk products: {len(highChurn)}")
        highChurnPct = len(highChurn) / len(analyses)
        riskScore += min(40, int(highChurnPct * 100))
    
    reduceProducts = [a for a in analyses if a.finalAction == ActionRecommendation.REDUCE]
    if len(reduceProducts) > len(analyses) * 0.3:
        riskFactors.append(f"Products to REDUCE: {len(reduceProducts)} ({len(reduceProducts)/len(analyses)*100:.0f}%)")
        riskScore += 15
    
    lowConf = [a for a in analyses if a.confidenceLevel in ["LOW", "VERY_LOW"]]
    if len(lowConf) > len(analyses) * 0.2:
        riskFactors.append(f"Low confidence data: {len(lowConf)} products")
        riskScore += 15
    
    riskScore = min(100, riskScore)
    
    if riskScore >= 50: riskLevel = "🔴 HIGH"
    elif riskScore >= 30: riskLevel = "🟡 MEDIUM"
    else: riskLevel = "🟢 LOW"
    
    if not riskFactors:
        riskFactors.append("Tidak ada risiko signifikan terdeteksi")
    
    return PortfolioHealth(
        healthScore=round(healthScore, 1),
        grade=grade,
        profitabilityScore=round(profitabilityScore, 1),
        roiScore=round(roiScore, 1),
        trendScore=round(trendScore, 1),
        dataQualityScore=round(dataQualityScore, 1),
        riskLevel=riskLevel,
        riskScore=riskScore,
        riskFactors=riskFactors
    )


def calculateProjection(analyses: List[ComprehensiveAnalysis]) -> FinancialProjection:
    """
    Calculate financial projections with CI.
    
    Sources of improvement:
    - Savings from STOP: Stop losing money on unprofitable products
    - Gains from Scale Up: Increase budget to high performers
    - Reallocation: Move budget from REDUCE to Scale Up
    """
    currentProfit = sum(a.totalProfit for a in analyses)
    
    # Savings from STOP
    stopProducts = [a for a in analyses if a.finalAction == ActionRecommendation.STOP]
    savingsFromStop = sum(abs(a.totalProfit) for a in stopProducts if a.totalProfit < 0)
    stopBudgetRecovered = sum(a.totalCost for a in stopProducts)
    
    # Gains from Scale Up
    scaleUp = [a for a in analyses if a.finalAction in [
        ActionRecommendation.SCALE_UP_AGGRESSIVE, ActionRecommendation.SCALE_UP
    ]]
    gainsFromScaleUp = sum(a.budgetRec.profitChange for a in scaleUp if a.budgetRec.profitChange > 0)
    
    # Reallocation potential
    if scaleUp:
        avgScaleUpRoas = np.mean([a.roas for a in scaleUp])
        potentialReallocation = stopBudgetRecovered * (avgScaleUpRoas - 1) * 0.5
    else:
        avgScaleUpRoas = 1
        potentialReallocation = 0
    
    reduceProducts = [a for a in analyses if a.finalAction == ActionRecommendation.REDUCE]
    freedBudget = sum(abs(a.budgetRec.budgetChange) for a in reduceProducts)
    reallocationGains = freedBudget * (avgScaleUpRoas - 1) * 0.7
    
    # Total projection
    totalGains = savingsFromStop + gainsFromScaleUp + reallocationGains + potentialReallocation
    projectedProfit = currentProfit + totalGains
    growthPct = ((totalGains) / currentProfit * 100) if currentProfit > 0 else 0
    
    # 95% CI
    uncertainties = [a.probability.uncertaintyPct for a in analyses if hasattr(a.probability, 'uncertaintyPct')]
    avgUncertainty = np.mean(uncertainties) if uncertainties else 20
    ciMargin = max(projectedProfit * (avgUncertainty / 100), totalGains * 0.3)
    
    return FinancialProjection(
        currentProfit=currentProfit,
        projectedProfit=projectedProfit,
        growthPct=round(growthPct, 2),
        ciLower=projectedProfit - ciMargin,
        ciUpper=projectedProfit + ciMargin,
        savingsFromStop=savingsFromStop + potentialReallocation,
        gainsFromScaleUp=gainsFromScaleUp,
        reallocationGains=reallocationGains,
        methodology="Composite: ROI × Conservative Factor (0.7) × Trend Adjustment"
    )
