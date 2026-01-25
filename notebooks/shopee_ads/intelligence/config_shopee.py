#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads Intelligence - Configuration
=======================================
Shopee-specific thresholds and configurations.
"""

from dataclasses import dataclass
from typing import Dict


@dataclass
class ShopeeIntelligenceConfig:
    """Shopee-specific intelligence configuration."""
    
    # Shopee ROAS thresholds (platform-specific)
    roas_excellent: float = 8.0  # Shopee has higher ROAS typically
    roas_good: float = 5.0
    roas_moderate: float = 2.5
    roas_poor: float = 1.0
    
    # Saturation thresholds
    saturation_warning: float = 0.7
    saturation_critical: float = 0.9
    
    # Fatigue thresholds (days)
    fatigue_warning_days: int = 21
    fatigue_critical_days: int = 45
    
    # Lifecycle parameters
    lifecycle_growth_threshold: float = 0.15
    lifecycle_decline_threshold: float = -0.10
    
    # Score weights for Shopee (same as TikTok for consistency)
    weight_roi: float = 0.30
    weight_profit: float = 0.25
    weight_momentum: float = 0.20
    weight_consistency: float = 0.15
    weight_trend: float = 0.10
    
    # Category thresholds
    score_excellent: float = 80
    score_good: float = 60
    score_moderate: float = 40
    score_poor: float = 20
    
    def get_weights(self) -> Dict[str, float]:
        """Get scoring weights as dictionary."""
        return {
            "roi": self.weight_roi,
            "profit": self.weight_profit,
            "momentum": self.weight_momentum,
            "consistency": self.weight_consistency,
            "trend": self.weight_trend,
        }
    
    def get_roas_thresholds(self) -> Dict[str, float]:
        """Get ROAS tier thresholds."""
        return {
            "excellent": self.roas_excellent,
            "good": self.roas_good,
            "moderate": self.roas_moderate,
            "poor": self.roas_poor,
        }


# Default config instance
DEFAULT_CONFIG = ShopeeIntelligenceConfig()
