#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads - Executive Summary Helpers
======================================
Helper functions untuk kalkulasi dan estimasi.
"""


def calculate_confidence_label(p):
    """
    Calculate confidence label based on data quality.
    
    Methodology:
    - HIGH: confidence >= 80% AND n_periods >= 6 → Rekomendasi reliable
    - MEDIUM: confidence >= 50% AND n_periods >= 4 → Perlu monitoring
    - LOW: data terbatas → Masih ragu, perlu evaluasi tambahan
    """
    confidence = p.get('confidence_level', 50)
    n_periods = p.get('n_periods', 0)
    
    if confidence >= 80 and n_periods >= 6:
        return "✅ HIGH", "Rekomendasi reliable"
    elif confidence >= 50 and n_periods >= 4:
        return "🔶 MEDIUM", "Perlu monitoring"
    else:
        return "⚠️ LOW (RAGU)", "Data terbatas, perlu evaluasi tambahan"


def estimate_with_methodology(p, budget_change_pct=0.5):
    """
    Estimate profit impact using strict statistical methodology.
    
    Methodology:
    1. Base ROI from historical data (conservative: 70% of current ROI)
    2. Adjust for trend direction (Mann-Kendall result)
    3. Adjust for momentum (recent vs historical)
    4. Apply confidence interval based on data quality
    
    Returns dict with estimates and methodology explanation.
    """
    current_cost = p['total_cost']
    current_roi = p['roi']
    
    # Base estimate with conservative factor
    conservative_factor = 0.7
    
    # Adjust for trend (Mann-Kendall result)
    trend_adj = 1.0
    if 'Naik' in p.get('trend', ''):
        trend_adj = 1.1
    elif 'Turun' in p.get('trend', ''):
        trend_adj = 0.85
    
    # Adjust for momentum (70/30 split recent vs historical)
    momentum_adj = 1.0
    momentum_pct = p.get('momentum_pct', 0) or 0
    if momentum_pct > 20:
        momentum_adj = 1.15
    elif momentum_pct > 5:
        momentum_adj = 1.05
    elif momentum_pct < -5:
        momentum_adj = 0.9
    elif momentum_pct < -20:
        momentum_adj = 0.75
    
    # Calculate adjusted ROI
    adjusted_roi = current_roi * conservative_factor * trend_adj * momentum_adj
    
    # Budget change calculation
    budget_change = current_cost * budget_change_pct
    new_budget = current_cost + budget_change
    
    # Estimate new revenue and profit
    estimated_revenue = new_budget * adjusted_roi
    estimated_profit = estimated_revenue - new_budget
    
    # Current profit for comparison
    current_profit = p['profit']
    profit_change = estimated_profit - current_profit
    profit_change_pct = (profit_change / max(1, abs(current_profit))) * 100 if current_profit != 0 else 0
    
    # Confidence interval based on data quality
    confidence = p.get('confidence_level', 50)
    margin = 0.3 if confidence < 50 else 0.2 if confidence < 70 else 0.1
    
    return {
        'budget_change': budget_change,
        'budget_change_pct': budget_change_pct * 100,
        'new_budget': new_budget,
        'estimated_profit': estimated_profit,
        'profit_change': profit_change,
        'profit_change_pct': profit_change_pct,
        'adjusted_roi': adjusted_roi,
        'margin_of_error': margin * 100,
        'profit_lower': estimated_profit * (1 - margin),
        'profit_upper': estimated_profit * (1 + margin),
        'methodology': f"ROI×{conservative_factor}×Trend({trend_adj})×Momentum({momentum_adj:.2f})"
    }
