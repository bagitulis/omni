#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Budget Optimizer
================
Budget recommendation engine with cap limits.
Following AGENTS.MD: Clean Code, DRY, SRP.
"""

from dataclasses import dataclass
from typing import List
from enum import Enum

from .config_intelligence import IntelligenceConfig, RecommendationAction


class BudgetAction(Enum):
    """Budget action types."""
    INCREASE_AGGRESSIVE = "INCREASE_AGGRESSIVE"  # +30%
    INCREASE_MODERATE = "INCREASE_MODERATE"      # +15%
    MAINTAIN = "MAINTAIN"                        # 0%
    DECREASE_MODERATE = "DECREASE_MODERATE"      # -20%
    DECREASE_AGGRESSIVE = "DECREASE_AGGRESSIVE"  # -50%
    STOP = "STOP"                                # Stop completely


@dataclass
class BudgetRecommendation:
    """Budget recommendation result."""
    action: BudgetAction
    currentBudget: float
    recommendedBudget: float
    changePercent: float
    changeAmount: float
    cappedByMax: bool
    reasons: List[str]
    confidenceLevel: str


class BudgetOptimizer:
    """
    Optimizes budget recommendations based on multiple factors.
    Respects maximum budget cap (Rp 2 juta per product).
    """
    
    def __init__(self, config: IntelligenceConfig = None):
        self.config = config or IntelligenceConfig()
        self.maxBudget = self.config.budget.maxBudgetPerProduct
    
    def recommend(self, currentBudget: float, compositeScore: float,
                  marginalRoas: float, volatilityScore: float,
                  trendScore: float) -> BudgetRecommendation:
        """
        Generate budget recommendation based on analysis results.
        
        Args:
            currentBudget: Current ad spend
            compositeScore: Final composite score (0-100)
            marginalRoas: Marginal ROAS from saturation model
            volatilityScore: Stability score (0-100)
            trendScore: Trend momentum score (0-100)
        
        Returns:
            BudgetRecommendation with action and new budget
        """
        reasons = []
        
        # Determine base action from composite score
        baseAction = self._getBaseAction(compositeScore)
        reasons.append(f"Composite Score: {compositeScore:.1f}/100")
        
        # Adjust based on marginal ROAS
        action = self._adjustByMarginalRoas(baseAction, marginalRoas, reasons)
        
        # Adjust based on volatility (risk)
        action = self._adjustByVolatility(action, volatilityScore, reasons)
        
        # Adjust based on trend
        action = self._adjustByTrend(action, trendScore, reasons)
        
        # Calculate new budget
        changePercent = self._getChangePercent(action)
        newBudget = currentBudget * (1 + changePercent)
        
        # Apply max budget cap
        cappedByMax = False
        if newBudget > self.maxBudget:
            newBudget = self.maxBudget
            cappedByMax = True
            reasons.append(f"⚠️ Budget dicap ke maksimum Rp {self.maxBudget:,.0f}")
        
        # Ensure minimum (don't go below 0 for non-stop)
        if action != BudgetAction.STOP and newBudget < 10000:
            newBudget = max(10000, currentBudget * 0.5)
        
        # Stop action
        if action == BudgetAction.STOP:
            newBudget = 0
        
        changeAmount = newBudget - currentBudget
        
        # Confidence level
        confidence = self._getConfidenceLevel(compositeScore, volatilityScore)
        
        return BudgetRecommendation(
            action=action,
            currentBudget=round(currentBudget, 0),
            recommendedBudget=round(newBudget, 0),
            changePercent=round(changePercent * 100, 1),
            changeAmount=round(changeAmount, 0),
            cappedByMax=cappedByMax,
            reasons=reasons,
            confidenceLevel=confidence
        )
    
    def _getBaseAction(self, score: float) -> BudgetAction:
        """Get base action from composite score."""
        thresholds = self.config.scores
        
        if score >= thresholds.star:
            return BudgetAction.INCREASE_AGGRESSIVE
        elif score >= thresholds.performer:
            return BudgetAction.INCREASE_MODERATE
        elif score >= thresholds.average:
            return BudgetAction.MAINTAIN
        elif score >= thresholds.underperformer:
            return BudgetAction.DECREASE_MODERATE
        elif score >= thresholds.problem:
            return BudgetAction.DECREASE_AGGRESSIVE
        else:
            return BudgetAction.STOP
    
    def _adjustByMarginalRoas(self, action: BudgetAction, 
                              marginalRoas: float,
                              reasons: List[str]) -> BudgetAction:
        """Adjust action based on marginal ROAS."""
        if marginalRoas < 1.0:
            # Marginal ROAS negative - don't increase
            if action in [BudgetAction.INCREASE_AGGRESSIVE, 
                         BudgetAction.INCREASE_MODERATE]:
                reasons.append(f"⚠️ mROAS rendah ({marginalRoas:.1f}x), "
                             "tidak naikkan budget")
                return BudgetAction.MAINTAIN
        
        elif marginalRoas >= 3.0:
            # Very good marginal ROAS - can be more aggressive
            if action == BudgetAction.INCREASE_MODERATE:
                reasons.append(f"✅ mROAS tinggi ({marginalRoas:.1f}x), "
                             "upgrade ke agresif")
                return BudgetAction.INCREASE_AGGRESSIVE
        
        return action
    
    def _adjustByVolatility(self, action: BudgetAction,
                            volatilityScore: float,
                            reasons: List[str]) -> BudgetAction:
        """Adjust action based on volatility/risk."""
        if volatilityScore < 50:  # High volatility
            # Don't be aggressive with unstable products
            if action == BudgetAction.INCREASE_AGGRESSIVE:
                reasons.append(f"⚠️ Volatilitas tinggi (score {volatilityScore:.0f}), "
                             "turun ke moderat")
                return BudgetAction.INCREASE_MODERATE
        
        return action
    
    def _adjustByTrend(self, action: BudgetAction,
                       trendScore: float,
                       reasons: List[str]) -> BudgetAction:
        """Adjust action based on trend momentum."""
        if trendScore < 30:  # Strong downtrend
            # Don't increase on downtrend
            if action in [BudgetAction.INCREASE_AGGRESSIVE,
                         BudgetAction.INCREASE_MODERATE]:
                reasons.append(f"📉 Trend turun (score {trendScore:.0f}), "
                             "tidak naikkan budget")
                return BudgetAction.MAINTAIN
        
        elif trendScore >= 70:  # Strong uptrend
            # Can be more confident
            if action == BudgetAction.MAINTAIN:
                reasons.append(f"📈 Trend naik (score {trendScore:.0f}), "
                             "bisa naikkan moderat")
                return BudgetAction.INCREASE_MODERATE
        
        return action
    
    def _getChangePercent(self, action: BudgetAction) -> float:
        """Get budget change percentage from action."""
        changeMap = {
            BudgetAction.INCREASE_AGGRESSIVE: 
                self.config.budget.scaleUpAggressive,
            BudgetAction.INCREASE_MODERATE: 
                self.config.budget.scaleUpModerate,
            BudgetAction.MAINTAIN: 0.0,
            BudgetAction.DECREASE_MODERATE: 
                self.config.budget.reduceBudget,
            BudgetAction.DECREASE_AGGRESSIVE: 
                self.config.budget.reduceAggressiveBudget,
            BudgetAction.STOP: -1.0,
        }
        return changeMap.get(action, 0.0)
    
    def _getConfidenceLevel(self, compositeScore: float, 
                            volatilityScore: float) -> str:
        """Determine confidence level of recommendation."""
        # High confidence if score is extreme and volatility is low
        if volatilityScore >= 70:
            if compositeScore >= 80 or compositeScore <= 30:
                return "HIGH"
            return "MEDIUM"
        elif volatilityScore >= 50:
            return "MEDIUM"
        else:
            return "LOW"
