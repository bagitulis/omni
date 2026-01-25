#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads Analysis - Main Analysis Functions
==============================================
Product analysis and quality assessment.
Uses shared core modules for statistical analysis.
Scoring functions are in scoring.py for SRP compliance.
"""

import sys
from pathlib import Path
import pandas as pd
from typing import List, Dict, Any

# Add parent to path for core module access
sys.path.insert(0, str(Path(__file__).parent.parent))

from .config import MIN_PERIODS_FOR_TREND, MIN_PERIODS_FOR_MOMENTUM, MIN_COST_FOR_VALID
from .data import filter_data_by_window
from .scoring import calculate_scores, determine_category, get_top_products, get_stop_products, get_monitor_products

# Import from core statistics module
from core.statistics import (
    mann_kendall_test,
    calculate_momentum,
    coefficient_of_variation,
    linear_trend_analysis,
)


def assess_data_quality(product_data: pd.DataFrame) -> Dict[str, Any]:
    """
    Comprehensive data quality assessment for a product.
    
    Returns:
        dict: overall_quality, confidence_level, n_periods, total_cost, warnings, is_reliable
    """
    n_periods = len(product_data)
    total_cost = product_data["cost"].sum()
    
    warnings = []
    confidence = 100.0
    
    # Check number of periods
    if n_periods < 2:
        overall = "❌ INSUFFICIENT"
        confidence = 0
        warnings.append(f"Data sangat terbatas (hanya {n_periods} periode)")
    elif n_periods < MIN_PERIODS_FOR_TREND:
        overall = "⚠️ LOW"
        confidence = 30
        warnings.append(f"Data kurang untuk trend analysis (butuh min {MIN_PERIODS_FOR_TREND} periode)")
    elif n_periods < 8:
        overall = "🔶 MEDIUM"
        confidence = 60
        warnings.append(f"Data cukup, tapi belum optimal ({n_periods}/8 periode)")
    else:
        overall = "✅ HIGH"
        confidence = 90
    
    # Check cost threshold
    if total_cost < MIN_COST_FOR_VALID:
        confidence -= 20
        warnings.append(f"Budget terlalu kecil (Rp {total_cost:,.0f} < Rp {MIN_COST_FOR_VALID:,.0f})")
    
    # Check for missing periods (gaps)
    if n_periods >= 2:
        period_dates = pd.to_datetime(product_data["period_start"]).sort_values()
        date_gaps = period_dates.diff().dt.days.dropna()
        if len(date_gaps) > 0 and date_gaps.max() > 14:
            confidence -= 10
            warnings.append("Terdapat gap data > 2 minggu")
    
    confidence = max(0, min(100, confidence))
    
    return {
        "overall_quality": overall,
        "confidence_level": round(confidence, 1),
        "n_periods": n_periods,
        "total_cost": total_cost,
        "warnings": warnings,
        "is_reliable": confidence >= 50
    }


def analyze_products(df: pd.DataFrame, verbose: bool = True, 
                    use_rolling_window: bool = False, window_months: int = 3) -> List[Dict]:
    """
    Analyze all products with comprehensive ML scoring.
    
    Uses multiple statistical methods from core module:
    - Mann-Kendall Test: Non-parametric trend detection
    - Linear Regression: Slope and R² analysis
    - Momentum Analysis: Recent vs historical comparison
    - Coefficient of Variation: Consistency measurement
    - Composite Scoring: Weighted multi-factor score
    """
    def log(msg):
        if verbose:
            print(msg)
    
    # Apply rolling window if requested
    if use_rolling_window:
        df = filter_data_by_window(df, n_months=window_months, verbose=verbose)
    
    log("📈 Analyzing products with ML methods...")
    
    # Get the latest period in the entire dataset
    latest_period_in_data = df["period_label"].max()
    
    results = []
    
    for product_id, product_df in df.groupby("product_id"):
        result = _analyze_single_product(product_id, product_df, latest_period_in_data)
        results.append(result)
    
    # Sort by composite score (descending)
    results = sorted(results, key=lambda x: x["score"], reverse=True)
    
    # Summary stats
    high_quality = len([r for r in results if "✅" in r["data_quality"]])
    medium_quality = len([r for r in results if "🔶" in r["data_quality"]])
    low_quality = len([r for r in results if "⚠️" in r["data_quality"] or "❌" in r["data_quality"]])
    still_active = len([r for r in results if r.get("is_still_active", True)])
    
    log(f"   ✅ Analyzed {len(results)} products")
    log(f"   📊 Data Quality: HIGH={high_quality}, MEDIUM={medium_quality}, LOW={low_quality}")
    log(f"   🔄 Active Products: {still_active}, Stopped: {len(results) - still_active}")
    
    return results


def _analyze_single_product(product_id: str, product_df: pd.DataFrame, 
                           latest_period_in_data: str = None) -> Dict:
    """Analyze a single product (internal helper)."""
    
    # Data quality assessment
    quality = assess_data_quality(product_df)
    
    # Get display name (use first row's product_name)
    display_name = product_df["product_name"].iloc[0]
    if len(display_name) > 55:
        display_name = display_name[:55]
    
    # Get bidding mode (like creative type in TikTok)
    bidding_mode = product_df["bidding_mode"].iloc[0] if "bidding_mode" in product_df.columns else "Unknown"
    
    # Basic statistics
    total_cost = product_df["cost"].sum()
    total_revenue = product_df["revenue"].sum()
    total_orders = product_df["units_sold"].sum() if "units_sold" in product_df.columns else 0
    profit = total_revenue - total_cost
    roi = total_revenue / total_cost if total_cost > 0 else 0
    
    # Detect if product is still active
    product_last_period = product_df["period_label"].max()
    is_still_active = (product_last_period == latest_period_in_data) if latest_period_in_data else True
    
    # Calculate average weekly metrics
    n_unique_periods = max(1, len(product_df["period_label"].unique()))
    avg_weekly_cost = total_cost / n_unique_periods
    avg_weekly_profit = profit / n_unique_periods
    
    # Period-wise aggregation
    period_data = product_df.groupby("period_start").agg({
        "cost": "sum",
        "revenue": "sum",
        "units_sold": "sum",
    }).reset_index().sort_values("period_start")
    
    n_periods = len(period_data)
    
    # Calculate period values for statistical analysis
    roi_values = (period_data["revenue"] / period_data["cost"].replace(0, 1)).values
    revenue_values = period_data["revenue"].values
    
    # Statistical analysis using core modules
    mk_result = mann_kendall_test(roi_values)
    lr_result = linear_trend_analysis(roi_values)
    momentum_result = calculate_momentum(revenue_values)
    cv_result = coefficient_of_variation(revenue_values)
    
    # Composite scoring
    scores = calculate_scores(roi, profit, momentum_result, cv_result, mk_result, lr_result)
    
    # Category & recommendation
    category, action = determine_category(scores["composite"], roi, profit, quality["is_reliable"])
    
    # Extract CV value (cv_result is a float, not dict)
    cv_value = cv_result if isinstance(cv_result, (int, float)) else cv_result.get("cv", 0)
    cv_consistency = "high" if cv_value < 30 else "medium" if cv_value < 50 else "low"
    
    return {
        # Basic Info
        "product_id": str(product_id),
        "product_name": display_name,
        "bidding_mode": bidding_mode,
        
        # Financial Metrics
        "total_cost": int(total_cost),
        "total_revenue": int(total_revenue),
        "profit": int(profit),
        "roi": round(roi, 2),
        "total_orders": int(total_orders),
        "n_periods": n_periods,
        
        # Mann-Kendall Results
        "trend": mk_result["trend"],
        "trend_p_value": mk_result["p_value"],
        "trend_tau": mk_result.get("tau", 0),
        "trend_confidence": mk_result["confidence"],
        
        # Linear Regression Results
        "lr_direction": lr_result["direction"],
        "lr_slope": lr_result["slope"],
        "lr_r_squared": lr_result["r_squared"],
        
        # Momentum Results (use correct keys from calculate_momentum)
        "momentum": momentum_result["direction"],
        "momentum_pct": momentum_result["momentum_pct"],
        
        # Consistency Results (cv_result is float, not dict)
        "consistency": cv_consistency,
        "cv": cv_value,
        
        # Data Quality Assessment
        "data_quality": quality["overall_quality"],
        "confidence_level": quality["confidence_level"],
        "is_reliable": quality["is_reliable"],
        "quality_warnings": quality["warnings"],
        
        # Scoring
        "roi_score": scores["roi"],
        "profit_score": scores["profit"],
        "momentum_score": scores["momentum"],
        "consistency_score": scores["consistency"],
        "trend_score": scores["trend"],
        "score": scores["composite"],
        
        # Activity Status
        "is_still_active": is_still_active,
        "last_active_period": product_last_period,
        "avg_weekly_cost": int(avg_weekly_cost),
        "avg_weekly_profit": int(avg_weekly_profit),
        
        # Recommendation
        "category": category,
        "action": action
    }


# Re-export from scoring module for backward compatibility
__all__ = [
    "assess_data_quality",
    "analyze_products",
    "get_top_products",
    "get_stop_products",
    "get_monitor_products",
]
