#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Probability Engine (Re-exported from Core + TikTok Extensions)
==============================================================
Re-exports core probability calculations and adds TikTok-specific wrappers.
Base logic is in `notebooks/core/intelligence/probability.py`.
"""

import sys
from pathlib import Path
from dataclasses import dataclass
from typing import List, Optional, Tuple
import pandas as pd
import numpy as np

# Add parent directory to sys.path for core imports
_parent_dir = Path(__file__).resolve().parent.parent.parent
if str(_parent_dir) not in sys.path:
    sys.path.insert(0, str(_parent_dir))

# Re-export from core
from core.intelligence.probability import (
    calculate_success_probability,
    calculate_failure_risk,
    bayesian_update,
    calculate_monte_carlo_projection,
)

from .config_intelligence import IntelligenceConfig


@dataclass
class ProbabilityResult:
    """Probability analysis result for TikTok."""
    successProbability: float    # P(ROAS > threshold)
    failureProbability: float    # P(ROAS < breakeven)
    confidenceLevel: str         # HIGH, MEDIUM, LOW
    ciLower: float              # 95% CI lower bound
    ciUpper: float              # 95% CI upper bound
    expectedRoas: float         # Expected ROAS
    description: str


class ProbabilityEngine:
    """
    TikTok-specific probability engine wrapping core logic.
    Bayesian probability estimation for ROAS performance.
    """
    
    def __init__(self, config: IntelligenceConfig = None):
        self.config = config or IntelligenceConfig()
    
    def analyze(
        self, 
        roasHistory: pd.Series,
        threshold: float = 2.0
    ) -> ProbabilityResult:
        """
        Calculate success probability based on ROAS history.
        
        Args:
            roasHistory: Historical ROAS values
            threshold: Success threshold (default 2.0 = 2x ROAS)
        
        Returns:
            ProbabilityResult with probability metrics
        """
        roas_values = roasHistory.dropna().tolist()
        
        if len(roas_values) < 3:
            return self._insufficientData()
        
        # Use core probability calculation
        success_result = calculate_success_probability(
            recent_roas=roas_values,
            threshold=threshold
        )
        
        failure_result = calculate_failure_risk(
            recent_roas=roas_values,
            threshold=1.0  # Breakeven
        )
        
        # Calculate confidence interval using Monte Carlo
        mc_result = calculate_monte_carlo_projection(
            historical_values=roas_values,
            n_days=1,  # Next day projection
            n_simulations=1000
        )
        
        # Determine confidence level
        n = len(roas_values)
        if n >= 14:
            confidence_level = "HIGH"
        elif n >= 7:
            confidence_level = "MEDIUM"
        else:
            confidence_level = "LOW"
        
        success_prob = success_result["probability"]
        failure_prob = 1 - calculate_success_probability(roas_values, threshold=1.0)["probability"]
        expected_roas = np.mean(roas_values)
        
        return ProbabilityResult(
            successProbability=round(success_prob * 100, 1),
            failureProbability=round(failure_prob * 100, 1),
            confidenceLevel=confidence_level,
            ciLower=mc_result.get("ci_lower", expected_roas * 0.8),
            ciUpper=mc_result.get("ci_upper", expected_roas * 1.2),
            expectedRoas=round(expected_roas, 2),
            description=self._getDescription(success_prob, confidence_level)
        )
    
    def _getDescription(self, successProb: float, confidence: str) -> str:
        """Get human-readable description."""
        if successProb >= 0.8:
            return f"✅ Sangat likely sukses ({successProb*100:.0f}%) - {confidence} confidence"
        elif successProb >= 0.6:
            return f"🔶 Likely sukses ({successProb*100:.0f}%) - {confidence} confidence"
        elif successProb >= 0.4:
            return f"⚠️ Moderate chance ({successProb*100:.0f}%) - {confidence} confidence"
        else:
            return f"❌ Likely gagal ({successProb*100:.0f}%) - {confidence} confidence"
    
    def _insufficientData(self) -> ProbabilityResult:
        """Return insufficient data result."""
        return ProbabilityResult(
            successProbability=50,
            failureProbability=50,
            confidenceLevel="LOW",
            ciLower=0,
            ciUpper=0,
            expectedRoas=0,
            description="Data tidak cukup untuk analisis probabilitas"
        )


__all__ = [
    "ProbabilityEngine",
    "ProbabilityResult",
    "calculate_success_probability",
    "calculate_failure_risk",
    "bayesian_update",
    "calculate_monte_carlo_projection",
]
