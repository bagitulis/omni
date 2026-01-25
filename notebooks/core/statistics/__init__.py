#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Statistics Module
=================
Statistical tests dan analysis untuk ads performance.
Platform-agnostic - bisa dipakai untuk TikTok, Shopee, Lazada, dll.
"""

from .trend import mann_kendall_test, linear_trend_analysis, macd_trend
from .momentum import calculate_momentum, momentum_label
from .variation import coefficient_of_variation, cv_label
from .confidence import (
    calculate_confidence_interval,
    bootstrap_ci,
    compare_groups_ttest,
)
from .correlation import (
    calculate_pearson,
    calculate_spearman,
    correlation_strength_label,
    correlation_matrix,
)
from .trend_momentum import (
    TrendMomentumAnalyzer,
    TrendDirection,
    TrendResult,
    analyze_trend_momentum,
    get_trend_direction,
    get_trend_score,
)

__all__ = [
    # Trend
    "mann_kendall_test",
    "linear_trend_analysis",
    "macd_trend",
    # Trend Momentum (MACD-style)
    "TrendMomentumAnalyzer",
    "TrendDirection",
    "TrendResult",
    "analyze_trend_momentum",
    "get_trend_direction",
    "get_trend_score",
    # Momentum
    "calculate_momentum",
    "momentum_label",
    # Variation
    "coefficient_of_variation",
    "cv_label",
    # Confidence
    "calculate_confidence_interval",
    "bootstrap_ci",
    "compare_groups_ttest",
    # Correlation
    "calculate_pearson",
    "calculate_spearman",
    "correlation_strength_label",
    "correlation_matrix",
]
