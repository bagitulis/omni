#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - AI Evaluation & Profit Estimation
========================================================
Main evaluation entry point and profit impact estimation.
"""

from .health import calculate_health_score, calculate_risk_assessment
from .advanced_stats import (
    calculate_profit_ci,
    compare_creative_types,
    analyze_cost_revenue_correlation,
    analyze_portfolio_churn_risk,
    optimize_portfolio_budget
)


def estimate_profit_impact(products, scale_up_pct=0.3, window_label="periode"):
    """
    Estimate profit impact if recommendations are followed.
    Includes 95% Confidence Interval for projected profit.
    """
    if not products:
        return {}
    
    current_profit = sum(p['profit'] for p in products)
    current_cost = sum(p['total_cost'] for p in products)
    
    top_products = [p for p in products if 'LANJUTKAN' in p['category']]
    stop_products = [p for p in products if 'HENTIKAN' in p['category']]
    
    # Savings from stopping losers
    loss_from_stop = sum(abs(p['profit']) for p in stop_products if p['profit'] < 0)
    cost_freed = sum(p['total_cost'] for p in stop_products)
    
    # Scale-up gains
    projected_additional_profit = 0
    scale_up_details = []
    
    for p in top_products[:10]:
        additional_budget = p['total_cost'] * scale_up_pct
        conservative_roi = p['roi'] * 0.8
        projected_additional = additional_budget * (conservative_roi - 1)
        projected_additional_profit += max(0, projected_additional)
        
        scale_up_details.append({
            "product": p['product_name'][:40],
            "current_profit": p['profit'],
            "additional_budget": int(additional_budget),
            "projected_additional": int(max(0, projected_additional)),
            "confidence": p.get('confidence_level', 50)
        })
    
    # Reallocation potential
    if cost_freed > 0 and top_products:
        avg_top_roi = sum(p['roi'] for p in top_products[:5]) / min(5, len(top_products))
        reallocation_profit = cost_freed * (avg_top_roi * 0.7 - 1)
    else:
        reallocation_profit = 0
    
    projected_profit = current_profit + loss_from_stop + projected_additional_profit + max(0, reallocation_profit)
    growth_pct = ((projected_profit - current_profit) / abs(current_profit) * 100) if current_profit != 0 else 0
    
    # Confidence Interval
    roi_values = [p['roi'] for p in products if p['roi'] > 0]
    total_gains = loss_from_stop + projected_additional_profit + max(0, reallocation_profit)
    
    profit_ci = calculate_profit_ci(current_profit, total_gains, roi_values, 0.95)
    
    return {
        "current": {"profit": int(current_profit), "cost": int(current_cost)},
        "projected": {
            "profit": int(projected_profit),
            "growth_pct": round(growth_pct, 1),
            "profit_lower": profit_ci['lower'],
            "profit_upper": profit_ci['upper'],
            "confidence_interval": f"Rp {profit_ci['lower']:,.0f} - Rp {profit_ci['upper']:,.0f}",
            "margin_of_error": profit_ci['margin_of_error']
        },
        "breakdown": {
            "savings_from_stop": int(loss_from_stop),
            "scale_up_gains": int(projected_additional_profit),
            "reallocation_potential": int(max(0, reallocation_profit)),
            "cost_freed": int(cost_freed)
        },
        "confidence_analysis": {
            "level": "95%",
            "roi_cv": profit_ci.get('roi_cv', 0),
            "uncertainty_factor": profit_ci.get('uncertainty_factor', 0),
            "is_valid": profit_ci.get('is_valid', False)
        },
        "assumptions": {
            "scale_up_percentage": f"{scale_up_pct:.0%}",
            "roi_conservation_factor": "80%",
            "reallocation_roi_factor": "70%"
        },
        "scale_up_details": scale_up_details[:5],
        "window": window_label
    }


def _generate_strategic_insights(products, video_roi, kartu_roi, correlation_analysis, churn_analysis):
    """Generate strategic insights from analysis results."""
    insights = []
    henti = len([p for p in products if 'HENTIKAN' in p['category']])
    
    if kartu_roi > video_roi * 1.5:
        insights.append({
            "type": "OPPORTUNITY",
            "insight": f"Kartu Produk ({kartu_roi:.1f}x) outperform Video ({video_roi:.1f}x) sebesar {(kartu_roi/video_roi-1)*100:.0f}%",
            "action": "Realokasi budget dari Video ke Kartu Produk",
            "methodology": "Perbandingan ROI agregat per creative type"
        })
    
    if henti > 0:
        total_loss = sum(abs(p['profit']) for p in products if p['profit'] < 0)
        insights.append({
            "type": "SAVINGS",
            "insight": f"{henti} produk dengan ROI negatif menyebabkan kerugian Rp {total_loss:,.0f}",
            "action": "Stop segera untuk menghentikan bleeding",
            "methodology": "Identifikasi produk dengan profit < 0 dan ROI < 1x"
        })
    
    top_momentum = [p for p in products if 'Sangat Positif' in p.get('momentum', '')]
    if top_momentum:
        insights.append({
            "type": "GROWTH",
            "insight": f"{len(top_momentum)} produk menunjukkan momentum sangat positif",
            "action": "Prioritaskan scale-up untuk produk dengan momentum positif",
            "methodology": "Momentum Analysis (70/30 recent vs historical)"
        })
    
    # Correlation insights
    if correlation_analysis.get('is_valid'):
        if correlation_analysis['cost_roi'].get('diminishing_returns'):
            insights.append({
                "type": "WARNING",
                "insight": "Diminishing returns: Higher spending = lower ROI",
                "action": "Review high-spend campaigns, consider redistributing",
                "methodology": "Pearson correlation between cost and ROI"
            })
        
        efficient_count = correlation_analysis['efficiency_analysis'].get('efficient_count', 0)
        if efficient_count > 2:
            insights.append({
                "type": "OPPORTUNITY",
                "insight": f"{efficient_count} produk low spend, high ROI teridentifikasi",
                "action": "Scale up produk ini sebagai prioritas utama",
                "methodology": "Efficiency analysis: below median cost, above-avg ROI"
            })
    
    # Churn insights
    if churn_analysis.get('is_valid'):
        high_risk = churn_analysis['distribution'].get('high_risk', 0)
        if high_risk > 0:
            insights.append({
                "type": "ALERT",
                "insight": f"{high_risk} produk memiliki risiko tinggi penurunan performa",
                "action": "Evaluasi dan optimasi segera",
                "methodology": "Churn Risk Model: trend + momentum + volatility"
            })
    
    return insights


def generate_ai_evaluation(products, period_info=None):
    """Generate comprehensive AI evaluation for the report."""
    health = calculate_health_score(products)
    risk = calculate_risk_assessment(products)
    impact = estimate_profit_impact(products)
    
    # Category distribution
    lanjut = len([p for p in products if 'LANJUTKAN' in p['category']])
    pantau = len([p for p in products if 'PANTAU' in p['category']])
    henti = len([p for p in products if 'HENTIKAN' in p['category']])
    
    # Creative type comparison
    video = [p for p in products if p['creative_type'] == 'Video']
    kartu = [p for p in products if p['creative_type'] == 'Kartu']
    
    video_roi = sum(p['total_revenue'] for p in video) / max(sum(p['total_cost'] for p in video), 1)
    kartu_roi = sum(p['total_revenue'] for p in kartu) / max(sum(p['total_cost'] for p in kartu), 1)
    
    # Advanced Analytics
    video_rois = [p['roi'] for p in video if p['roi'] > 0]
    kartu_rois = [p['roi'] for p in kartu if p['roi'] > 0]
    creative_stats = compare_creative_types(video_rois, kartu_rois)
    
    correlation_analysis = analyze_cost_revenue_correlation(products)
    churn_analysis = analyze_portfolio_churn_risk(products)
    budget_optimization = optimize_portfolio_budget(products)
    
    insights = _generate_strategic_insights(products, video_roi, kartu_roi, correlation_analysis, churn_analysis)
    
    return {
        "health": health,
        "risk": risk,
        "impact": impact,
        "distribution": {"lanjutkan": lanjut, "pantau": pantau, "hentikan": henti},
        "creative_comparison": {
            "video": {"count": len(video), "roi": round(video_roi, 2)},
            "kartu": {"count": len(kartu), "roi": round(kartu_roi, 2)}
        },
        "creative_stats": creative_stats,
        "correlation_analysis": correlation_analysis,
        "churn_analysis": churn_analysis,
        "budget_optimization": budget_optimization,
        "strategic_insights": insights,
        "period_info": period_info
    }
