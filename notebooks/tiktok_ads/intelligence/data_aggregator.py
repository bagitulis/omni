#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Data Aggregator
===============
Aggregates raw data into analysis-ready format.
Following AGENTS.MD: Clean Code, DRY, SRP.
"""

from dataclasses import dataclass
from typing import Dict, List, Optional
import pandas as pd
import numpy as np
import sqlite3
from pathlib import Path


@dataclass
class ProductMetrics:
    """Aggregated metrics for a single product."""
    productId: str
    productName: str
    totalCost: float
    totalRevenue: float
    totalProfit: float
    totalOrders: int
    avgRoas: float
    totalImpressions: int
    totalClicks: int
    avgCtr: float
    avgCvr: float
    periodsCount: int
    firstPeriod: str
    lastPeriod: str


class DataAggregator:
    """
    Aggregates raw TikTok ads data into analysis-ready format.
    Handles time-windowed aggregation for trend analysis.
    """
    
    def __init__(self, dbPath: str, tenantId: str = "yumna_bertigamart"):
        self.dbPath = Path(dbPath)
        self.tenantId = tenantId
    
    def loadRawData(self, minCost: float = 0) -> pd.DataFrame:
        """Load raw data from database."""
        conn = sqlite3.connect(str(self.dbPath))
        
        query = """
            SELECT 
                productId, campaignName, creativeType, videoTitle,
                periodStart, periodEnd, periodLabel,
                cost, grossRevenue, ordersSku, roi,
                impressions, clicks, ctr, conversionRate
            FROM TiktokAdsCreativeData 
            WHERE tenantId = ? AND cost >= ?
            ORDER BY periodStart
        """
        
        df = pd.read_sql_query(query, conn, params=(self.tenantId, minCost))
        conn.close()
        
        # Parse dates
        df['periodStart'] = pd.to_datetime(df['periodStart'])
        df['periodEnd'] = pd.to_datetime(df['periodEnd'])
        
        return df
    
    def aggregateByProduct(self, df: pd.DataFrame = None) -> pd.DataFrame:
        """Aggregate data by product across all periods."""
        if df is None:
            df = self.loadRawData()
        
        if df.empty:
            return pd.DataFrame()
        
        agg = df.groupby('productId').agg({
            'campaignName': 'first',
            'videoTitle': 'first',
            'cost': 'sum',
            'grossRevenue': 'sum',
            'ordersSku': 'sum',
            'impressions': 'sum',
            'clicks': 'sum',
            'periodStart': ['min', 'max', 'count'],
        }).reset_index()
        
        # Flatten column names
        agg.columns = [
            'productId', 'campaignName', 'videoTitle',
            'totalCost', 'totalRevenue', 'totalOrders',
            'totalImpressions', 'totalClicks',
            'firstPeriod', 'lastPeriod', 'periodsCount'
        ]
        
        # Calculate derived metrics
        agg['totalProfit'] = agg['totalRevenue'] - agg['totalCost']
        agg['avgRoas'] = np.where(
            agg['totalCost'] > 0,
            agg['totalRevenue'] / agg['totalCost'],
            0
        )
        agg['avgCtr'] = np.where(
            agg['totalImpressions'] > 0,
            (agg['totalClicks'] / agg['totalImpressions']) * 100,
            0
        )
        agg['avgCvr'] = np.where(
            agg['totalClicks'] > 0,
            (agg['totalOrders'] / agg['totalClicks']) * 100,
            0
        )
        
        return agg
    
    def getProductTimeSeries(self, productId: str, 
                             df: pd.DataFrame = None) -> pd.DataFrame:
        """Get time series data for a specific product."""
        if df is None:
            df = self.loadRawData()
        
        productDf = df[df['productId'] == productId].copy()
        
        if productDf.empty:
            return pd.DataFrame()
        
        # Aggregate by period for this product
        timeSeries = productDf.groupby('periodStart').agg({
            'cost': 'sum',
            'grossRevenue': 'sum',
            'ordersSku': 'sum',
            'impressions': 'sum',
            'clicks': 'sum',
        }).reset_index()
        
        timeSeries = timeSeries.sort_values('periodStart')
        
        # Calculate per-period metrics
        timeSeries['roi'] = np.where(
            timeSeries['cost'] > 0,
            timeSeries['grossRevenue'] / timeSeries['cost'],
            0
        )
        timeSeries['ctr'] = np.where(
            timeSeries['impressions'] > 0,
            (timeSeries['clicks'] / timeSeries['impressions']) * 100,
            0
        )
        timeSeries['cvr'] = np.where(
            timeSeries['clicks'] > 0,
            (timeSeries['ordersSku'] / timeSeries['clicks']) * 100,
            0
        )
        
        return timeSeries
    
    def getRecentWindow(self, df: pd.DataFrame, 
                        periods: int = 4) -> pd.DataFrame:
        """Get data for most recent N periods."""
        if df.empty:
            return df
        
        # Get unique periods sorted
        uniquePeriods = df['periodStart'].sort_values().unique()
        
        if len(uniquePeriods) <= periods:
            return df
        
        cutoffDate = uniquePeriods[-periods]
        return df[df['periodStart'] >= cutoffDate]
    
    def getWindowedStats(self, productId: str,
                         df: pd.DataFrame = None) -> Dict[str, pd.DataFrame]:
        """
        Get statistics for different time windows.
        Returns dict with 'short', 'medium', 'long' DataFrames.
        """
        timeSeries = self.getProductTimeSeries(productId, df)
        
        if timeSeries.empty:
            return {'short': pd.DataFrame(), 
                    'medium': pd.DataFrame(), 
                    'long': timeSeries}
        
        return {
            'short': self.getRecentWindow(timeSeries, 3),   # ~3 weeks
            'medium': self.getRecentWindow(timeSeries, 8),  # ~2 months
            'long': timeSeries,                              # All data
        }
    
    def getSummaryStats(self, df: pd.DataFrame = None) -> Dict:
        """Get overall summary statistics."""
        if df is None:
            df = self.loadRawData()
        
        if df.empty:
            return {}
        
        totalCost = df['cost'].sum()
        totalRevenue = df['grossRevenue'].sum()
        
        return {
            'totalCost': totalCost,
            'totalRevenue': totalRevenue,
            'totalProfit': totalRevenue - totalCost,
            'overallRoas': totalRevenue / totalCost if totalCost > 0 else 0,
            'totalProducts': df['productId'].nunique(),
            'totalPeriods': df['periodStart'].nunique(),
            'dateRange': {
                'start': df['periodStart'].min(),
                'end': df['periodEnd'].max(),
            }
        }
