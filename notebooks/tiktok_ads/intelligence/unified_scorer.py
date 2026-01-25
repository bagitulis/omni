#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Unified Scorer (Re-exported from Core + TikTok Extensions)
==========================================================
Re-exports core unified scorer and adds TikTok-specific formatting.
Base logic is in `notebooks/core/intelligence/unified_scorer.py`.
"""

import sys
from pathlib import Path
from typing import List

# Add parent directory to sys.path for core imports
_parent_dir = Path(__file__).resolve().parent.parent.parent
if str(_parent_dir) not in sys.path:
    sys.path.insert(0, str(_parent_dir))

# Re-export from core
from core.intelligence.unified_types import (
    ActionRecommendation,
    ScoreCategory,
    ScoringWeightsConfig,
    ActionThresholds,
    UnifiedScoreResult,
    ScoreInputs,
)
from core.intelligence.unified_scorer import UnifiedScorer as CoreUnifiedScorer

# TikTok-specific alias for backward compatibility  
UnifiedScore = UnifiedScoreResult


def get_category_emoji(category: ScoreCategory) -> str:
    """Get emoji for score category."""
    emojis = {
        ScoreCategory.STAR: "⭐",
        ScoreCategory.GROWTH: "🚀",
        ScoreCategory.STABLE: "✅",
        ScoreCategory.WATCH: "👀",
        ScoreCategory.PROBLEM: "⚠️",
    }
    return emojis.get(category, "")


def get_category_label(category: ScoreCategory) -> str:
    """Get Indonesian label for category."""
    labels = {
        ScoreCategory.STAR: "⭐ STAR",
        ScoreCategory.GROWTH: "🚀 GROWTH",
        ScoreCategory.STABLE: "✅ STABLE",
        ScoreCategory.WATCH: "👀 WATCH",
        ScoreCategory.PROBLEM: "⚠️ PROBLEM",
    }
    return labels.get(category, str(category))


def get_action_label(action: ActionRecommendation) -> str:
    """Get Indonesian action label."""
    labels = {
        ActionRecommendation.SCALE_UP_AGGRESSIVE: "🚀 SCALE UP AGRESIF (+30%)",
        ActionRecommendation.SCALE_UP: "📈 SCALE UP (+10-30%)",
        ActionRecommendation.MAINTAIN: "✅ MAINTAIN (0%)",
        ActionRecommendation.REDUCE: "📉 REDUCE (-10-30%)",
        ActionRecommendation.STOP: "⛔ STOP (-100%)",
    }
    return labels.get(action, str(action))


class UnifiedScorer(CoreUnifiedScorer):
    """
    TikTok-specific unified scorer with backward-compatible API.
    Wraps core UnifiedScorer but accepts keyword arguments directly.
    """
    
    def calculate(self, 
                  productId: str,
                  productName: str,
                  roas: float,
                  mkTrend: str = "",
                  mkPValue: float = 1.0,
                  trendDirection = None,
                  trendRoc: float = 0.0,
                  legacyMomentumPct: float = 0.0,
                  elasticity: float = 0.0,
                  saturationStatus: str = "",
                  marginalRoas: float = 0.0,
                  cv: float = 0.0,
                  intelVolatilityScore: float = 50.0,
                  fatigueStatus: str = "",
                  fatiguePct: float = 0.0,
                  churnRiskScore: float = 0.0,
                  eventMultiplier: float = 1.0,
                  eventNames: List[str] = None,
                  nPeriods: int = 0,
                  dataQuality: str = "MEDIUM") -> UnifiedScoreResult:
        """
        Calculate unified score (backward-compatible API).
        Accepts keyword arguments and converts to ScoreInputs.
        """
        # Convert trendDirection enum to string if needed
        trendDirStr = ""
        if trendDirection is not None:
            if hasattr(trendDirection, 'value'):
                trendDirStr = trendDirection.value
            elif hasattr(trendDirection, 'name'):
                # Check for UP/DOWN in the name
                name = trendDirection.name
                if "UP" in name:
                    trendDirStr = "UP"
                elif "DOWN" in name:
                    trendDirStr = "DOWN"
                else:
                    trendDirStr = "STABLE"
            else:
                trendDirStr = str(trendDirection)
        
        # Create ScoreInputs
        inputs = ScoreInputs(
            productId=productId,
            productName=productName,
            roas=roas,
            mkTrend=mkTrend,
            mkPValue=mkPValue,
            trendDirection=trendDirStr,
            trendRoc=trendRoc,
            legacyMomentumPct=legacyMomentumPct,
            elasticity=elasticity,
            saturationStatus=saturationStatus,
            marginalRoas=marginalRoas,
            cv=cv,
            intelVolatilityScore=intelVolatilityScore,
            fatigueStatus=fatigueStatus,
            fatiguePct=fatiguePct,
            churnRiskScore=churnRiskScore,
            eventMultiplier=eventMultiplier,
            eventNames=eventNames or [],
            nPeriods=nPeriods,
            dataQuality=dataQuality,
        )
        
        # Call parent method
        return super().calculate(inputs)


__all__ = [
    # Core exports
    "ActionRecommendation",
    "ScoreCategory",
    "ScoringWeightsConfig",
    "ActionThresholds",
    "UnifiedScoreResult",
    "ScoreInputs",
    "UnifiedScorer",
    # TikTok-specific
    "UnifiedScore",  # Alias
    "get_category_emoji",
    "get_category_label",
    "get_action_label",
]
