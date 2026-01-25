#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads - Statistics Wrapper
===============================
TikTok-specific wrapper using core statistics modules.
Following AGENTS.MD: Clean Code, DRY, SRP.
"""

import numpy as np
from typing import List, Union

# Import core statistics
from core.statistics.trend import mann_kendall_test as core_mann_kendall, linear_trend_analysis as core_linear_trend
from core.statistics.momentum import calculate_momentum as core_momentum
from core.statistics.variation import coefficient_of_variation as core_cv

# TikTok-specific config
from .config import (
    MIN_PERIODS_FOR_TREND,
    MIN_PERIODS_FOR_MOMENTUM,
    MIN_SAMPLES_HIGH_CONFIDENCE
)

__all__ = [
    'mann_kendall_test',
    'linear_trend_analysis',
    'calculate_momentum',
    'coefficient_of_variation'
]


def _data_quality_label(n: int) -> str:
    """Generate TikTok-specific data quality label."""
    if n >= MIN_SAMPLES_HIGH_CONFIDENCE:
        return "✅ HIGH"
    elif n >= MIN_PERIODS_FOR_TREND:
        return f"🔶 MEDIUM (n={n})"
    else:
        return f"⚠️ LOW (n={n})"


def mann_kendall_test(data: Union[List[float], np.ndarray]) -> dict:
    """
    Mann-Kendall Trend Test - TikTok wrapper.
    
    Args:
        data: Array of values over time (minimum 4 points recommended)
        
    Returns:
        dict: trend, p_value, tau, z_score, confidence, data_quality, n_samples, is_valid
    """
    data = np.array(data)
    n = len(data)
    
    # Insufficient data
    if n < 2:
        return {
            "trend": "➖ N/A", 
            "trend_en": "no_data",
            "p_value": 1.0, 
            "tau": 0, 
            "z_score": 0,
            "confidence": 0,
            "data_quality": "❌ NO DATA",
            "n_samples": n,
            "is_valid": False
        }
    
    if n < MIN_PERIODS_FOR_TREND:
        return {
            "trend": "⚠️ Data Kurang", 
            "trend_en": "insufficient_data",
            "p_value": 1.0, 
            "tau": 0, 
            "z_score": 0,
            "confidence": 0,
            "data_quality": f"⚠️ LOW (n={n}, need {MIN_PERIODS_FOR_TREND}+)",
            "n_samples": n,
            "is_valid": False
        }
    
    # Use core implementation
    result = core_mann_kendall(data)
    
    # TikTok-specific formatting
    trend_emoji = {
        "increasing": "↗️ Naik",
        "decreasing": "↘️ Turun",
        "no_trend": "➡️ Stabil"
    }
    
    return {
        "trend": trend_emoji.get(result.get("trend", "no_trend"), "➡️ Stabil"),
        "trend_en": result.get("trend", "no_trend"),
        "p_value": result.get("p_value", 1.0),
        "tau": result.get("tau", 0),
        "z_score": result.get("z_score", 0),
        "confidence": result.get("confidence", 0),
        "data_quality": _data_quality_label(n),
        "n_samples": n,
        "is_valid": True
    }


def linear_trend_analysis(data: Union[List[float], np.ndarray]) -> dict:
    """
    Linear Regression Trend Analysis - TikTok wrapper.
    
    Args:
        data: Array of values over time
        
    Returns:
        dict: slope, r_squared, p_value, direction, etc.
    """
    data = np.array(data)
    n = len(data)
    
    if n < 2:
        return {
            'slope': 0, 
            'r_squared': 0, 
            'p_value': 1.0,
            'direction': 'insufficient_data',
            'direction_en': 'no_data',
            'intercept': 0,
            'std_err': 0,
            'data_quality': '❌ NO DATA',
            'n_samples': n,
            'is_valid': False
        }
    
    # Use core implementation
    result = core_linear_trend(data)
    
    # TikTok-specific direction mapping
    direction_map = {
        "up": ("naik", "up"),
        "down": ("turun", "down"),
        "flat": ("stabil", "flat")
    }
    
    dir_result = direction_map.get(result.get("direction", "flat"), ("stabil", "flat"))
    
    return {
        'slope': result.get("slope", 0),
        'r_squared': result.get("r_squared", 0),
        'p_value': result.get("p_value", 1.0),
        'direction': dir_result[0],
        'direction_en': dir_result[1],
        'intercept': result.get("intercept", 0),
        'std_err': result.get("std_err", 0),
        'data_quality': _data_quality_label(n),
        'n_samples': n,
        'is_valid': n >= MIN_PERIODS_FOR_TREND
    }


def calculate_momentum(values: Union[List[float], np.ndarray]) -> dict:
    """
    Calculate momentum - TikTok wrapper with emoji labels.
    
    Args:
        values: Array of values over time
        
    Returns:
        dict: momentum, momentum_en, change_pct, etc.
    """
    values = np.array(values)
    n = len(values)
    
    if n < 2:
        return {
            "momentum": "N/A", 
            "momentum_en": "no_data",
            "change_pct": 0,
            "recent_avg": 0,
            "historical_avg": 0,
            "data_quality": "❌ NO DATA",
            "n_samples": n,
            "is_valid": False
        }
    
    # Use core momentum
    result = core_momentum(values, recent_weight=0.3, min_periods=MIN_PERIODS_FOR_MOMENTUM)
    
    # TikTok-specific emoji labels
    change_pct = result.get("momentum_pct", 0)
    
    if change_pct > 20:
        momentum = "🚀 Sangat Positif"
        momentum_en = "strongly_improving"
    elif change_pct > 5:
        momentum = "📈 Positif"
        momentum_en = "improving"
    elif change_pct < -20:
        momentum = "📉 Sangat Negatif"
        momentum_en = "strongly_declining"
    elif change_pct < -5:
        momentum = "⚠️ Negatif"
        momentum_en = "declining"
    else:
        momentum = "➡️ Stabil"
        momentum_en = "stable"
    
    return {
        "momentum": momentum,
        "momentum_en": momentum_en,
        "change_pct": change_pct,
        "recent_avg": result.get("recent_avg", 0),
        "historical_avg": result.get("historical_avg", 0),
        "data_quality": _data_quality_label(n),
        "n_samples": n,
        "is_valid": n >= MIN_PERIODS_FOR_MOMENTUM
    }


def coefficient_of_variation(data: Union[List[float], np.ndarray]) -> dict:
    """
    Calculate Coefficient of Variation - TikTok wrapper.
    
    Args:
        data: Array of values
        
    Returns:
        dict: cv, consistency, std, mean, etc.
    """
    data = np.array(data)
    n = len(data)
    
    if n < 2:
        return {
            "cv": None,
            "consistency": "❌ Data Kurang",
            "std": 0,
            "mean": 0,
            "data_quality": "❌ NO DATA",
            "n_samples": n,
            "is_valid": False
        }
    
    # Use core CV
    cv = core_cv(data)
    mean_val = float(np.mean(data))
    std_val = float(np.std(data))
    
    # TikTok consistency label
    if cv is None or cv == float('inf'):
        consistency = "⚠️ Tidak dapat dihitung"
    elif cv < 20:
        consistency = "✅ Sangat Konsisten"
    elif cv < 50:
        consistency = "🔶 Konsisten"
    else:
        consistency = "⚠️ Tidak Konsisten"
    
    return {
        "cv": cv if cv != float('inf') else None,
        "consistency": consistency,
        "std": round(std_val, 4),
        "mean": round(mean_val, 4),
        "data_quality": _data_quality_label(n),
        "n_samples": n,
        "is_valid": n >= MIN_PERIODS_FOR_MOMENTUM
    }
