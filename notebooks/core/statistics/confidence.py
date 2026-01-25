#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Confidence Interval Module
==========================
Statistical confidence intervals and hypothesis tests.
Platform-agnostic.
"""

import numpy as np
from scipy import stats
from typing import List, Dict, Any, Union, Tuple


def calculate_confidence_interval(
    values: Union[List[float], np.ndarray],
    confidence_level: float = 0.95,
    min_samples: int = 3
) -> Dict[str, Any]:
    """
    Calculate confidence interval for mean using t-distribution.
    
    Args:
        values: Array of values
        confidence_level: Confidence level (default 0.95 = 95%)
        min_samples: Minimum samples required
        
    Returns:
        dict with: mean, ci_lower, ci_upper, std_error, margin
    """
    values = np.array(values)
    n = len(values)
    
    if n < min_samples:
        return {
            "mean": 0,
            "ci_lower": 0,
            "ci_upper": 0,
            "std_error": 0,
            "margin": 0,
            "confidence_level": confidence_level,
            "n_samples": n,
            "is_valid": False
        }
    
    mean = np.mean(values)
    std = np.std(values, ddof=1)  # Sample std
    std_error = std / np.sqrt(n)
    
    # t-critical value
    alpha = 1 - confidence_level
    t_critical = stats.t.ppf(1 - alpha / 2, df=n - 1)
    margin = t_critical * std_error
    
    return {
        "mean": round(mean, 4),
        "ci_lower": round(mean - margin, 4),
        "ci_upper": round(mean + margin, 4),
        "std_error": round(std_error, 4),
        "margin": round(margin, 4),
        "confidence_level": confidence_level,
        "n_samples": n,
        "is_valid": True
    }


def bootstrap_ci(
    values: Union[List[float], np.ndarray],
    n_bootstrap: int = 1000,
    confidence_level: float = 0.95,
    statistic: str = "mean"
) -> Dict[str, Any]:
    """
    Calculate confidence interval using bootstrap resampling.
    More robust for non-normal distributions.
    
    Args:
        values: Array of values
        n_bootstrap: Number of bootstrap samples
        confidence_level: Confidence level
        statistic: Statistic to calculate ("mean", "median")
        
    Returns:
        dict with: estimate, ci_lower, ci_upper
    """
    values = np.array(values)
    n = len(values)
    
    if n < 3:
        return {
            "estimate": 0,
            "ci_lower": 0,
            "ci_upper": 0,
            "confidence_level": confidence_level,
            "n_samples": n,
            "is_valid": False
        }
    
    # Bootstrap resampling
    stat_func = np.mean if statistic == "mean" else np.median
    bootstrap_stats = []
    
    for _ in range(n_bootstrap):
        sample = np.random.choice(values, size=n, replace=True)
        bootstrap_stats.append(stat_func(sample))
    
    # Calculate percentiles
    alpha = 1 - confidence_level
    ci_lower = np.percentile(bootstrap_stats, alpha / 2 * 100)
    ci_upper = np.percentile(bootstrap_stats, (1 - alpha / 2) * 100)
    estimate = stat_func(values)
    
    return {
        "estimate": round(estimate, 4),
        "ci_lower": round(ci_lower, 4),
        "ci_upper": round(ci_upper, 4),
        "confidence_level": confidence_level,
        "n_samples": n,
        "is_valid": True
    }


def compare_groups_ttest(
    group_a: Union[List[float], np.ndarray],
    group_b: Union[List[float], np.ndarray],
    alpha: float = 0.05
) -> Dict[str, Any]:
    """
    Compare two groups using independent samples t-test.
    Useful for comparing Video vs Kartu creative types.
    
    Args:
        group_a: First group values
        group_b: Second group values
        alpha: Significance level (default 0.05)
        
    Returns:
        dict with: t_statistic, p_value, significant, effect_size (Cohen's d)
    """
    group_a = np.array(group_a)
    group_b = np.array(group_b)
    
    n_a, n_b = len(group_a), len(group_b)
    
    if n_a < 2 or n_b < 2:
        return {
            "t_statistic": 0,
            "p_value": 1.0,
            "significant": False,
            "effect_size": 0,
            "effect_label": "insufficient_data",
            "group_a_mean": 0,
            "group_b_mean": 0,
            "is_valid": False
        }
    
    # T-test
    t_stat, p_value = stats.ttest_ind(group_a, group_b)
    
    # Cohen's d (effect size)
    pooled_std = np.sqrt(
        ((n_a - 1) * np.var(group_a) + (n_b - 1) * np.var(group_b)) / 
        (n_a + n_b - 2)
    )
    
    if pooled_std > 0:
        cohens_d = (np.mean(group_a) - np.mean(group_b)) / pooled_std
    else:
        cohens_d = 0
    
    # Effect size label
    abs_d = abs(cohens_d)
    if abs_d < 0.2:
        effect_label = "negligible"
    elif abs_d < 0.5:
        effect_label = "small"
    elif abs_d < 0.8:
        effect_label = "medium"
    else:
        effect_label = "large"
    
    return {
        "t_statistic": round(t_stat, 4),
        "p_value": round(p_value, 4),
        "significant": p_value < alpha,
        "effect_size": round(cohens_d, 4),
        "effect_label": effect_label,
        "group_a_mean": round(np.mean(group_a), 4),
        "group_b_mean": round(np.mean(group_b), 4),
        "n_a": n_a,
        "n_b": n_b,
        "is_valid": True
    }
