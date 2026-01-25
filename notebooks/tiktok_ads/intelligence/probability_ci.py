#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Probability with CI (Re-exported from Core + TikTok Extensions)
===============================================================
Re-exports core probability engine and adds TikTok-specific formatting.
Base logic is in `notebooks/core/intelligence/probability_ci.py`.
"""

import sys
from pathlib import Path

# Add parent directory to sys.path for core imports
_parent_dir = Path(__file__).resolve().parent.parent.parent
if str(_parent_dir) not in sys.path:
    sys.path.insert(0, str(_parent_dir))

# Re-export from core
from core.intelligence.probability_ci import (
    ProbabilityWithCI,
    ProbabilityEstimate,
    estimate_probability,
)


def get_confidence_label(level: str) -> str:
    """Get Indonesian confidence label with emoji."""
    labels = {
        "HIGH": "✅ TINGGI",
        "MEDIUM": "🔶 SEDANG",
        "LOW": "⚠️ RENDAH",
        "VERY_LOW": "❌ SANGAT RENDAH",
    }
    return labels.get(level, level)


def get_uncertainty_label(level: str) -> str:
    """Get Indonesian uncertainty label."""
    labels = {
        "LOW": "📊 Prediksi stabil",
        "MEDIUM": "📈 Variasi moderat",
        "HIGH": "📉 Variasi tinggi",
    }
    return labels.get(level, level)


def format_probability_summary(estimate: ProbabilityEstimate) -> str:
    """Format probability estimate as Indonesian summary."""
    lines = [
        f"Probabilitas Sukses: {estimate.successProbability:.1f}%",
        f"  CI 95%: [{estimate.successCI[0]:.1f}%, {estimate.successCI[1]:.1f}%]",
        f"Probabilitas Profit: {estimate.profitProbability:.1f}%",
        f"Probabilitas ROAS Target: {estimate.roasTargetProbability:.1f}%",
        f"Confidence: {get_confidence_label(estimate.confidenceLevel)}",
        f"Uncertainty: {get_uncertainty_label(estimate.uncertaintyLevel)}",
    ]
    
    if estimate.positiveFactors:
        lines.append(f"✅ Faktor Positif: {', '.join(estimate.positiveFactors[:3])}")
    
    if estimate.negativeFactors:
        lines.append(f"⚠️ Faktor Negatif: {', '.join(estimate.negativeFactors[:3])}")
    
    return "\n".join(lines)


__all__ = [
    # Core exports
    "ProbabilityWithCI",
    "ProbabilityEstimate",
    "estimate_probability",
    # TikTok-specific
    "get_confidence_label",
    "get_uncertainty_label",
    "format_probability_summary",
]
