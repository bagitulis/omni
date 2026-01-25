#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Saturation Model (Re-exported from Core + TikTok Extensions)
============================================================
Re-exports core saturation detection and adds TikTok-specific wrappers.
Base logic is in `notebooks/core/intelligence/saturation.py`.
"""

import sys
from pathlib import Path
from dataclasses import dataclass
from enum import Enum
from typing import Optional, Tuple
import pandas as pd
import numpy as np

# Add parent directory to sys.path for core imports
_parent_dir = Path(__file__).resolve().parent.parent.parent
if str(_parent_dir) not in sys.path:
    sys.path.insert(0, str(_parent_dir))

# Re-export from core
from core.intelligence.saturation import (
    SaturationModel as CoreSaturationModel,
    detect_saturation_point,
    calculate_marginal_returns,
)


class SaturationStatus(Enum):
    """Budget saturation status for TikTok."""
    HIGH_ELASTICITY = "HIGH_ELASTICITY"     # Tambah budget = profit tinggi
    MODERATE = "MODERATE"                    # Tambah budget = profit OK
    APPROACHING_SATURATION = "APPROACHING_SATURATION"  # Hati-hati
    SATURATED = "SATURATED"                  # Jangan tambah
    OVER_SATURATED = "OVER_SATURATED"        # Kurangi budget
    INSUFFICIENT_DATA = "INSUFFICIENT_DATA"


@dataclass
class SaturationResult:
    """Saturation analysis result for TikTok."""
    status: SaturationStatus
    marginalRoas: float      # Marginal ROAS at current spend
    saturationPoint: float   # Estimated saturation budget
    currentSpend: float
    headroom: float          # % budget dapat ditambah
    score: float             # 0-100
    description: str


class SaturationModel:
    """
    TikTok-specific saturation model wrapping core logic.
    Models diminishing returns using logistic growth function.
    Answers: "If I add Rp X to budget, how much more revenue?"
    """
    
    def __init__(self, maxBudget: int = 2_000_000):
        self.maxBudget = maxBudget
        self.core_model = CoreSaturationModel()
    
    def analyze(
        self, 
        spendHistory: pd.Series, 
        revenueHistory: pd.Series
    ) -> SaturationResult:
        """
        Analyze spend-revenue relationship to find saturation.
        
        Args:
            spendHistory: Historical spend values
            revenueHistory: Corresponding revenue values
        
        Returns:
            SaturationResult with saturation status and metrics
        """
        # Clean data
        spend = spendHistory.dropna().tolist()
        revenue = revenueHistory.dropna().tolist()
        
        if len(spend) < 5 or len(revenue) < 5:
            return self._insufficientData()
        
        # Ensure same length
        min_len = min(len(spend), len(revenue))
        spend = spend[:min_len]
        revenue = revenue[:min_len]
        
        # Use core saturation detection
        core_result = detect_saturation_point(spend, revenue)
        
        if not core_result.get("is_saturated") and core_result.get("saturation_spend", 0) == 0:
            return self._insufficientData()
        
        # Calculate metrics
        current_spend = spend[-1] if spend else 0
        saturation_point = core_result.get("saturation_spend", self.maxBudget)
        
        # Calculate marginal ROAS
        marginals = calculate_marginal_returns(spend, revenue)
        if marginals["marginal_returns"]:
            marginal_roas = marginals["marginal_returns"][-1]["marginal_return"]
        else:
            marginal_roas = 1.0
        
        # Calculate headroom
        if saturation_point > 0 and current_spend > 0:
            headroom = ((saturation_point - current_spend) / current_spend) * 100
        else:
            headroom = 100
        
        # Determine status
        status = self._determineStatus(marginal_roas, headroom, core_result.get("is_saturated", False))
        score = self._calculateScore(status, marginal_roas, headroom)
        
        return SaturationResult(
            status=status,
            marginalRoas=round(marginal_roas, 2),
            saturationPoint=round(saturation_point, 0),
            currentSpend=round(current_spend, 0),
            headroom=round(max(0, headroom), 1),
            score=round(score, 1),
            description=self._getDescription(status, marginal_roas, headroom)
        )
    
    def _determineStatus(
        self, 
        marginalRoas: float, 
        headroom: float, 
        isSaturated: bool
    ) -> SaturationStatus:
        """Determine saturation status."""
        if isSaturated or headroom <= 0:
            return SaturationStatus.OVER_SATURATED
        elif headroom < 10:
            return SaturationStatus.SATURATED
        elif headroom < 30:
            return SaturationStatus.APPROACHING_SATURATION
        elif marginalRoas >= 2:
            return SaturationStatus.HIGH_ELASTICITY
        else:
            return SaturationStatus.MODERATE
    
    def _calculateScore(
        self, 
        status: SaturationStatus, 
        marginalRoas: float, 
        headroom: float
    ) -> float:
        """Calculate saturation score (0-100, higher = more room to grow)."""
        base_scores = {
            SaturationStatus.HIGH_ELASTICITY: 90,
            SaturationStatus.MODERATE: 70,
            SaturationStatus.APPROACHING_SATURATION: 50,
            SaturationStatus.SATURATED: 30,
            SaturationStatus.OVER_SATURATED: 10,
            SaturationStatus.INSUFFICIENT_DATA: 50
        }
        return base_scores.get(status, 50)
    
    def _getDescription(
        self, 
        status: SaturationStatus, 
        marginalRoas: float, 
        headroom: float
    ) -> str:
        """Get human-readable description."""
        descs = {
            SaturationStatus.HIGH_ELASTICITY: 
                f"✅ High elasticity! Budget dapat ditambah {headroom:.0f}% (MROAS: {marginalRoas:.1f}x)",
            SaturationStatus.MODERATE: 
                f"🔶 Moderate response. Headroom {headroom:.0f}%",
            SaturationStatus.APPROACHING_SATURATION: 
                f"⚠️ Mendekati saturasi. Hanya {headroom:.0f}% headroom tersisa",
            SaturationStatus.SATURATED: 
                f"🔴 Saturated! Marginal ROAS hanya {marginalRoas:.1f}x",
            SaturationStatus.OVER_SATURATED: 
                f"❌ Over-saturated! Kurangi budget",
            SaturationStatus.INSUFFICIENT_DATA: 
                "❓ Data tidak cukup"
        }
        return descs.get(status, "Unknown")
    
    def _insufficientData(self) -> SaturationResult:
        """Return insufficient data result."""
        return SaturationResult(
            status=SaturationStatus.INSUFFICIENT_DATA,
            marginalRoas=0,
            saturationPoint=0,
            currentSpend=0,
            headroom=0,
            score=50,
            description="Data tidak cukup untuk analisis saturasi"
        )


__all__ = [
    "SaturationModel",
    "SaturationStatus",
    "SaturationResult",
    "detect_saturation_point",
    "calculate_marginal_returns",
]
