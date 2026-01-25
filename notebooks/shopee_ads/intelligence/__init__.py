#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads Intelligence System
==============================
Comprehensive analysis using shared core intelligence modules.
Same methodology as TikTok Ads for consistency.
"""

__version__ = "1.0.0"
__author__ = "YN Digital"

# Re-export from core intelligence (platform-agnostic)
from core.intelligence import (
    # Scoring
    CompositeScorer,
    ScoringWeights,
    calculate_unified_score,
    normalize_metric,
    get_risk_level,
    get_recommendation_action,
    
    # Probability
    calculate_success_probability,
    calculate_failure_risk,
    bayesian_update,
    calculate_monte_carlo_projection,
    
    # Saturation
    SaturationModel,
    detect_saturation_point,
    calculate_marginal_returns,
    
    # Fatigue
    FatigueDetector,
    FatigueThresholds,
    detect_creative_fatigue,
    calculate_fatigue_index,
    
    # Lifecycle
    ProductLifecycle,
    LifecycleStage,
    detect_lifecycle_stage,
    predict_lifecycle_remaining,
    calculate_churn_risk,
    
    # ROAS
    RoasClassifier,
    RoasTier,
    RoasThresholds,
    RoasResult,
    RecommendedAction,
    classify_roas,
    get_roas_tier,
    get_roas_score,
    
    # Unified Types
    ActionRecommendation,
    ScoreCategory,
    ScoringWeightsConfig,
    ActionThresholds,
    UnifiedScoreResult,
    ScoreInputs,
    
    # Unified Scorer
    UnifiedScorer,
)

# Shopee-specific configuration
from .config_shopee import ShopeeIntelligenceConfig

# Shopee analysis engine
from .shopee_engine import ShopeeIntelligenceEngine

# Shopee comprehensive engine (same methodology as TikTok)
from .comprehensive_engine import ShopeeComprehensiveEngine, ComprehensiveAnalysis

# Unified scorer (wrapper around core)
from .unified_scorer import (
    get_category_emoji,
    get_category_label,
    get_action_label,
    get_action_simple,
)

__all__ = [
    # Core re-exports
    "CompositeScorer", "ScoringWeights", "calculate_unified_score",
    "SaturationModel", "detect_saturation_point",
    "FatigueDetector", "detect_creative_fatigue", "calculate_fatigue_index",
    "ProductLifecycle", "detect_lifecycle_stage", "calculate_churn_risk",
    "RoasClassifier", "classify_roas", "get_roas_tier",
    "UnifiedScorer", "ActionRecommendation",
    
    # Shopee-specific
    "ShopeeIntelligenceConfig", "ShopeeIntelligenceEngine",
    "ShopeeComprehensiveEngine", "ComprehensiveAnalysis",
    
    # Helpers
    "get_category_emoji", "get_category_label", "get_action_label", "get_action_simple",
]
