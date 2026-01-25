"""
TikTok Ads ML - ROI Predictor
Train and predict ROI using LightGBM/XGBoost
"""

import json
import joblib
import numpy as np
import pandas as pd
from datetime import datetime
from typing import Dict, Any, Optional, List
from pathlib import Path

try:
    import lightgbm as lgb
    HAS_LIGHTGBM = True
except ImportError:
    HAS_LIGHTGBM = False

try:
    import xgboost as xgb
    HAS_XGBOOST = True
except ImportError:
    HAS_XGBOOST = False

from sklearn.ensemble import RandomForestRegressor
from sklearn.metrics import mean_squared_error, mean_absolute_error, r2_score

from config import (
    FEATURE_COLUMNS,
    TARGET_COLUMN,
    LIGHTGBM_PARAMS,
    XGBOOST_PARAMS,
    MODEL_VERSION,
    ensure_model_dir,
)
from data_loader import (
    load_creative_data,
    prepare_features,
    aggregate_by_product,
    split_train_test,
    get_feature_target,
)


class ROIPredictor:
    """
    ROI Prediction Model using ensemble methods
    """
    
    def __init__(self, model_type: str = "auto"):
        """
        Initialize predictor
        
        Args:
            model_type: "lightgbm", "xgboost", "random_forest", or "auto"
        """
        self.model_type = self._select_model_type(model_type)
        self.model = None
        self.feature_columns = FEATURE_COLUMNS.copy()
        self.is_trained = False
        self.metrics: Dict[str, float] = {}
    
    def _select_model_type(self, requested: str) -> str:
        """Select best available model type"""
        if requested == "auto":
            if HAS_LIGHTGBM:
                return "lightgbm"
            elif HAS_XGBOOST:
                return "xgboost"
            else:
                return "random_forest"
        return requested
    
    def _create_model(self):
        """Create model instance based on type"""
        if self.model_type == "lightgbm" and HAS_LIGHTGBM:
            return lgb.LGBMRegressor(**LIGHTGBM_PARAMS)
        elif self.model_type == "xgboost" and HAS_XGBOOST:
            return xgb.XGBRegressor(**XGBOOST_PARAMS)
        else:
            return RandomForestRegressor(
                n_estimators=100,
                max_depth=10,
                random_state=42
            )
    
    def train(self, df: pd.DataFrame) -> Dict[str, float]:
        """
        Train the model on prepared data
        
        Args:
            df: Prepared dataframe with features
            
        Returns:
            Dictionary of evaluation metrics
        """
        # Split data
        train_df, test_df = split_train_test(df)
        
        # Get features and target
        X_train, y_train = get_feature_target(
            train_df, self.feature_columns, TARGET_COLUMN
        )
        X_test, y_test = get_feature_target(
            test_df, self.feature_columns, TARGET_COLUMN
        )
        
        # Create and train model
        self.model = self._create_model()
        self.model.fit(X_train, y_train)
        
        # Evaluate
        y_pred = self.model.predict(X_test)
        
        self.metrics = {
            "rmse": float(np.sqrt(mean_squared_error(y_test, y_pred))),
            "mae": float(mean_absolute_error(y_test, y_pred)),
            "r2": float(r2_score(y_test, y_pred)),
            "train_samples": len(train_df),
            "test_samples": len(test_df),
        }
        
        self.is_trained = True
        return self.metrics
    
    def predict(self, df: pd.DataFrame) -> np.ndarray:
        """
        Predict ROI for new data
        
        Args:
            df: Dataframe with features
            
        Returns:
            Array of predicted ROI values
        """
        if not self.is_trained:
            raise ValueError("Model not trained. Call train() first.")
        
        X = df[self.feature_columns]
        return self.model.predict(X)
    
    def predict_with_confidence(
        self, df: pd.DataFrame
    ) -> List[Dict[str, Any]]:
        """
        Predict ROI with confidence scores
        
        Returns list of predictions with:
        - predicted_roi
        - confidence (based on feature similarity to training data)
        """
        predictions = self.predict(df)
        
        # Simple confidence based on data characteristics
        results = []
        for i, pred in enumerate(predictions):
            row = df.iloc[i]
            
            # Higher confidence if cost and impressions are substantial
            cost_factor = min(1.0, row["cost"] / 1000000)  # Normalize
            imp_factor = min(1.0, row["impressions"] / 10000)
            confidence = (cost_factor + imp_factor) / 2 * 0.8 + 0.2
            
            results.append({
                "predicted_roi": float(pred),
                "confidence": float(confidence),
            })
        
        return results
    
    def save(self, tenant_id: str) -> Path:
        """Save trained model to disk"""
        model_dir = ensure_model_dir()
        model_path = model_dir / f"roi_predictor_{tenant_id}.joblib"
        
        model_data = {
            "model": self.model,
            "model_type": self.model_type,
            "feature_columns": self.feature_columns,
            "metrics": self.metrics,
            "version": MODEL_VERSION,
            "trained_at": datetime.now().isoformat(),
        }
        
        joblib.dump(model_data, model_path)
        return model_path
    
    def load(self, tenant_id: str) -> bool:
        """Load model from disk"""
        model_dir = ensure_model_dir()
        model_path = model_dir / f"roi_predictor_{tenant_id}.joblib"
        
        if not model_path.exists():
            return False
        
        model_data = joblib.load(model_path)
        self.model = model_data["model"]
        self.model_type = model_data["model_type"]
        self.feature_columns = model_data["feature_columns"]
        self.metrics = model_data["metrics"]
        self.is_trained = True
        
        return True


def train_roi_model(tenant_id: str) -> Dict[str, Any]:
    """
    Train ROI prediction model for a tenant
    
    Args:
        tenant_id: Tenant ID
        
    Returns:
        Training results including metrics
    """
    # Load and prepare data
    df = load_creative_data(tenant_id)
    
    if len(df) < 50:
        return {
            "success": False,
            "error": f"Insufficient data: {len(df)} rows (need at least 50)",
        }
    
    df = prepare_features(df)
    df = aggregate_by_product(df)
    
    # Train model
    predictor = ROIPredictor(model_type="auto")
    metrics = predictor.train(df)
    
    # Save model
    model_path = predictor.save(tenant_id)
    
    return {
        "success": True,
        "model_type": predictor.model_type,
        "model_path": str(model_path),
        "metrics": metrics,
        "version": MODEL_VERSION,
    }


def predict_roi(tenant_id: str, product_ids: Optional[List[str]] = None) -> List[Dict]:
    """
    Predict ROI for products
    
    Args:
        tenant_id: Tenant ID
        product_ids: Optional list of product IDs to predict for
        
    Returns:
        List of predictions per product
    """
    predictor = ROIPredictor()
    
    if not predictor.load(tenant_id):
        return [{"error": "Model not found. Train first."}]
    
    # Load latest data
    df = load_creative_data(tenant_id)
    df = prepare_features(df)
    df = aggregate_by_product(df)
    
    if product_ids:
        df = df[df["product_id"].isin(product_ids)]
    
    if len(df) == 0:
        return []
    
    predictions = predictor.predict_with_confidence(df)
    
    results = []
    for i, row in df.iterrows():
        pred = predictions[i] if i < len(predictions) else predictions[0]
        results.append({
            "product_id": row["product_id"],
            "current_roi": float(row["roi"]),
            "predicted_roi": pred["predicted_roi"],
            "confidence": pred["confidence"],
        })
    
    return results


if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 3:
        print("Usage: python roi_predictor.py <train|predict> <tenant_id>")
        sys.exit(1)
    
    action = sys.argv[1]
    tenant_id = sys.argv[2]
    
    if action == "train":
        result = train_roi_model(tenant_id)
        print(json.dumps(result, indent=2))
    elif action == "predict":
        product_ids = sys.argv[3:] if len(sys.argv) > 3 else None
        results = predict_roi(tenant_id, product_ids)
        print(json.dumps(results, indent=2))
    else:
        print(f"Unknown action: {action}")
        sys.exit(1)
