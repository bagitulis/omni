#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Unified Scorer Engine (Generic Core Module)
===========================================
Main scoring engine that combines multiple methodologies.
Platform-agnostic: Works for TikTok, Shopee, Lazada, etc.
Following AGENTS.MD: Clean Code, DRY, SRP, max 300 lines.
"""

from typing import Dict, List
import numpy as np

from .unified_types import (
    ActionRecommendation,
    ScoreCategory,
    ScoringWeightsConfig,
    ActionThresholds,
    UnifiedScoreResult,
    ScoreInputs,
)


class UnifiedScorer:
    """
    Unified scoring system combining multiple methodologies.
    
    Default Weights (configurable):
    - ROAS: 20% (core profitability)
    - Trend: 15% (direction)
    - Momentum: 10% (recent performance)
    - Elasticity: 15% (budget response)
    - Volatility: 10% (stability)
    - Fatigue: 10% (creative health)
    - Churn Risk: 10% (decline risk)
    - Event Context: 10% (timing)
    """
    
    def __init__(self, 
                 weights: ScoringWeightsConfig = None,
                 thresholds: ActionThresholds = None):
        self.weights = weights or ScoringWeightsConfig()
        self.thresholds = thresholds or ActionThresholds()
    
    def calculate(self, inputs: ScoreInputs) -> UnifiedScoreResult:
        """Calculate unified score from inputs."""
        scores = {}
        topFactors = []
        warningFactors = []
        
        # 1. ROAS Score
        scores['roas'] = self._calculateRoasScore(inputs.roas)
        if inputs.roas >= 5:
            topFactors.append(f"Excellent ROAS: {inputs.roas:.1f}x")
        elif inputs.roas < 1:
            warningFactors.append(f"Losing money: ROAS {inputs.roas:.2f}x")
        
        # 2. Trend Score
        scores['trend'] = self._calculateTrendScore(
            inputs.mkTrend, inputs.mkPValue, 
            inputs.trendDirection, inputs.trendRoc
        )
        if inputs.trendDirection == "UP":
            topFactors.append("Upward trend")
        elif inputs.trendDirection == "DOWN":
            warningFactors.append("Downward trend")
        
        # 3. Momentum Score
        scores['momentum'] = self._calculateMomentumScore(inputs.legacyMomentumPct)
        if inputs.legacyMomentumPct > 20:
            topFactors.append(f"Strong momentum: +{inputs.legacyMomentumPct:.0f}%")
        elif inputs.legacyMomentumPct < -20:
            warningFactors.append(f"Declining: {inputs.legacyMomentumPct:.0f}%")
        
        # 4. Elasticity Score
        scores['elasticity'] = self._calculateElasticityScore(
            inputs.elasticity, inputs.saturationStatus, inputs.marginalRoas
        )
        if inputs.elasticity > 1.0:
            topFactors.append(f"High elasticity: {inputs.elasticity:.2f}")
        elif inputs.saturationStatus == "SATURATED":
            warningFactors.append("Budget saturated")
        
        # 5. Volatility Score
        scores['volatility'] = self._calculateVolatilityScore(
            inputs.cv, inputs.intelVolatilityScore
        )
        if inputs.cv > 80:
            warningFactors.append(f"High volatility: CV={inputs.cv:.0f}%")
        
        # 6. Fatigue Score
        scores['fatigue'] = self._calculateFatigueScore(
            inputs.fatigueStatus, inputs.fatiguePct
        )
        if inputs.fatigueStatus in ["FATIGUED", "DEAD"]:
            warningFactors.append(f"Creative fatigue: {inputs.fatigueStatus}")
        
        # 7. Churn Risk Score (inverted)
        scores['churn'] = 100 - min(100, inputs.churnRiskScore)
        if inputs.churnRiskScore > 60:
            warningFactors.append(f"High churn risk: {inputs.churnRiskScore:.0f}%")
        
        # 8. Event Score
        scores['event'] = self._calculateEventScore(inputs.eventMultiplier)
        if inputs.eventMultiplier > 1.3 and inputs.eventNames:
            topFactors.append(f"Event boost: {', '.join(inputs.eventNames[:2])}")
        
        # Calculate weighted composite
        compositeScore = (
            scores['roas'] * self.weights.roas +
            scores['trend'] * self.weights.trend +
            scores['momentum'] * self.weights.momentum +
            scores['elasticity'] * self.weights.elasticity +
            scores['volatility'] * self.weights.volatility +
            scores['fatigue'] * self.weights.fatigue +
            scores['churn'] * self.weights.churn +
            scores['event'] * self.weights.event
        )
        
        # Adjust for data quality
        if inputs.nPeriods < 4:
            compositeScore *= 0.85
            warningFactors.append(f"Limited data: {inputs.nPeriods} periods")
        
        # Determine category and action
        category = self._determineCategory(compositeScore)
        action = self._determineAction(compositeScore, inputs.roas, scores)
        
        # Calculate confidence
        confidence = self._calculateConfidence(
            inputs.nPeriods, inputs.cv, inputs.dataQuality
        )
        confidenceLevel = self._getConfidenceLevel(confidence)
        
        return UnifiedScoreResult(
            productId=inputs.productId,
            productName=inputs.productName,
            roasScore=scores['roas'],
            trendScore=scores['trend'],
            momentumScore=scores['momentum'],
            elasticityScore=scores['elasticity'],
            volatilityScore=scores['volatility'],
            fatigueScore=scores['fatigue'],
            churnRiskScore=scores['churn'],
            eventScore=scores['event'],
            compositeScore=round(compositeScore, 1),
            category=category,
            action=action,
            confidence=confidence,
            confidenceLevel=confidenceLevel,
            topFactors=topFactors[:3],
            warningFactors=warningFactors[:3]
        )
    
    def _calculateRoasScore(self, roas: float) -> float:
        """Convert ROAS to 0-100 score."""
        if roas >= 10: return 100
        if roas >= 5: return 80 + (roas - 5) * 4
        if roas >= 3: return 60 + (roas - 3) * 10
        if roas >= 2: return 45 + (roas - 2) * 15
        if roas >= 1: return 25 + (roas - 1) * 20
        return max(0, roas * 25)
    
    def _calculateTrendScore(self, mkTrend: str, mkPValue: float,
                              trendDir: str, roc: float) -> float:
        """Combine Mann-Kendall and MACD trend scores."""
        mkScore = 25
        if "increasing" in mkTrend.lower() or "naik" in mkTrend.lower():
            mkScore = 50 if mkPValue < 0.05 else 40
        elif "decreasing" in mkTrend.lower() or "turun" in mkTrend.lower():
            mkScore = 0 if mkPValue < 0.05 else 10
        
        macdScore = 25
        if trendDir == "UP":
            macdScore = min(50, 30 + roc)
        elif trendDir == "DOWN":
            macdScore = max(0, 20 + roc)
        
        return mkScore + macdScore
    
    def _calculateMomentumScore(self, momentumPct: float) -> float:
        """Convert momentum % to score."""
        return max(0, min(100, 50 + momentumPct * 0.5))
    
    def _calculateElasticityScore(self, elasticity: float,
                                   satStatus: str, mRoas: float) -> float:
        """Combine elasticity and saturation."""
        if elasticity > 1.2:
            eScore = 60
        elif elasticity > 0.8:
            eScore = 45
        elif elasticity > 0:
            eScore = 30
        else:
            eScore = 10
        
        satScores = {
            "GROWTH_PHASE": 40, "HIGH_ELASTICITY": 40,
            "APPROACHING_SATURATION": 30, "MODERATE": 30,
            "SATURATED": 15,
            "DECLINING": 5, "OVER_SATURATED": 5,
            "INSUFFICIENT_DATA": 25
        }
        sScore = satScores.get(satStatus, 25)
        
        if mRoas > 2:
            sScore = min(40, sScore + 10)
        elif mRoas < 1:
            sScore = max(0, sScore - 10)
        
        return eScore + sScore
    
    def _calculateVolatilityScore(self, cv: float, intelScore: float) -> float:
        """Higher = more stable. Adjusted for ad metrics."""
        if cv < 100:
            cvScore = 50
        elif cv < 200:
            cvScore = 40
        elif cv < 400:
            cvScore = 30
        elif cv < 800:
            cvScore = 20
        else:
            cvScore = 10
        
        return cvScore + intelScore * 0.5
    
    def _calculateFatigueScore(self, status: str, pct: float) -> float:
        """Higher = fresher."""
        statusScores = {
            "FRESH": 90,
            "AGING": 70,
            "FATIGUED": 50,
            "DEAD": 30,
            "INSUFFICIENT_DATA": 60
        }
        base = statusScores.get(status, 50)
        
        if pct < 15:
            base = min(100, base + 10)
        elif pct > 40:
            base = max(0, base - 10)
        
        return base
    
    def _calculateEventScore(self, multiplier: float) -> float:
        """Convert event multiplier to score."""
        if multiplier >= 1.5:
            return 100
        elif multiplier >= 1.3:
            return 80
        elif multiplier >= 1.0:
            return 60
        else:
            return max(20, 60 + (multiplier - 1) * 100)
    
    def _determineCategory(self, score: float) -> ScoreCategory:
        """Determine product category from score."""
        if score >= 80:
            return ScoreCategory.STAR
        elif score >= 65:
            return ScoreCategory.GROWTH
        elif score >= 50:
            return ScoreCategory.STABLE
        elif score >= 35:
            return ScoreCategory.WATCH
        else:
            return ScoreCategory.PROBLEM
    
    def _determineAction(self, score: float, roas: float,
                         scores: Dict) -> ActionRecommendation:
        """Determine action recommendation."""
        if roas < 0.8:
            return ActionRecommendation.STOP
        
        if roas < 1.5 and score < 50:
            return ActionRecommendation.REDUCE
        
        hasHighRoas = roas >= 3.0
        hasExcellentRoas = roas >= 5.0
        
        if hasExcellentRoas:
            if score >= 60:
                return ActionRecommendation.SCALE_UP_AGGRESSIVE
            elif score >= 50:
                return ActionRecommendation.SCALE_UP
        
        if hasHighRoas:
            if score >= 65:
                return ActionRecommendation.SCALE_UP
            elif score >= 45:
                return ActionRecommendation.MAINTAIN
        
        if score >= self.thresholds.scaleAggressive:
            return ActionRecommendation.SCALE_UP_AGGRESSIVE
        elif score >= self.thresholds.scale:
            return ActionRecommendation.SCALE_UP
        elif score >= self.thresholds.maintain:
            return ActionRecommendation.MAINTAIN
        elif score >= self.thresholds.reduce:
            return ActionRecommendation.REDUCE
        else:
            return ActionRecommendation.STOP
    
    def _calculateConfidence(self, nPeriods: int, cv: float,
                              dataQuality: str) -> float:
        """Calculate overall confidence."""
        if nPeriods >= 12:
            dataConf = 40
        elif nPeriods >= 8:
            dataConf = 30
        elif nPeriods >= 4:
            dataConf = 20
        else:
            dataConf = 10
        
        if cv < 30:
            stabConf = 30
        elif cv < 60:
            stabConf = 20
        else:
            stabConf = 10
        
        if "HIGH" in dataQuality:
            qualConf = 30
        elif "MEDIUM" in dataQuality:
            qualConf = 20
        else:
            qualConf = 10
        
        return dataConf + stabConf + qualConf
    
    def _getConfidenceLevel(self, confidence: float) -> str:
        """Get confidence level label."""
        if confidence >= 80:
            return "HIGH"
        elif confidence >= 60:
            return "MEDIUM"
        elif confidence >= 40:
            return "LOW"
        else:
            return "VERY_LOW"
