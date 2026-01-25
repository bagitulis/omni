#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Budget Elasticity Analysis (Generic Core Module)
================================================
Analyzes responsiveness of revenue to budget changes.
Platform-agnostic: Works for TikTok, Shopee, Lazada, etc.
Following AGENTS.MD: Clean Code, DRY, SRP.
"""

import numpy as np
import pandas as pd
from dataclasses import dataclass
from typing import Dict, List, Optional
from enum import Enum


class ElasticityLevel(Enum):
    """Elasticity classification levels."""
    HIGHLY_ELASTIC = "HIGHLY_ELASTIC"  # E > 1.2
    ELASTIC = "ELASTIC"                 # 0.8 < E <= 1.2
    INELASTIC = "INELASTIC"            # 0 < E <= 0.8
    NEGATIVE = "NEGATIVE"               # E <= 0
    UNKNOWN = "UNKNOWN"                 # Insufficient data


@dataclass
class ElasticityResult:
    """Elasticity analysis result."""
    elasticity: float
    recentElasticity: float
    level: ElasticityLevel
    isValid: bool
    nSamples: int
    description: str


def calculate_budget_elasticity(
    df: pd.DataFrame,
    costColumn: str = 'cost',
    revenueColumn: str = 'grossRevenue',
    timeColumn: str = 'periodStart',
    minChangeThreshold: float = 0.05
) -> ElasticityResult:
    """
    Calculate budget elasticity from time series data.
    
    Elasticity (E) = (% Change Revenue) / (% Change Cost)
    
    Interpretation:
    - E > 1.0: Elastic - revenue grows faster than cost
    - 0 < E < 1.0: Inelastic - diminishing returns
    - E < 0: Negative - increasing cost reduces revenue
    
    Args:
        df: DataFrame with cost and revenue columns
        costColumn: Name of cost column
        revenueColumn: Name of revenue column
        timeColumn: Name of time column for sorting
        minChangeThreshold: Minimum % change to consider significant
    
    Returns:
        ElasticityResult with elasticity value and classification
    """
    if len(df) < 3:
        return ElasticityResult(
            elasticity=0,
            recentElasticity=0,
            level=ElasticityLevel.UNKNOWN,
            isValid=False,
            nSamples=len(df),
            description="Insufficient data for elasticity analysis"
        )
    
    # Sort by time
    sorted_df = df.sort_values(timeColumn).copy()
    
    # Calculate % changes
    sorted_df['cost_pct_change'] = sorted_df[costColumn].pct_change()
    sorted_df['rev_pct_change'] = sorted_df[revenueColumn].pct_change()
    
    # Filter significant changes
    significant = sorted_df[abs(sorted_df['cost_pct_change']) > minChangeThreshold].copy()
    
    if len(significant) == 0:
        return ElasticityResult(
            elasticity=1.0,
            recentElasticity=0,
            level=ElasticityLevel.UNKNOWN,
            isValid=False,
            nSamples=0,
            description="No significant budget changes detected"
        )
    
    # Calculate elasticity per period
    significant['elasticity'] = significant['rev_pct_change'] / significant['cost_pct_change']
    
    # Use median for robustness
    medianElasticity = significant['elasticity'].median()
    recentElasticity = significant['elasticity'].iloc[-1] if len(significant) > 0 else 0
    
    # Classification
    level = _classify_elasticity(medianElasticity)
    description = _get_elasticity_description(level, medianElasticity)
    
    return ElasticityResult(
        elasticity=round(medianElasticity, 2),
        recentElasticity=round(recentElasticity, 2),
        level=level,
        isValid=True,
        nSamples=len(significant),
        description=description
    )


def calculate_marginal_roi(
    df: pd.DataFrame,
    costColumn: str = 'cost',
    revenueColumn: str = 'grossRevenue',
    timeColumn: str = 'periodStart',
    window: int = 4
) -> float:
    """
    Calculate Marginal ROI for the last n periods.
    mROI = (Total Delta Revenue) / (Total Delta Cost)
    
    Measures efficiency of RECENT budget additions.
    
    Args:
        df: DataFrame with cost and revenue columns
        costColumn: Name of cost column
        revenueColumn: Name of revenue column
        timeColumn: Name of time column
        window: Number of periods to consider
    
    Returns:
        Marginal ROI value
    """
    if len(df) < 2:
        return 0.0
    
    sorted_df = df.sort_values(timeColumn).tail(window + 1).copy()
    
    deltaCost = sorted_df[costColumn].diff().sum()
    deltaRevenue = sorted_df[revenueColumn].diff().sum()
    
    if abs(deltaCost) < 1000:  # Avoid division by small numbers
        return 0.0
    
    return round(deltaRevenue / deltaCost, 2)


def _classify_elasticity(elasticity: float) -> ElasticityLevel:
    """Classify elasticity value into level."""
    if elasticity > 1.2:
        return ElasticityLevel.HIGHLY_ELASTIC
    elif elasticity > 0.8:
        return ElasticityLevel.ELASTIC
    elif elasticity > 0:
        return ElasticityLevel.INELASTIC
    else:
        return ElasticityLevel.NEGATIVE


def _get_elasticity_description(level: ElasticityLevel, value: float) -> str:
    """Get description for elasticity level."""
    descriptions = {
        ElasticityLevel.HIGHLY_ELASTIC: f"Highly elastic ({value:.2f}) - Strong response to budget increase",
        ElasticityLevel.ELASTIC: f"Elastic ({value:.2f}) - Good response to budget",
        ElasticityLevel.INELASTIC: f"Inelastic ({value:.2f}) - Diminishing returns",
        ElasticityLevel.NEGATIVE: f"Negative ({value:.2f}) - Budget increase hurts revenue",
        ElasticityLevel.UNKNOWN: "Insufficient data",
    }
    return descriptions.get(level, f"Elasticity: {value:.2f}")


# Convenience aliases
def get_elasticity_score(df: pd.DataFrame) -> float:
    """Get elasticity value as score (0-100)."""
    result = calculate_budget_elasticity(df)
    if not result.isValid:
        return 50  # Neutral
    
    # Map elasticity to 0-100 score
    e = result.elasticity
    if e >= 1.5:
        return 100
    elif e >= 1.0:
        return 70 + (e - 1.0) * 60  # 70-100 for 1.0-1.5
    elif e >= 0.5:
        return 40 + (e - 0.5) * 60  # 40-70 for 0.5-1.0
    elif e >= 0:
        return 20 + e * 40          # 20-40 for 0-0.5
    else:
        return max(0, 20 + e * 20)  # 0-20 for negative
