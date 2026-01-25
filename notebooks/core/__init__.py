#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Core Analytics Library
======================
Shared analytics library untuk semua platform (TikTok, Shopee, Lazada, dll).

Modules:
- statistics: Statistical tests (Mann-Kendall, CI, momentum)
- intelligence: ML/AI analysis (scoring, probability, saturation)
- calendar: Calendar events (Indonesian holidays, payday)

Usage:
    from core.statistics import mann_kendall_test, calculate_momentum
    from core.intelligence import CompositeScorer, calculate_success_probability
    from core.calendar import IndonesianCalendar
"""

__version__ = "1.0.0"
__author__ = "Omni Analytics"

# Statistics
from .statistics import (
    mann_kendall_test,
    linear_trend_analysis,
    macd_trend,
    calculate_momentum,
    momentum_label,
    coefficient_of_variation,
    cv_label,
    calculate_confidence_interval,
    bootstrap_ci,
    compare_groups_ttest,
    calculate_pearson,
    calculate_spearman,
)
from .intelligence import (
    CompositeScorer,
    ScoringWeights,
    calculate_unified_score,
    normalize_metric,
    get_risk_level,
    calculate_success_probability,
    calculate_failure_risk,
    SaturationModel,
    detect_saturation_point,
    FatigueDetector,
    detect_creative_fatigue,
    calculate_fatigue_index,
    ProductLifecycle,
    detect_lifecycle_stage,
    calculate_churn_risk,
)

# Calendar
from .calendar import (
    IndonesianCalendar,
    is_payday,
    is_twin_date,
    is_holiday,
    get_event_boost,
    get_upcoming_events,
)

__all__ = [
    # Statistics
    "mann_kendall_test",
    "linear_trend_analysis",
    "macd_trend",
    "calculate_momentum",
    "momentum_label",
    "coefficient_of_variation",
    "cv_label",
    "calculate_confidence_interval",
    "bootstrap_ci",
    "compare_groups_ttest",
    "calculate_pearson",
    "calculate_spearman",
    # Intelligence - Scoring
    "CompositeScorer",
    "ScoringWeights",
    "calculate_unified_score",
    "normalize_metric",
    "get_risk_level",
    # Intelligence - Probability
    "calculate_success_probability",
    "calculate_failure_risk",
    # Intelligence - Saturation
    "SaturationModel",
    "detect_saturation_point",
    # Intelligence - Fatigue
    "FatigueDetector",
    "detect_creative_fatigue",
    "calculate_fatigue_index",
    # Intelligence - Lifecycle
    "ProductLifecycle",
    "detect_lifecycle_stage",
    "calculate_churn_risk",
    # Calendar
    "IndonesianCalendar",
    "is_payday",
    "is_twin_date",
    "is_holiday",
    "get_event_boost",
    "get_upcoming_events",
]
