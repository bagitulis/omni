#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Probability with Confidence Interval (Generic Core Module)
==========================================================
Enhanced probability estimation with statistical rigor.
Platform-agnostic: Works for TikTok, Shopee, Lazada, etc.
Following AGENTS.MD: Clean Code, DRY, SRP, max 300 lines.
"""

from dataclasses import dataclass
from typing import List, Tuple, Dict
import numpy as np
from scipy import stats


@dataclass
class ProbabilityEstimate:
    """Complete probability estimate with CI."""
    successProbability: float
    profitProbability: float
    roasTargetProbability: float
    
    successCI: Tuple[float, float]
    profitCI: Tuple[float, float]
    roasCI: Tuple[float, float]
    
    confidenceLevel: str
    confidencePct: float
    
    positiveFactors: List[str]
    negativeFactors: List[str]
    
    uncertaintyLevel: str
    uncertaintyPct: float


class ProbabilityWithCI:
    """
    Enhanced probability engine with confidence intervals.
    
    Methods:
    1. Bayesian-style weighted probability
    2. Bootstrap CI for uncertainty quantification
    3. Monte Carlo for scenario simulation
    """
    
    def __init__(self, targetRoas: float = 2.0, nBootstrap: int = 1000):
        self.targetRoas = targetRoas
        self.nBootstrap = nBootstrap
    
    def estimate(self,
                 compositeScore: float,
                 roasScore: float,
                 trendScore: float,
                 elasticityScore: float,
                 historicalRoas: List[float],
                 historicalProfits: List[float],
                 eventMultiplier: float,
                 churnRiskScore: float,
                 fatigueStatus: str,
                 nPeriods: int,
                 dataQuality: str) -> ProbabilityEstimate:
        """Generate probability estimate with confidence intervals."""
        
        positiveFactors = []
        negativeFactors = []
        
        baseProb = self._calculateBaseSuccess(
            compositeScore, roasScore, trendScore, elasticityScore
        )
        
        profitProb, profitCI = self._calculateProfitProbability(historicalProfits)
        roasProb, roasCI = self._calculateRoasProbability(historicalRoas)
        
        adjustedProb, factors = self._adjustForContext(
            baseProb, eventMultiplier, churnRiskScore, fatigueStatus
        )
        positiveFactors.extend(factors['positive'])
        negativeFactors.extend(factors['negative'])
        
        successCI = self._bootstrapCI(
            adjustedProb, historicalRoas, historicalProfits, nPeriods
        )
        
        confidencePct = self._calculateConfidence(nPeriods, dataQuality, historicalRoas)
        confidenceLevel = self._getConfidenceLevel(confidencePct)
        
        uncertaintyPct, uncertaintyLevel = self._calculateUncertainty(
            successCI, nPeriods, dataQuality
        )
        
        if compositeScore >= 70:
            positiveFactors.append(f"High composite score: {compositeScore:.0f}")
        elif compositeScore < 40:
            negativeFactors.append(f"Low composite score: {compositeScore:.0f}")
        
        if roasScore >= 70:
            positiveFactors.append("Strong ROAS performance")
        
        if trendScore >= 70:
            positiveFactors.append("Positive trend direction")
        elif trendScore < 30:
            negativeFactors.append("Negative trend direction")
        
        return ProbabilityEstimate(
            successProbability=round(adjustedProb, 1),
            profitProbability=round(profitProb, 1),
            roasTargetProbability=round(roasProb, 1),
            successCI=(round(successCI[0], 1), round(successCI[1], 1)),
            profitCI=(round(profitCI[0], 1), round(profitCI[1], 1)),
            roasCI=(round(roasCI[0], 1), round(roasCI[1], 1)),
            confidenceLevel=confidenceLevel,
            confidencePct=round(confidencePct, 1),
            positiveFactors=positiveFactors[:5],
            negativeFactors=negativeFactors[:5],
            uncertaintyLevel=uncertaintyLevel,
            uncertaintyPct=round(uncertaintyPct, 1)
        )
    
    def _calculateBaseSuccess(self, composite: float, roas: float,
                               trend: float, elasticity: float) -> float:
        """Calculate base success probability from scores."""
        weights = {'composite': 0.40, 'roas': 0.25, 'trend': 0.20, 'elasticity': 0.15}
        
        weightedScore = (
            composite * weights['composite'] +
            roas * weights['roas'] +
            trend * weights['trend'] +
            elasticity * weights['elasticity']
        )
        
        if weightedScore >= 50:
            prob = 50 + (weightedScore - 50) * 0.9
        else:
            prob = weightedScore
        
        return min(95, max(5, prob))
    
    def _calculateProfitProbability(self, profits: List[float]
                                     ) -> Tuple[float, Tuple[float, float]]:
        """Calculate probability of profit > 0 with CI."""
        if len(profits) < 2:
            return 50.0, (30.0, 70.0)
        
        profits_arr = np.array(profits)
        n_profitable = np.sum(profits_arr > 0)
        prob = (n_profitable / len(profits_arr)) * 100
        
        ci = self._wilsonScoreInterval(n_profitable, len(profits_arr))
        return prob, ci
    
    def _calculateRoasProbability(self, roas_list: List[float]
                                   ) -> Tuple[float, Tuple[float, float]]:
        """Calculate probability of ROAS > target with CI."""
        if len(roas_list) < 2:
            return 50.0, (30.0, 70.0)
        
        roas_arr = np.array(roas_list)
        n_above_target = np.sum(roas_arr > self.targetRoas)
        prob = (n_above_target / len(roas_arr)) * 100
        
        ci = self._wilsonScoreInterval(n_above_target, len(roas_arr))
        return prob, ci
    
    def _wilsonScoreInterval(self, successes: int, total: int,
                              confidence: float = 0.95) -> Tuple[float, float]:
        """Calculate Wilson score confidence interval."""
        if total == 0:
            return (0.0, 100.0)
        
        p = successes / total
        z = stats.norm.ppf(1 - (1 - confidence) / 2)
        
        denominator = 1 + z**2 / total
        center = (p + z**2 / (2 * total)) / denominator
        margin = z * np.sqrt((p * (1 - p) + z**2 / (4 * total)) / total) / denominator
        
        lower = max(0, center - margin) * 100
        upper = min(1, center + margin) * 100
        
        return (lower, upper)
    
    def _adjustForContext(self, baseProb: float, eventMult: float,
                           churnRisk: float, fatigue: str
                           ) -> Tuple[float, Dict]:
        """Adjust probability for contextual factors."""
        prob = baseProb
        factors = {'positive': [], 'negative': []}
        
        if eventMult > 1.3:
            prob = min(95, prob + 5)
            factors['positive'].append(f"Event boost: x{eventMult:.2f}")
        elif eventMult < 0.9:
            prob = max(5, prob - 5)
            factors['negative'].append("Dry season timing")
        
        if churnRisk > 60:
            prob = max(5, prob - 10)
            factors['negative'].append(f"High churn risk: {churnRisk:.0f}%")
        elif churnRisk < 20:
            factors['positive'].append("Low churn risk")
        
        fatigueImpact = {
            'FRESH': (5, 'positive', "Fresh creative"),
            'AGING': (0, None, None),
            'FATIGUED': (-10, 'negative', "Creative fatigue"),
            'DEAD': (-15, 'negative', "Creative exhausted"),
            'INSUFFICIENT_DATA': (0, None, None)
        }
        
        impact = fatigueImpact.get(fatigue, (0, None, None))
        prob = max(5, min(95, prob + impact[0]))
        if impact[1]:
            factors[impact[1]].append(impact[2])
        
        return prob, factors
    
    def _bootstrapCI(self, baseProb: float, historicalRoas: List[float],
                      historicalProfits: List[float], nPeriods: int
                      ) -> Tuple[float, float]:
        """Calculate CI using bootstrap method."""
        if len(historicalRoas) < 3:
            margin = 20 - min(15, nPeriods * 2)
            return (max(5, baseProb - margin), min(95, baseProb + margin))
        
        roas_arr = np.array(historicalRoas)
        bootstrap_probs = []
        
        for _ in range(self.nBootstrap):
            sample = np.random.choice(roas_arr, size=len(roas_arr), replace=True)
            sample_above = np.sum(sample > self.targetRoas) / len(sample)
            sample_prob = (baseProb * 0.6 + sample_above * 100 * 0.4)
            bootstrap_probs.append(sample_prob)
        
        lower = np.percentile(bootstrap_probs, 2.5)
        upper = np.percentile(bootstrap_probs, 97.5)
        
        return (max(5, lower), min(95, upper))
    
    def _calculateConfidence(self, nPeriods: int, dataQuality: str,
                              historicalRoas: List[float]) -> float:
        """Calculate overall confidence."""
        dataConf = min(40, nPeriods * 4)
        
        if "HIGH" in dataQuality:
            qualConf = 30
        elif "MEDIUM" in dataQuality:
            qualConf = 20
        else:
            qualConf = 10
        
        if len(historicalRoas) >= 3:
            mean_roas = np.mean(historicalRoas)
            cv = np.std(historicalRoas) / mean_roas if mean_roas > 0 else 1
            varConf = max(0, 30 - cv * 30)
        else:
            varConf = 10
        
        return dataConf + qualConf + varConf
    
    def _getConfidenceLevel(self, confidence: float) -> str:
        """Get confidence level label."""
        if confidence >= 80:
            return "HIGH"
        elif confidence >= 60:
            return "MEDIUM"
        elif confidence >= 40:
            return "LOW"
        else:
            return "VERY_LOW"
    
    def _calculateUncertainty(self, successCI: Tuple[float, float],
                               nPeriods: int, dataQuality: str
                               ) -> Tuple[float, str]:
        """Calculate uncertainty level."""
        ciWidth = successCI[1] - successCI[0]
        uncertainty = ciWidth / 2
        
        if "LOW" in dataQuality:
            uncertainty *= 1.3
        
        if nPeriods < 4:
            uncertainty *= 1.2
        
        if uncertainty < 10:
            level = "LOW"
        elif uncertainty < 20:
            level = "MEDIUM"
        else:
            level = "HIGH"
        
        return uncertainty, level


# Convenience function
def estimate_probability(compositeScore: float,
                         historicalRoas: List[float],
                         historicalProfits: List[float],
                         targetRoas: float = 2.0) -> ProbabilityEstimate:
    """Quick probability estimation."""
    engine = ProbabilityWithCI(targetRoas=targetRoas)
    return engine.estimate(
        compositeScore=compositeScore,
        roasScore=compositeScore,
        trendScore=50,
        elasticityScore=50,
        historicalRoas=historicalRoas,
        historicalProfits=historicalProfits,
        eventMultiplier=1.0,
        churnRiskScore=30,
        fatigueStatus="AGING",
        nPeriods=len(historicalRoas),
        dataQuality="MEDIUM"
    )
