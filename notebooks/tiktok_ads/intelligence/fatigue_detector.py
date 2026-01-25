#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Fatigue Detector (Re-exported from Core + TikTok Extensions)
============================================================
Re-exports core fatigue detection and adds TikTok-specific wrappers.
Base logic is in `notebooks/core/intelligence/fatigue.py`.
"""

import sys
from pathlib import Path
from dataclasses import dataclass
from enum import Enum
from typing import List, Optional
import pandas as pd
import numpy as np

# Add parent directory to sys.path for core imports
_parent_dir = Path(__file__).resolve().parent.parent.parent
if str(_parent_dir) not in sys.path:
    sys.path.insert(0, str(_parent_dir))

# Re-export from core
from core.intelligence.fatigue import (
    FatigueDetector as CoreFatigueDetector,
    FatigueThresholds,
    detect_creative_fatigue,
    calculate_fatigue_index,
)


class FatigueStatus(Enum):
    """Creative fatigue status for TikTok."""
    FRESH = "FRESH"           # CTR decay < 10%
    AGING = "AGING"           # CTR decay 10-25%
    FATIGUED = "FATIGUED"     # CTR decay 25-50%
    DEAD = "DEAD"             # CTR decay > 50%
    INSUFFICIENT_DATA = "INSUFFICIENT_DATA"


@dataclass
class FatigueResult:
    """Fatigue analysis result for TikTok."""
    status: FatigueStatus
    ctrDecay: float          # CTR decay percentage
    peakCtr: float           # Highest CTR recorded
    currentCtr: float        # Most recent CTR
    daysSincePeak: int       # Days since peak performance
    score: float             # 0-100 (100 = fresh)
    description: str
    action: str


class FatigueDetector:
    """
    TikTok-specific fatigue detector wrapping core logic.
    Detects creative fatigue by analyzing CTR decay over time.
    """
    
    def __init__(self, config=None):
        self.core_detector = CoreFatigueDetector()
        self.config = config
    
    def analyzeFromDf(self, df: pd.DataFrame) -> FatigueResult:
        """
        Analyze creative fatigue from DataFrame.
        
        Args:
            df: DataFrame with 'ctr' column and optionally 'periodStart'
        
        Returns:
            FatigueResult with fatigue status and score
        """
        if 'ctr' not in df.columns or len(df) < 3:
            return self._insufficientData()
        
        ctrHistory = df['ctr'].dropna()
        dateHistory = df.get('periodStart', None)
        
        return self.analyze(ctrHistory, dateHistory)
    
    def analyze(
        self, 
        ctrHistory: pd.Series, 
        dateHistory: pd.Series = None
    ) -> FatigueResult:
        """
        Analyze creative fatigue from CTR history.
        
        Args:
            ctrHistory: Time-ordered CTR values (%)
            dateHistory: Corresponding dates (optional)
        
        Returns:
            FatigueResult with fatigue status and score
        """
        ctrHistory = ctrHistory.dropna()
        
        if len(ctrHistory) < 3:
            return self._insufficientData()
        
        # Use core detector
        ctr_list = ctrHistory.tolist()
        days_running = len(ctrHistory)
        
        core_result = self.core_detector.analyze(
            ctr_history=ctr_list,
            days_running=days_running
        )
        
        # Extract metrics
        peakCtr = max(ctr_list)
        currentCtr = ctr_list[-1]
        ctrDecay = ((peakCtr - currentCtr) / peakCtr * 100) if peakCtr > 0 else 0
        
        # Map core result to TikTok format
        fatigue_index = core_result["fatigue_index"]
        status = self._mapStatus(fatigue_index)
        score = (1 - fatigue_index) * 100
        
        return FatigueResult(
            status=status,
            ctrDecay=round(ctrDecay, 1),
            peakCtr=round(peakCtr, 2),
            currentCtr=round(currentCtr, 2),
            daysSincePeak=days_running,
            score=round(score, 1),
            description=self._getDescription(status, ctrDecay),
            action=core_result["recommendation"]
        )
    
    def _mapStatus(self, fatigue_index: float) -> FatigueStatus:
        """Map fatigue index to status."""
        if fatigue_index < 0.1:
            return FatigueStatus.FRESH
        elif fatigue_index < 0.25:
            return FatigueStatus.AGING
        elif fatigue_index < 0.5:
            return FatigueStatus.FATIGUED
        else:
            return FatigueStatus.DEAD
    
    def _getDescription(self, status: FatigueStatus, decay: float) -> str:
        """Get human-readable description."""
        descs = {
            FatigueStatus.FRESH: f"✅ Creative masih fresh (decay {decay:.1f}%)",
            FatigueStatus.AGING: f"🔶 Creative mulai aging (decay {decay:.1f}%)",
            FatigueStatus.FATIGUED: f"⚠️ Creative fatigued (decay {decay:.1f}%)",
            FatigueStatus.DEAD: f"❌ Creative mati (decay {decay:.1f}%)",
            FatigueStatus.INSUFFICIENT_DATA: "❓ Data tidak cukup"
        }
        return descs.get(status, "Unknown")
    
    def _insufficientData(self) -> FatigueResult:
        """Return insufficient data result."""
        return FatigueResult(
            status=FatigueStatus.INSUFFICIENT_DATA,
            ctrDecay=0,
            peakCtr=0,
            currentCtr=0,
            daysSincePeak=0,
            score=50,
            description="Data tidak cukup untuk analisis fatigue",
            action="gather_more_data"
        )


__all__ = [
    "FatigueDetector",
    "FatigueStatus",
    "FatigueResult",
    "FatigueThresholds",
    "detect_creative_fatigue",
    "calculate_fatigue_index",
]
