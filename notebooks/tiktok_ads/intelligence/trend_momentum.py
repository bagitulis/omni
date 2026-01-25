#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Trend Momentum (Re-exported from Core + TikTok Extensions)
==========================================================
Re-exports core trend momentum and adds TikTok-specific formatting.
Base logic is in `notebooks/core/statistics/trend_momentum.py`.
"""

import sys
from pathlib import Path

# Add parent directory to sys.path for core imports
_parent_dir = Path(__file__).resolve().parent.parent.parent
if str(_parent_dir) not in sys.path:
    sys.path.insert(0, str(_parent_dir))

# Re-export from core
from core.statistics.trend_momentum import (
    TrendMomentumAnalyzer,
    TrendDirection,
    TrendResult,
    analyze_trend_momentum,
    get_trend_direction,
    get_trend_score,
)

# TikTok-specific alias for backward compatibility
TrendMomentum = TrendMomentumAnalyzer


def get_tiktok_trend_description(direction: TrendDirection, roc: float) -> str:
    """Get TikTok-specific trend description in Indonesian."""
    descriptions = {
        TrendDirection.STRONG_UPTREND: f"📈 Momentum sangat kuat (ROC: {roc:+.1f}%)",
        TrendDirection.UPTREND: f"↗️ Trend naik (ROC: {roc:+.1f}%)",
        TrendDirection.WEAKENING: f"⚠️ Momentum melemah (ROC: {roc:+.1f}%)",
        TrendDirection.RECOVERING: f"🔄 Sedang recovery (ROC: {roc:+.1f}%)",
        TrendDirection.DOWNTREND: f"↘️ Trend turun (ROC: {roc:+.1f}%)",
        TrendDirection.STRONG_DOWNTREND: f"📉 Momentum sangat lemah (ROC: {roc:+.1f}%)",
        TrendDirection.INSUFFICIENT_DATA: "❓ Data tidak cukup",
    }
    return descriptions.get(direction, f"ROC: {roc:+.1f}%")


__all__ = [
    # Core exports
    "TrendMomentumAnalyzer",
    "TrendDirection",
    "TrendResult",
    "analyze_trend_momentum",
    "get_trend_direction",
    "get_trend_score",
    # TikTok-specific
    "TrendMomentum",  # Alias
    "get_tiktok_trend_description",
]
