#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Health Score & Risk Assessment
=====================================================
Portfolio health scoring and risk evaluation.
"""

import math


def calculate_health_score(products):
    """
    Calculate overall portfolio health score (0-100).
    
    Components:
    - Profitability Rate (40%): % produk profitable
    - ROI Performance (30%): Normalized ROI score
    - Trend Health (15%): % produk dengan trend naik/stabil
    - Data Quality (15%): % produk dengan data HIGH/MEDIUM
    """
    if not products:
        return {"health_score": 0, "grade": "N/A", "interpretation": "Tidak ada data"}
    
    n = len(products)
    
    # 1. Profitability Rate (40%)
    profitable = len([p for p in products if p['profit'] > 0])
    profitability_rate = (profitable / n) * 100
    profitability_score = min(profitability_rate, 100)
    
    # 2. ROI Performance (30%) - Log scale
    total_cost = sum(p['total_cost'] for p in products)
    total_revenue = sum(p['total_revenue'] for p in products)
    overall_roi = total_revenue / total_cost if total_cost > 0 else 0
    
    if overall_roi <= 0:
        roi_score = 0
    elif overall_roi < 1:
        roi_score = overall_roi * 50
    else:
        roi_score = min(50 + (math.log10(overall_roi) * 30), 100)
    
    # 3. Trend Health (15%)
    trend_up = len([p for p in products if 'Naik' in p.get('trend', '')])
    trend_stable = len([p for p in products if 'Stabil' in p.get('trend', '')])
    trend_health = ((trend_up + trend_stable * 0.7) / n) * 100
    
    # 4. Data Quality (15%)
    high_quality = len([p for p in products if '✅' in p.get('data_quality', '')])
    medium_quality = len([p for p in products if '🔶' in p.get('data_quality', '')])
    data_quality_score = ((high_quality + medium_quality * 0.6) / n) * 100
    
    # Weighted composite
    health_score = (
        profitability_score * 0.40 +
        roi_score * 0.30 +
        trend_health * 0.15 +
        data_quality_score * 0.15
    )
    
    # Grade
    if health_score >= 85:
        grade, interp = "A+", "EXCELLENT - Portfolio dalam kondisi sangat baik"
    elif health_score >= 75:
        grade, interp = "A", "SANGAT BAIK - Mayoritas iklan profitable"
    elif health_score >= 65:
        grade, interp = "B+", "BAIK - Performa solid dengan ruang improvement"
    elif health_score >= 55:
        grade, interp = "B", "CUKUP BAIK - Beberapa iklan perlu optimasi"
    elif health_score >= 45:
        grade, interp = "C", "PERLU PERHATIAN - Banyak iklan underperforming"
    else:
        grade, interp = "D", "KRITIS - Perlu evaluasi menyeluruh"
    
    return {
        "health_score": round(health_score, 1),
        "grade": grade,
        "interpretation": interp,
        "components": {
            "profitability": {"score": round(profitability_score, 1), "weight": "40%", "value": f"{profitability_rate:.1f}%"},
            "roi_performance": {"score": round(roi_score, 1), "weight": "30%", "value": f"{overall_roi:.2f}x"},
            "trend_health": {"score": round(trend_health, 1), "weight": "15%", "value": f"{(trend_up/n)*100:.1f}% naik"},
            "data_quality": {"score": round(data_quality_score, 1), "weight": "15%", "value": f"{(high_quality/n)*100:.1f}% high"}
        }
    }


def calculate_risk_assessment(products):
    """
    Calculate risk level and identify risk factors.
    
    Risk Factors:
    - High loss concentration (max 30 pts)
    - Too many low-quality data products (max 25 pts)
    - Declining trends (max 25 pts)
    - Over-reliance on few products (max 20 pts)
    """
    if not products:
        return {"risk_level": "N/A", "risk_score": 0}
    
    n = len(products)
    risk_score = 0
    factors = []
    
    # 1. Loss Concentration Risk
    total_loss = sum(abs(p['profit']) for p in products if p['profit'] < 0)
    total_gain = sum(p['profit'] for p in products if p['profit'] > 0)
    
    if total_gain > 0:
        loss_ratio = total_loss / total_gain
        if loss_ratio > 0.3:
            risk_score += 30
            factors.append(f"Rasio kerugian tinggi ({loss_ratio:.1%} dari total gain)")
        elif loss_ratio > 0.15:
            risk_score += 15
            factors.append(f"Rasio kerugian moderat ({loss_ratio:.1%})")
    
    # 2. Data Quality Risk
    low_quality = len([p for p in products if '⚠️' in p.get('data_quality', '') or '❌' in p.get('data_quality', '')])
    low_quality_ratio = low_quality / n
    if low_quality_ratio > 0.4:
        risk_score += 25
        factors.append(f"{low_quality_ratio:.0%} produk dengan data tidak reliable")
    elif low_quality_ratio > 0.2:
        risk_score += 12
        factors.append(f"{low_quality_ratio:.0%} produk perlu data lebih banyak")
    
    # 3. Declining Trend Risk
    declining = len([p for p in products if 'Turun' in p.get('trend', '')])
    declining_ratio = declining / n
    if declining_ratio > 0.3:
        risk_score += 25
        factors.append(f"{declining_ratio:.0%} produk menunjukkan trend turun")
    elif declining_ratio > 0.15:
        risk_score += 12
        factors.append(f"{declining_ratio:.0%} produk dengan trend menurun")
    
    # 4. Concentration Risk
    sorted_by_revenue = sorted(products, key=lambda x: x['total_revenue'], reverse=True)
    total_revenue = sum(p['total_revenue'] for p in products)
    if total_revenue > 0:
        top3_revenue = sum(p['total_revenue'] for p in sorted_by_revenue[:3])
        concentration = top3_revenue / total_revenue
        if concentration > 0.7:
            risk_score += 20
            factors.append(f"Konsentrasi tinggi: Top 3 = {concentration:.0%} revenue")
        elif concentration > 0.5:
            risk_score += 10
            factors.append(f"Konsentrasi moderat: Top 3 = {concentration:.0%} revenue")
    
    # Risk Level
    if risk_score >= 70:
        risk_level = "🔴 TINGGI"
    elif risk_score >= 40:
        risk_level = "🟡 SEDANG"
    else:
        risk_level = "🟢 RENDAH"
    
    return {
        "risk_level": risk_level,
        "risk_score": risk_score,
        "max_score": 100,
        "factors": factors if factors else ["Tidak ada risiko signifikan terdeteksi"]
    }
