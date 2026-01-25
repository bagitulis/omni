#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Funnel Analyzer
===============
Analyzes marketing funnel health (CTR, CVR, CPM, CPA).
Following AGENTS.MD: Clean Code, DRY, SRP.
"""

from dataclasses import dataclass
from typing import Dict, Optional
from enum import Enum
import pandas as pd
import numpy as np


class FunnelHealth(Enum):
    """Funnel health status."""
    EXCELLENT = "EXCELLENT"
    GOOD = "GOOD"
    NEEDS_ATTENTION = "NEEDS_ATTENTION"
    CRITICAL = "CRITICAL"


@dataclass
class FunnelMetrics:
    """Funnel metrics container."""
    ctr: float           # Click-Through Rate (%)
    cvr: float           # Conversion Rate (%)
    cpm: float           # Cost Per Mille (Rp)
    cpa: float           # Cost Per Acquisition (Rp)
    health: FunnelHealth
    diagnosis: str


class FunnelAnalyzer:
    """
    Analyzes marketing funnel metrics.
    Provides diagnosis based on CTR and CVR combination.
    """
    
    # Thresholds based on TikTok Ads typical performance
    CTR_EXCELLENT = 3.0    # > 3% CTR is excellent
    CTR_GOOD = 1.5         # > 1.5% CTR is good
    CTR_LOW = 0.5          # < 0.5% CTR is low
    
    CVR_EXCELLENT = 5.0    # > 5% CVR is excellent
    CVR_GOOD = 2.0         # > 2% CVR is good
    CVR_LOW = 0.5          # < 0.5% CVR is low
    
    CPM_BENCHMARK = 30000  # Rp 30K CPM as benchmark
    CPA_BENCHMARK = 50000  # Rp 50K CPA as benchmark
    
    def analyze(self, impressions: float, clicks: float, 
                conversions: float, cost: float) -> FunnelMetrics:
        """
        Analyze funnel metrics from raw data.
        
        Args:
            impressions: Total ad impressions
            clicks: Total clicks
            conversions: Total orders/conversions
            cost: Total ad spend (Rp)
        
        Returns:
            FunnelMetrics with calculated values and diagnosis
        """
        # Calculate metrics with safe division
        ctr = self._safeDivide(clicks, impressions) * 100
        cvr = self._safeDivide(conversions, clicks) * 100
        cpm = self._safeDivide(cost, impressions) * 1000
        cpa = self._safeDivide(cost, conversions)
        
        # Determine health and diagnosis
        health, diagnosis = self._diagnose(ctr, cvr)
        
        return FunnelMetrics(
            ctr=round(ctr, 2),
            cvr=round(cvr, 2),
            cpm=round(cpm, 0),
            cpa=round(cpa, 0),
            health=health,
            diagnosis=diagnosis
        )
    
    def analyzeFromDf(self, df: pd.DataFrame) -> FunnelMetrics:
        """Analyze funnel from DataFrame with standard columns."""
        impressions = df['impressions'].sum() if 'impressions' in df else 0
        clicks = df['clicks'].sum() if 'clicks' in df else 0
        conversions = df['ordersSku'].sum() if 'ordersSku' in df else 0
        cost = df['cost'].sum() if 'cost' in df else 0
        
        return self.analyze(impressions, clicks, conversions, cost)
    
    def _safeDivide(self, numerator: float, denominator: float) -> float:
        """Safe division to avoid ZeroDivisionError."""
        if denominator == 0 or pd.isna(denominator):
            return 0.0
        return numerator / denominator
    
    def _diagnose(self, ctr: float, cvr: float) -> tuple:
        """
        Diagnose funnel health based on CTR and CVR.
        
        Returns:
            Tuple of (FunnelHealth, diagnosis_string)
        """
        ctrHigh = ctr >= self.CTR_GOOD
        ctrLow = ctr < self.CTR_LOW
        cvrHigh = cvr >= self.CVR_GOOD
        cvrLow = cvr < self.CVR_LOW
        
        # Decision matrix
        if ctrHigh and cvrHigh:
            return (FunnelHealth.EXCELLENT,
                    "✅ Funnel sehat. Kreatif menarik & produk menjual.")
        
        elif ctrHigh and cvrLow:
            return (FunnelHealth.NEEDS_ATTENTION,
                    "⚠️ CTR tinggi tapi CVR rendah. Kreatif bagus, "
                    "tapi harga/produk kurang menarik. Review landing page.")
        
        elif ctrLow and cvrHigh:
            return (FunnelHealth.NEEDS_ATTENTION,
                    "⚠️ CTR rendah tapi CVR tinggi. Produk bagus, "
                    "tapi kreatif kurang menarik. Ganti video/thumbnail.")
        
        elif ctrLow and cvrLow:
            return (FunnelHealth.CRITICAL,
                    "❌ CTR & CVR rendah. Review total: kreatif, "
                    "target audience, dan penawaran produk.")
        
        else:  # Middle ground
            return (FunnelHealth.GOOD,
                    "👍 Funnel cukup baik. Ada ruang optimasi.")
    
    def getHealthScore(self, metrics: FunnelMetrics) -> float:
        """
        Convert funnel health to numeric score (0-100).
        Used for composite scoring.
        """
        # CTR Score (0-40 points)
        if metrics.ctr >= self.CTR_EXCELLENT:
            ctrScore = 40
        elif metrics.ctr >= self.CTR_GOOD:
            ctrScore = 30
        elif metrics.ctr >= self.CTR_LOW:
            ctrScore = 15
        else:
            ctrScore = 5
        
        # CVR Score (0-60 points)
        if metrics.cvr >= self.CVR_EXCELLENT:
            cvrScore = 60
        elif metrics.cvr >= self.CVR_GOOD:
            cvrScore = 45
        elif metrics.cvr >= self.CVR_LOW:
            cvrScore = 20
        else:
            cvrScore = 5
        
        return min(100, ctrScore + cvrScore)
