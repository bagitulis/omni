#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads - Scoring Functions
==============================
Composite scoring calculation and category determination.
"""

from typing import Dict, List

from .config import (
    WEIGHT_ROI, WEIGHT_PROFIT, WEIGHT_MOMENTUM, WEIGHT_CONSISTENCY, WEIGHT_TREND,
    SCORE_THRESHOLD_GOOD, SCORE_THRESHOLD_BAD, ROI_THRESHOLD_GOOD, ROI_THRESHOLD_BAD, LOSS_THRESHOLD
)


def calculate_scores(roi: float, profit: float, momentum_result: Dict, 
                    cv_result, mk_result: Dict, lr_result: Dict) -> Dict:
    """Calculate component and composite scores."""
    
    # ROI Score (30% weight)
    roi_score = min(roi * 10, 100)
    
    # Profit Score (25% weight)
    if profit > 0:
        profit_score = 100
    else:
        profit_score = max(0, 50 + (profit / 2000000) * 50)
    
    # Momentum Score (20% weight)
    if momentum_result["is_valid"] and momentum_result.get("momentum_pct"):
        momentum_score = 50 + (momentum_result["momentum_pct"] / 2)
        momentum_score = min(max(momentum_score, 0), 100)
    else:
        momentum_score = 50
    
    # Consistency Score (15% weight)
    if isinstance(cv_result, dict):
        cv_value = cv_result.get("cv", 50) if cv_result.get("is_valid", False) else 50
    else:
        cv_value = cv_result if cv_result > 0 else 50
    consistency_score = max(0, 100 - cv_value)
    
    # Trend Score (10% weight)
    if mk_result["is_valid"]:
        if "Naik" in mk_result["trend"] or lr_result["direction"] == "naik":
            trend_score = 90
        elif "Turun" in mk_result["trend"] or lr_result["direction"] == "turun":
            trend_score = 20
        else:
            trend_score = 50
    else:
        trend_score = 50
    
    # Composite weighted score
    composite = (
        roi_score * WEIGHT_ROI +
        profit_score * WEIGHT_PROFIT +
        momentum_score * WEIGHT_MOMENTUM +
        consistency_score * WEIGHT_CONSISTENCY +
        trend_score * WEIGHT_TREND
    )
    
    return {
        "roi": round(roi_score, 1),
        "profit": round(profit_score, 1),
        "momentum": round(momentum_score, 1),
        "consistency": round(consistency_score, 1),
        "trend": round(trend_score, 1),
        "composite": round(composite, 1),
    }


def determine_category(score: float, roi: float, profit: float, is_reliable: bool) -> tuple:
    """Determine category and action based on score and metrics."""
    
    if not is_reliable:
        if profit < LOSS_THRESHOLD:
            return "🛑 HENTIKAN (Data Limited)", "Stop - rugi signifikan"
        return "⏸️ PANTAU (Data Limited)", "Monitor - data kurang"
    
    if score >= SCORE_THRESHOLD_GOOD and roi >= ROI_THRESHOLD_GOOD:
        return "✅ LANJUTKAN", "Scale up budget 20-50%"
    elif score >= SCORE_THRESHOLD_GOOD:
        return "✅ LANJUTKAN (Moderat)", "Maintain budget"
    elif score >= SCORE_THRESHOLD_BAD:
        return "⏸️ PANTAU", "Monitor & optimize"
    elif profit < LOSS_THRESHOLD or roi < ROI_THRESHOLD_BAD:
        return "🛑 HENTIKAN", "Stop immediately"
    else:
        return "⚠️ EVALUASI", "Review & decide"


def get_top_products(products: List[Dict], n: int = 10) -> List[Dict]:
    """Get top N performing products by score."""
    sorted_products = sorted(products, key=lambda x: x.get("score", 0), reverse=True)
    return [p for p in sorted_products if "LANJUTKAN" in p.get("category", "")][:n]


def get_stop_products(products: List[Dict]) -> List[Dict]:
    """Get products that should be stopped."""
    return [p for p in products if "HENTIKAN" in p.get("category", "")]


def get_monitor_products(products: List[Dict]) -> List[Dict]:
    """Get products that should be monitored."""
    return [p for p in products if "PANTAU" in p.get("category", "")]
