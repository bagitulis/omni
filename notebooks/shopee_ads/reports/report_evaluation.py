#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads - Report Evaluation
==============================
Generate AI evaluation with health score and risk assessment.
"""

import math
from typing import List, Dict


def generate_evaluation(products: List[Dict], period_info: Dict = None) -> Dict:
    """Generate AI evaluation summary from products including health score."""
    
    total_cost = sum(p["total_cost"] for p in products)
    total_revenue = sum(p["total_revenue"] for p in products)
    total_profit = total_revenue - total_cost
    roi = total_revenue / total_cost if total_cost > 0 else 0
    n_products = len(products)
    
    # Category distribution
    lanjut = len([p for p in products if "LANJUTKAN" in p["category"]])
    pantau = len([p for p in products if "PANTAU" in p["category"]])
    henti = len([p for p in products if "HENTIKAN" in p["category"]])
    
    # Potential savings
    savings = sum(abs(p["profit"]) for p in products if p["profit"] < 0)
    
    # Calculate Health Score
    health_data = _calculate_health_score(products, n_products, roi)
    
    # Risk assessment
    risk_data = _assess_risk(n_products, henti, roi, savings, total_cost)
    
    # Status determination
    status, status_emoji = _determine_status(roi)
    
    return {
        "status": status,
        "status_emoji": status_emoji,
        "total_cost": total_cost,
        "total_revenue": total_revenue,
        "total_profit": total_profit,
        "roi": roi,
        "total_products": n_products,
        "distribution": {"LANJUTKAN": lanjut, "PANTAU": pantau, "HENTIKAN": henti},
        "potential_savings": savings,
        "period_info": period_info,
        "health": health_data,
        "risk": risk_data,
    }


def _calculate_health_score(products: List[Dict], n_products: int, roi: float) -> Dict:
    """Calculate portfolio health score with components."""
    if n_products == 0:
        return {"score": 0, "grade": "N/A", "interpretation": "No data", "components": {}}
    
    # Profitability score
    profitable = len([p for p in products if p["profit"] > 0])
    profitability = (profitable / n_products * 100)
    
    # ROI performance score
    roi_score = min(100, max(0, 50 + math.log10(max(roi, 0.1)) * 30))
    
    # Trend health
    up = len([p for p in products if "Naik" in str(p.get("trend", "")) or "increasing" in str(p.get("trend", ""))])
    stable = len([p for p in products if "stable" in str(p.get("trend", "")) or "Stabil" in str(p.get("trend", ""))])
    trend = ((up + stable * 0.7) / n_products * 100)
    
    # Data quality
    high = len([p for p in products if p.get("data_quality", "").startswith("✅")])
    med = len([p for p in products if p.get("data_quality", "").startswith("⚠️")])
    quality = ((high + med * 0.6) / n_products * 100)
    
    # Weighted composite
    health_score = profitability * 0.35 + roi_score * 0.30 + trend * 0.20 + quality * 0.15
    
    # Grade
    if health_score >= 80:
        grade = "A (Excellent)"
    elif health_score >= 65:
        grade = "B (Good)"
    elif health_score >= 50:
        grade = "C (Fair)"
    else:
        grade = "D (Needs Improvement)"
    
    # Calculate actual values for display
    up_pct = up / n_products * 100 if n_products > 0 else 0
    high_pct = high / n_products * 100 if n_products > 0 else 0
    
    return {
        "score": round(health_score, 1),
        "grade": grade,
        "interpretation": f"Portfolio dalam kondisi {'sangat baik' if health_score >= 80 else 'baik' if health_score >= 65 else 'cukup' if health_score >= 50 else 'perlu perhatian'}",
        "components": {
            "Tingkat Profitabilitas": {
                "score": round(profitability, 1), 
                "weight": 35, 
                "value": f"{profitable}/{n_products} produk",
                "formula": "(Produk Untung / Total) × 100"
            },
            "Performa ROI": {
                "score": round(roi_score, 1), 
                "weight": 30, 
                "value": f"{roi:.2f}x",
                "formula": "50 + log₁₀(ROI) × 30"
            },
            "Kesehatan Trend": {
                "score": round(trend, 1), 
                "weight": 20, 
                "value": f"{up_pct:.1f}% naik",
                "formula": "(Naik + Stabil×0.7) / Total"
            },
            "Kualitas Data": {
                "score": round(quality, 1), 
                "weight": 15, 
                "value": f"{high_pct:.1f}% tinggi",
                "formula": "(High + Med×0.6) / Total"
            },
        }
    }


def _assess_risk(n_products: int, henti: int, roi: float, savings: float, total_cost: float) -> Dict:
    """Assess portfolio risk level."""
    risk_factors = []
    risk_score = 0
    
    if n_products > 0 and henti > n_products * 0.2:
        risk_factors.append(f"Tinggi: {henti} produk ({henti/n_products*100:.0f}%) harus dihentikan")
        risk_score += 25
    
    if roi < 2:
        risk_factors.append("ROI keseluruhan rendah (< 2x)")
        risk_score += 20
    
    if total_cost > 0 and savings > total_cost * 0.1:
        risk_factors.append(f"Kerugian signifikan: Rp {savings:,.0f}".replace(",", "."))
        risk_score += 15
    
    # Determine level and emoji
    if risk_score > 40:
        level = "HIGH"
        level_emoji = "🔴"
    elif risk_score > 20:
        level = "MEDIUM"
        level_emoji = "🟡"
    else:
        level = "LOW"
        level_emoji = "🟢"
    
    return {
        "level": level,
        "level_emoji": level_emoji,
        "score": risk_score,
        "max_score": 100,
        "factors": risk_factors if risk_factors else ["Tidak ada risiko signifikan teridentifikasi"],
    }


def _determine_status(roi: float) -> tuple:
    """Determine portfolio status based on ROI."""
    if roi > 5:
        return "SANGAT BAIK", "🌟"
    elif roi > 2:
        return "BAIK", "✅"
    elif roi > 1:
        return "CUKUP", "⚠️"
    else:
        return "PERLU PERHATIAN", "🚨"


def generate_strategic_insights(products: List[Dict]) -> List[Dict]:
    """Generate strategic insights based on data analysis."""
    insights = []
    n_products = len(products)
    if n_products == 0:
        return insights
    
    # 1. SAVINGS: Products to stop
    henti = [p for p in products if "HENTIKAN" in p.get("category", "")]
    if henti:
        total_loss = sum(abs(p["profit"]) for p in henti if p["profit"] < 0)
        insights.append({
            "type": "SAVINGS",
            "insight": f"{len(henti)} produk dengan ROI rendah menyebabkan kerugian Rp {total_loss:,.0f}".replace(",", "."),
            "action": "Hentikan segera untuk menghentikan pemborosan budget",
            "methodology": "Identifikasi produk dengan profit < 0 dan ROI < threshold"
        })
    
    # 2. GROWTH: Products with good momentum
    momentum_up = [p for p in products if "Naik" in str(p.get("trend", "")) 
                   or "Positif" in str(p.get("momentum", ""))]
    if len(momentum_up) >= 3:
        insights.append({
            "type": "GROWTH",
            "insight": f"{len(momentum_up)} produk menunjukkan tren positif",
            "action": "Prioritaskan scale-up untuk produk dengan momentum positif",
            "methodology": "Trend Analysis: Mann-Kendall + Linear Regression"
        })
    
    # 3. OPPORTUNITY: High ROI products with low spend
    lanjut = [p for p in products if "LANJUTKAN" in p.get("category", "")]
    if lanjut:
        median_cost = sorted([p["total_cost"] for p in products])[n_products // 2]
        efficient = [p for p in lanjut if p["total_cost"] < median_cost and p["roi"] > 3]
        if efficient:
            insights.append({
                "type": "OPPORTUNITY",
                "insight": f"{len(efficient)} produk low-spend tapi high-ROI teridentifikasi",
                "action": "Scale up produk ini sebagai prioritas utama",
                "methodology": "Efficiency Analysis: below median cost, above-avg ROI"
            })
    
    # 4. ALERT: High concentration risk
    if henti and len(henti) > n_products * 0.3:
        insights.append({
            "type": "ALERT",
            "insight": f"Konsentrasi tinggi produk bermasalah ({len(henti)/n_products*100:.0f}%)",
            "action": "Evaluasi strategi pemilihan produk untuk iklan",
            "methodology": "Risk Assessment: Distribution Analysis"
        })
    
    # 5. OPPORTUNITY: Low spend diversification
    total_cost = sum(p["total_cost"] for p in products)
    top_3_cost = sum(p["total_cost"] for p in sorted(products, key=lambda x: x["total_cost"], reverse=True)[:3])
    if total_cost > 0 and top_3_cost / total_cost > 0.7:
        insights.append({
            "type": "ALERT",
            "insight": f"Budget terkonsentrasi: Top 3 produk menghabiskan {top_3_cost/total_cost*100:.0f}% budget",
            "action": "Diversifikasi budget ke produk lain yang potensial",
            "methodology": "Budget Concentration Analysis"
        })
    
    return insights
