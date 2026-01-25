#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Churn Risk Analysis - Core Module
=================================
Platform-agnostic churn/decline risk assessment.
"""

from dataclasses import dataclass, field
from enum import Enum
from typing import List, Dict, Any, Optional, Callable


class RiskSeverity(Enum):
    """Risk factor severity levels."""
    LOW = "LOW"
    MEDIUM = "MEDIUM"
    HIGH = "HIGH"
    CRITICAL = "CRITICAL"


class ChurnRiskLevel(Enum):
    """Overall churn risk classification."""
    MINIMAL = "MINIMAL"
    LOW = "LOW"
    MEDIUM = "MEDIUM"
    HIGH = "HIGH"
    CRITICAL = "CRITICAL"


@dataclass
class RiskFactor:
    """A single risk factor contributing to churn score."""
    name: str
    value: str
    points: int
    severity: RiskSeverity
    max_points: int = 0
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "factor": self.name,
            "value": self.value,
            "points": self.points,
            "severity": self.severity.value
        }


@dataclass
class ChurnRiskResult:
    """Result of churn risk analysis."""
    entity_id: str
    entity_name: str
    risk_score: int
    max_score: int
    risk_level: ChurnRiskLevel
    prediction: str
    risk_factors: List[RiskFactor] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    @property
    def risk_percentage(self) -> float:
        return round(self.risk_score / self.max_score * 100, 1) if self.max_score > 0 else 0
    
    @property
    def n_factors(self) -> int:
        return len(self.risk_factors)
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "entity_id": self.entity_id,
            "entity_name": self.entity_name,
            "risk_score": self.risk_score,
            "max_score": self.max_score,
            "risk_percentage": self.risk_percentage,
            "risk_level": self.risk_level.value,
            "prediction": self.prediction,
            "risk_factors": [f.to_dict() for f in self.risk_factors],
            "n_factors": self.n_factors,
            **self.metadata
        }


@dataclass
class RiskThresholds:
    """Thresholds for risk level classification."""
    minimal_max: int = 15
    low_max: int = 35
    medium_max: int = 60
    high_max: int = 85
    # Above high_max = CRITICAL
    
    def get_level(self, score: int) -> ChurnRiskLevel:
        if score < self.minimal_max:
            return ChurnRiskLevel.MINIMAL
        elif score < self.low_max:
            return ChurnRiskLevel.LOW
        elif score < self.medium_max:
            return ChurnRiskLevel.MEDIUM
        elif score < self.high_max:
            return ChurnRiskLevel.HIGH
        else:
            return ChurnRiskLevel.CRITICAL


@dataclass
class RiskPredictions:
    """Prediction messages for each risk level."""
    minimal: str = "Very stable performance expected"
    low: str = "Stable performance expected"
    medium: str = "Moderate risk - monitor closely"
    high: str = "High probability of decline"
    critical: str = "Critical - immediate action required"
    
    def get_prediction(self, level: ChurnRiskLevel) -> str:
        return {
            ChurnRiskLevel.MINIMAL: self.minimal,
            ChurnRiskLevel.LOW: self.low,
            ChurnRiskLevel.MEDIUM: self.medium,
            ChurnRiskLevel.HIGH: self.high,
            ChurnRiskLevel.CRITICAL: self.critical
        }.get(level, self.medium)


class ChurnRiskCalculator:
    """
    Generic churn risk calculator.
    
    Subclass and override `_evaluate_factors()` for platform-specific logic.
    
    Example:
        ```python
        class TikTokChurnRisk(ChurnRiskCalculator):
            def _evaluate_factors(self, entity: Dict) -> List[RiskFactor]:
                factors = []
                momentum = entity.get('momentum_pct', 0)
                if momentum < -30:
                    factors.append(RiskFactor(
                        name="Severe Momentum Decline",
                        value=f"{momentum:.1f}%",
                        points=35,
                        severity=RiskSeverity.HIGH
                    ))
                return factors
        ```
    """
    
    def __init__(
        self,
        max_score: int = 100,
        thresholds: Optional[RiskThresholds] = None,
        predictions: Optional[RiskPredictions] = None
    ):
        self.max_score = max_score
        self.thresholds = thresholds or RiskThresholds()
        self.predictions = predictions or RiskPredictions()
    
    def calculate(
        self,
        entity: Dict[str, Any],
        id_field: str = "id",
        name_field: str = "name"
    ) -> ChurnRiskResult:
        """
        Calculate churn risk for a single entity.
        
        Args:
            entity: Dict with entity data
            id_field: Key for entity ID
            name_field: Key for entity name
            
        Returns:
            ChurnRiskResult
        """
        factors = self._evaluate_factors(entity)
        total_score = min(sum(f.points for f in factors), self.max_score)
        
        risk_level = self.thresholds.get_level(total_score)
        prediction = self.predictions.get_prediction(risk_level)
        
        return ChurnRiskResult(
            entity_id=str(entity.get(id_field, "")),
            entity_name=str(entity.get(name_field, "")),
            risk_score=total_score,
            max_score=self.max_score,
            risk_level=risk_level,
            prediction=prediction,
            risk_factors=factors,
            metadata=self._get_metadata(entity)
        )
    
    def _evaluate_factors(self, entity: Dict[str, Any]) -> List[RiskFactor]:
        """
        Override this method to implement platform-specific risk evaluation.
        
        Args:
            entity: Dict with entity data
            
        Returns:
            List of RiskFactor
        """
        raise NotImplementedError("Subclass must implement _evaluate_factors()")
    
    def _get_metadata(self, entity: Dict[str, Any]) -> Dict[str, Any]:
        """
        Override to add extra metadata to result.
        
        Args:
            entity: Dict with entity data
            
        Returns:
            Dict with metadata
        """
        return {}
    
    def analyze_portfolio(
        self,
        entities: List[Dict[str, Any]],
        weight_field: Optional[str] = None,
        id_field: str = "id",
        name_field: str = "name"
    ) -> Dict[str, Any]:
        """
        Analyze churn risk across portfolio.
        
        Args:
            entities: List of entity dicts
            weight_field: Field to use for weighted average (e.g., 'total_cost')
            id_field: Key for entity ID
            name_field: Key for entity name
            
        Returns:
            Portfolio analysis result
        """
        if not entities:
            return {"is_valid": False, "reason": "No entities provided"}
        
        assessments = [
            self.calculate(e, id_field, name_field)
            for e in entities
        ]
        
        # Distribution
        critical = [a for a in assessments if a.risk_level == ChurnRiskLevel.CRITICAL]
        high = [a for a in assessments if a.risk_level == ChurnRiskLevel.HIGH]
        medium = [a for a in assessments if a.risk_level == ChurnRiskLevel.MEDIUM]
        low = [a for a in assessments if a.risk_level in (ChurnRiskLevel.LOW, ChurnRiskLevel.MINIMAL)]
        
        # Weighted portfolio score
        if weight_field:
            total_weight = sum(e.get(weight_field, 0) for e in entities)
            if total_weight > 0:
                portfolio_score = sum(
                    assessments[i].risk_score * (e.get(weight_field, 0) / total_weight)
                    for i, e in enumerate(entities)
                )
            else:
                portfolio_score = sum(a.risk_score for a in assessments) / len(assessments)
        else:
            portfolio_score = sum(a.risk_score for a in assessments) / len(assessments)
        
        return {
            "is_valid": True,
            "portfolio_risk_score": round(portfolio_score, 1),
            "portfolio_risk_level": self.thresholds.get_level(int(portfolio_score)).value,
            "distribution": {
                "critical": len(critical),
                "high_risk": len(high),
                "medium_risk": len(medium),
                "low_risk": len(low)
            },
            "high_risk_pct": round((len(critical) + len(high)) / len(entities) * 100, 1),
            "early_warnings": [
                a.to_dict() for a in sorted(
                    critical + high,
                    key=lambda x: x.risk_score,
                    reverse=True
                )[:10]
            ],
            "total_entities": len(entities),
            "all_assessments": [a.to_dict() for a in assessments]
        }


def calculate_cost_revenue_correlation_analysis(
    entities: List[Dict[str, Any]],
    cost_field: str = "total_cost",
    revenue_field: str = "total_revenue",
    roi_field: str = "roi",
    min_entities: int = 5
) -> Dict[str, Any]:
    """
    Generic cost-revenue correlation analysis.
    
    Args:
        entities: List of entity dicts with cost/revenue/roi
        cost_field: Field name for cost
        revenue_field: Field name for revenue
        roi_field: Field name for ROI
        min_entities: Minimum entities required
        
    Returns:
        Correlation analysis result
    """
    import numpy as np
    from scipy import stats
    
    if len(entities) < min_entities:
        return {"is_valid": False, "note": f"Need at least {min_entities} entities"}
    
    costs = np.array([e.get(cost_field, 0) for e in entities])
    revenues = np.array([e.get(revenue_field, 0) for e in entities])
    rois = np.array([e.get(roi_field, 0) for e in entities])
    
    # Correlations
    pearson_r, pearson_p = stats.pearsonr(costs, revenues)
    spearman_r, spearman_p = stats.spearmanr(costs, revenues)
    cost_roi_r, cost_roi_p = stats.pearsonr(costs, rois)
    
    def interpret_r(r: float) -> str:
        if abs(r) >= 0.7:
            return "Strong"
        if abs(r) >= 0.4:
            return "Moderate"
        if abs(r) >= 0.2:
            return "Weak"
        return "Very Weak"
    
    has_diminishing_returns = cost_roi_r < -0.3 and cost_roi_p < 0.1
    avg_roi = float(np.mean(rois))
    median_cost = float(np.median(costs))
    
    return {
        "is_valid": True,
        "cost_revenue": {
            "pearson_r": round(pearson_r, 3),
            "pearson_p": round(pearson_p, 4),
            "spearman_r": round(spearman_r, 3),
            "spearman_p": round(spearman_p, 4),
            "interpretation": interpret_r(pearson_r),
            "significant": pearson_p < 0.05
        },
        "cost_roi": {
            "correlation": round(cost_roi_r, 3),
            "p_value": round(cost_roi_p, 4),
            "interpretation": interpret_r(cost_roi_r),
            "diminishing_returns": has_diminishing_returns
        },
        "summary_stats": {
            "avg_roi": round(avg_roi, 2),
            "median_cost": round(median_cost, 2),
            "n_entities": len(entities)
        }
    }
