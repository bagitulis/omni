#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Probability Engine Module
=========================
Bayesian probability calculations for ads performance.
Platform-agnostic.
"""

import numpy as np
from scipy import stats
from typing import Dict, List, Any, Optional, Union, Tuple


def calculate_success_probability(
    recent_roas: List[float],
    threshold: float = 1.0,
    prior_alpha: float = 1.0,
    prior_beta: float = 1.0
) -> Dict[str, Any]:
    """
    Calculate probability of success (ROAS > threshold) using Beta distribution.
    
    Uses Bayesian approach with conjugate prior:
    - Prior: Beta(alpha, beta)
    - Likelihood: Bernoulli (success/failure per observation)
    - Posterior: Beta(alpha + successes, beta + failures)
    
    Args:
        recent_roas: List of recent ROAS values
        threshold: Success threshold (default 1.0 = breakeven)
        prior_alpha: Prior alpha (default 1.0 = uninformative)
        prior_beta: Prior beta (default 1.0 = uninformative)
        
    Returns:
        dict with: probability, confidence, posterior_params
    """
    if len(recent_roas) < 1:
        return {
            "probability": 0.5,
            "confidence": "low",
            "posterior_alpha": prior_alpha,
            "posterior_beta": prior_beta,
            "n_observations": 0,
            "n_successes": 0
        }
    
    # Count successes
    successes = sum(1 for r in recent_roas if r >= threshold)
    failures = len(recent_roas) - successes
    
    # Posterior parameters
    post_alpha = prior_alpha + successes
    post_beta = prior_beta + failures
    
    # Mean of posterior = probability estimate
    probability = post_alpha / (post_alpha + post_beta)
    
    # Confidence based on sample size
    n = len(recent_roas)
    if n >= 14:
        confidence = "high"
    elif n >= 7:
        confidence = "medium"
    else:
        confidence = "low"
    
    return {
        "probability": round(probability, 4),
        "confidence": confidence,
        "posterior_alpha": post_alpha,
        "posterior_beta": post_beta,
        "n_observations": n,
        "n_successes": successes
    }


def calculate_failure_risk(
    recent_roas: List[float],
    threshold: float = 0.5,
    lookback_days: int = 7
) -> Dict[str, Any]:
    """
    Calculate risk of failure (sustained poor performance).
    
    Uses exponentially weighted moving average to give
    more weight to recent observations.
    
    Args:
        recent_roas: List of recent ROAS values (newest last)
        threshold: Failure threshold
        lookback_days: Days to analyze
        
    Returns:
        dict with: risk_score, risk_level, ewma_roas
    """
    if len(recent_roas) < 1:
        return {
            "risk_score": 0.5,
            "risk_level": "unknown",
            "ewma_roas": 0
        }
    
    # Limit to lookback
    values = recent_roas[-lookback_days:]
    
    # Calculate EWMA (span = half of lookback)
    span = max(2, len(values) // 2)
    alpha = 2 / (span + 1)
    
    ewma = values[0]
    for v in values[1:]:
        ewma = alpha * v + (1 - alpha) * ewma
    
    # Risk score (0-1, higher = more risk)
    if ewma <= 0:
        risk_score = 1.0
    elif ewma >= threshold * 2:
        risk_score = 0
    else:
        risk_score = 1 - (ewma / (threshold * 2))
    
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
        "risk_level": risk_level,
        "ewma_roas": round(ewma, 4)
    }


def bayesian_update(
    prior_mean: float,
    prior_var: float,
    observation: float,
    observation_var: float
) -> Tuple[float, float]:
    """
    Bayesian update for Gaussian distributions.
    
    Useful for updating predictions with new observations.
    
    Args:
        prior_mean: Prior mean estimate
        prior_var: Prior variance
        observation: New observation
        observation_var: Observation variance
        
    Returns:
        tuple: (posterior_mean, posterior_var)
    """
    # Kalman gain
    K = prior_var / (prior_var + observation_var)
    
    # Posterior
    post_mean = prior_mean + K * (observation - prior_mean)
    post_var = (1 - K) * prior_var
    
    return round(post_mean, 4), round(post_var, 4)


def calculate_monte_carlo_projection(
    historical_values: List[float],
    n_days: int = 7,
    n_simulations: int = 1000,
    confidence_level: float = 0.95
) -> Dict[str, Any]:
    """
    Monte Carlo simulation for future projections.
    
    Samples from historical distribution to project future values.
    
    Args:
        historical_values: Historical metric values
        n_days: Days to project
        n_simulations: Number of simulations
        confidence_level: Confidence level for intervals
        
    Returns:
        dict with: projected_mean, ci_lower, ci_upper, scenarios
    """
    if len(historical_values) < 5:
        return {
            "projected_mean": 0,
            "ci_lower": 0,
            "ci_upper": 0,
            "is_valid": False
        }
    
    values = np.array(historical_values)
    mean = np.mean(values)
    std = np.std(values)
    
    # Simulate
    simulations = []
    for _ in range(n_simulations):
        # Random walk with drift
        path = [values[-1]]
        for _ in range(n_days):
            next_val = path[-1] + np.random.normal(0, std / np.sqrt(n_days))
            path.append(max(0, next_val))  # Non-negative
        simulations.append(path[-1])
    
    # Calculate percentiles
    alpha = 1 - confidence_level
    ci_lower = np.percentile(simulations, alpha / 2 * 100)
    ci_upper = np.percentile(simulations, (1 - alpha / 2) * 100)
    projected_mean = np.mean(simulations)
    
    return {
        "projected_mean": round(projected_mean, 4),
        "ci_lower": round(ci_lower, 4),
        "ci_upper": round(ci_upper, 4),
        "confidence_level": confidence_level,
        "n_simulations": n_simulations,
        "is_valid": True
    }
