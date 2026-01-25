#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Main Analysis Functions
==============================================
Product analysis, quality assessment, dan scoring.
"""

import pandas as pd

from .config import (
    MIN_PERIODS_FOR_TREND, MIN_PERIODS_FOR_MOMENTUM, MIN_COST_FOR_VALID,
    WEIGHT_ROI, WEIGHT_PROFIT, WEIGHT_MOMENTUM, WEIGHT_CONSISTENCY, WEIGHT_TREND,
    SCORE_THRESHOLD_GOOD, SCORE_THRESHOLD_BAD, ROI_THRESHOLD_GOOD, ROI_THRESHOLD_BAD, LOSS_THRESHOLD
)
from .stats import mann_kendall_test, linear_trend_analysis, calculate_momentum, coefficient_of_variation
from .data import get_product_display_name
from .period import filter_data_by_window


def assess_data_quality(product_data):
    """
    Comprehensive data quality assessment for a product.
    
    Returns:
        dict: overall_quality, confidence_level, n_periods, total_cost, warnings, is_reliable
    """
    n_periods = len(product_data)
    total_cost = product_data['cost'].sum()
    
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
    elif n_periods < 8:  # MIN_SAMPLES_HIGH_CONFIDENCE
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
        period_dates = pd.to_datetime(product_data['periodStart']).sort_values()
        date_gaps = period_dates.diff().dt.days.dropna()
        if date_gaps.max() > 14:  # Gap > 2 weeks
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


def analyze_products(df, product_map, verbose=True, use_rolling_window=False, window_months=3):
    """
    Analyze all products with comprehensive ML scoring.
    
    Uses multiple statistical methods:
    - Mann-Kendall Test: Non-parametric trend detection
    - Linear Regression: Slope and R² analysis
    - Momentum Analysis: Recent vs historical comparison
    - Coefficient of Variation: Consistency measurement
    - Composite Scoring: Weighted multi-factor score
    
    Args:
        df: DataFrame from load_data()
        product_map: dict from load_product_names()
        verbose: Print progress
        use_rolling_window: If True, only analyze last n months
        window_months: Window size for rolling analysis (default: 3)
        
    Returns:
        list[dict]: Sorted list of product analysis results
    """
    def log(msg):
        if verbose:
            print(msg)
    
    # Apply rolling window if requested
    if use_rolling_window:
        df = filter_data_by_window(df, n_months=window_months, verbose=verbose)
    
    log("📈 Analyzing products with ML methods...")
    
    # Get the latest period in the entire dataset (to detect active/stopped products)
    latest_period_in_data = df['periodLabel'].max()
    
    results = []
    
    for product_id, product_df in df.groupby('productId'):
        result = _analyze_single_product(product_id, product_df, product_map, latest_period_in_data)
        results.append(result)
    
    # Sort by composite score (descending)
    results = sorted(results, key=lambda x: x['score'], reverse=True)
    
    # Summary stats
    high_quality = len([r for r in results if '✅' in r['data_quality']])
    medium_quality = len([r for r in results if '🔶' in r['data_quality']])
    low_quality = len([r for r in results if '⚠️' in r['data_quality'] or '❌' in r['data_quality']])
    still_active = len([r for r in results if r.get('is_still_active', True)])
    
    log(f"   ✅ Analyzed {len(results)} products")
    log(f"   📊 Data Quality: HIGH={high_quality}, MEDIUM={medium_quality}, LOW={low_quality}")
    log(f"   🔄 Active Products: {still_active}, Stopped: {len(results) - still_active}")
    
    return results


def _analyze_single_product(product_id, product_df, product_map, latest_period_in_data=None):
    """Analyze a single product (internal helper)"""
    
    # Data quality assessment
    quality = assess_data_quality(product_df)
    
    # Get display name
    campaign_name = product_df['campaignName'].iloc[0]
    video_title = product_df['videoTitle'].iloc[0]
    display_name = get_product_display_name(product_id, campaign_name, video_title, product_map)
    creative_type = product_df['creativeType'].iloc[0]
    
    # Basic statistics
    total_cost = product_df['cost'].sum()
    total_revenue = product_df['grossRevenue'].sum()
    total_orders = product_df['ordersSku'].sum()
    profit = total_revenue - total_cost
    roi = total_revenue / total_cost if total_cost > 0 else 0
    
    # Detect if product is still active (has data in latest period)
    product_last_period = product_df['periodLabel'].max()
    is_still_active = (product_last_period == latest_period_in_data) if latest_period_in_data else True
    
    # Calculate average weekly metrics for projections
    avg_weekly_cost = total_cost / max(1, len(product_df['periodLabel'].unique()))
    avg_weekly_profit = profit / max(1, len(product_df['periodLabel'].unique()))
    
    # Period-wise aggregation
    period_data = product_df.groupby('periodStart').agg({
        'cost': 'sum',
        'grossRevenue': 'sum',
        'ordersSku': 'sum',
        'roi': 'mean'
    }).reset_index().sort_values('periodStart')
    
    n_periods = len(period_data)
    
    # Calculate period values
    roi_values = (period_data['grossRevenue'] / period_data['cost'].replace(0, 1)).values
    revenue_values = period_data['grossRevenue'].values
    
    # Statistical analysis
    mk_result = mann_kendall_test(roi_values)
    lr_result = linear_trend_analysis(roi_values)
    momentum_result = calculate_momentum(revenue_values)
    cv_result = coefficient_of_variation(revenue_values)
    
    # Composite scoring
    scores = _calculate_scores(roi, profit, momentum_result, cv_result, mk_result, lr_result)
    
    # Category & recommendation
    category, action = _determine_category(scores['composite'], roi, profit, quality['is_reliable'])
    
    # Format creative type
    creative_type_display = "Video" if creative_type == "Video" else "Kartu"
    
    return {
        # Basic Info
        'product_id': str(product_id),
        'product_name': display_name[:55] if len(display_name) > 55 else display_name,
        'creative_type': creative_type_display,
        
        # Financial Metrics
        'total_cost': int(total_cost),
        'total_revenue': int(total_revenue),
        'profit': int(profit),
        'roi': round(roi, 2),
        'total_orders': int(total_orders),
        'n_periods': n_periods,
        
        # Mann-Kendall Results
        'trend': mk_result['trend'],
        'trend_p_value': mk_result['p_value'],
        'trend_tau': mk_result.get('tau', 0),
        'trend_confidence': mk_result['confidence'],
        'trend_data_quality': mk_result.get('data_quality', 'N/A'),
        
        # Linear Regression Results
        'lr_direction': lr_result['direction'],
        'lr_slope': lr_result['slope'],
        'lr_r_squared': lr_result['r_squared'],
        'lr_p_value': lr_result['p_value'],
        
        # Momentum Results
        'momentum': momentum_result['momentum'],
        'momentum_pct': momentum_result['change_pct'],
        'recent_avg': momentum_result.get('recent_avg', 0),
        'historical_avg': momentum_result.get('historical_avg', 0),
        'momentum_data_quality': momentum_result.get('data_quality', 'N/A'),
        
        # Consistency Results
        'consistency': cv_result['consistency'],
        'cv': cv_result['cv'],
        
        # Data Quality Assessment
        'data_quality': quality['overall_quality'],
        'confidence_level': quality['confidence_level'],
        'is_reliable': quality['is_reliable'],
        'quality_warnings': quality['warnings'],
        
        # Scoring
        'roi_score': scores['roi'],
        'profit_score': scores['profit'],
        'momentum_score': scores['momentum'],
        'consistency_score': scores['consistency'],
        'trend_score': scores['trend'],
        'score': scores['composite'],
        
        # Activity Status (for stop/active detection)
        'is_still_active': is_still_active,
        'last_active_period': product_last_period,
        'avg_weekly_cost': int(avg_weekly_cost),
        'avg_weekly_profit': int(avg_weekly_profit),
        
        # Recommendation
        'category': category,
        'action': action
    }


def _calculate_scores(roi, profit, momentum_result, cv_result, mk_result, lr_result):
    """Calculate component and composite scores"""
    
    # ROI Score (30% weight)
    roi_score = min(roi * 10, 100)
    
    # Profit Score (25% weight)
    if profit > 0:
        profit_score = 100
    else:
        profit_score = max(0, 50 + (profit / 2000000) * 50)
    
    # Momentum Score (20% weight)
    if momentum_result['is_valid'] and momentum_result['change_pct']:
        momentum_score = 50 + (momentum_result['change_pct'] / 2)
        momentum_score = min(max(momentum_score, 0), 100)
    else:
        momentum_score = 50
    
    # Consistency Score (15% weight)
    if cv_result['is_valid'] and cv_result['cv'] is not None:
        consistency_score = max(0, 100 - cv_result['cv'])
    else:
        consistency_score = 50
    
    # Trend Score (10% weight)
    if mk_result['is_valid']:
        if 'Naik' in mk_result['trend'] or lr_result['direction'] == 'naik':
            trend_score = 90
        elif 'Turun' in mk_result['trend'] or lr_result['direction'] == 'turun':
            trend_score = 20
        else:
            trend_score = 50
    else:
        trend_score = 50
    
    # Composite
    composite = (
        roi_score * WEIGHT_ROI +
        profit_score * WEIGHT_PROFIT +
        momentum_score * WEIGHT_MOMENTUM +
        consistency_score * WEIGHT_CONSISTENCY +
        trend_score * WEIGHT_TREND
    )
    
    return {
        'roi': round(roi_score, 1),
        'profit': round(profit_score, 1),
        'momentum': round(momentum_score, 1),
        'consistency': round(consistency_score, 1),
        'trend': round(trend_score, 1),
        'composite': round(composite, 1)
    }


def _determine_category(score, roi, profit, is_reliable):
    """Determine product category and action"""
    
    if score >= SCORE_THRESHOLD_GOOD and roi >= ROI_THRESHOLD_GOOD and profit > 0:
        category = "🟢 LANJUTKAN"
        action = "Scale up budget 30-50%"
    elif score < SCORE_THRESHOLD_BAD or roi < ROI_THRESHOLD_BAD or profit < LOSS_THRESHOLD:
        category = "🔴 HENTIKAN"
        if profit < 0:
            action = f"STOP! ROI {roi:.2f}x, Rugi Rp {abs(profit):,.0f}"
        else:
            action = f"STOP! ROI {roi:.2f}x terlalu rendah"
    else:
        category = "🟡 PANTAU"
        action = "Monitor 7 hari, evaluasi ulang"
    
    # Add quality caveat
    if not is_reliable:
        action = f"[DATA TERBATAS] {action}"
    
    return category, action


# Quick access functions
def get_top_products(products, n=10, creative_type=None):
    """Get top N products by score"""
    filtered = products
    if creative_type:
        filtered = [p for p in products if p['creative_type'] == creative_type]
    return [p for p in filtered if 'LANJUTKAN' in p['category']][:n]


def get_stop_products(products, creative_type=None):
    """Get products that should be stopped"""
    filtered = products
    if creative_type:
        filtered = [p for p in products if p['creative_type'] == creative_type]
    return [p for p in filtered if 'HENTIKAN' in p['category']]


def get_monitor_products(products, creative_type=None):
    """Get products that need monitoring"""
    filtered = products
    if creative_type:
        filtered = [p for p in products if p['creative_type'] == creative_type]
    return [p for p in filtered if 'PANTAU' in p['category']]
