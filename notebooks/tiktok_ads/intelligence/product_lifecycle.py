#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Product Lifecycle (Re-exported from Core + TikTok Extensions)
=============================================================
Re-exports core lifecycle detection and adds TikTok-specific wrappers.
Base logic is in `notebooks/core/intelligence/lifecycle.py`.
"""

import sys
from pathlib import Path
from dataclasses import dataclass
from enum import Enum
from datetime import datetime, timedelta
from typing import Optional
import pandas as pd
import numpy as np

# Add parent directory to sys.path for core imports
_parent_dir = Path(__file__).resolve().parent.parent.parent
if str(_parent_dir) not in sys.path:
    sys.path.insert(0, str(_parent_dir))

# Re-export from core
from core.intelligence.lifecycle import (
    ProductLifecycle as CoreProductLifecycle,
    LifecycleStage as CoreLifecycleStage,
    detect_lifecycle_stage,
    predict_lifecycle_remaining,
    calculate_churn_risk,
)

from .config_intelligence import IntelligenceConfig, ProductStage


@dataclass
class LifecycleResult:
    """Product lifecycle analysis result for TikTok."""
    stage: ProductStage
    daysActive: int
    trendDirection: str    # "UP", "DOWN", "STABLE"
    avgRoas: float
    recentRoas: float
    isEvergreen: bool      # Stabil sepanjang waktu
    score: float           # 0-100
    description: str


class ProductLifecycle:
    """
    TikTok-specific product lifecycle detector wrapping core logic.
    Determines product lifecycle stage based on historical performance.
    Stages: LAUNCH → GROWTH → MATURE → DECLINE
    """
    
    def __init__(self, config: IntelligenceConfig = None):
        self.config = config or IntelligenceConfig()
        self.core_lifecycle = CoreProductLifecycle()
    
    def analyze(self, df: pd.DataFrame) -> LifecycleResult:
        """
        Analyze product lifecycle from performance DataFrame.
        
        Args:
            df: DataFrame with 'periodStart', 'roi' columns
        
        Returns:
            LifecycleResult with stage and metrics
        """
        if len(df) < 2:
            return self._insufficientData()
        
        # Sort by date
        df = df.sort_values('periodStart').copy()
        
        # Calculate days active
        firstDate = pd.to_datetime(df['periodStart'].iloc[0])
        lastDate = pd.to_datetime(df['periodStart'].iloc[-1])
        daysActive = (lastDate - firstDate).days
        
        # Get ROI history
        roi_values = df['roi'].dropna().tolist() if 'roi' in df.columns else []
        
        if len(roi_values) < 2:
            return self._insufficientData()
        
        # Use core lifecycle detection
        core_result = self.core_lifecycle.analyze(
            sales_history=roi_values,
            days_since_launch=daysActive
        )
        
        # Calculate averages
        avgRoas = np.mean(roi_values)
        recentRoas = np.mean(roi_values[-3:]) if len(roi_values) >= 3 else roi_values[-1]
        
        # Map core stage to TikTok stage
        stage = self._mapStage(core_result["stage"])
        
        # Determine trend
        trendDirection = self._determineTrend(roi_values)
        
        # Check if evergreen (stable performance)
        cv = np.std(roi_values) / np.mean(roi_values) if np.mean(roi_values) > 0 else 1
        isEvergreen = cv < 0.3 and stage in [ProductStage.MATURE, ProductStage.GROWTH]
        
        # Calculate score
        score = self._calculateScore(stage, avgRoas, trendDirection, isEvergreen)
        
        return LifecycleResult(
            stage=stage,
            daysActive=daysActive,
            trendDirection=trendDirection,
            avgRoas=round(avgRoas, 2),
            recentRoas=round(recentRoas, 2),
            isEvergreen=isEvergreen,
            score=round(score, 1),
            description=self._getDescription(stage, daysActive, isEvergreen)
        )
    
    def _mapStage(self, core_stage: str) -> ProductStage:
        """Map core lifecycle stage to TikTok ProductStage."""
        mapping = {
            "introduction": ProductStage.LAUNCH,
            "growth": ProductStage.GROWTH,
            "maturity": ProductStage.MATURE,
            "decline": ProductStage.DECLINE,
            "end_of_life": ProductStage.DECLINE
        }
        return mapping.get(core_stage, ProductStage.LAUNCH)
    
    def _determineTrend(self, values: list) -> str:
        """Determine trend direction."""
        if len(values) < 3:
            return "STABLE"
        
        n = len(values)
        first_half = np.mean(values[:n//2])
        second_half = np.mean(values[n//2:])
        
        change = (second_half - first_half) / first_half if first_half > 0 else 0
        
        if change > 0.1:
            return "UP"
        elif change < -0.1:
            return "DOWN"
        return "STABLE"
    
    def _calculateScore(
        self, 
        stage: ProductStage, 
        avgRoas: float, 
        trend: str,
        isEvergreen: bool
    ) -> float:
        """Calculate lifecycle score."""
        base_scores = {
            ProductStage.LAUNCH: 60,
            ProductStage.GROWTH: 85,
            ProductStage.MATURE: 70,
            ProductStage.DECLINE: 40
        }
        score = base_scores.get(stage, 50)
        
        # Adjust for trend
        if trend == "UP":
            score += 10
        elif trend == "DOWN":
            score -= 10
        
        # Bonus for evergreen
        if isEvergreen:
            score += 5
        
        # Bonus for high ROAS
        if avgRoas >= 3:
            score += 5
        
        return min(100, max(0, score))
    
    def _getDescription(
        self, 
        stage: ProductStage, 
        daysActive: int,
        isEvergreen: bool
    ) -> str:
        """Get human-readable description."""
        evergreen_suffix = " (Evergreen ✨)" if isEvergreen else ""
        
        descs = {
            ProductStage.LAUNCH: f"🚀 Launch phase ({daysActive} hari){evergreen_suffix}",
            ProductStage.GROWTH: f"📈 Growth phase ({daysActive} hari){evergreen_suffix}",
            ProductStage.MATURE: f"✅ Mature phase ({daysActive} hari){evergreen_suffix}",
            ProductStage.DECLINE: f"📉 Decline phase ({daysActive} hari){evergreen_suffix}"
        }
        return descs.get(stage, f"Unknown ({daysActive} hari)")
    
    def _insufficientData(self) -> LifecycleResult:
        """Return insufficient data result."""
        return LifecycleResult(
            stage=ProductStage.LAUNCH,
            daysActive=0,
            trendDirection="STABLE",
            avgRoas=0,
            recentRoas=0,
            isEvergreen=False,
            score=50,
            description="Data tidak cukup untuk analisis lifecycle"
        )


__all__ = [
    "ProductLifecycle",
    "LifecycleResult",
    "detect_lifecycle_stage",
    "predict_lifecycle_remaining",
    "calculate_churn_risk",
]
