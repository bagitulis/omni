#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads - Executive Summary Categorizers
============================================
Functions untuk kategorisasi produk berdasarkan rekomendasi.
"""

from ..analysis import get_stop_products
from .exec_helpers import calculate_confidence_label, estimate_with_methodology


def get_high_impact_products(products, max_products=10):
    """
    Get products with HIGH IMPACT potential for budget increase.
    
    Criteria:
    1. Must be active (is_still_active = True)
    2. ROI > 1.5x
    3. Profit > 0
    4. Momentum not negative
    5. Trend not declining
    """
    candidates = []
    
    for p in products:
        if not p.get('is_still_active', True):
            continue
        if p['roi'] < 1.5:
            continue
        if p['profit'] <= 0:
            continue
        if 'Negatif' in p.get('momentum', ''):
            continue
        if 'Turun' in p.get('trend', ''):
            continue
        
        estimate = estimate_with_methodology(p, budget_change_pct=0.5)
        conf_label, conf_note = calculate_confidence_label(p)
        
        if estimate['profit_change'] <= 0:
            continue
        
        candidates.append({
            **p,
            'budget_increase': estimate['budget_change'],
            'budget_increase_pct': estimate['budget_change_pct'],
            'estimated_additional_profit': estimate['profit_change'],
            'profit_increase_pct': estimate['profit_change_pct'],
            'profit_lower': estimate['profit_lower'],
            'profit_upper': estimate['profit_upper'],
            'margin_of_error': estimate['margin_of_error'],
            'methodology': estimate['methodology'],
            'confidence_label': conf_label,
            'confidence_note': conf_note,
            'impact_score': p['score'] * (1 + estimate['profit_change'] / 100000)
        })
    
    candidates.sort(key=lambda x: x['impact_score'], reverse=True)
    return candidates[:max_products]


def get_potential_restart_products(products):
    """
    Get stopped products that have potential and should be restarted.
    
    Criteria:
    1. Must be stopped (is_still_active = False)
    2. ROI >= 0.8 (at least 80% return)
    3. Not too much loss (profit > -50000)
    4. Had positive momentum OR stable trend before stopping
    """
    candidates = []
    
    for p in products:
        if p.get('is_still_active', True):
            continue
        if p['roi'] < 0.8:
            continue
        if p['profit'] < -50000:
            continue
        
        momentum_ok = 'Positif' in p.get('momentum', '') or 'Stabil' in p.get('momentum', '')
        trend_ok = 'Naik' in p.get('trend', '') or 'Stabil' in p.get('trend', '')
        
        if not (momentum_ok or trend_ok):
            continue
        
        conf_label, conf_note = calculate_confidence_label(p)
        avg_weekly_cost = p.get('avg_weekly_cost', p['total_cost'] / max(1, p['n_periods']))
        recommended_budget = avg_weekly_cost * 0.7
        
        potential_roi = p['roi'] * 0.8
        potential_profit = (recommended_budget * potential_roi) - recommended_budget
        
        if potential_profit <= 0:
            continue
        
        candidates.append({
            **p,
            'recommended_restart_budget': recommended_budget,
            'potential_weekly_profit': potential_profit,
            'potential_monthly_profit': potential_profit * 4,
            'confidence_label': conf_label,
            'confidence_note': conf_note,
            'restart_reason': 'ROI positif dengan trend/momentum baik sebelum di-stop'
        })
    
    candidates.sort(key=lambda x: x['potential_monthly_profit'], reverse=True)
    return candidates[:5]


def get_maintain_budget_products(products, high_impact_ids=None):
    """
    Get products that should maintain current budget (stable performers).
    
    Criteria:
    1. Must be active
    2. Profit positive
    3. Trend not declining
    4. NOT already in high_impact list (to avoid duplicates)
    5. Score between 50-80 OR ROI between 1.2-10.0 (broader range)
    """
    high_impact_ids = high_impact_ids or set()
    candidates = []
    
    for p in products:
        if not p.get('is_still_active', True):
            continue
        if p['profit'] <= 0:
            continue
        if 'Turun' in p.get('trend', ''):
            continue
        
        # Skip if already in high_impact
        product_id = (p['product_name'], p['creative_type'])
        if product_id in high_impact_ids:
            continue
        
        # Broader criteria: moderate score OR moderate ROI
        score_moderate = 50 <= p['score'] <= 80
        roi_moderate = 1.2 <= p['roi'] <= 10.0
        
        if not (score_moderate or roi_moderate):
            continue
        
        conf_label, conf_note = calculate_confidence_label(p)
        
        candidates.append({
            **p,
            'recommendation': 'PERTAHANKAN',
            'reason': f"ROI {p['roi']:.1f}x, score {p['score']:.0f}",
            'confidence_label': conf_label,
            'confidence_note': conf_note
        })
    
    candidates.sort(key=lambda x: x['profit'], reverse=True)
    return candidates


def get_all_stop_products_categorized(products):
    """
    Categorize ALL stop products with detailed analysis.
    
    Returns:
        dict: {
            'active': list of products still active (need action),
            'stopped': list of products already stopped (info only),
            'total_potential_loss': potential loss if active products continue
        }
    """
    stop = get_stop_products(products)
    
    active_stop = []
    already_stopped = []
    
    for p in stop:
        loss = abs(p['profit']) if p['profit'] < 0 else 0
        weekly_burn = p.get('avg_weekly_cost', p['total_cost'] / max(1, p['n_periods']))
        weekly_loss = p.get('avg_weekly_profit', p['profit'] / max(1, p['n_periods']))
        
        potential_monthly_loss = abs(weekly_loss) * 4 if weekly_loss < 0 else weekly_burn * 0.3
        loss_pct = (loss / max(1, p['total_cost'])) * 100
        
        conf_label, conf_note = calculate_confidence_label(p)
        
        enriched = {
            **p,
            'loss': loss,
            'loss_pct': loss_pct,
            'weekly_burn': weekly_burn,
            'potential_monthly_loss': potential_monthly_loss,
            'urgency': 'CRITICAL' if loss > 500000 else 'HIGH' if loss > 100000 else 'MEDIUM',
            'confidence_label': conf_label,
            'confidence_note': conf_note
        }
        
        if p.get('is_still_active', True):
            active_stop.append(enriched)
        else:
            already_stopped.append(enriched)
    
    active_stop.sort(key=lambda x: x['loss'], reverse=True)
    already_stopped.sort(key=lambda x: x['loss'], reverse=True)
    
    return {
        'active': active_stop,
        'stopped': already_stopped,
        'total_potential_loss': sum(p['potential_monthly_loss'] for p in active_stop)
    }
