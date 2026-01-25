#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Momentum Analysis Module
========================
Calculate momentum by comparing recent vs historical performance.
Platform-agnostic - works for any e-commerce ads data.

Method: 70/30 split (historical 70% vs recent 30%)
"""

import numpy as np
from typing import List, Dict, Any, Union


DEFAULT_MIN_PERIODS = 3


def calculate_momentum(
    values: Union[List[float], np.ndarray],
    recent_weight: float = 0.3,
    min_periods: int = DEFAULT_MIN_PERIODS
) -> Dict[str, Any]:
    """
    Calculate momentum by comparing recent vs historical performance.
    Uses 70/30 split (historical/recent) by default.
    
    Args:
        values: Array of values over time
        recent_weight: Weight for recent period (default 0.3 = 30%)
        min_periods: Minimum periods required
        
    Returns:
        dict with: momentum_pct, label, direction, historical_avg, recent_avg, is_valid
    """
    values = np.array(values)
    n = len(values)
    
    if n < min_periods:
        return {
            "momentum_pct": 0,
            "label": "insufficient_data",
            "label_display": "⚠️ Data Kurang",
            "direction": "neutral",
            "historical_avg": 0,
            "recent_avg": 0,
            "n_samples": n,
            "is_valid": False
        }
    
    # Split point (30% recent, 70% historical)
    split_idx = max(1, int(n * (1 - recent_weight)))
    
    historical = values[:split_idx]
    recent = values[split_idx:]
    
    historical_avg = np.mean(historical) if len(historical) > 0 else 0
    recent_avg = np.mean(recent) if len(recent) > 0 else 0
    
    # Calculate momentum percentage
    if historical_avg != 0:
        momentum_pct = ((recent_avg - historical_avg) / abs(historical_avg)) * 100
    else:
        momentum_pct = 100 if recent_avg > 0 else -100 if recent_avg < 0 else 0
    
    # Determine label and direction
    label, label_display, direction = momentum_label(momentum_pct)
    
    return {
        "momentum_pct": round(momentum_pct, 2),
        "label": label,
        "label_display": label_display,
        "direction": direction,
        "historical_avg": round(historical_avg, 2),
        "recent_avg": round(recent_avg, 2),
        "n_samples": n,
        "is_valid": True
    }


def momentum_label(momentum_pct: float) -> tuple:
    """
    Convert momentum percentage to human-readable label.
    
    Thresholds:
    - > +20%: Sangat Positif (strong_positive)
    - +5% to +20%: Positif (positive)
    - -5% to +5%: Stabil (stable)
    - -20% to -5%: Negatif (negative)
    - < -20%: Sangat Negatif (strong_negative)
    
    Returns:
        tuple: (label_en, label_display, direction)
    """
    if momentum_pct > 20:
        return ("strong_positive", "🚀 Sangat Positif", "up")
    elif momentum_pct > 5:
        return ("positive", "📈 Positif", "up")
    elif momentum_pct >= -5:
        return ("stable", "➡️ Stabil", "neutral")
    elif momentum_pct >= -20:
        return ("negative", "📉 Negatif", "down")
    else:
        return ("strong_negative", "⚠️ Sangat Negatif", "down")
