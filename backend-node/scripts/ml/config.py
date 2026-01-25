"""
TikTok Ads ML - Configuration
Configuration and constants for ML models
"""

import os
from pathlib import Path

# Database configuration
DATABASE_DIR = Path(__file__).parent.parent.parent / "config" / "databases"

# Model configuration
MODEL_VERSION = "v1.0"
MODEL_DIR = Path(__file__).parent / "models"

# Feature columns for training
FEATURE_COLUMNS = [
    "cost",
    "impressions", 
    "clicks",
    "ctr",
    "conversion_rate",
    "creative_type_encoded",
    "month",
]

# Target column
TARGET_COLUMN = "roi"

# Model hyperparameters
LIGHTGBM_PARAMS = {
    "objective": "regression",
    "metric": "rmse",
    "boosting_type": "gbdt",
    "num_leaves": 31,
    "learning_rate": 0.05,
    "feature_fraction": 0.9,
    "bagging_fraction": 0.8,
    "bagging_freq": 5,
    "verbose": -1,
    "n_estimators": 100,
}

XGBOOST_PARAMS = {
    "objective": "reg:squarederror",
    "max_depth": 6,
    "learning_rate": 0.1,
    "n_estimators": 100,
    "subsample": 0.8,
    "colsample_bytree": 0.8,
    "random_state": 42,
}

# Performance score thresholds
PERFORMANCE_THRESHOLDS = {
    "A": 10.0,   # ROI >= 10
    "B": 5.0,    # ROI >= 5
    "C": 2.0,    # ROI >= 2
    "D": 1.0,    # ROI >= 1
    "F": 0.0,    # ROI < 1
}

# Anomaly detection threshold (z-score)
ANOMALY_THRESHOLD = 3.0

# Creative type encoding
CREATIVE_TYPE_MAP = {
    "Video": 1,
    "Kartu produk": 0,
    "Unknown": 0,
}


def get_database_path(tenant_id: str) -> Path:
    """Get database file path for tenant"""
    return DATABASE_DIR / f"{tenant_id}.db"


def ensure_model_dir():
    """Ensure model directory exists"""
    MODEL_DIR.mkdir(parents=True, exist_ok=True)
    return MODEL_DIR
