#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Trend Analysis Module
=====================
Statistical tests for trend detection.
Platform-agnostic - works for any e-commerce ads data.

Methods:
- Mann-Kendall Test: Non-parametric monotonic trend test
- Linear Regression: Parametric trend with R²
- MACD-Style: Exponential moving average crossover

Reference: Mann (1945), Kendall (1975)
"""

import numpy as np
from scipy import stats
from typing import List, Dict, Any, Union


# Default thresholds (can be overridden)
DEFAULT_MIN_PERIODS = 4
DEFAULT_HIGH_CONFIDENCE_SAMPLES = 8


def mann_kendall_test(
    data: Union[List[float], np.ndarray],
    min_periods: int = DEFAULT_MIN_PERIODS,
    high_confidence_samples: int = DEFAULT_HIGH_CONFIDENCE_SAMPLES
) -> Dict[str, Any]:
    """
    Mann-Kendall Trend Test - Non-parametric test for monotonic trend.
    
    Args:
        data: Array of values over time (minimum 4 points recommended)
        min_periods: Minimum periods for valid analysis
        high_confidence_samples: Samples needed for high confidence
        
    Returns:
        dict with: trend, p_value, tau, z_score, confidence, data_quality, n_samples, is_valid
    """
    data = np.array(data)
    n = len(data)
    
    # Data quality assessment
    if n < 2:
        return {
            "trend": "no_data",
            "trend_label": "➖ N/A",
            "p_value": 1.0,
            "tau": 0,
            "z_score": 0,
            "confidence": 0,
            "data_quality": "no_data",
            "n_samples": n,
            "is_valid": False
        }
    elif n < min_periods:
        return {
            "trend": "insufficient_data",
            "trend_label": "⚠️ Data Kurang",
            "p_value": 1.0,
            "tau": 0,
            "z_score": 0,
            "confidence": 0,
            "data_quality": "low",
            "n_samples": n,
            "is_valid": False
        }
    
    # Calculate S statistic (sum of signs of all pairwise differences)
    s = 0
    for i in range(n - 1):
        for j in range(i + 1, n):
            s += np.sign(data[j] - data[i])
    
    # Variance of S (assuming no ties)
    var_s = (n * (n - 1) * (2 * n + 5)) / 18
    
    # Z statistic (normalized)
    if s > 0:
        z = (s - 1) / np.sqrt(var_s)
    elif s < 0:
        z = (s + 1) / np.sqrt(var_s)
    else:
        z = 0
    
    # P-value (two-tailed test)
    p_value = 2 * (1 - stats.norm.cdf(abs(z)))
    
    # Kendall's tau (correlation coefficient)
    tau = s / (n * (n - 1) / 2)
    
    # Confidence level
    confidence = round((1 - p_value) * 100, 1) if p_value < 1 else 0
    
    # Data quality
    if n >= high_confidence_samples:
        data_quality = "high"
    elif n >= min_periods:
        data_quality = "medium"
    else:
        data_quality = "low"
    
    # Determine trend (α = 0.05)
    if p_value < 0.05:
        if s > 0:
            trend = "increasing"
            trend_label = "↗️ Naik"
        else:
            trend = "decreasing"
            trend_label = "↘️ Turun"
    else:
        trend = "stable"
        trend_label = "➡️ Stabil"
    
    return {
        "trend": trend,
        "trend_label": trend_label,
        "p_value": round(p_value, 4),
        "tau": round(tau, 4),
        "z_score": round(z, 4),
        "confidence": confidence,
        "data_quality": data_quality,
        "n_samples": n,
        "is_valid": True
    }


def linear_trend_analysis(
    data: Union[List[float], np.ndarray],
    min_periods: int = DEFAULT_MIN_PERIODS,
    high_confidence_samples: int = DEFAULT_HIGH_CONFIDENCE_SAMPLES
) -> Dict[str, Any]:
    """
    Linear Regression Trend Analysis.
    
    Args:
        data: Array of values over time
        
    Returns:
        dict with: slope, r_squared, p_value, direction, intercept, std_err
    """
    data = np.array(data)
    n = len(data)
    
    if n < 2:
        return {
            'slope': 0,
            'r_squared': 0,
            'p_value': 1.0,
            'direction': 'no_data',
            'intercept': 0,
            'std_err': 0,
            'data_quality': 'no_data',
            'n_samples': n,
            'is_valid': False
        }
    
    x = np.arange(n)
    slope, intercept, r_value, p_value, std_err = stats.linregress(x, data)
    
    # Data quality
    if n >= high_confidence_samples:
        data_quality = "high"
    elif n >= min_periods:
        data_quality = "medium"
    else:
        data_quality = "low"
    
    # Determine direction
    mean_val = np.mean(data) or 1
    if p_value > 0.1 or abs(slope) < 0.01 * abs(mean_val):
        direction = 'stable'
    elif slope > 0:
        direction = 'up'
    else:
        direction = 'down'
    
    return {
        'slope': round(slope, 6),
        'r_squared': round(r_value ** 2, 4),
        'p_value': round(p_value, 4),
        'direction': direction,
        'intercept': round(intercept, 4),
        'std_err': round(std_err, 6),
        'data_quality': data_quality,
        'n_samples': n,
        'is_valid': n >= min_periods
    }


def macd_trend(
    data: Union[List[float], np.ndarray],
    short_period: int = 3,
    long_period: int = 8
) -> Dict[str, Any]:
    """
    MACD-style trend analysis using EMA crossover.
    
    Args:
        data: Array of values over time
        short_period: Short EMA period (default 3)
        long_period: Long EMA period (default 8)
        
    Returns:
        dict with: signal, direction, strength, ema_short, ema_long
    """
    data = np.array(data)
    n = len(data)
    
    if n < long_period:
        return {
            'signal': 0,
            'direction': 'insufficient_data',
            'strength': 0,
            'ema_short': 0,
            'ema_long': 0,
            'is_valid': False
        }
    
    def ema(values, period):
        """Calculate Exponential Moving Average."""
        alpha = 2 / (period + 1)
        result = np.zeros_like(values, dtype=float)
        result[0] = values[0]
        for i in range(1, len(values)):
            result[i] = alpha * values[i] + (1 - alpha) * result[i - 1]
        return result
    
    ema_short = ema(data, short_period)
    ema_long = ema(data, long_period)
    
    # MACD line = EMA_short - EMA_long
    signal = ema_short[-1] - ema_long[-1]
    
    # Normalize signal relative to average
    avg = np.mean(data) or 1
    strength = abs(signal / avg) * 100
    
    # Direction
    if signal > 0 and strength > 5:
        direction = 'bullish'
    elif signal < 0 and strength > 5:
        direction = 'bearish'
    else:
        direction = 'neutral'
    
    return {
        'signal': round(signal, 4),
        'direction': direction,
        'strength': round(strength, 2),
        'ema_short': round(ema_short[-1], 4),
        'ema_long': round(ema_long[-1], 4),
        'is_valid': True
    }
