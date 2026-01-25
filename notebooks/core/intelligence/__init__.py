#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Intelligence Subpackage
=======================
ML/AI analysis engine for ads optimization.
Platform-agnostic.
"""

from .scoring import (
    CompositeScorer,
    ScoringWeights,
    calculate_unified_score,
    normalize_metric,
    get_risk_level,
    get_recommendation_action
)
from .probability import (
    calculate_success_probability,
    calculate_failure_risk,
    bayesian_update,
    calculate_monte_carlo_projection
)
from .saturation import (
    SaturationModel,
    detect_saturation_point,
    calculate_marginal_returns
)
from .fatigue import (
    FatigueDetector,
    FatigueThresholds,
    detect_creative_fatigue,
    calculate_fatigue_index
)
from .lifecycle import (
    ProductLifecycle,
    LifecycleStage,
    detect_lifecycle_stage,
    predict_lifecycle_remaining,
    calculate_churn_risk
)
from .roas import (
    RoasClassifier,
    RoasTier,
    RoasThresholds,
    RoasResult,
    RecommendedAction,
    classify_roas,
    get_roas_tier,
    get_roas_score,
)
from .unified_types import (
    ActionRecommendation,
    ScoreCategory,
    ScoringWeightsConfig,
    ActionThresholds,
    UnifiedScoreResult,
    ScoreInputs,
)
from .unified_scorer import UnifiedScorer
from .probability_ci import (
    ProbabilityWithCI,
    ProbabilityEstimate,
    estimate_probability,
)
from .elasticity import (
    ElasticityLevel,
    ElasticityResult,
    calculate_budget_elasticity,
    calculate_marginal_roi,
    get_elasticity_score,
)
from .churn_risk import (
    ChurnRiskCalculator,
    RiskFactor,
    RiskSeverity,
    RiskThresholds,
    RiskPredictions,
    ChurnRiskLevel,
    ChurnRiskResult,
    calculate_cost_revenue_correlation_analysis,
)

__all__ = [
    # Scoring
    "CompositeScorer",
    "ScoringWeights",
    "calculate_unified_score",
    "normalize_metric",
    "get_risk_level",
    "get_recommendation_action",
    # Probability
    "calculate_success_probability",
    "calculate_failure_risk",
    "bayesian_update",
    "calculate_monte_carlo_projection",
    # Saturation
    "SaturationModel",
    "detect_saturation_point",
    "calculate_marginal_returns",
    # Fatigue
    "FatigueDetector",
    "FatigueThresholds",
    "detect_creative_fatigue",
    "calculate_fatigue_index",
    # Lifecycle
    "ProductLifecycle",
    "LifecycleStage",
    "detect_lifecycle_stage",
    "predict_lifecycle_remaining",
    "calculate_churn_risk",
    # ROAS
    "RoasClassifier",
    "RoasTier",
    "RoasThresholds",
    "RoasResult",
    "RecommendedAction",
    "classify_roas",
    "get_roas_tier",
    "get_roas_score",
    # Unified Scorer
    "ActionRecommendation",
    "ScoreCategory",
    "ScoringWeightsConfig",
    "ActionThresholds",
    "UnifiedScoreResult",
    "ScoreInputs",
    "UnifiedScorer",
    # Probability with CI
    "ProbabilityWithCI",
    "ProbabilityEstimate",
    "estimate_probability",
    # Elasticity
    "ElasticityLevel",
    "ElasticityResult",
    "calculate_budget_elasticity",
    "calculate_marginal_roi",
    "get_elasticity_score",
    # Churn Risk
    "ChurnRiskCalculator",
    "RiskFactor",
    "RiskSeverity",
    "RiskThresholds",
    "RiskPredictions",
    "ChurnRiskLevel",
    "ChurnRiskResult",
    "calculate_cost_revenue_correlation_analysis",
]
