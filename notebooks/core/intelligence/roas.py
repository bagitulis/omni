#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
ROAS Classifier (Generic Core Module)
=====================================
Tiered classification of Return on Ad Spend.
Platform-agnostic: Works for TikTok, Shopee, Lazada, etc.
Following AGENTS.MD: Clean Code, DRY, SRP.
"""

from dataclasses import dataclass
from enum import Enum
import numpy as np


class RoasTier(Enum):
    """ROAS tier classifications."""
    EXCELLENT = "EXCELLENT"    # > 5.0x
    GOOD = "GOOD"              # 3.0x - 5.0x
    MODERATE = "MODERATE"      # 2.0x - 3.0x
    MARGINAL = "MARGINAL"      # 1.0x - 2.0x
    LOSS = "LOSS"              # < 1.0x


class RecommendedAction(Enum):
    """Generic recommended actions based on ROAS."""
    SCALE_UP_AGGRESSIVE = "SCALE_UP_AGGRESSIVE"
    SCALE_UP_MODERATE = "SCALE_UP_MODERATE"
    MAINTAIN = "MAINTAIN"
    REDUCE_BUDGET = "REDUCE_BUDGET"
    STOP_IMMEDIATELY = "STOP_IMMEDIATELY"
    MONITOR_CLOSELY = "MONITOR_CLOSELY"


@dataclass
class RoasThresholds:
    """Configurable ROAS thresholds."""
    excellent: float = 5.0
    good: float = 3.0
    moderate: float = 2.0
    marginal: float = 1.0


@dataclass
class RoasResult:
    """ROAS classification result."""
    roas: float
    tier: RoasTier
    action: RecommendedAction
    score: float
    description: str


class RoasClassifier:
    """
    Classifies ROAS into tiers and provides recommendations.
    Thresholds are configurable per platform.
    """
    
    def __init__(self, thresholds: RoasThresholds = None):
        self.thresholds = thresholds or RoasThresholds()
    
    def classify(self, revenue: float, cost: float) -> RoasResult:
        """
        Classify ROAS from revenue and cost.
        
        Args:
            revenue: Total gross revenue
            cost: Total ad spend
        
        Returns:
            RoasResult with tier, action, and score
        """
        roas = self._calculateRoas(revenue, cost)
        tier = self._getTier(roas)
        action = self._getAction(tier)
        score = self._getScore(roas)
        description = self._getDescription(tier, roas)
        
        return RoasResult(
            roas=round(roas, 2),
            tier=tier,
            action=action,
            score=round(score, 1),
            description=description
        )
    
    def classifyFromRoas(self, roas: float) -> RoasResult:
        """Classify from pre-calculated ROAS value."""
        tier = self._getTier(roas)
        action = self._getAction(tier)
        score = self._getScore(roas)
        description = self._getDescription(tier, roas)
        
        return RoasResult(
            roas=round(roas, 2),
            tier=tier,
            action=action,
            score=round(score, 1),
            description=description
        )
    
    def _calculateRoas(self, revenue: float, cost: float) -> float:
        """Calculate ROAS with safe division."""
        if cost <= 0:
            return 0.0
        return revenue / cost
    
    def _getTier(self, roas: float) -> RoasTier:
        """Get ROAS tier from value."""
        if roas >= self.thresholds.excellent:
            return RoasTier.EXCELLENT
        elif roas >= self.thresholds.good:
            return RoasTier.GOOD
        elif roas >= self.thresholds.moderate:
            return RoasTier.MODERATE
        elif roas >= self.thresholds.marginal:
            return RoasTier.MARGINAL
        else:
            return RoasTier.LOSS
    
    def _getAction(self, tier: RoasTier) -> RecommendedAction:
        """Get recommended action from tier."""
        actionMap = {
            RoasTier.EXCELLENT: RecommendedAction.SCALE_UP_AGGRESSIVE,
            RoasTier.GOOD: RecommendedAction.SCALE_UP_MODERATE,
            RoasTier.MODERATE: RecommendedAction.MAINTAIN,
            RoasTier.MARGINAL: RecommendedAction.REDUCE_BUDGET,
            RoasTier.LOSS: RecommendedAction.STOP_IMMEDIATELY,
        }
        return actionMap.get(tier, RecommendedAction.MONITOR_CLOSELY)
    
    def _getScore(self, roas: float) -> float:
        """
        Convert ROAS to score (0-100).
        Uses logarithmic scaling for better distribution.
        """
        if roas <= 0:
            return 0
        
        # Logarithmic scaling
        baseScore = 30 + 25 * np.log10(max(roas, 0.1))
        
        # Bonus for high ROAS
        if roas >= 10:
            baseScore += 20
        elif roas >= 5:
            baseScore += 10
        
        return min(100, max(0, baseScore))
    
    def _getDescription(self, tier: RoasTier, roas: float) -> str:
        """Get human-readable description."""
        descriptions = {
            RoasTier.EXCELLENT: f"ROAS {roas:.1f}x excellent - aggressive scale up",
            RoasTier.GOOD: f"ROAS {roas:.1f}x good - moderate scale up",
            RoasTier.MODERATE: f"ROAS {roas:.1f}x moderate - maintain and optimize",
            RoasTier.MARGINAL: f"ROAS {roas:.1f}x marginal - reduce budget",
            RoasTier.LOSS: f"ROAS {roas:.1f}x loss - stop or pivot",
        }
        return descriptions.get(tier, f"ROAS: {roas:.1f}x")


# Convenience functions
def classify_roas(revenue: float, cost: float, 
                  thresholds: RoasThresholds = None) -> RoasResult:
    """Classify ROAS from revenue and cost."""
    classifier = RoasClassifier(thresholds)
    return classifier.classify(revenue, cost)


def get_roas_tier(roas: float) -> str:
    """Get ROAS tier string."""
    classifier = RoasClassifier()
    result = classifier.classifyFromRoas(roas)
    return result.tier.value


def get_roas_score(roas: float) -> float:
    """Get ROAS score (0-100)."""
    classifier = RoasClassifier()
    result = classifier.classifyFromRoas(roas)
    return result.score
