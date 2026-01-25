#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads - Correlation & Churn Risk Wrapper
==============================================
TikTok-specific implementation using core churn risk calculator.
"""

from typing import List, Dict, Any

# Import core components
from core.intelligence.churn_risk import (
    ChurnRiskCalculator,
    RiskFactor,
    RiskSeverity,
    RiskThresholds,
    RiskPredictions,
    ChurnRiskLevel,
    ChurnRiskResult,
    calculate_cost_revenue_correlation_analysis
)

# Re-export core types for backward compatibility
__all__ = [
    'TikTokChurnRiskCalculator',
    'analyze_cost_revenue_correlation',
    'calculate_churn_risk',
    'analyze_portfolio_churn_risk',
    # Core types
    'ChurnRiskLevel',
    'RiskSeverity',
    'RiskFactor',
    'ChurnRiskResult'
]


# TikTok-specific thresholds
TIKTOK_RISK_THRESHOLDS = RiskThresholds(
    minimal_max=15,
    low_max=35,
    medium_max=60,
    high_max=85
)

TIKTOK_PREDICTIONS = RiskPredictions(
    minimal="✅ Very stable",
    low="🟢 Stable performance expected",
    medium="🟡 Moderate risk - monitor closely",
    high="🔴 High probability of decline",
    critical="⛔ Critical - immediate action required"
)


class TikTokChurnRiskCalculator(ChurnRiskCalculator):
    """
    TikTok-specific churn risk calculator.
    
    Evaluates risk factors based on TikTok ads metrics:
    - Momentum (0-35 points)
    - Trend (0-30 points)
    - Volatility (0-20 points)
    - Data quality (0-15 points)
    """
    
    def __init__(self):
        super().__init__(
            max_score=100,
            thresholds=TIKTOK_RISK_THRESHOLDS,
            predictions=TIKTOK_PREDICTIONS
        )
    
    def _evaluate_factors(self, entity: Dict[str, Any]) -> List[RiskFactor]:
        """Evaluate TikTok-specific risk factors."""
        factors = []
        
        # 1. Momentum risk (max 35 points)
        momentum_pct = entity.get('momentum_pct', 0)
        if momentum_pct < -30:
            factors.append(RiskFactor(
                name="Severe Momentum Decline",
                value=f"{momentum_pct:.1f}%",
                points=35,
                severity=RiskSeverity.HIGH,
                max_points=35
            ))
        elif momentum_pct < -15:
            factors.append(RiskFactor(
                name="Significant Momentum Decline",
                value=f"{momentum_pct:.1f}%",
                points=25,
                severity=RiskSeverity.MEDIUM,
                max_points=35
            ))
        elif momentum_pct < -5:
            factors.append(RiskFactor(
                name="Mild Momentum Decline",
                value=f"{momentum_pct:.1f}%",
                points=15,
                severity=RiskSeverity.LOW,
                max_points=35
            ))
        
        # 2. Trend risk (max 30 points)
        trend = entity.get('trend', '')
        trend_p = entity.get('trend_p_value', 1.0)
        if 'Turun' in trend and trend_p < 0.1:
            factors.append(RiskFactor(
                name="Significant Downtrend",
                value=f"p={trend_p:.3f}",
                points=30,
                severity=RiskSeverity.HIGH,
                max_points=30
            ))
        elif 'Turun' in trend:
            factors.append(RiskFactor(
                name="Declining Trend",
                value=f"p={trend_p:.3f}",
                points=20,
                severity=RiskSeverity.MEDIUM,
                max_points=30
            ))
        
        # 3. Volatility risk (max 20 points)
        cv = entity.get('cv', 50)
        if cv and cv > 80:
            factors.append(RiskFactor(
                name="Very High Volatility",
                value=f"CV={cv:.0f}%",
                points=20,
                severity=RiskSeverity.HIGH,
                max_points=20
            ))
        elif cv and cv > 50:
            factors.append(RiskFactor(
                name="High Volatility",
                value=f"CV={cv:.0f}%",
                points=10,
                severity=RiskSeverity.MEDIUM,
                max_points=20
            ))
        
        # 4. Data quality risk (max 15 points)
        data_quality = entity.get('data_quality', '')
        if '❌' in data_quality or '⚠️' in data_quality:
            factors.append(RiskFactor(
                name="Limited Data",
                value=data_quality,
                points=15,
                severity=RiskSeverity.MEDIUM,
                max_points=15
            ))
        
        return factors
    
    def _get_metadata(self, entity: Dict[str, Any]) -> Dict[str, Any]:
        """Add TikTok-specific metadata."""
        return {
            "creative_type": entity.get('creative_type', 'N/A')
        }


# Singleton instance for convenience
_calculator = TikTokChurnRiskCalculator()


def calculate_churn_risk(product: Dict[str, Any]) -> Dict[str, Any]:
    """
    Calculate churn/decline risk for a single TikTok product.
    
    Backward compatible wrapper function.
    
    Args:
        product: Product dict with momentum_pct, trend, cv, etc.
        
    Returns:
        Risk assessment dict
    """
    result = _calculator.calculate(
        product,
        id_field='product_id',
        name_field='product_name'
    )
    
    # Format for backward compatibility
    level_emoji = {
        ChurnRiskLevel.MINIMAL: "✅ MINIMAL",
        ChurnRiskLevel.LOW: "🟢 LOW",
        ChurnRiskLevel.MEDIUM: "🟡 MEDIUM",
        ChurnRiskLevel.HIGH: "🔴 HIGH",
        ChurnRiskLevel.CRITICAL: "⛔ CRITICAL"
    }
    
    return {
        "product_id": result.entity_id,
        "product_name": result.entity_name,
        "creative_type": result.metadata.get('creative_type', 'N/A'),
        "risk_score": result.risk_score,
        "max_score": result.max_score,
        "risk_level": level_emoji.get(result.risk_level, "🟡 MEDIUM"),
        "prediction": result.prediction,
        "risk_factors": [f.to_dict() for f in result.risk_factors],
        "n_factors": result.n_factors
    }


def analyze_portfolio_churn_risk(products: List[Dict[str, Any]]) -> Dict[str, Any]:
    """
    Analyze churn risk across entire TikTok portfolio.
    
    Backward compatible wrapper function.
    
    Args:
        products: List of product dicts
        
    Returns:
        Portfolio analysis result
    """
    return _calculator.analyze_portfolio(
        products,
        weight_field='total_cost',
        id_field='product_id',
        name_field='product_name'
    )


def analyze_cost_revenue_correlation(products: List[Dict[str, Any]]) -> Dict[str, Any]:
    """
    Analyze correlation between cost (spending) and revenue for TikTok products.
    
    Identifies diminishing returns and efficiency patterns.
    
    Args:
        products: List of product dicts with total_cost, total_revenue, roi
        
    Returns:
        Correlation analysis result with TikTok-specific insights
    """
    import numpy as np
    
    # Get base correlation analysis from core
    result = calculate_cost_revenue_correlation_analysis(
        products,
        cost_field='total_cost',
        revenue_field='total_revenue',
        roi_field='roi',
        min_entities=5
    )
    
    if not result.get('is_valid'):
        return result
    
    # Add TikTok-specific efficiency analysis
    avg_roi = result['summary_stats']['avg_roi']
    median_cost = result['summary_stats']['median_cost']
    
    inefficient = [
        {
            "product_id": p['product_id'],
            "product_name": p['product_name'],
            "creative_type": p.get('creative_type', 'N/A'),
            "cost": p['total_cost'],
            "roi": p['roi'],
            "issue": "High spend, below-average ROI"
        }
        for p in products
        if p['total_cost'] > median_cost and p['roi'] < avg_roi
    ]
    
    efficient = [
        {
            "product_id": p['product_id'],
            "product_name": p['product_name'],
            "creative_type": p.get('creative_type', 'N/A'),
            "cost": p['total_cost'],
            "roi": p['roi'],
            "opportunity": "Low spend, high ROI - scale up"
        }
        for p in products
        if p['total_cost'] < median_cost and p['roi'] > avg_roi * 1.5
    ]
    
    result['efficiency_analysis'] = {
        "avg_roi": avg_roi,
        "median_cost": median_cost,
        "inefficient_count": len(inefficient),
        "efficient_count": len(efficient),
        "inefficient_products": inefficient[:5],
        "efficient_products": efficient[:5]
    }
    
    # Generate TikTok-specific insights
    result['insights'] = _generate_tiktok_insights(
        result['cost_revenue']['pearson_r'],
        result['cost_roi']['correlation'],
        result['cost_roi']['diminishing_returns'],
        len(inefficient),
        len(efficient)
    )
    
    return result


def _generate_tiktok_insights(
    pearson_r: float,
    cost_roi_r: float,
    diminishing: bool,
    inefficient: int,
    efficient: int
) -> List[Dict[str, str]]:
    """Generate TikTok-specific insights from correlation analysis."""
    insights = []
    
    if pearson_r > 0.7:
        insights.append({
            "type": "POSITIVE",
            "insight": "Strong correlation: Higher spending → higher revenue",
            "action": "Budget increases likely effective"
        })
    elif pearson_r < 0.3:
        insights.append({
            "type": "WARNING",
            "insight": "Weak correlation: Spending more ≠ more revenue",
            "action": "Optimize before increasing budget"
        })
    
    if diminishing:
        insights.append({
            "type": "WARNING",
            "insight": "Diminishing returns: Higher spend = lower ROI",
            "action": "Redistribute to smaller, high-ROI campaigns"
        })
    
    if inefficient > 5:
        insights.append({
            "type": "ALERT",
            "insight": f"{inefficient} products: high spend, low ROI",
            "action": "Review and optimize or reduce budget"
        })
    
    if efficient > 3:
        insights.append({
            "type": "OPPORTUNITY",
            "insight": f"{efficient} products: untapped potential",
            "action": "Prioritize scaling up"
        })
    
    return insights
