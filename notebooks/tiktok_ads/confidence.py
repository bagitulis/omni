#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads - Confidence Interval Wrapper
========================================
TikTok-specific wrapper using core statistics.
Following AGENTS.MD: Clean Code, DRY, SRP.
"""

import numpy as np
from scipy import stats
from typing import List, Dict

# Import core statistics
from core.statistics.confidence import (
    calculate_confidence_interval as core_ci,
    bootstrap_ci as core_bootstrap
)

__all__ = [
    'calculate_confidence_interval',
    'calculate_profit_ci',
    'compare_creative_types'
]


def calculate_confidence_interval(
    values: List[float], 
    confidence: float = 0.95,
    method: str = "bootstrap"
) -> Dict:
    """
    Calculate confidence interval - TikTok wrapper.
    
    Args:
        values: List of values
        confidence: Confidence level (0-1)
        method: "bootstrap" or "normal"
        
    Returns:
        Dict with point_estimate, lower, upper, etc.
    """
    values = np.array([v for v in values if v is not None and not np.isnan(v)])
    n = len(values)
    
    if n < 2:
        return {
            "point_estimate": values[0] if n == 1 else 0,
            "lower": values[0] if n == 1 else 0,
            "upper": values[0] if n == 1 else 0,
            "margin_of_error": 0,
            "confidence_level": confidence,
            "method": method,
            "n_samples": n,
            "is_valid": False
        }
    
    # Use core implementation
    if method == "bootstrap":
        result = core_bootstrap(values, confidence_level=confidence)
    else:
        result = core_ci(values, confidence_level=confidence)
    
    point_estimate = result.get("mean", np.mean(values))
    lower = result.get("ci_lower", point_estimate)
    upper = result.get("ci_upper", point_estimate)
    margin = (upper - lower) / 2
    
    return {
        "point_estimate": round(point_estimate, 2),
        "lower": round(lower, 2),
        "upper": round(upper, 2),
        "margin_of_error": round(margin, 2),
        "confidence_level": confidence,
        "confidence_pct": f"{confidence*100:.0f}%",
        "method": method,
        "n_samples": n,
        "is_valid": True
    }


def calculate_profit_ci(
    current_profit: float,
    projected_gains: float,
    roi_values: List[float],
    confidence: float = 0.95
) -> Dict:
    """
    Calculate confidence interval for projected profit using ROI variability.
    TikTok-specific projection model.
    
    Args:
        current_profit: Current profit amount
        projected_gains: Expected additional profit
        roi_values: Historical ROI values for uncertainty estimation
        confidence: Confidence level
        
    Returns:
        Dict with profit CI
    """
    roi_array = np.array([r for r in roi_values if r > 0])
    
    if len(roi_array) < 3:
        return {
            "point_estimate": current_profit + projected_gains,
            "lower": current_profit + projected_gains * 0.7,
            "upper": current_profit + projected_gains * 1.3,
            "confidence_level": confidence,
            "is_valid": False,
            "note": "Limited data - using conservative range"
        }
    
    # Use CV to estimate uncertainty
    roi_cv = np.std(roi_array) / np.mean(roi_array) if np.mean(roi_array) > 0 else 0.5
    uncertainty_factor = min(roi_cv, 0.5)  # Cap at 50%
    
    point_estimate = current_profit + projected_gains
    lower = current_profit + projected_gains * (1 - uncertainty_factor)
    upper = current_profit + projected_gains * (1 + uncertainty_factor)
    
    return {
        "point_estimate": round(point_estimate),
        "lower": round(lower),
        "upper": round(upper),
        "margin_of_error": round((upper - lower) / 2),
        "confidence_level": confidence,
        "confidence_pct": f"{confidence*100:.0f}%",
        "roi_cv": round(roi_cv * 100, 1),
        "uncertainty_factor": round(uncertainty_factor * 100, 1),
        "is_valid": True
    }


def compare_creative_types(video_rois: List[float], kartu_rois: List[float]) -> Dict:
    """
    Statistical comparison between Video and Kartu creative types.
    TikTok-specific: Uses T-Test and Mann-Whitney U.
    
    Args:
        video_rois: ROI values for Video creative type
        kartu_rois: ROI values for Kartu creative type
        
    Returns:
        Dict with statistical comparison results
    """
    video = np.array([r for r in video_rois if r > 0])
    kartu = np.array([r for r in kartu_rois if r > 0])
    
    if len(video) < 3 or len(kartu) < 3:
        return {
            "is_valid": False,
            "note": "Insufficient data for statistical comparison",
            "video_n": len(video),
            "kartu_n": len(kartu)
        }
    
    # Basic stats
    video_mean, kartu_mean = np.mean(video), np.mean(kartu)
    video_std, kartu_std = np.std(video), np.std(kartu)
    diff_pct = ((kartu_mean - video_mean) / video_mean * 100) if video_mean > 0 else 0
    
    # T-Test (Welch's, unequal variance)
    t_stat, t_pvalue = stats.ttest_ind(kartu, video, equal_var=False)
    
    # Mann-Whitney U (non-parametric)
    u_stat, u_pvalue = stats.mannwhitneyu(kartu, video, alternative='two-sided')
    
    # Effect size (Cohen's d)
    pooled_std = np.sqrt((video_std**2 + kartu_std**2) / 2)
    cohens_d = (kartu_mean - video_mean) / pooled_std if pooled_std > 0 else 0
    
    effect_interpretation = (
        "Negligible" if abs(cohens_d) < 0.2 else
        "Small" if abs(cohens_d) < 0.5 else
        "Medium" if abs(cohens_d) < 0.8 else "Large"
    )
    
    # Overall significance (use more conservative p-value)
    p_value = max(t_pvalue, u_pvalue)
    
    if p_value < 0.01:
        significance, stars = "Highly Significant (p < 0.01)", "***"
    elif p_value < 0.05:
        significance, stars = "Significant (p < 0.05)", "**"
    elif p_value < 0.10:
        significance, stars = "Marginally Significant (p < 0.10)", "*"
    else:
        significance, stars = "Not Significant (p ≥ 0.10)", ""
    
    # Determine winner
    if kartu_mean > video_mean and p_value < 0.05:
        winner, conf = "Kartu", "Statistically confirmed"
    elif video_mean > kartu_mean and p_value < 0.05:
        winner, conf = "Video", "Statistically confirmed"
    else:
        winner = "Kartu" if kartu_mean > video_mean else "Video"
        conf = "Not statistically significant"
    
    return {
        "is_valid": True,
        "video": {
            "n": len(video),
            "mean_roi": round(video_mean, 2),
            "std_roi": round(video_std, 2),
            "median_roi": round(np.median(video), 2)
        },
        "kartu": {
            "n": len(kartu),
            "mean_roi": round(kartu_mean, 2),
            "std_roi": round(kartu_std, 2),
            "median_roi": round(np.median(kartu), 2)
        },
        "comparison": {
            "difference_pct": round(diff_pct, 1),
            "winner": winner,
            "winner_confidence": conf
        },
        "t_test": {
            "statistic": round(t_stat, 4),
            "p_value": round(t_pvalue, 4),
            "significant": t_pvalue < 0.05
        },
        "mann_whitney": {
            "statistic": round(u_stat, 4),
            "p_value": round(u_pvalue, 4),
            "significant": u_pvalue < 0.05
        },
        "effect_size": {
            "cohens_d": round(cohens_d, 3),
            "interpretation": effect_interpretation
        },
        "overall": {
            "p_value": round(p_value, 4),
            "significance": significance,
            "significance_stars": stars
        }
    }
