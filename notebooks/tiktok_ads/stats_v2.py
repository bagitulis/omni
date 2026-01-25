#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Statistical Functions (Re-exported from Core)
====================================================================
This module re-exports statistical functions from the shared core library
and adds TikTok-specific formatting/labels.

All statistical logic is in `notebooks/core/statistics/`.
This module only provides TikTok-specific wrappers.
"""

# Re-export from core library
import sys
from pathlib import Path

# Add parent directory to sys.path for core imports
parent_dir = Path(__file__).parent.parent
if str(parent_dir) not in sys.path:
    sys.path.insert(0, str(parent_dir))

from core.statistics import (
    mann_kendall_test as _mann_kendall_test,
    linear_trend_analysis as _linear_trend_analysis,
    macd_trend as _macd_trend,
    calculate_momentum as _calculate_momentum,
    coefficient_of_variation as _coefficient_of_variation,
    cv_label,
    calculate_confidence_interval,
    bootstrap_ci,
    compare_groups_ttest,
    calculate_pearson,
    calculate_spearman,
)

from .config import (
    MIN_PERIODS_FOR_TREND,
    MIN_PERIODS_FOR_MOMENTUM,
    MIN_SAMPLES_HIGH_CONFIDENCE
)


def mann_kendall_test(data):
    """
    Mann-Kendall Trend Test with TikTok-specific labels.
    
    Wraps core.statistics.mann_kendall_test() and adds
    Indonesian labels for TikTok dashboard.
    
    Args:
        data: Array of values over time
        
    Returns:
        dict: trend, trend_en, p_value, tau, z_score, confidence, data_quality, n_samples, is_valid
    """
    result = _mann_kendall_test(
        data,
        min_periods=MIN_PERIODS_FOR_TREND,
        high_confidence_samples=MIN_SAMPLES_HIGH_CONFIDENCE
    )
    
    # Add TikTok-specific Indonesian labels
    trend_labels = {
        "increasing": "📈 Naik",
        "decreasing": "📉 Turun",
        "stable": "➖ Stabil",
        "no_data": "➖ N/A",
        "insufficient_data": "⚠️ Data Kurang"
    }
    
    result["trend"] = trend_labels.get(result.get("trend", "stable"), "➖ N/A")
    result["trend_en"] = result.get("trend", "stable")
    
    return result


def linear_trend_analysis(data):
    """
    Linear regression trend analysis with TikTok labels.
    
    Args:
        data: Array of values over time
        
    Returns:
        dict: slope, intercept, r_squared, trend, p_value, data_quality
    """
    result = _linear_trend_analysis(
        data,
        min_periods=MIN_PERIODS_FOR_TREND,
        high_confidence_samples=MIN_SAMPLES_HIGH_CONFIDENCE
    )
    
    # Add TikTok-specific labels
    trend_labels = {
        "strong_up": "🚀 Naik Kuat",
        "up": "📈 Naik",
        "stable": "➖ Stabil",
        "down": "📉 Turun",
        "strong_down": "💀 Turun Kuat"
    }
    
    if result.get("is_valid"):
        result["trend"] = trend_labels.get(result.get("trend", "stable"), "➖ Stabil")
    else:
        result["trend"] = "⚠️ Data Kurang"
    
    return result


def calculate_momentum(data, split_ratio=0.7):
    """
    70/30 Momentum analysis with TikTok labels.
    
    Args:
        data: Array of values over time
        split_ratio: Ratio for baseline/recent split (default 0.7)
        
    Returns:
        dict: ratio, direction, baseline_avg, recent_avg, change_pct
    """
    # split_ratio 0.7 means 70% historical, 30% recent
    # core uses recent_weight which is the complement
    recent_weight = 1.0 - split_ratio
    
    result = _calculate_momentum(
        data,
        recent_weight=recent_weight,
        min_periods=MIN_PERIODS_FOR_MOMENTUM
    )
    
    # Add TikTok-specific labels
    direction_labels = {
        "accelerating": "🔥 Akselerasi",
        "stable": "➖ Stabil",
        "decelerating": "⚠️ Deselerasi",
        "insufficient_data": "⚠️ Data Kurang"
    }
    
    result["direction"] = direction_labels.get(result.get("direction", "stable"), "➖ Stabil")
    result["direction_en"] = result.get("direction", "stable")
    
    return result


def coefficient_of_variation(data):
    """
    Coefficient of Variation (CV) with TikTok labels.
    
    Args:
        data: Array of values
        
    Returns:
        dict: cv, label, quality
    """
    cv_value = _coefficient_of_variation(data)
    label_en, label_display = cv_label(cv_value)
    
    return {
        "cv": cv_value,
        "label": label_display,
        "label_en": label_en,
        "quality": label_en
    }


# Re-export other functions as-is
macd_trend = _macd_trend

__all__ = [
    "mann_kendall_test",
    "linear_trend_analysis",
    "calculate_momentum",
    "coefficient_of_variation",
    "macd_trend",
    "calculate_confidence_interval",
    "bootstrap_ci",
    "compare_groups_ttest",
    "calculate_pearson",
    "calculate_spearman",
]
