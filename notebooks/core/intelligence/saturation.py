#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Saturation Detection Module
===========================
Detect diminishing returns and optimal spend levels.
Platform-agnostic.
"""

import numpy as np
from scipy.optimize import curve_fit
from typing import Dict, List, Any, Optional, Union, Tuple


class SaturationModel:
    """
    Models relationship between spend and returns.
    
    Uses log-linear model: returns = a * ln(spend + 1) + b
    This captures diminishing returns as spend increases.
    """
    
    def __init__(self):
        self.params = None
        self.r_squared = 0
        self.fitted = False
    
    def fit(
        self,
        spend: List[float],
        returns: List[float],
        min_samples: int = 5
    ) -> bool:
        """
        Fit saturation model to data.
        
        Args:
            spend: List of spend amounts
            returns: List of return values (revenue, conversions, etc.)
            min_samples: Minimum samples required
            
        Returns:
            bool: True if fit successful
        """
        spend = np.array(spend)
        returns = np.array(returns)
        
        if len(spend) < min_samples:
            return False
        
        try:
            # Log-linear model
            def log_model(x, a, b):
                return a * np.log(x + 1) + b
            
            popt, _ = curve_fit(log_model, spend, returns, maxfev=1000)
            self.params = {"a": popt[0], "b": popt[1]}
            
            # Calculate R-squared
            predicted = log_model(spend, *popt)
            ss_res = np.sum((returns - predicted) ** 2)
            ss_tot = np.sum((returns - np.mean(returns)) ** 2)
            self.r_squared = 1 - (ss_res / ss_tot) if ss_tot > 0 else 0
            
            self.fitted = True
            return True
            
        except Exception:
            return False
    
    def predict(self, spend: float) -> float:
        """Predict returns for given spend."""
        if not self.fitted:
            return 0
        
        a, b = self.params["a"], self.params["b"]
        return a * np.log(spend + 1) + b
    
    def marginal_return(self, spend: float) -> float:
        """Calculate marginal return at spend level (derivative)."""
        if not self.fitted:
            return 0
        
        a = self.params["a"]
        return a / (spend + 1)
    
    def optimal_spend(self, marginal_threshold: float = 0.5) -> float:
        """
        Find optimal spend where marginal return equals threshold.
        
        Args:
            marginal_threshold: Minimum acceptable marginal return
            
        Returns:
            float: Optimal spend amount
        """
        if not self.fitted:
            return 0
        
        a = self.params["a"]
        # Solve: a / (spend + 1) = threshold
        # spend = a / threshold - 1
        return max(0, a / marginal_threshold - 1)


def detect_saturation_point(
    spend: List[float],
    returns: List[float],
    threshold_pct: float = 0.1
) -> Dict[str, Any]:
    """
    Detect saturation point in spend-returns relationship.
    
    Saturation = point where additional spend yields < threshold improvement.
    
    Args:
        spend: List of spend amounts (sorted ascending)
        returns: List of return values
        threshold_pct: Threshold for diminishing returns (default 10%)
        
    Returns:
        dict with: saturation_spend, saturation_return, is_saturated
    """
    if len(spend) < 5:
        return {
            "saturation_spend": 0,
            "saturation_return": 0,
            "is_saturated": False,
            "confidence": "low"
        }
    
    model = SaturationModel()
    if not model.fit(spend, returns):
        return {
            "saturation_spend": 0,
            "saturation_return": 0,
            "is_saturated": False,
            "confidence": "low"
        }
    
    # Find where marginal return drops below threshold
    current_spend = max(spend)
    initial_marginal = model.marginal_return(min(spend))
    
    if initial_marginal <= 0:
        return {
            "saturation_spend": current_spend,
            "saturation_return": model.predict(current_spend),
            "is_saturated": True,
            "confidence": "medium"
        }
    
    threshold_marginal = initial_marginal * threshold_pct
    saturation_spend = model.optimal_spend(threshold_marginal)
    
    is_saturated = current_spend >= saturation_spend
    
    return {
        "saturation_spend": round(saturation_spend, 2),
        "saturation_return": round(model.predict(saturation_spend), 2),
        "current_spend": round(current_spend, 2),
        "is_saturated": is_saturated,
        "r_squared": round(model.r_squared, 4),
        "confidence": "high" if model.r_squared > 0.7 else "medium"
    }


def calculate_marginal_returns(
    spend: List[float],
    returns: List[float]
) -> Dict[str, Any]:
    """
    Calculate marginal returns at different spend levels.
    
    Args:
        spend: List of spend amounts
        returns: List of return values
        
    Returns:
        dict with: marginal_returns (list), efficiency_trend
    """
    if len(spend) < 3:
        return {
            "marginal_returns": [],
            "efficiency_trend": "unknown"
        }
    
    # Sort by spend
    pairs = sorted(zip(spend, returns))
    spend_sorted = [p[0] for p in pairs]
    returns_sorted = [p[1] for p in pairs]
    
    # Calculate marginal returns
    marginals = []
    for i in range(1, len(spend_sorted)):
        delta_spend = spend_sorted[i] - spend_sorted[i-1]
        delta_return = returns_sorted[i] - returns_sorted[i-1]
        
        if delta_spend > 0:
            marginal = delta_return / delta_spend
            marginals.append({
                "spend_level": spend_sorted[i],
                "marginal_return": round(marginal, 4)
            })
    
    # Determine trend
    if len(marginals) >= 2:
        first_half = [m["marginal_return"] for m in marginals[:len(marginals)//2]]
        second_half = [m["marginal_return"] for m in marginals[len(marginals)//2:]]
        
        if np.mean(second_half) < np.mean(first_half) * 0.8:
            trend = "diminishing"
        elif np.mean(second_half) > np.mean(first_half) * 1.2:
            trend = "increasing"
        else:
            trend = "stable"
    else:
        trend = "unknown"
    
    return {
        "marginal_returns": marginals,
        "efficiency_trend": trend
    }
