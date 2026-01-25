#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
ROAS Classifier (Re-exported from Core + TikTok Extensions)
===========================================================
Re-exports core ROAS classifier and adds TikTok-specific formatting.
Base logic is in `notebooks/core/intelligence/roas.py`.
"""

import sys
from pathlib import Path

# Add parent directory to sys.path for core imports
_parent_dir = Path(__file__).resolve().parent.parent.parent
if str(_parent_dir) not in sys.path:
    sys.path.insert(0, str(_parent_dir))

# Re-export from core
from core.intelligence.roas import (
    RoasClassifier,
    RoasTier,
    RoasThresholds,
    RoasResult,
    RecommendedAction,
    classify_roas,
    get_roas_tier,
    get_roas_score,
)

# TikTok-specific config import
from .config_intelligence import IntelligenceConfig, RecommendationAction


def get_tiktok_roas_description(tier: RoasTier, roas: float) -> str:
    """Get TikTok-specific ROAS description in Indonesian."""
    descriptions = {
        RoasTier.EXCELLENT: f"🌟 ROAS {roas:.1f}x sangat bagus! Scale up agresif.",
        RoasTier.GOOD: f"✅ ROAS {roas:.1f}x bagus. Scale up moderat.",
        RoasTier.MODERATE: f"👍 ROAS {roas:.1f}x cukup. Maintain, optimasi kreatif.",
        RoasTier.MARGINAL: f"⚠️ ROAS {roas:.1f}x marginal. Kurangi budget.",
        RoasTier.LOSS: f"❌ ROAS {roas:.1f}x rugi. Stop atau pivot.",
    }
    return descriptions.get(tier, f"ROAS: {roas:.1f}x")


class TikTokRoasClassifier(RoasClassifier):
    """TikTok-specific ROAS classifier with Indonesian descriptions."""
    
    def __init__(self, config: IntelligenceConfig = None):
        self.config = config or IntelligenceConfig()
        # Use config thresholds
        thresholds = RoasThresholds(
            excellent=self.config.roas.excellent,
            good=self.config.roas.good,
            moderate=self.config.roas.moderate,
            marginal=self.config.roas.marginal,
        )
        super().__init__(thresholds)
    
    def _getDescription(self, tier: RoasTier, roas: float) -> str:
        """Override with Indonesian descriptions."""
        return get_tiktok_roas_description(tier, roas)


__all__ = [
    # Core exports
    "RoasClassifier",
    "RoasTier",
    "RoasThresholds",
    "RoasResult",
    "RecommendedAction",
    "classify_roas",
    "get_roas_tier",
    "get_roas_score",
    # TikTok-specific
    "TikTokRoasClassifier",
    "get_tiktok_roas_description",
]
