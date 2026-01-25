#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Intelligence Configuration
==========================
Centralized configuration for all thresholds, weights, and constants.
Following AGENTS.MD: Clean Code, DRY, SRP.
"""

from dataclasses import dataclass, field
from typing import Dict, List
from enum import Enum


class RecommendationAction(Enum):
    """Possible recommendation actions."""
    SCALE_UP_AGGRESSIVE = "SCALE_UP_AGGRESSIVE"
    SCALE_UP_MODERATE = "SCALE_UP_MODERATE"
    MAINTAIN = "MAINTAIN"
    REDUCE_BUDGET = "REDUCE_BUDGET"
    STOP_IMMEDIATELY = "STOP_IMMEDIATELY"
    MONITOR_CLOSELY = "MONITOR_CLOSELY"


class RiskLevel(Enum):
    """Risk level classifications."""
    LOW = "LOW"
    MODERATE = "MODERATE"
    MEDIUM = "MEDIUM"
    HIGH = "HIGH"
    EXTREME = "EXTREME"
    CRITICAL = "CRITICAL"
    UNKNOWN = "UNKNOWN"


class ProductStage(Enum):
    """Product lifecycle stages."""
    LAUNCH = "LAUNCH"
    GROWTH = "GROWTH"
    MATURE = "MATURE"
    DECLINE = "DECLINE"


@dataclass
class RoasThresholds:
    """ROAS classification thresholds."""
    excellent: float = 5.0
    good: float = 3.0
    moderate: float = 2.0
    marginal: float = 1.0


@dataclass
class BudgetConfig:
    """Budget-related configurations."""
    maxBudgetPerProduct: int = 2_000_000  # Rp 2 juta max per produk
    scaleUpAggressive: float = 0.30       # +30%
    scaleUpModerate: float = 0.15         # +15%
    reduceBudget: float = -0.20           # -20%
    reduceAggressiveBudget: float = -0.50 # -50%


@dataclass
class ScoringWeights:
    """Weights for composite scoring."""
    roas: float = 0.25
    trend: float = 0.20
    elasticity: float = 0.15
    volatility: float = 0.15
    fatigue: float = 0.10
    indonesianContext: float = 0.15


@dataclass
class TrendConfig:
    """Trend analysis configurations."""
    shortWindow: int = 7    # EMA short period (days/periods)
    longWindow: int = 30    # EMA long period
    minPeriodsForTrend: int = 4


@dataclass
class VolatilityThresholds:
    """Volatility risk thresholds (Coefficient of Variation %).
    NOTE: TikTok Ads typically have high CV (100-300%), adjusted accordingly."""
    lowRisk: float = 100.0     # TikTok: <100% is actually stable
    mediumRisk: float = 200.0  # TikTok: 100-200% is normal


@dataclass
class FatigueThresholds:
    """Creative fatigue thresholds (CTR decay %).
    NOTE: TikTok CTR fluctuates heavily, adjusted to be more lenient."""
    fresh: float = 20.0       # <20% decay = still fresh
    aging: float = 40.0       # 20-40% = aging
    fatigued: float = 60.0    # 40-60% = fatigued (but not dead yet)


@dataclass
class LifecycleThresholds:
    """Product lifecycle thresholds (in days)."""
    launchPeriod: int = 60      # 0-60 days = LAUNCH
    growthPeriod: int = 120     # 60-120 days = GROWTH
    maturePeriod: int = 180     # 120-180 days = MATURE
    declineThreshold: float = -0.30  # -30% trend = DECLINE


@dataclass
class ConfidenceThresholds:
    """Data confidence thresholds."""
    highConfidenceDays: int = 90
    mediumConfidenceDays: int = 30
    lowConfidenceDays: int = 7


@dataclass
class ScoreThresholds:
    """Final score thresholds for recommendations."""
    star: int = 85          # 85-100: STAR
    performer: int = 70     # 70-84: PERFORMER
    average: int = 55       # 55-69: AVERAGE
    underperformer: int = 40  # 40-54: UNDERPERFORMER
    problem: int = 25       # 25-39: PROBLEM
    # 0-24: STOP


class IntelligenceConfig:
    """
    Main configuration container for the Intelligence System.
    Singleton pattern to ensure consistent config across modules.
    """
    _instance = None
    
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance._initialize()
        return cls._instance
    
    def _initialize(self):
        """Initialize all configuration objects."""
        self.roas = RoasThresholds()
        self.budget = BudgetConfig()
        self.weights = ScoringWeights()
        self.trend = TrendConfig()
        self.volatility = VolatilityThresholds()
        self.fatigue = FatigueThresholds()
        self.lifecycle = LifecycleThresholds()
        self.confidence = ConfidenceThresholds()
        self.scores = ScoreThresholds()
    
    def getActionFromScore(self, score: float) -> RecommendationAction:
        """Get recommended action based on final score."""
        if score >= self.scores.star:
            return RecommendationAction.SCALE_UP_AGGRESSIVE
        elif score >= self.scores.performer:
            return RecommendationAction.SCALE_UP_MODERATE
        elif score >= self.scores.average:
            return RecommendationAction.MAINTAIN
        elif score >= self.scores.underperformer:
            return RecommendationAction.REDUCE_BUDGET
        elif score >= self.scores.problem:
            return RecommendationAction.MONITOR_CLOSELY
        else:
            return RecommendationAction.STOP_IMMEDIATELY
    
    def getRiskLevel(self, volatilityCv: float) -> RiskLevel:
        """Get risk level from volatility coefficient of variation."""
        if volatilityCv < self.volatility.lowRisk:
            return RiskLevel.LOW
        elif volatilityCv < self.volatility.mediumRisk:
            return RiskLevel.MEDIUM
        else:
            return RiskLevel.HIGH
