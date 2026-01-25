#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Variation Analysis Module
=========================
Coefficient of Variation and other dispersion metrics.
Platform-agnostic.
"""

import numpy as np
from typing import List, Union


def coefficient_of_variation(
    values: Union[List[float], np.ndarray],
    min_samples: int = 2
) -> float:
    """
    Calculate Coefficient of Variation (CV).
    CV = (Standard Deviation / Mean) * 100
    
    Lower CV = more consistent, higher CV = more volatile
    
    Args:
        values: Array of values
        min_samples: Minimum samples required
        
    Returns:
        CV percentage (0-100+), 0 if insufficient data
    """
    values = np.array(values)
    n = len(values)
    
    if n < min_samples:
        return 0.0
    
    mean_val = np.mean(values)
    std_val = np.std(values)
    
    if mean_val == 0:
        return 100.0 if std_val > 0 else 0.0
    
    cv = (std_val / abs(mean_val)) * 100
    return round(cv, 2)


def cv_label(cv: float) -> tuple:
    """
    Convert CV to human-readable label.
    
    Thresholds:
    - < 20%: Sangat Konsisten (very consistent)
    - 20-35%: Konsisten (consistent)
    - 35-50%: Cukup Variabel (moderate)
    - > 50%: Volatil (volatile)
    
    Returns:
        tuple: (label_en, label_display)
    """
    if cv < 20:
        return ("very_consistent", "✅ Sangat Konsisten")
    elif cv < 35:
        return ("consistent", "🔶 Konsisten")
    elif cv < 50:
        return ("moderate", "⚠️ Cukup Variabel")
    else:
        return ("volatile", "🔴 Volatil")
