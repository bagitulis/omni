#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Unified Scoring Module
======================
Composite scoring system for ads performance.
Platform-agnostic.
"""

import numpy as np
from typing import Dict, List, Any, Optional, Union
from dataclasses import dataclass, field


@dataclass
class ScoringWeights:
    """Default weights for composite scoring."""
    roas: float = 0.20
    trend: float = 0.15
    momentum: float = 0.10
    elasticity: float = 0.15
    volatility: float = 0.10
    fatigue: float = 0.10
    churn: float = 0.10
    event: float = 0.10
    
    def to_dict(self) -> Dict[str, float]:
        return {
            "roas": self.roas,
            "trend": self.trend,
            "momentum": self.momentum,
            "elasticity": self.elasticity,
            "volatility": self.volatility,
            "fatigue": self.fatigue,
            "churn": self.churn,
            "event": self.event,
        }
    
    def validate(self) -> bool:
        """Check weights sum to 1.0."""
        total = sum(self.to_dict().values())
        return abs(total - 1.0) < 0.001


class CompositeScorer:
    """
    Unified scoring engine for ads performance.
    
    Uses weighted composite of multiple metrics:
    - ROAS (20%): Return efficiency
    - Trend (15%): Direction from Mann-Kendall
    - Momentum (10%): 70/30 split ratio
    - Elasticity (15%): Budget responsiveness
    - Volatility (10%): CV stability (inverse)
    - Fatigue (10%): Creative freshness (inverse)
    - Churn (10%): Customer retention (inverse)
    - Event (10%): Calendar boost
    """
    
    def __init__(self, weights: Optional[ScoringWeights] = None):
        self.weights = weights or ScoringWeights()
    
    def calculate(self, metrics: Dict[str, float]) -> Dict[str, Any]:
        """
        Calculate composite score from multiple metrics.
        
        Args:
            metrics: Dict with keys matching weight names
                     Values should be normalized 0-100
                     
        Returns:
            dict with: score, components, risk_level
        """
        w = self.weights.to_dict()
        
        # Extract components
        components = {}
        weighted_sum = 0
        total_weight = 0
        
        for metric_name, weight in w.items():
            value = metrics.get(metric_name, 50)  # Default to neutral
            
            # Inverse metrics (lower is better)
            if metric_name in ["volatility", "fatigue", "churn"]:
                normalized = 100 - value
            else:
                normalized = value
            
            components[metric_name] = {
                "raw": value,
                "normalized": normalized,
                "weighted": normalized * weight
            }
            weighted_sum += normalized * weight
            total_weight += weight
        
        # Final score
        score = weighted_sum / total_weight if total_weight > 0 else 50
        
        return {
            "score": round(score, 2),
            "components": components,
            "risk_level": get_risk_level(score),
            "weights_used": w
        }
    
    def rank_products(
        self,
        products: List[Dict[str, Any]],
        id_key: str = "product_id"
    ) -> List[Dict[str, Any]]:
        """
        Score and rank multiple products.
        
        Args:
            products: List of product dicts with metrics
            id_key: Key for product identifier
            
        Returns:
            Sorted list (highest score first)
        """
        results = []
        
        for product in products:
            product_id = product.get(id_key, "unknown")
            score_result = self.calculate(product)
            
            results.append({
                "product_id": product_id,
                **score_result,
                "original_data": product
            })
        
        # Sort by score descending
        results.sort(key=lambda x: x["score"], reverse=True)
        
        # Add rank
        for i, result in enumerate(results):
            result["rank"] = i + 1
        
        return results


def calculate_unified_score(metrics: Dict[str, float]) -> float:
    """
    Quick helper to calculate unified score.
    
    Args:
        metrics: Dict with metric values (0-100)
        
    Returns:
        float: Composite score (0-100)
    """
    scorer = CompositeScorer()
    result = scorer.calculate(metrics)
    return result["score"]


def normalize_metric(
    value: float,
    min_val: float,
    max_val: float,
    invert: bool = False
) -> float:
    """
    Normalize metric to 0-100 scale.
    
    Args:
        value: Raw metric value
        min_val: Minimum expected value
        max_val: Maximum expected value
        invert: If True, lower values become higher scores
        
    Returns:
        float: Normalized value (0-100)
    """
    if max_val == min_val:
        return 50
    
    # Clamp to range
    value = max(min_val, min(max_val, value))
    
    # Normalize
    normalized = (value - min_val) / (max_val - min_val) * 100
    
    if invert:
        normalized = 100 - normalized
    
    return round(normalized, 2)


def get_risk_level(score: float) -> str:
    """
    Get risk level from score.
    
    Args:
        score: Composite score (0-100)
        
    Returns:
        str: Risk level (safe, moderate, high, critical)
    """
    if score >= 70:
        return "safe"
    elif score >= 50:
        return "moderate"
    elif score >= 30:
        return "high"
    else:
        return "critical"


def get_recommendation_action(score: float, trend: str) -> str:
    """
    Get recommended action based on score and trend.
    
    Args:
        score: Composite score (0-100)
        trend: Trend direction ("increasing", "stable", "decreasing")
        
    Returns:
        str: Recommended action
    """
    if score >= 70:
        if trend == "increasing":
            return "scale_up"
        return "maintain"
    elif score >= 50:
        if trend == "increasing":
            return "optimize"
        elif trend == "decreasing":
            return "investigate"
        return "monitor"
    elif score >= 30:
        if trend == "decreasing":
            return "reduce_budget"
        return "optimize_urgently"
    else:
        return "stop"
