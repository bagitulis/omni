#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Creative Fatigue Detection Module
==================================
Detect when ads/creatives are losing effectiveness.
Platform-agnostic.
"""

import numpy as np
from typing import Dict, List, Any, Optional, Union
from dataclasses import dataclass


@dataclass
class FatigueThresholds:
    """Configurable thresholds for fatigue detection."""
    ctr_decline_pct: float = 0.20     # CTR drops 20%
    cvr_decline_pct: float = 0.15     # CVR drops 15%
    cpc_increase_pct: float = 0.25    # CPC increases 25%
    frequency_limit: float = 3.0       # Max impressions per user
    days_running: int = 14            # Days before fatigue risk
    

class FatigueDetector:
    """
    Detect creative fatigue in ads.
    
    Fatigue indicators:
    - CTR declining over time
    - CVR declining over time
    - CPC increasing over time
    - High frequency (user sees ad too often)
    - Long running duration
    """
    
    def __init__(self, thresholds: Optional[FatigueThresholds] = None):
        self.thresholds = thresholds or FatigueThresholds()
    
    def analyze(
        self,
        ctr_history: List[float],
        cvr_history: Optional[List[float]] = None,
        cpc_history: Optional[List[float]] = None,
        frequency: Optional[float] = None,
        days_running: Optional[int] = None
    ) -> Dict[str, Any]:
        """
        Analyze fatigue indicators.
        
        Args:
            ctr_history: List of CTR values (newest last)
            cvr_history: List of CVR values
            cpc_history: List of CPC values
            frequency: Current ad frequency
            days_running: Days since ad started
            
        Returns:
            dict with: fatigue_index, is_fatigued, indicators
        """
        indicators = {}
        fatigue_score = 0
        
        # CTR analysis
        ctr_result = self._analyze_metric_decline(
            ctr_history, 
            self.thresholds.ctr_decline_pct
        )
        indicators["ctr"] = ctr_result
        if ctr_result["is_declining"]:
            fatigue_score += 0.3
        
        # CVR analysis
        if cvr_history:
            cvr_result = self._analyze_metric_decline(
                cvr_history,
                self.thresholds.cvr_decline_pct
            )
            indicators["cvr"] = cvr_result
            if cvr_result["is_declining"]:
                fatigue_score += 0.25
        
        # CPC analysis (increasing = fatigue)
        if cpc_history:
            cpc_result = self._analyze_metric_increase(
                cpc_history,
                self.thresholds.cpc_increase_pct
            )
            indicators["cpc"] = cpc_result
            if cpc_result["is_increasing"]:
                fatigue_score += 0.2
        
        # Frequency check
        if frequency is not None:
            freq_fatigued = frequency > self.thresholds.frequency_limit
            indicators["frequency"] = {
                "value": frequency,
                "limit": self.thresholds.frequency_limit,
                "is_high": freq_fatigued
            }
            if freq_fatigued:
                fatigue_score += 0.15
        
        # Duration check
        if days_running is not None:
            duration_fatigued = days_running > self.thresholds.days_running
            indicators["duration"] = {
                "days_running": days_running,
                "limit": self.thresholds.days_running,
                "is_long": duration_fatigued
            }
            if duration_fatigued:
                fatigue_score += 0.1
        
        # Final assessment
        fatigue_index = min(1.0, fatigue_score)
        
        return {
            "fatigue_index": round(fatigue_index, 4),
            "fatigue_score_100": round(fatigue_index * 100, 2),
            "is_fatigued": fatigue_index >= 0.5,
            "severity": self._get_severity(fatigue_index),
            "indicators": indicators,
            "recommendation": self._get_recommendation(fatigue_index)
        }
    
    def _analyze_metric_decline(
        self,
        values: List[float],
        threshold_pct: float
    ) -> Dict[str, Any]:
        """Check if metric is declining beyond threshold."""
        if len(values) < 3:
            return {"is_declining": False, "change_pct": 0}
        
        # Compare recent (last 30%) vs baseline (first 30%)
        n = len(values)
        baseline = np.mean(values[:max(1, n // 3)])
        recent = np.mean(values[-max(1, n // 3):])
        
        if baseline == 0:
            return {"is_declining": False, "change_pct": 0}
        
        change_pct = (recent - baseline) / baseline
        is_declining = change_pct < -threshold_pct
        
        return {
            "is_declining": is_declining,
            "baseline": round(baseline, 4),
            "recent": round(recent, 4),
            "change_pct": round(change_pct * 100, 2)
        }
    
    def _analyze_metric_increase(
        self,
        values: List[float],
        threshold_pct: float
    ) -> Dict[str, Any]:
        """Check if metric is increasing beyond threshold."""
        if len(values) < 3:
            return {"is_increasing": False, "change_pct": 0}
        
        n = len(values)
        baseline = np.mean(values[:max(1, n // 3)])
        recent = np.mean(values[-max(1, n // 3):])
        
        if baseline == 0:
            return {"is_increasing": False, "change_pct": 0}
        
        change_pct = (recent - baseline) / baseline
        is_increasing = change_pct > threshold_pct
        
        return {
            "is_increasing": is_increasing,
            "baseline": round(baseline, 4),
            "recent": round(recent, 4),
            "change_pct": round(change_pct * 100, 2)
        }
    
    def _get_severity(self, index: float) -> str:
        """Get fatigue severity label."""
        if index >= 0.75:
            return "critical"
        elif index >= 0.5:
            return "high"
        elif index >= 0.25:
            return "moderate"
        else:
            return "low"
    
    def _get_recommendation(self, index: float) -> str:
        """Get recommendation based on fatigue level."""
        if index >= 0.75:
            return "replace_creative_immediately"
        elif index >= 0.5:
            return "prepare_new_creative"
        elif index >= 0.25:
            return "monitor_closely"
        else:
            return "continue_running"


def detect_creative_fatigue(
    ctr_history: List[float],
    cvr_history: Optional[List[float]] = None,
    days_running: Optional[int] = None
) -> Dict[str, Any]:
    """
    Quick helper to detect creative fatigue.
    
    Args:
        ctr_history: List of CTR values
        cvr_history: List of CVR values
        days_running: Days since ad started
        
    Returns:
        dict with fatigue analysis
    """
    detector = FatigueDetector()
    return detector.analyze(
        ctr_history=ctr_history,
        cvr_history=cvr_history,
        days_running=days_running
    )


def calculate_fatigue_index(
    baseline_ctr: float,
    current_ctr: float,
    days_running: int,
    frequency: float = 1.0
) -> float:
    """
    Calculate simple fatigue index (0-100).
    
    Args:
        baseline_ctr: Initial CTR
        current_ctr: Current CTR
        days_running: Days running
        frequency: Ad frequency
        
    Returns:
        float: Fatigue index (0-100)
    """
    score = 0
    
    # CTR decline factor
    if baseline_ctr > 0:
        ctr_change = (current_ctr - baseline_ctr) / baseline_ctr
        if ctr_change < -0.1:
            score += min(40, abs(ctr_change) * 100)
    
    # Duration factor
    if days_running > 7:
        score += min(30, (days_running - 7) * 2)
    
    # Frequency factor
    if frequency > 2:
        score += min(30, (frequency - 2) * 10)
    
    return min(100, score)
