#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Advanced Statistical Methods
===================================================
Re-exports from submodules for backward compatibility.

Modules:
- confidence: CI, T-Test, Mann-Whitney
- correlation: Cost-Revenue correlation, Churn Risk
- optimization: Budget optimization
"""

# Re-export from confidence module
from .confidence import (
    calculate_confidence_interval,
    calculate_profit_ci,
    compare_creative_types,
)

# Re-export from correlation module
from .correlation import (
    analyze_cost_revenue_correlation,
    calculate_churn_risk,
    analyze_portfolio_churn_risk,
)

# Re-export from optimization module
from .optimization import (
    calculate_optimal_budget,
    optimize_portfolio_budget,
)

__all__ = [
    # Confidence
    "calculate_confidence_interval",
    "calculate_profit_ci", 
    "compare_creative_types",
    # Correlation
    "analyze_cost_revenue_correlation",
    "calculate_churn_risk",
    "analyze_portfolio_churn_risk",
    # Optimization
    "calculate_optimal_budget",
    "optimize_portfolio_budget",
]
