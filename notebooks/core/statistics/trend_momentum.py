#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Trend Momentum Analyzer (Generic Core Module)
=============================================
MACD-style trend analysis for performance metrics.
Platform-agnostic: Works for TikTok, Shopee, Lazada, etc.
Following AGENTS.MD: Clean Code, DRY, SRP.
"""

from dataclasses import dataclass
from typing import Optional
from enum import Enum
import pandas as pd
import numpy as np


class TrendDirection(Enum):
    """Trend direction classification."""
    STRONG_UPTREND = "STRONG_UPTREND"
    UPTREND = "UPTREND"
    WEAKENING = "WEAKENING"
    RECOVERING = "RECOVERING"
    DOWNTREND = "DOWNTREND"
    STRONG_DOWNTREND = "STRONG_DOWNTREND"
    INSUFFICIENT_DATA = "INSUFFICIENT_DATA"


@dataclass
class TrendResult:
    """Trend analysis result."""
    direction: TrendDirection
    signal: float           # EMA_short - EMA_long
    signalChange: float     # Change in signal
    roc: float              # Rate of Change (%)
    score: float            # 0-100 score
    description: str


class TrendMomentumAnalyzer:
    """
    MACD-style momentum analysis for performance metrics.
    Uses Exponential Moving Averages to detect trends.
    """
    
    def __init__(self, shortWindow: int = 3, longWindow: int = 8):
        """
        Initialize with window sizes.
        
        Args:
            shortWindow: Short EMA window (default 3 periods)
            longWindow: Long EMA window (default 8 periods)
        """
        self.shortWindow = shortWindow
        self.longWindow = longWindow
        self.minPeriods = longWindow + 1
    
    def analyze(self, values: pd.Series) -> TrendResult:
        """
        Analyze trend from time series of values.
        
        Args:
            values: Time-ordered series of metric values
        
        Returns:
            TrendResult with direction, signal, and score
        """
        if len(values) < self.minPeriods:
            return TrendResult(
                direction=TrendDirection.INSUFFICIENT_DATA,
                signal=0,
                signalChange=0,
                roc=0,
                score=50,
                description="Insufficient data for trend analysis"
            )
        
        # Calculate EMAs
        emaShort = values.ewm(span=self.shortWindow, adjust=False).mean()
        emaLong = values.ewm(span=self.longWindow, adjust=False).mean()
        
        # MACD Signal
        signal = emaShort.iloc[-1] - emaLong.iloc[-1]
        prevSignal = emaShort.iloc[-2] - emaLong.iloc[-2]
        signalChange = signal - prevSignal
        
        # Rate of Change (ROC)
        roc = self._calculateRoc(values, self.shortWindow)
        
        # Determine direction
        direction = self._getDirection(signal, signalChange)
        
        # Calculate score
        score = self._calculateScore(signal, signalChange, roc, emaLong.iloc[-1])
        
        # Generate description
        description = self._getDescription(direction, roc)
        
        return TrendResult(
            direction=direction,
            signal=round(signal, 4),
            signalChange=round(signalChange, 4),
            roc=round(roc, 2),
            score=round(score, 1),
            description=description
        )
    
    def analyzeFromDf(self, df: pd.DataFrame, 
                      metricColumn: str = 'roi') -> TrendResult:
        """
        Analyze trend from DataFrame.
        
        Args:
            df: DataFrame sorted by time with metric column
            metricColumn: Column name to analyze
        
        Returns:
            TrendResult
        """
        if metricColumn not in df.columns:
            return TrendResult(
                direction=TrendDirection.INSUFFICIENT_DATA,
                signal=0, signalChange=0, roc=0, score=50,
                description=f"Column '{metricColumn}' not found"
            )
        
        values = df[metricColumn].dropna()
        return self.analyze(values)
    
    def _calculateRoc(self, values: pd.Series, periods: int) -> float:
        """Calculate Rate of Change percentage."""
        if len(values) < periods + 1:
            return 0
        
        current = values.iloc[-1]
        previous = values.iloc[-(periods + 1)]
        
        if previous == 0:
            return 0
        
        return ((current - previous) / abs(previous)) * 100
    
    def _getDirection(self, signal: float, signalChange: float) -> TrendDirection:
        """Determine trend direction from signal and change."""
        if signal > 0:
            if signalChange > 0:
                return TrendDirection.STRONG_UPTREND
            elif signalChange < -0.01:
                return TrendDirection.WEAKENING
            else:
                return TrendDirection.UPTREND
        else:
            if signalChange > 0.01:
                return TrendDirection.RECOVERING
            elif signalChange < 0:
                return TrendDirection.STRONG_DOWNTREND
            else:
                return TrendDirection.DOWNTREND
    
    def _calculateScore(self, signal: float, signalChange: float, 
                        roc: float, emaLong: float) -> float:
        """Calculate trend score (0-100)."""
        score = 50
        
        if emaLong != 0:
            signalRatio = signal / abs(emaLong)
            score += min(25, max(-25, signalRatio * 100))
        
        rocContribution = min(25, max(-25, roc / 4))
        score += rocContribution
        
        return min(100, max(0, score))
    
    def _getDescription(self, direction: TrendDirection, roc: float) -> str:
        """Get human-readable description."""
        descriptions = {
            TrendDirection.STRONG_UPTREND: f"Strong upward momentum (ROC: {roc:+.1f}%)",
            TrendDirection.UPTREND: f"Upward trend (ROC: {roc:+.1f}%)",
            TrendDirection.WEAKENING: f"Weakening momentum (ROC: {roc:+.1f}%)",
            TrendDirection.RECOVERING: f"Recovery phase (ROC: {roc:+.1f}%)",
            TrendDirection.DOWNTREND: f"Downward trend (ROC: {roc:+.1f}%)",
            TrendDirection.STRONG_DOWNTREND: f"Strong downward momentum (ROC: {roc:+.1f}%)",
            TrendDirection.INSUFFICIENT_DATA: "Insufficient data",
        }
        return descriptions.get(direction, f"ROC: {roc:+.1f}%")


# Convenience functions
def analyze_trend_momentum(values: pd.Series, 
                           shortWindow: int = 3, 
                           longWindow: int = 8) -> TrendResult:
    """Analyze trend momentum from series."""
    analyzer = TrendMomentumAnalyzer(shortWindow, longWindow)
    return analyzer.analyze(values)


def get_trend_direction(values: pd.Series) -> str:
    """Get simple trend direction string."""
    result = analyze_trend_momentum(values)
    return result.direction.value


def get_trend_score(values: pd.Series) -> float:
    """Get trend score (0-100)."""
    result = analyze_trend_momentum(values)
    return result.score
