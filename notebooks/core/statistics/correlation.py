#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Correlation Module
==================
Correlation analysis methods.
Platform-agnostic.
"""

import numpy as np
from scipy import stats
from typing import List, Dict, Any, Union


def calculate_pearson(
    x: Union[List[float], np.ndarray],
    y: Union[List[float], np.ndarray],
    min_samples: int = 5
) -> Dict[str, Any]:
    """
    Calculate Pearson correlation coefficient.
    Best for linear relationships with normal distributions.
    
    Args:
        x: First variable
        y: Second variable
        min_samples: Minimum samples required
        
    Returns:
        dict with: correlation, p_value, strength_label
    """
    x = np.array(x)
    y = np.array(y)
    
    if len(x) != len(y) or len(x) < min_samples:
        return {
            "correlation": 0,
            "p_value": 1.0,
            "strength_label": "insufficient_data",
            "is_valid": False
        }
    
    # Remove nan pairs
    mask = ~(np.isnan(x) | np.isnan(y))
    x, y = x[mask], y[mask]
    
    if len(x) < min_samples:
        return {
            "correlation": 0,
            "p_value": 1.0,
            "strength_label": "insufficient_data",
            "is_valid": False
        }
    
    r, p_value = stats.pearsonr(x, y)
    
    return {
        "correlation": round(r, 4),
        "p_value": round(p_value, 4),
        "strength_label": correlation_strength_label(r),
        "n_samples": len(x),
        "is_valid": True
    }


def calculate_spearman(
    x: Union[List[float], np.ndarray],
    y: Union[List[float], np.ndarray],
    min_samples: int = 5
) -> Dict[str, Any]:
    """
    Calculate Spearman rank correlation coefficient.
    More robust for non-linear relationships and outliers.
    
    Args:
        x: First variable
        y: Second variable
        min_samples: Minimum samples required
        
    Returns:
        dict with: correlation, p_value, strength_label
    """
    x = np.array(x)
    y = np.array(y)
    
    if len(x) != len(y) or len(x) < min_samples:
        return {
            "correlation": 0,
            "p_value": 1.0,
            "strength_label": "insufficient_data",
            "is_valid": False
        }
    
    # Remove nan pairs
    mask = ~(np.isnan(x) | np.isnan(y))
    x, y = x[mask], y[mask]
    
    if len(x) < min_samples:
        return {
            "correlation": 0,
            "p_value": 1.0,
            "strength_label": "insufficient_data",
            "is_valid": False
        }
    
    rho, p_value = stats.spearmanr(x, y)
    
    return {
        "correlation": round(rho, 4),
        "p_value": round(p_value, 4),
        "strength_label": correlation_strength_label(rho),
        "n_samples": len(x),
        "is_valid": True
    }


def correlation_strength_label(r: float) -> str:
    """
    Get correlation strength label based on absolute value.
    
    Args:
        r: Correlation coefficient (-1 to 1)
        
    Returns:
        str: Strength label
    """
    abs_r = abs(r)
    
    if abs_r < 0.1:
        return "negligible"
    elif abs_r < 0.3:
        return "weak"
    elif abs_r < 0.5:
        return "moderate"
    elif abs_r < 0.7:
        return "strong"
    else:
        return "very_strong"


def correlation_matrix(
    data: Dict[str, List[float]],
    method: str = "pearson"
) -> Dict[str, Any]:
    """
    Calculate correlation matrix for multiple variables.
    
    Args:
        data: Dict of variable_name: values
        method: "pearson" or "spearman"
        
    Returns:
        dict with: matrix (nested dict), variables
    """
    variables = list(data.keys())
    n_vars = len(variables)
    
    calc_func = calculate_pearson if method == "pearson" else calculate_spearman
    
    matrix = {}
    for var1 in variables:
        matrix[var1] = {}
        for var2 in variables:
            if var1 == var2:
                matrix[var1][var2] = {"correlation": 1.0, "p_value": 0.0}
            else:
                result = calc_func(data[var1], data[var2])
                matrix[var1][var2] = {
                    "correlation": result["correlation"],
                    "p_value": result.get("p_value", 1.0)
                }
    
    return {
        "matrix": matrix,
        "variables": variables,
        "method": method
    }
