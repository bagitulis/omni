#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads - Budget Elasticity Wrapper
======================================
TikTok-specific wrapper using core elasticity module.
Following AGENTS.MD: Clean Code, DRY, SRP.
"""

import pandas as pd
from typing import Dict

# Import core elasticity
from core.intelligence.elasticity import (
    calculate_budget_elasticity as core_calculate_elasticity,
    calculate_marginal_roi as core_calculate_mroi,
    ElasticityLevel,
    ElasticityResult
)

# Re-export for backward compatibility
__all__ = [
    'calculate_budget_elasticity',
    'calculate_marginal_roi',
    'determine_scaling_strategy',
    'ElasticityLevel',
    'ElasticityResult'
]


# TikTok-specific emoji labels
TIKTOK_ELASTICITY_LABELS = {
    ElasticityLevel.HIGHLY_ELASTIC: "HIGHLY ELASTIC (🚀)",
    ElasticityLevel.ELASTIC: "ELASTIC (✅)",
    ElasticityLevel.INELASTIC: "INELASTIC (⚠️)",
    ElasticityLevel.NEGATIVE: "NEGATIVE (⛔)",
    ElasticityLevel.UNKNOWN: "UNKNOWN (Not enough data)"
}


def calculate_budget_elasticity(product_df: pd.DataFrame) -> Dict:
    """
    Calculate budget elasticity for TikTok products.
    Wrapper with TikTok-specific formatting.
    
    Args:
        product_df: DataFrame with 'cost', 'grossRevenue', 'periodStart' columns
        
    Returns:
        Dict with elasticity, classification, etc.
    """
    result = core_calculate_elasticity(
        df=product_df,
        costColumn='cost',
        revenueColumn='grossRevenue',
        timeColumn='periodStart',
        minChangeThreshold=0.05
    )
    
    # Format for backward compatibility with TikTok labels
    return {
        "elasticity": result.elasticity,
        "recent_elasticity": result.recentElasticity,
        "is_valid": result.isValid,
        "classification": TIKTOK_ELASTICITY_LABELS.get(result.level, "UNKNOWN"),
        "n_samples": result.nSamples
    }


def calculate_marginal_roi(product_df: pd.DataFrame, window: int = 4) -> float:
    """
    Calculate Marginal ROI for TikTok products.
    
    Args:
        product_df: DataFrame with 'cost', 'grossRevenue', 'periodStart' columns
        window: Number of recent periods to analyze
        
    Returns:
        mROI value
    """
    return core_calculate_mroi(
        df=product_df,
        costColumn='cost',
        revenueColumn='grossRevenue',
        timeColumn='periodStart',
        window=window
    )


def determine_scaling_strategy(
    current_roi: float, 
    elasticity_data: Dict, 
    mroi: float,
    current_budget: float
) -> Dict:
    """
    Menentukan strategi scaling 2-Tahap (Berdasarkan Request User).
    
    Stage 1 (Conservative): ~10% increase
    Stage 2 (Aggressive): ~30% increase
    """
    
    elasticity = elasticity_data.get('elasticity', 0)
    is_valid_elasticity = elasticity_data.get('is_valid', False)
    
    strategy = "MAINTAIN"
    pct_change = 0.0
    reason = []
    
    # Logic Decision Tree
    
    # 1. Cek ROI Dasar dulu
    if current_roi < 2.0: # ROI Buruk
        strategy = "JANGAN NAIKKAN"
        pct_change = 0.0
        if current_roi < 1.0:
            strategy = "DECREASE/STOP"
            pct_change = -0.5
            reason.append("ROI sangat rendah (< 1.0)")
        else:
            reason.append("ROI belum aman (< 2.0)")
            
    else: # ROI Bagus (>= 2.0)
        
        # 2. Cek Elastisitas
        if is_valid_elasticity:
            if elasticity > 1.0:
                # Stage 2: Aggressive
                strategy = "SCALE UP (STAGE 2)"
                pct_change = 0.30  # 30%
                reason.append(f"Elastisitas Tinggi ({elasticity}). Market masih responsif.")
            elif elasticity > 0.5:
                # Stage 1: Conservative
                strategy = "SCALE UP (STAGE 1)"
                pct_change = 0.10 # 10%
                reason.append(f"Elastisitas Moderat ({elasticity}). Scale up pelan-pelan.")
            else:
                # Diminishing returns warning, despite good ROI
                strategy = "MAINTAIN / OPTIMIZE"
                pct_change = 0.0
                reason.append("Elastisitas Rendah. Kenaikan budget mungkin tidak profit.")
        else:
            # Tidak cukup data elastisitas -> Default ke Stage 1 jika ROI sangat bagus
            if current_roi > 5.0:
                strategy = "SCALE UP (STAGE 1)"
                pct_change = 0.10
                reason.append("ROI Sangat Bagus (>5), data elastisitas belum cukup.")
            else:
                strategy = "MAINTAIN"
                reason.append("ROI Bagus, tapi butuh data elastisitas untuk scale up.")
                
    # 3. Cek mROI (Marginal Return)
    # Hanya gunakan mROI sebagai blocker jika ada data tes budget baru-baru ini (delta cost signifikan)
    # Jika mROI 0 (tidak ada perubahan budget), abaikan check ini.
    if mroi < 0.8 and mroi != 0 and pct_change > 0:
        # Jika penambahan budget terakhir KURANG EFISIEN, batalkan kenaikan agresif
        if mroi < 0:
             strategy = "HOLD (RISK)"
             pct_change = 0.0
             reason.append(f"⚠️ Blocked: Penambahan budget terakhir merugi/negatif (mROI {mroi}).")
        else:
             # Masih profit tapi diminishing (0 < mROI < 0.8)
             # Turunkan level agresivitas
             if pct_change > 0.15: # Jika tadinya mau agresif 30%
                 strategy = "SCALE UP (CAUTIOUS)"
                 pct_change = 0.10
                 reason.append(f"⚠️ Adjusted: Marginal ROI rendah ({mroi}). Turun ke Stage 1.")
             else:
                 # Sudah konservatif, biarkan tapi kasih warning
                 reason.append(f"ℹ️ Note: Marginal ROI ({mroi}) mulai menurun.")

    recommended_budget = current_budget * (1 + pct_change)
    
    return {
        "strategy": strategy,
        "recommended_pct": f"{pct_change*100:+.0f}%",
        "recommended_budget": int(recommended_budget),
        "reason": "; ".join(reason),
        "metrics": {
            "elasticity": elasticity,
            "mROI": mroi,
            "current_roi": current_roi
        }
    }
