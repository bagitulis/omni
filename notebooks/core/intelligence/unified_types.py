#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Unified Scorer (Generic Core Module)
====================================
Combines multiple scoring methodologies into a composite score.
Platform-agnostic: Works for TikTok, Shopee, Lazada, etc.
Following AGENTS.MD: Clean Code, DRY, SRP.

This file is split into 2 parts for SRP:
- Part 1: Data classes and enums (this file)
- Part 2: UnifiedScorer class (unified_scorer_engine.py)
"""

from dataclasses import dataclass, field
from typing import Dict, List
from enum import Enum


class ActionRecommendation(Enum):
    """Final action recommendations."""
    SCALE_UP_AGGRESSIVE = "SCALE_UP_AGGRESSIVE"
    SCALE_UP = "SCALE_UP"
    MAINTAIN = "MAINTAIN"
    REDUCE = "REDUCE"
    STOP = "STOP"


class ScoreCategory(Enum):
    """Product score categories."""
    STAR = "STAR"           # 80+
    GROWTH = "GROWTH"       # 65-79
    STABLE = "STABLE"       # 50-64
    WATCH = "WATCH"         # 35-49
    PROBLEM = "PROBLEM"     # <35


@dataclass
class ScoringWeightsConfig:
    """Configurable scoring weights (must sum to 1.0)."""
    roas: float = 0.20
    trend: float = 0.15
    momentum: float = 0.10
    elasticity: float = 0.15
    volatility: float = 0.10
    fatigue: float = 0.10
    churn: float = 0.10
    event: float = 0.10
    
    def validate(self) -> bool:
        """Validate weights sum to 1.0."""
        total = (self.roas + self.trend + self.momentum + self.elasticity +
                 self.volatility + self.fatigue + self.churn + self.event)
        return abs(total - 1.0) < 0.001


@dataclass
class ActionThresholds:
    """Configurable action thresholds."""
    scaleAggressive: float = 80
    scale: float = 65
    maintain: float = 50
    reduce: float = 35


@dataclass
class UnifiedScoreResult:
    """Complete unified score result."""
    productId: str
    productName: str
    
    # Individual scores (0-100)
    roasScore: float
    trendScore: float
    momentumScore: float
    elasticityScore: float
    volatilityScore: float
    fatigueScore: float
    churnRiskScore: float
    eventScore: float
    
    # Final composite
    compositeScore: float
    category: ScoreCategory
    action: ActionRecommendation
    
    # Confidence
    confidence: float
    confidenceLevel: str
    
    # Factors
    topFactors: List[str] = field(default_factory=list)
    warningFactors: List[str] = field(default_factory=list)


@dataclass
class ScoreInputs:
    """Input data for unified scoring."""
    # Identity
    productId: str
    productName: str
    
    # ROAS inputs
    roas: float
    
    # Trend inputs
    mkTrend: str = ""           # Mann-Kendall result
    mkPValue: float = 1.0
    trendDirection: str = ""    # UP, DOWN, STABLE
    trendRoc: float = 0.0       # Rate of change
    
    # Momentum inputs
    legacyMomentumPct: float = 0.0
    
    # Elasticity inputs
    elasticity: float = 0.0
    saturationStatus: str = ""
    marginalRoas: float = 0.0
    
    # Volatility inputs
    cv: float = 0.0             # Coefficient of variation
    intelVolatilityScore: float = 50.0
    
    # Fatigue inputs
    fatigueStatus: str = ""
    fatiguePct: float = 0.0
    
    # Churn Risk inputs
    churnRiskScore: float = 0.0
    
    # Event inputs
    eventMultiplier: float = 1.0
    eventNames: List[str] = field(default_factory=list)
    
    # Data quality
    nPeriods: int = 0
    dataQuality: str = "MEDIUM"
