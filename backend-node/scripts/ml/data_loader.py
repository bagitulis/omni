"""
TikTok Ads ML - Data Loader
Load and prepare data from SQLite database for ML training
"""

import sqlite3
import pandas as pd
import numpy as np
from typing import Optional, Tuple
from datetime import datetime

from config import get_database_path, CREATIVE_TYPE_MAP


def load_creative_data(
    tenant_id: str,
    period_start: Optional[datetime] = None,
    period_end: Optional[datetime] = None,
) -> pd.DataFrame:
    """
    Load creative data from database for ML training
    """
    db_path = get_database_path(tenant_id)
    
    if not db_path.exists():
        raise FileNotFoundError(f"Database not found: {db_path}")
    
    conn = sqlite3.connect(str(db_path))
    
    query = """
    SELECT 
        productId as product_id,
        campaignId as campaign_id,
        creativeType as creative_type,
        cost,
        ordersSku as orders,
        grossRevenue as revenue,
        roi,
        impressions,
        clicks,
        ctr,
        conversionRate as conversion_rate,
        watchRate2s as watch_rate_2s,
        watchRate6s as watch_rate_6s,
        periodStart as period_start,
        periodEnd as period_end
    FROM TiktokAdsCreativeData
    WHERE tenantId = ?
    """
    
    params = [tenant_id]
    
    if period_start:
        query += " AND periodStart >= ?"
        params.append(period_start.isoformat())
    
    if period_end:
        query += " AND periodEnd <= ?"
        params.append(period_end.isoformat())
    
    df = pd.read_sql_query(query, conn, params=params)
    conn.close()
    
    return df


def prepare_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Prepare features for ML model
    """
    df = df.copy()
    
    # Encode creative type
    df["creative_type_encoded"] = df["creative_type"].map(CREATIVE_TYPE_MAP).fillna(0)
    
    # Extract month from period
    df["period_start"] = pd.to_datetime(df["period_start"])
    df["month"] = df["period_start"].dt.month
    
    # Fill NaN values
    numeric_cols = ["cost", "orders", "revenue", "roi", "impressions", 
                    "clicks", "ctr", "conversion_rate"]
    for col in numeric_cols:
        if col in df.columns:
            df[col] = df[col].fillna(0)
    
    # Remove invalid rows (zero cost or negative values)
    df = df[df["cost"] > 0]
    
    return df


def aggregate_by_product(df: pd.DataFrame) -> pd.DataFrame:
    """
    Aggregate data by product for product-level predictions
    """
    agg_df = df.groupby("product_id").agg({
        "cost": "sum",
        "orders": "sum",
        "revenue": "sum",
        "impressions": "sum",
        "clicks": "sum",
        "creative_type_encoded": "mean",  # Average (0-1 scale)
        "month": "first",  # Just take first month
    }).reset_index()
    
    # Recalculate derived metrics
    agg_df["roi"] = np.where(
        agg_df["cost"] > 0,
        agg_df["revenue"] / agg_df["cost"],
        0
    )
    agg_df["ctr"] = np.where(
        agg_df["impressions"] > 0,
        agg_df["clicks"] / agg_df["impressions"],
        0
    )
    agg_df["conversion_rate"] = np.where(
        agg_df["clicks"] > 0,
        agg_df["orders"] / agg_df["clicks"],
        0
    )
    
    return agg_df


def split_train_test(
    df: pd.DataFrame,
    test_size: float = 0.2,
    random_state: int = 42
) -> Tuple[pd.DataFrame, pd.DataFrame]:
    """
    Split data into training and test sets
    """
    from sklearn.model_selection import train_test_split
    
    train_df, test_df = train_test_split(
        df, 
        test_size=test_size, 
        random_state=random_state
    )
    
    return train_df, test_df


def get_feature_target(
    df: pd.DataFrame,
    feature_cols: list,
    target_col: str
) -> Tuple[pd.DataFrame, pd.Series]:
    """
    Extract features and target from dataframe
    """
    X = df[feature_cols].copy()
    y = df[target_col].copy()
    
    return X, y
