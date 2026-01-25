#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Enhanced Budget Recommender
===========================
Combines all budget optimization logic with precise nominal calculations.
Following AGENTS.MD: Clean Code, DRY, SRP, max 300 lines.

Outputs:
- Action: SCALE_UP_AGGRESSIVE (+30%), SCALE_UP (+10%), MAINTAIN (0%), REDUCE (-20%), STOP (-100%)
- Budget Change: Percentage and nominal amount
- Expected Revenue Impact with Confidence Interval
"""

from dataclasses import dataclass
from typing import Dict, List, Tuple, Optional
from enum import Enum
import numpy as np

from .unified_scorer import ActionRecommendation


@dataclass
class BudgetRecommendation:
    """Complete budget recommendation with all details."""
    productId: str
    productName: str
    
    # Current state
    currentBudget: float
    currentRevenue: float
    currentProfit: float
    currentRoas: float
    
    # Recommendation
    action: ActionRecommendation
    recommendedBudget: float
    budgetChange: float
    budgetChangePct: float
    
    # Expected impact
    expectedRevenue: float
    expectedProfit: float
    revenueChange: float
    profitChange: float
    
    # Confidence Interval
    revenueCI: Tuple[float, float]  # (lower, upper)
    profitCI: Tuple[float, float]
    confidencePct: float
    
    # Rationale
    reasons: List[str]
    warnings: List[str]
    
    # Methodology
    methodology: str


class EnhancedBudgetRecommender:
    """
    Enhanced budget optimizer that combines:
    - Legacy optimization.py logic
    - Intelligence budget_optimizer.py
    - Elasticity analysis
    - Saturation model
    - Event impact
    
    Constraints:
    - Max budget per product: Rp 2,000,000/day (configurable)
    - Min meaningful change: Rp 50,000 (below = maintain)
    - Changes are percentage-based for methodology, capped by constraints
    """
    
    MAX_BUDGET = 2_000_000  # Rp 2 juta per produk
    MIN_CHANGE = 50_000    # Minimum perubahan yang meaningful
    
    # Action to budget change mapping (base percentages)
    # These will be adjusted by 7 factors: elasticity, saturation, marginal ROAS,
    # events, trend, fatigue, churn risk
    ACTION_CHANGES = {
        ActionRecommendation.SCALE_UP_AGGRESSIVE: 0.50,  # +50% (was 30%)
        ActionRecommendation.SCALE_UP: 0.25,              # +25% (was 10%)
        ActionRecommendation.MAINTAIN: 0.00,              # 0%
        ActionRecommendation.REDUCE: -0.30,               # -30% (was 20%)
        ActionRecommendation.STOP: -1.00,                 # -100%
    }
    
    def __init__(self, maxBudget: float = None):
        self.maxBudget = maxBudget or self.MAX_BUDGET
    
    def recommend(self,
                  productId: str,
                  productName: str,
                  # Current metrics
                  currentBudget: float,
                  currentRevenue: float,
                  roas: float,
                  # Score-based action
                  action: ActionRecommendation,
                  compositeScore: float,
                  # Elasticity data
                  elasticity: float,
                  saturationStatus: str,
                  marginalRoas: float,
                  # Event impact
                  eventMultiplier: float,
                  eventNames: List[str],
                  # Confidence
                  confidence: float,
                  # Additional factors
                  trendDirection: str,
                  fatigueStatus: str,
                  churnRiskScore: float
                  ) -> BudgetRecommendation:
        """Generate comprehensive budget recommendation."""
        
        reasons = []
        warnings = []
        
        # Get base change from action
        baseChange = self.ACTION_CHANGES[action]
        
        # Adjust based on factors
        adjustedChange, adjReasons, adjWarnings = self._adjustBudgetChange(
            baseChange=baseChange,
            action=action,
            elasticity=elasticity,
            saturationStatus=saturationStatus,
            marginalRoas=marginalRoas,
            eventMultiplier=eventMultiplier,
            trendDirection=trendDirection,
            fatigueStatus=fatigueStatus,
            churnRiskScore=churnRiskScore,
            compositeScore=compositeScore
        )
        
        reasons.extend(adjReasons)
        warnings.extend(adjWarnings)
        
        # Calculate recommended budget
        proposedBudget = currentBudget * (1 + adjustedChange)
        
        # Apply constraints
        constraintApplied = False
        
        # Max budget constraint
        if proposedBudget > self.maxBudget:
            proposedBudget = self.maxBudget
            if adjustedChange > 0:
                warnings.append(f"Budget capped at Rp {self.maxBudget:,.0f}")
                constraintApplied = True
        
        recommendedBudget = max(0, proposedBudget)
        budgetChange = recommendedBudget - currentBudget
        
        # Minimum meaningful change constraint
        # Logic: Jika perubahan terlalu kecil (< Rp 50k), evaluasi apakah layak
        # - Jika score tinggi (>=60) dan perubahan positif kecil → tetap terapkan (tidak memaksa ke 50k)
        # - Jika perubahan sangat kecil dan tidak signifikan → maintain
        if 0 < abs(budgetChange) < self.MIN_CHANGE and action != ActionRecommendation.STOP:
            if adjustedChange > 0:
                # Perubahan positif kecil - evaluasi berdasarkan estimasi
                # Hitung apakah perubahan kecil ini masih menguntungkan
                estimatedAdditionalProfit = budgetChange * max(0, roas - 1)  # Profit = Revenue - Cost
                
                if estimatedAdditionalProfit > budgetChange * 0.3:  # Minimal 30% profit margin
                    # Perubahan kecil tapi masih menguntungkan - terapkan apa adanya
                    reasons.append(f"Perubahan kecil (+Rp {budgetChange:,.0f}) tapi tetap menguntungkan")
                elif compositeScore >= 70:
                    # Score sangat tinggi - beri minimum increase
                    recommendedBudget = currentBudget + self.MIN_CHANGE
                    budgetChange = self.MIN_CHANGE
                    reasons.append(f"Score tinggi ({compositeScore:.0f}) - minimum +Rp {self.MIN_CHANGE:,.0f}")
                else:
                    # Tidak cukup meyakinkan - maintain saja
                    recommendedBudget = currentBudget
                    budgetChange = 0
                    adjustedChange = 0
                    warnings.append(f"Perubahan < Rp {self.MIN_CHANGE:,.0f} & profit margin rendah → pertahankan")
            else:
                # Perubahan negatif kecil
                if compositeScore < 40:
                    # Score rendah - terapkan minimum reduction
                    recommendedBudget = max(0, currentBudget - self.MIN_CHANGE)
                    budgetChange = recommendedBudget - currentBudget
                    reasons.append(f"Score rendah ({compositeScore:.0f}) - kurangi Rp {self.MIN_CHANGE:,.0f}")
                else:
                    # Score cukup - maintain
                    recommendedBudget = currentBudget
                    budgetChange = 0
                    adjustedChange = 0
        
        # Calculate expected impact
        expectedRevenue, expectedProfit = self._calculateExpectedImpact(
            currentBudget=currentBudget,
            currentRevenue=currentRevenue,
            recommendedBudget=recommendedBudget,
            roas=roas,
            elasticity=elasticity,
            eventMultiplier=eventMultiplier
        )
        
        # Calculate confidence intervals
        revenueCI, profitCI = self._calculateConfidenceIntervals(
            expectedRevenue=expectedRevenue,
            expectedProfit=expectedProfit,
            confidence=confidence,
            elasticity=elasticity
        )
        
        # Generate methodology string
        methodology = self._generateMethodology(action, adjustedChange, reasons)
        
        return BudgetRecommendation(
            productId=productId,
            productName=productName,
            currentBudget=currentBudget,
            currentRevenue=currentRevenue,
            currentProfit=currentRevenue - currentBudget,
            currentRoas=roas,
            action=action,
            recommendedBudget=recommendedBudget,
            budgetChange=budgetChange,
            budgetChangePct=adjustedChange * 100,
            expectedRevenue=expectedRevenue,
            expectedProfit=expectedProfit,
            revenueChange=expectedRevenue - currentRevenue,
            profitChange=expectedProfit - (currentRevenue - currentBudget),
            revenueCI=revenueCI,
            profitCI=profitCI,
            confidencePct=confidence,
            reasons=reasons,
            warnings=warnings,
            methodology=methodology
        )
    
    def _adjustBudgetChange(self, baseChange: float, action: ActionRecommendation,
                            elasticity: float, saturationStatus: str,
                            marginalRoas: float, eventMultiplier: float,
                            trendDirection: str, fatigueStatus: str,
                            churnRiskScore: float, compositeScore: float
                            ) -> Tuple[float, List[str], List[str]]:
        """
        Adjust budget change based on 7 methodological factors.
        
        Formula: Final Change = Base Change + Σ(Factor Adjustments)
        
        Factor Adjustments:
        1. Elasticity: +15% if E>1.2, -15% if E<0.5
        2. Saturation: -20% if saturated, +10% if growth phase
        3. Marginal ROAS: +10% if mROAS>2, -15% if mROAS<0.8
        4. Event: +15% if multiplier>1.3, -10% if <0.9
        5. Trend: +10% if UP, -15% if DOWN
        6. Fatigue: force 0% if FATIGUED/DEAD
        7. Churn Risk: -15% if score>60
        
        Max positive change: 100% (double budget)
        Max negative change: -100% (stop)
        """
        
        change = baseChange
        reasons = []
        warnings = []
        adjustments = []  # Track all adjustments
        
        # 1. Elasticity adjustment (±15%)
        if elasticity > 1.2 and baseChange > 0:
            adj = 0.15
            change = min(1.0, change + adj)
            reasons.append(f"Elastic ({elasticity:.2f}) +{adj*100:.0f}%")
            adjustments.append(("Elasticity", adj))
        elif elasticity < 0.5 and baseChange > 0:
            adj = -0.15
            change = max(0, change + adj)
            warnings.append(f"Inelastic ({elasticity:.2f}) {adj*100:.0f}%")
            adjustments.append(("Elasticity", adj))
        
        # 2. Saturation adjustment (±20%/-10%)
        if saturationStatus == "SATURATED" and baseChange > 0:
            adj = -0.20
            change = max(0, change + adj)
            warnings.append(f"Saturated {adj*100:.0f}%")
            adjustments.append(("Saturation", adj))
        elif saturationStatus in ["GROWTH_PHASE", "HIGH_ELASTICITY"] and baseChange > 0:
            adj = 0.10
            change = min(1.0, change + adj)
            reasons.append(f"Growth phase +{adj*100:.0f}%")
            adjustments.append(("Saturation", adj))
        
        # 3. Marginal ROAS adjustment (±15%/+10%)
        if marginalRoas < 0.8 and baseChange > 0:
            adj = -0.15
            change = max(0, change + adj)
            warnings.append(f"Low mROAS ({marginalRoas:.2f}) {adj*100:.0f}%")
            adjustments.append(("Marginal ROAS", adj))
        elif marginalRoas > 2.0 and baseChange > 0:
            adj = 0.10
            change = min(1.0, change + adj)
            reasons.append(f"High mROAS ({marginalRoas:.2f}) +{adj*100:.0f}%")
            adjustments.append(("Marginal ROAS", adj))
        
        # 4. Event timing adjustment (±15%/-10%)
        if eventMultiplier > 1.3 and baseChange >= 0:
            adj = 0.15
            change = min(1.0, change + adj)
            reasons.append(f"Event x{eventMultiplier:.2f} +{adj*100:.0f}%")
            adjustments.append(("Event", adj))
        elif eventMultiplier < 0.9 and baseChange > 0:
            adj = -0.10
            change = max(-1, change + adj)
            warnings.append(f"Dry season {adj*100:.0f}%")
            adjustments.append(("Event", adj))
        
        # 5. Trend adjustment (±15%/+10%)
        if trendDirection == "DOWN" and baseChange > 0:
            adj = -0.15
            change = max(0, change + adj)
            warnings.append(f"Downtrend {adj*100:.0f}%")
            adjustments.append(("Trend", adj))
        elif trendDirection == "UP" and baseChange > 0:
            adj = 0.10
            change = min(1.0, change + adj)
            reasons.append(f"Uptrend +{adj*100:.0f}%")
            adjustments.append(("Trend", adj))
        
        # 6. Fatigue adjustment (force 0 or reduce more)
        if fatigueStatus in ["FATIGUED", "DEAD"] and baseChange > 0:
            oldChange = change
            change = 0
            warnings.append(f"Fatigue ({fatigueStatus}) → Maintain")
            adjustments.append(("Fatigue", -oldChange))
        
        # 7. Churn risk adjustment (-15%)
        if churnRiskScore > 60 and baseChange > 0:
            adj = -0.15
            change = max(0, change + adj)
            warnings.append(f"Churn risk ({churnRiskScore:.0f}%) {adj*100:.0f}%")
            adjustments.append(("Churn", adj))
        
        # Add base reason with formula explanation
        totalAdj = sum(a[1] for a in adjustments)
        if baseChange > 0:
            reasons.insert(0, f"Base {baseChange*100:+.0f}% + Adj {totalAdj*100:+.0f}% = {change*100:+.0f}%")
        elif baseChange < 0:
            reasons.insert(0, f"Score {compositeScore:.0f} → Reduce/Stop")
        else:
            reasons.insert(0, f"Score {compositeScore:.0f} → Maintain")
        
        return change, reasons, warnings
    
    def _calculateExpectedImpact(self, currentBudget: float, currentRevenue: float,
                                  recommendedBudget: float, roas: float,
                                  elasticity: float, eventMultiplier: float
                                  ) -> Tuple[float, float]:
        """Calculate expected revenue and profit."""
        
        budgetChange = recommendedBudget - currentBudget
        
        if budgetChange == 0:
            return currentRevenue, currentRevenue - currentBudget
        
        if recommendedBudget == 0:  # STOP
            return 0, 0
        
        # Calculate expected revenue considering elasticity
        if budgetChange > 0:
            # Scale up: Apply elasticity and conservation factor
            conservationFactor = 0.85  # Conservative estimate
            effectiveRoas = roas * conservationFactor * min(elasticity, 1.2)
            additionalRevenue = budgetChange * effectiveRoas * eventMultiplier
            expectedRevenue = currentRevenue + additionalRevenue
        else:
            # Scale down: Proportional reduction
            reductionRatio = recommendedBudget / currentBudget
            expectedRevenue = currentRevenue * reductionRatio
        
        expectedProfit = expectedRevenue - recommendedBudget
        
        return expectedRevenue, expectedProfit
    
    def _calculateConfidenceIntervals(self, expectedRevenue: float,
                                       expectedProfit: float, confidence: float,
                                       elasticity: float
                                       ) -> Tuple[Tuple[float, float], Tuple[float, float]]:
        """Calculate 95% confidence intervals."""
        
        # Uncertainty factor based on confidence and elasticity variability
        baseUncertainty = 0.25 - (confidence / 100 * 0.15)  # 10-25%
        elasticityUncertainty = 0.1 if elasticity > 0.5 else 0.2
        totalUncertainty = baseUncertainty + elasticityUncertainty
        
        # Revenue CI
        revLower = expectedRevenue * (1 - totalUncertainty)
        revUpper = expectedRevenue * (1 + totalUncertainty)
        
        # Profit CI (wider due to cost uncertainty)
        profitUncertainty = totalUncertainty * 1.2
        profitLower = expectedProfit * (1 - profitUncertainty) if expectedProfit > 0 else expectedProfit * (1 + profitUncertainty)
        profitUpper = expectedProfit * (1 + profitUncertainty) if expectedProfit > 0 else expectedProfit * (1 - profitUncertainty)
        
        return (revLower, revUpper), (min(profitLower, profitUpper), max(profitLower, profitUpper))
    
    def _generateMethodology(self, action: ActionRecommendation,
                              change: float, reasons: List[str]) -> str:
        """
        Generate methodology explanation for transparency.
        
        Shows the complete formula used:
        Final Change = Base(Action) + Elasticity + Saturation + mROAS + Event + Trend - Fatigue - Churn
        """
        
        # Extract base from reasons (first item usually contains formula)
        formula = reasons[0] if reasons else f"{action.value}"
        
        # Build methodology string
        parts = [
            f"Action: {action.value}",
            f"Formula: {formula}",
            f"Final: {change*100:+.0f}%",
            f"Constraints: Max Rp 2M, Min change Rp 50K"
        ]
        
        return " | ".join(parts)
