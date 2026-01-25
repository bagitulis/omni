#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Product Lifecycle Module
========================
Detect and predict product lifecycle stages.
Platform-agnostic.
"""

import numpy as np
from typing import Dict, List, Any, Optional, Union
from enum import Enum


class LifecycleStage(Enum):
    """Product lifecycle stages."""
    INTRODUCTION = "introduction"
    GROWTH = "growth"
    MATURITY = "maturity"
    DECLINE = "decline"
    END_OF_LIFE = "end_of_life"


class ProductLifecycle:
    """
    Analyze product lifecycle stage and predict remaining life.
    
    Uses sales velocity, trend, and variance to determine stage:
    - Introduction: Low sales, high variance, unknown trend
    - Growth: Increasing sales, positive trend
    - Maturity: Stable sales, flat trend, low variance
    - Decline: Decreasing sales, negative trend
    - End of Life: Very low sales, sustained decline
    """
    
    def __init__(
        self,
        growth_threshold: float = 0.1,
        decline_threshold: float = -0.1,
        min_days_for_stage: int = 7
    ):
        self.growth_threshold = growth_threshold
        self.decline_threshold = decline_threshold
        self.min_days = min_days_for_stage
    
    def analyze(
        self,
        sales_history: List[float],
        days_since_launch: Optional[int] = None
    ) -> Dict[str, Any]:
        """
        Analyze product lifecycle stage.
        
        Args:
            sales_history: Daily sales values (newest last)
            days_since_launch: Days since product launched
            
        Returns:
            dict with: stage, confidence, indicators
        """
        n = len(sales_history)
        
        if n < 3:
            return {
                "stage": LifecycleStage.INTRODUCTION.value,
                "confidence": "low",
                "is_valid": False
            }
        
        values = np.array(sales_history)
        
        # Calculate metrics
        mean_sales = np.mean(values)
        cv = np.std(values) / mean_sales if mean_sales > 0 else 0
        
        # Trend (simple slope)
        x = np.arange(n)
        slope = np.polyfit(x, values, 1)[0] if n >= 2 else 0
        trend_pct = slope / mean_sales if mean_sales > 0 else 0
        
        # Recent vs baseline
        baseline = np.mean(values[:n//2]) if n >= 4 else values[0]
        recent = np.mean(values[n//2:]) if n >= 4 else values[-1]
        change_pct = (recent - baseline) / baseline if baseline > 0 else 0
        
        # Determine stage
        stage = self._determine_stage(
            trend_pct=trend_pct,
            change_pct=change_pct,
            cv=cv,
            mean_sales=mean_sales,
            n_days=n,
            days_since_launch=days_since_launch
        )
        
        # Confidence
        confidence = "high" if n >= 14 else "medium" if n >= 7 else "low"
        
        return {
            "stage": stage.value,
            "confidence": confidence,
            "indicators": {
                "mean_sales": round(mean_sales, 2),
                "trend_pct": round(trend_pct * 100, 2),
                "change_pct": round(change_pct * 100, 2),
                "cv": round(cv, 4),
                "n_days": n
            },
            "is_valid": True
        }
    
    def _determine_stage(
        self,
        trend_pct: float,
        change_pct: float,
        cv: float,
        mean_sales: float,
        n_days: int,
        days_since_launch: Optional[int]
    ) -> LifecycleStage:
        """Determine lifecycle stage from metrics."""
        
        # End of Life: Very low and declining
        if mean_sales < 0.1 and trend_pct < self.decline_threshold:
            return LifecycleStage.END_OF_LIFE
        
        # Introduction: Early days or high variance
        if days_since_launch is not None and days_since_launch < self.min_days:
            return LifecycleStage.INTRODUCTION
        if n_days < self.min_days:
            return LifecycleStage.INTRODUCTION
        
        # Growth: Strong positive trend
        if trend_pct > self.growth_threshold:
            return LifecycleStage.GROWTH
        
        # Decline: Strong negative trend
        if trend_pct < self.decline_threshold:
            return LifecycleStage.DECLINE
        
        # Maturity: Stable with low variance
        if cv < 0.3 and abs(trend_pct) < 0.05:
            return LifecycleStage.MATURITY
        
        # Default to maturity for stable products
        return LifecycleStage.MATURITY


def detect_lifecycle_stage(
    sales_history: List[float],
    days_since_launch: Optional[int] = None
) -> str:
    """
    Quick helper to detect lifecycle stage.
    
    Args:
        sales_history: Daily sales values
        days_since_launch: Days since launch
        
    Returns:
        str: Lifecycle stage name
    """
    lifecycle = ProductLifecycle()
    result = lifecycle.analyze(sales_history, days_since_launch)
    return result["stage"]


def predict_lifecycle_remaining(
    sales_history: List[float],
    decline_threshold: float = 0.1
) -> Dict[str, Any]:
    """
    Predict remaining days until end of life.
    
    Uses linear extrapolation of decline trend.
    
    Args:
        sales_history: Daily sales values
        decline_threshold: Sales level considered EOL
        
    Returns:
        dict with: days_remaining, confidence
    """
    if len(sales_history) < 7:
        return {
            "days_remaining": None,
            "confidence": "low",
            "is_valid": False
        }
    
    values = np.array(sales_history)
    n = len(values)
    current = values[-1]
    
    # Calculate trend
    x = np.arange(n)
    coeffs = np.polyfit(x, values, 1)
    slope = coeffs[0]
    
    # If not declining, return None
    if slope >= 0:
        return {
            "days_remaining": None,
            "trend": "stable_or_growing",
            "confidence": "medium",
            "is_valid": True
        }
    
    # Extrapolate to threshold
    # current + slope * days = threshold
    days_remaining = (decline_threshold - current) / slope
    
    if days_remaining < 0:
        days_remaining = 0
    
    return {
        "days_remaining": int(days_remaining),
        "current_sales": round(current, 2),
        "daily_decline": round(-slope, 2),
        "confidence": "medium",
        "is_valid": True
    }


def calculate_churn_risk(
    conversion_history: List[float],
    min_threshold: float = 0.5
) -> Dict[str, Any]:
    """
    Calculate churn risk based on conversion decline.
    
    Args:
        conversion_history: Daily conversion values
        min_threshold: Minimum conversion threshold
        
    Returns:
        dict with: risk_score, risk_level
    """
    if len(conversion_history) < 3:
        return {
            "risk_score": 0.5,
            "risk_level": "unknown",
            "is_valid": False
        }
    
    values = np.array(conversion_history)
    current = values[-1]
    peak = np.max(values)
    
    if peak == 0:
        return {
            "risk_score": 1.0,
            "risk_level": "critical",
            "is_valid": True
        }
    
    # Decline from peak
    decline_from_peak = (peak - current) / peak
    
    # Trend component
    n = len(values)
    slope = np.polyfit(np.arange(n), values, 1)[0]
    trend_negative = slope < 0
    
    # Calculate risk
    risk_score = decline_from_peak * 0.6
    if trend_negative:
        risk_score += 0.3
    if current < min_threshold:
        risk_score += 0.1
    
    risk_score = min(1.0, risk_score)
    
    # Risk level
    if risk_score >= 0.7:
        risk_level = "critical"
    elif risk_score >= 0.5:
        risk_level = "high"
    elif risk_score >= 0.3:
        risk_level = "moderate"
    else:
        risk_level = "low"
    
    return {
        "risk_score": round(risk_score, 4),
        "risk_score_100": round(risk_score * 100, 2),
        "risk_level": risk_level,
        "decline_from_peak_pct": round(decline_from_peak * 100, 2),
        "is_valid": True
    }
