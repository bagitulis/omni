#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Budget Optimization
==========================================
- Optimal budget recommendation per product
- Portfolio budget reallocation
"""

from typing import List, Dict


def calculate_optimal_budget(product: Dict, total_budget_available: float = None) -> Dict:
    """
    Calculate optimal budget recommendation for a single product.
    Based on ROI, trend, momentum, and data reliability.
    """
    current_cost = product['total_cost']
    current_roi = product['roi']
    category = product.get('category', '')
    momentum_pct = product.get('momentum_pct', 0)
    
    # Base recommendation
    if 'LANJUTKAN' in category:
        base_change = 0.30
    elif 'HENTIKAN' in category:
        base_change = -1.0
    else:
        base_change = 0.0
    
    # Adjust for momentum and ROI
    if momentum_pct > 20:
        base_change += 0.20
    elif momentum_pct < -20:
        base_change -= 0.15
    
    if current_roi > 10:
        base_change += 0.15
    elif current_roi < 2:
        base_change -= 0.20
    
    base_change = max(-1.0, min(0.50, base_change))
    
    budget_change = current_cost * base_change
    recommended_budget = max(0, current_cost + budget_change)
    
    roi_conservation = 0.85 if base_change > 0 else 1.0
    estimated_revenue_change = budget_change * current_roi * roi_conservation
    
    # Confidence
    dq = product.get('data_quality', '')
    confidence = 90 if '✅' in dq else 70 if '🔶' in dq else 50
    
    return {
        "product_id": product['product_id'],
        "product_name": product['product_name'],
        "creative_type": product.get('creative_type', 'N/A'),
        "current_budget": round(current_cost),
        "recommended_budget": round(recommended_budget),
        "budget_change": round(budget_change),
        "change_pct": round(base_change * 100, 1),
        "action": "INCREASE" if base_change > 0 else "DECREASE" if base_change < 0 else "MAINTAIN",
        "estimated_revenue_impact": round(estimated_revenue_change),
        "confidence": confidence,
        "rationale": _generate_budget_rationale(product, base_change)
    }


def _generate_budget_rationale(product: Dict, change: float) -> str:
    """Generate explanation for budget recommendation."""
    reasons = []
    if 'LANJUTKAN' in product.get('category', ''):
        reasons.append(f"High score ({product.get('score', 0):.0f})")
    if product.get('roi', 0) > 5:
        reasons.append(f"Strong ROI ({product.get('roi', 0):.1f}x)")
    if product.get('momentum_pct', 0) > 10:
        reasons.append(f"Positive momentum (+{product.get('momentum_pct', 0):.0f}%)")
    if product.get('momentum_pct', 0) < -10:
        reasons.append(f"Declining ({product.get('momentum_pct', 0):.0f}%)")
    if 'HENTIKAN' in product.get('category', ''):
        reasons.append("Below threshold - stop")
    return "; ".join(reasons) if reasons else "Standard"


def optimize_portfolio_budget(
    products: List[Dict],
    total_budget: float = None,
    max_increase_pct: float = 0.50,
    preserve_minimum: float = 0.0
) -> Dict:
    """
    Optimize budget allocation across entire portfolio.
    
    Strategy: Free budget from STOP products, reallocate to top performers.
    """
    if not products:
        return {"is_valid": False}
    
    current_total = sum(p['total_cost'] for p in products)
    if total_budget is None:
        total_budget = current_total
    
    recommendations = [calculate_optimal_budget(p) for p in products]
    
    to_stop = [r for r in recommendations if r['action'] == 'DECREASE' and r['change_pct'] <= -50]
    to_reduce = [r for r in recommendations if r['action'] == 'DECREASE' and r['change_pct'] > -50]
    to_maintain = [r for r in recommendations if r['action'] == 'MAINTAIN']
    to_increase = [r for r in recommendations if r['action'] == 'INCREASE']
    
    budget_freed = sum(r['current_budget'] for r in to_stop)
    budget_from_reductions = sum(abs(r['budget_change']) for r in to_reduce)
    total_available = budget_freed + budget_from_reductions
    
    # Reallocate to top performers
    if to_increase:
        profits = {r['product_id']: next(
            (p['profit'] for p in products if p['product_id'] == r['product_id']), 0
        ) for r in to_increase}
        total_weight = sum(max(0, p) for p in profits.values())
        
        for rec in to_increase:
            if total_weight > 0:
                weight = max(0, profits[rec['product_id']]) / total_weight
                additional = total_available * weight
                rec['additional_from_reallocation'] = round(additional)
                rec['final_recommended_budget'] = round(rec['recommended_budget'] + additional)
            else:
                rec['additional_from_reallocation'] = 0
                rec['final_recommended_budget'] = rec['recommended_budget']
    
    new_total = sum(r.get('final_recommended_budget', r['recommended_budget']) for r in recommendations)
    
    return {
        "is_valid": True,
        "current_total_budget": round(current_total),
        "recommended_total_budget": round(new_total),
        "budget_change": round(new_total - current_total),
        "budget_freed": round(budget_freed),
        "budget_reallocated": round(total_available),
        "estimated_revenue_impact": round(sum(r['estimated_revenue_impact'] for r in recommendations)),
        "summary": {
            "products_to_stop": len(to_stop),
            "products_to_reduce": len(to_reduce),
            "products_to_maintain": len(to_maintain),
            "products_to_increase": len(to_increase)
        },
        "top_increases": sorted(
            to_increase,
            key=lambda x: x.get('final_recommended_budget', x['recommended_budget']) - x['current_budget'],
            reverse=True
        )[:10],
        "to_stop": to_stop,
        "all_recommendations": recommendations
    }
