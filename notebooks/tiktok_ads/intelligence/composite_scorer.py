#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Composite Scorer (Re-exported from Core + TikTok Extensions)
============================================================
Re-exports core scoring and adds TikTok-specific formatting.
Base logic is in `notebooks/core/intelligence/scoring.py`.
"""

import sys
from pathlib import Path
from dataclasses import dataclass
from typing import Dict, List
from enum import Enum

# Add parent directory to sys.path for core imports
_parent_dir = Path(__file__).resolve().parent.parent.parent
if str(_parent_dir) not in sys.path:
    sys.path.insert(0, str(_parent_dir))

# Re-export from core
from core.intelligence.scoring import (
    CompositeScorer as CoreCompositeScorer,
    ScoringWeights,
    calculate_unified_score,
    normalize_metric,
    get_risk_level,
    get_recommendation_action,
)

# TikTok-specific config
from .config_intelligence import IntelligenceConfig, RecommendationAction


class ScoreCategory(Enum):
    """Final score categories for TikTok."""
    STAR = "STAR"                    # 85-100
    PERFORMER = "PERFORMER"          # 70-84
    AVERAGE = "AVERAGE"              # 55-69
    UNDERPERFORMER = "UNDERPERFORMER"  # 40-54
    PROBLEM = "PROBLEM"              # 25-39
    STOP = "STOP"                    # 0-24


@dataclass
class ComponentScore:
    """Individual component score."""
    name: str
    score: float
    weight: float
    weightedScore: float
    description: str


@dataclass
class CompositeResult:
    """Composite scoring result for TikTok."""
    finalScore: float
    category: ScoreCategory
    action: RecommendationAction
    components: List[ComponentScore]
    breakdown: Dict[str, float]
    description: str


class CompositeScorer(CoreCompositeScorer):
    """
    TikTok-specific composite scorer with Indonesian descriptions.
    Combines all analysis scores into a final composite score.
    """
    
    def __init__(self, config: IntelligenceConfig = None):
        self.config = config or IntelligenceConfig()
        super().__init__()
    
    def score(self, 
              roasScore: float,
              trendScore: float,
              volatilityScore: float,
              fatigueScore: float,
              elasticityScore: float,
              eventScore: float = 50) -> CompositeResult:
        """
        Calculate composite score from individual components.
        
        Default weights:
        - ROAS: 30%
        - Trend: 20%
        - Volatility: 15%
        - Fatigue: 15%
        - Elasticity: 10%
        - Event: 10%
        """
        weights = {
            'roas': 0.30,
            'trend': 0.20,
            'volatility': 0.15,
            'fatigue': 0.15,
            'elasticity': 0.10,
            'event': 0.10
        }
        
        scores = {
            'roas': roasScore,
            'trend': trendScore,
            'volatility': volatilityScore,
            'fatigue': fatigueScore,
            'elasticity': elasticityScore,
            'event': eventScore
        }
        
        # Calculate weighted score
        finalScore = sum(scores[k] * weights[k] for k in weights)
        
        # Build components
        components = [
            ComponentScore(
                name=k.upper(),
                score=scores[k],
                weight=weights[k],
                weightedScore=scores[k] * weights[k],
                description=self._getComponentDescription(k, scores[k])
            )
            for k in weights
        ]
        
        # Determine category and action
        category = self._getCategory(finalScore)
        action = self._getAction(finalScore, roasScore)
        description = self._getDescription(category, finalScore)
        
        return CompositeResult(
            finalScore=round(finalScore, 1),
            category=category,
            action=action,
            components=components,
            breakdown=scores,
            description=description
        )
    
    def _getCategory(self, score: float) -> ScoreCategory:
        """Get score category."""
        if score >= 85:
            return ScoreCategory.STAR
        elif score >= 70:
            return ScoreCategory.PERFORMER
        elif score >= 55:
            return ScoreCategory.AVERAGE
        elif score >= 40:
            return ScoreCategory.UNDERPERFORMER
        elif score >= 25:
            return ScoreCategory.PROBLEM
        else:
            return ScoreCategory.STOP
    
    def _getAction(self, score: float, roasScore: float) -> RecommendationAction:
        """Get recommended action."""
        # ROAS override
        if roasScore < 20:
            return RecommendationAction.STOP_IMMEDIATELY
        
        if score >= 80:
            return RecommendationAction.SCALE_UP_AGGRESSIVE
        elif score >= 65:
            return RecommendationAction.SCALE_UP_MODERATE
        elif score >= 50:
            return RecommendationAction.MAINTAIN
        elif score >= 35:
            return RecommendationAction.REDUCE_BUDGET
        else:
            return RecommendationAction.STOP_IMMEDIATELY
    
    def _getComponentDescription(self, component: str, score: float) -> str:
        """Get Indonesian description for component."""
        if score >= 80:
            return "Sangat baik"
        elif score >= 60:
            return "Baik"
        elif score >= 40:
            return "Cukup"
        elif score >= 20:
            return "Kurang"
        else:
            return "Buruk"
    
    def _getDescription(self, category: ScoreCategory, score: float) -> str:
        """Get Indonesian description for category."""
        descriptions = {
            ScoreCategory.STAR: f"⭐ STAR ({score:.0f}) - Produk unggulan, scale up agresif!",
            ScoreCategory.PERFORMER: f"🚀 PERFORMER ({score:.0f}) - Performa bagus, scale up moderat.",
            ScoreCategory.AVERAGE: f"✅ AVERAGE ({score:.0f}) - Rata-rata, maintain dan optimasi.",
            ScoreCategory.UNDERPERFORMER: f"👀 UNDERPERFORMER ({score:.0f}) - Di bawah rata-rata, monitor.",
            ScoreCategory.PROBLEM: f"⚠️ PROBLEM ({score:.0f}) - Perlu perbaikan segera.",
            ScoreCategory.STOP: f"⛔ STOP ({score:.0f}) - Stop iklan, evaluasi total.",
        }
        return descriptions.get(category, f"Score: {score:.0f}")


__all__ = [
    # Core exports
    "ScoringWeights",
    "calculate_unified_score",
    "normalize_metric",
    "get_risk_level",
    "get_recommendation_action",
    # TikTok-specific
    "CompositeScorer",
    "CompositeResult",
    "ComponentScore",
    "ScoreCategory",
]
