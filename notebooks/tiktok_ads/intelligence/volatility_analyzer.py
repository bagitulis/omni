#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Volatility Analyzer (Re-exported from Core + TikTok Extensions)
===============================================================
Re-exports core variation analysis and adds TikTok-specific formatting.
Base logic is in `notebooks/core/statistics/variation.py`.
"""

import sys
from pathlib import Path
from dataclasses import dataclass
from enum import Enum
import pandas as pd
import numpy as np

# Add parent directory to sys.path for core imports
_parent_dir = Path(__file__).resolve().parent.parent.parent
if str(_parent_dir) not in sys.path:
    sys.path.insert(0, str(_parent_dir))

# Re-export from core
from core.statistics.variation import coefficient_of_variation, cv_label

# TikTok-specific config
from .config_intelligence import IntelligenceConfig, RiskLevel


@dataclass
class VolatilityResult:
    """Volatility analysis result for TikTok."""
    cv: float               # Coefficient of Variation (%)
    stdDev: float          # Standard Deviation
    mean: float            # Mean value
    riskLevel: RiskLevel
    score: float           # 0-100 score (higher = more stable)
    description: str


class VolatilityAnalyzer:
    """
    Analyzes ROAS volatility to measure risk.
    Uses Coefficient of Variation (CV = StdDev / Mean * 100).
    
    NOTE: TikTok Ads naturally have high CV (100-300% is normal).
    Thresholds are adjusted accordingly.
    """
    
    # TikTok-adjusted thresholds (standard CV thresholds are too strict)
    CV_LOW = 100       # <100% is actually very stable for TikTok
    CV_MODERATE = 200  # 100-200% is normal
    CV_HIGH = 400      # 200-400% is moderate risk
    CV_EXTREME = 800   # >800% is genuinely problematic
    
    def __init__(self, config: IntelligenceConfig = None):
        self.config = config or IntelligenceConfig()
    
    def analyze(self, values: pd.Series) -> VolatilityResult:
        """
        Analyze volatility from series of values.
        
        Args:
            values: Series of metric values (e.g., ROAS per period)
        
        Returns:
            VolatilityResult with CV, risk level, and score
        """
        values = values.dropna()
        
        if len(values) < 2:
            return VolatilityResult(
                cv=0,
                stdDev=0,
                mean=0,
                riskLevel=RiskLevel.UNKNOWN,
                score=50,
                description="Data tidak cukup untuk analisis volatilitas"
            )
        
        mean = values.mean()
        stdDev = values.std()
        
        # Use core function
        cv = coefficient_of_variation(values)
        
        # Get risk level (TikTok-adjusted)
        riskLevel = self._getRiskLevel(cv)
        
        # Calculate score (higher = more stable)
        score = self._calculateScore(cv)
        
        # Generate description
        description = self._getDescription(cv, riskLevel)
        
        return VolatilityResult(
            cv=round(cv, 1),
            stdDev=round(stdDev, 2),
            mean=round(mean, 2),
            riskLevel=riskLevel,
            score=round(score, 1),
            description=description
        )
    
    def _getRiskLevel(self, cv: float) -> RiskLevel:
        """Get risk level from CV (TikTok-adjusted)."""
        if cv < self.CV_LOW:
            return RiskLevel.LOW
        elif cv < self.CV_MODERATE:
            return RiskLevel.MODERATE
        elif cv < self.CV_HIGH:
            return RiskLevel.HIGH
        elif cv < self.CV_EXTREME:
            return RiskLevel.HIGH
        else:
            return RiskLevel.EXTREME
    
    def _calculateScore(self, cv: float) -> float:
        """Convert CV to stability score (0-100, higher = more stable)."""
        if cv < self.CV_LOW:
            return 90 + (self.CV_LOW - cv) / self.CV_LOW * 10
        elif cv < self.CV_MODERATE:
            return 70 + (self.CV_MODERATE - cv) / (self.CV_MODERATE - self.CV_LOW) * 20
        elif cv < self.CV_HIGH:
            return 40 + (self.CV_HIGH - cv) / (self.CV_HIGH - self.CV_MODERATE) * 30
        elif cv < self.CV_EXTREME:
            return 20 + (self.CV_EXTREME - cv) / (self.CV_EXTREME - self.CV_HIGH) * 20
        else:
            return max(0, 20 - (cv - self.CV_EXTREME) / 100)
    
    def _getDescription(self, cv: float, riskLevel: RiskLevel) -> str:
        """Get Indonesian description."""
        descriptions = {
            RiskLevel.LOW: f"📊 Volatilitas rendah (CV: {cv:.0f}%) - Sangat stabil",
            RiskLevel.MODERATE: f"📈 Volatilitas normal (CV: {cv:.0f}%) - Normal untuk TikTok",
            RiskLevel.HIGH: f"📉 Volatilitas tinggi (CV: {cv:.0f}%) - Monitor ketat",
            RiskLevel.EXTREME: f"⚠️ Volatilitas ekstrem (CV: {cv:.0f}%) - Risiko tinggi",
            RiskLevel.UNKNOWN: "❓ Data tidak cukup",
        }
        return descriptions.get(riskLevel, f"CV: {cv:.0f}%")


__all__ = [
    # Core exports
    "coefficient_of_variation",
    "cv_label",
    # TikTok-specific
    "VolatilityAnalyzer",
    "VolatilityResult",
]
