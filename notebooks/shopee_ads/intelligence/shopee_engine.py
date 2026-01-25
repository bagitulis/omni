#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads Intelligence Engine
==============================
Main orchestrator for comprehensive product analysis.
Uses core intelligence modules with Shopee-specific config.
"""

from typing import Dict, List, Any, Optional
from dataclasses import dataclass
import pandas as pd

from .config_shopee import ShopeeIntelligenceConfig, DEFAULT_CONFIG

# Import from core
from core.intelligence import (
    classify_roas,
    get_roas_tier,
    detect_lifecycle_stage,
    calculate_churn_risk,
    detect_saturation_point,
    calculate_fatigue_index,
    calculate_success_probability,
    get_risk_level,
    get_recommendation_action,
)
from core.statistics import (
    mann_kendall_test,
    calculate_momentum,
    coefficient_of_variation,
)


@dataclass
class ProductIntelligence:
    """Intelligence analysis result for a product."""
    product_id: str
    product_name: str
    
    # Core metrics
    roi: float
    profit: float
    total_cost: float
    total_revenue: float
    
    # ML Analysis
    roas_tier: str
    roas_score: float
    lifecycle_stage: str
    lifecycle_remaining_days: Optional[int]
    churn_risk: float
    saturation_level: float
    fatigue_index: float
    success_probability: float
    
    # Risk Assessment
    risk_level: str
    risk_factors: List[str]
    
    # Recommendation
    action: str
    action_priority: int
    action_details: str
    
    # Composite Score
    unified_score: float
    score_breakdown: Dict[str, float]


class ShopeeIntelligenceEngine:
    """Main intelligence engine for Shopee Ads analysis."""
    
    def __init__(self, config: ShopeeIntelligenceConfig = None):
        self.config = config or DEFAULT_CONFIG
    
    def analyze_product(self, product_data: pd.DataFrame, product_id: str) -> ProductIntelligence:
        """Run comprehensive intelligence analysis on a product."""
        
        # Basic metrics
        total_cost = product_data["cost"].sum()
        total_revenue = product_data["revenue"].sum()
        profit = total_revenue - total_cost
        roi = total_revenue / total_cost if total_cost > 0 else 0
        
        # Extract time series
        period_data = product_data.groupby("period_start").agg({
            "cost": "sum", "revenue": "sum", "units_sold": "sum"
        }).reset_index().sort_values("period_start")
        
        revenue_series = period_data["revenue"].values
        roi_series = (period_data["revenue"] / period_data["cost"].replace(0, 1)).values
        n_periods = len(period_data)
        
        # ROAS Analysis
        roas_thresholds = self.config.get_roas_thresholds()
        roas_result = classify_roas(roi, roas_thresholds)
        roas_tier = get_roas_tier(roi, roas_thresholds)
        
        # Lifecycle Analysis
        lifecycle = detect_lifecycle_stage(
            revenue_series,
            growth_threshold=self.config.lifecycle_growth_threshold,
            decline_threshold=self.config.lifecycle_decline_threshold
        )
        churn_risk = calculate_churn_risk(revenue_series, roi_series)
        
        # Saturation Analysis
        saturation = detect_saturation_point(revenue_series, period_data["cost"].values)
        
        # Fatigue Analysis
        fatigue = calculate_fatigue_index(
            n_periods,
            self.config.fatigue_warning_days,
            self.config.fatigue_critical_days
        )
        
        # Success Probability
        success_prob = calculate_success_probability(roi, profit, n_periods)
        
        # Risk Assessment
        risk_level, risk_factors = self._assess_risk(roi, profit, churn_risk, saturation, fatigue)
        
        # Unified Score
        score_breakdown = self._calculate_score_breakdown(roi, profit, revenue_series, roi_series)
        unified_score = sum(score_breakdown.values())
        
        # Recommendation
        action, priority, details = get_recommendation_action(
            unified_score, roi, profit, risk_level
        )
        
        # Get product name
        product_name = product_data["product_name"].iloc[0] if "product_name" in product_data.columns else str(product_id)
        
        return ProductIntelligence(
            product_id=str(product_id),
            product_name=product_name[:50],
            roi=round(roi, 2),
            profit=int(profit),
            total_cost=int(total_cost),
            total_revenue=int(total_revenue),
            roas_tier=roas_tier,
            roas_score=roas_result.get("score", 0),
            lifecycle_stage=lifecycle.get("stage", "unknown"),
            lifecycle_remaining_days=lifecycle.get("remaining_days"),
            churn_risk=round(churn_risk, 2),
            saturation_level=round(saturation.get("level", 0), 2),
            fatigue_index=round(fatigue, 2),
            success_probability=round(success_prob, 2),
            risk_level=risk_level,
            risk_factors=risk_factors,
            action=action,
            action_priority=priority,
            action_details=details,
            unified_score=round(unified_score, 1),
            score_breakdown=score_breakdown,
        )
    
    def _assess_risk(self, roi: float, profit: float, churn: float, 
                    saturation: dict, fatigue: float) -> tuple:
        """Assess overall risk level and factors."""
        factors = []
        risk_score = 0
        
        if roi < 1.0:
            factors.append("ROI negatif (rugi)")
            risk_score += 30
        elif roi < 2.0:
            factors.append("ROI rendah (< 2x)")
            risk_score += 15
        
        if profit < 0:
            factors.append(f"Kerugian aktif: Rp {abs(profit):,.0f}")
            risk_score += 25
        
        if churn > 0.7:
            factors.append("Risiko churn tinggi")
            risk_score += 20
        
        if saturation.get("level", 0) > 0.8:
            factors.append("Mendekati saturasi")
            risk_score += 15
        
        if fatigue > 0.7:
            factors.append("Kelelahan iklan terdeteksi")
            risk_score += 10
        
        level = get_risk_level(risk_score)
        return level, factors
    
    def _calculate_score_breakdown(self, roi: float, profit: float,
                                   revenue_series, roi_series) -> Dict[str, float]:
        """Calculate individual score components."""
        weights = self.config.get_weights()
        
        # ROI Score (0-100, scaled)
        roi_raw = min(roi * 10, 100)
        roi_score = roi_raw * weights["roi"]
        
        # Profit Score
        profit_raw = 100 if profit > 0 else max(0, 50 + profit / 2000000 * 50)
        profit_score = profit_raw * weights["profit"]
        
        # Momentum Score
        momentum = calculate_momentum(revenue_series)
        momentum_raw = 50 + (momentum.get("momentum_pct", 0) / 2) if momentum.get("is_valid") else 50
        momentum_raw = min(max(momentum_raw, 0), 100)
        momentum_score = momentum_raw * weights["momentum"]
        
        # Consistency Score
        cv = coefficient_of_variation(revenue_series)
        consistency_raw = max(0, 100 - cv)
        consistency_score = consistency_raw * weights["consistency"]
        
        # Trend Score
        mk = mann_kendall_test(roi_series)
        if mk.get("is_valid"):
            if "Naik" in mk.get("trend", ""):
                trend_raw = 90
            elif "Turun" in mk.get("trend", ""):
                trend_raw = 20
            else:
                trend_raw = 50
        else:
            trend_raw = 50
        trend_score = trend_raw * weights["trend"]
        
        return {
            "roi": round(roi_score, 1),
            "profit": round(profit_score, 1),
            "momentum": round(momentum_score, 1),
            "consistency": round(consistency_score, 1),
            "trend": round(trend_score, 1),
        }
    
    def analyze_portfolio(self, df: pd.DataFrame, verbose: bool = True) -> List[ProductIntelligence]:
        """Analyze all products in portfolio."""
        results = []
        
        product_ids = df["product_id"].unique()
        if verbose:
            print(f"🧠 Running intelligence analysis on {len(product_ids)} products...")
        
        for pid in product_ids:
            product_df = df[df["product_id"] == pid]
            try:
                intel = self.analyze_product(product_df, pid)
                results.append(intel)
            except Exception as e:
                if verbose:
                    print(f"   ⚠️ Error analyzing {pid}: {e}")
        
        if verbose:
            high_risk = len([r for r in results if r.risk_level == "HIGH"])
            print(f"   ✅ Analysis complete. High-risk products: {high_risk}")
        
        return results
