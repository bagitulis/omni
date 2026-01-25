"""
TikTok Ads ML - Creative Scorer
Score creative performance and provide recommendations
"""

import json
import numpy as np
import pandas as pd
from typing import Dict, Any, List
from datetime import datetime

from config import PERFORMANCE_THRESHOLDS, ANOMALY_THRESHOLD
from data_loader import load_creative_data, prepare_features


def calculate_performance_score(roi: float) -> str:
    """
    Calculate performance grade based on ROI
    
    Args:
        roi: Return on Investment value
        
    Returns:
        Grade A/B/C/D/F
    """
    for grade, threshold in PERFORMANCE_THRESHOLDS.items():
        if roi >= threshold:
            return grade
    return "F"


def calculate_creative_score(row: pd.Series) -> float:
    """
    Calculate creative score (0-100) based on multiple metrics
    
    Weights:
    - ROI: 40%
    - CTR: 20%
    - Conversion Rate: 25%
    - Watch Rate (if video): 15%
    """
    score = 0.0
    
    # ROI score (0-40)
    roi = row.get("roi", 0)
    roi_score = min(40, roi * 4)  # Cap at 40
    score += roi_score
    
    # CTR score (0-20) - assume good CTR is > 5%
    ctr = row.get("ctr", 0)
    ctr_score = min(20, ctr * 400)  # 5% CTR = 20 points
    score += ctr_score
    
    # Conversion score (0-25) - assume good conversion is > 10%
    conv = row.get("conversion_rate", 0)
    conv_score = min(25, conv * 250)  # 10% conv = 25 points
    score += conv_score
    
    # Watch rate score (0-15) - only for videos
    if row.get("creative_type") == "Video":
        watch_rate = row.get("watch_rate_6s", 0) or 0
        watch_score = min(15, watch_rate * 50)  # 30% watch = 15 points
        score += watch_score
    else:
        # Non-video gets partial score
        score += 7.5
    
    return min(100, max(0, score))


def detect_anomalies(df: pd.DataFrame) -> pd.DataFrame:
    """
    Detect anomalies in spending/performance
    
    Uses z-score method to find outliers
    """
    df = df.copy()
    
    # Calculate z-scores for key metrics
    for col in ["cost", "roi", "ctr"]:
        if col in df.columns:
            mean = df[col].mean()
            std = df[col].std()
            if std > 0:
                df[f"{col}_zscore"] = (df[col] - mean) / std
            else:
                df[f"{col}_zscore"] = 0
    
    # Flag anomalies
    df["is_anomaly"] = False
    df["anomaly_type"] = None
    df["anomaly_description"] = None
    
    # High cost, low ROI anomaly
    high_cost_low_roi = (
        (df["cost_zscore"] > ANOMALY_THRESHOLD) & 
        (df["roi_zscore"] < -1)
    )
    df.loc[high_cost_low_roi, "is_anomaly"] = True
    df.loc[high_cost_low_roi, "anomaly_type"] = "high_cost_low_roi"
    df.loc[high_cost_low_roi, "anomaly_description"] = (
        "Pengeluaran tinggi dengan ROI rendah"
    )
    
    # Unusually high ROI (might be data error)
    very_high_roi = df["roi_zscore"] > ANOMALY_THRESHOLD
    df.loc[very_high_roi, "is_anomaly"] = True
    df.loc[very_high_roi, "anomaly_type"] = "unusually_high_roi"
    df.loc[very_high_roi, "anomaly_description"] = (
        "ROI sangat tinggi - perlu verifikasi data"
    )
    
    return df


def generate_insights(row: pd.Series) -> List[str]:
    """
    Generate actionable insights for a product/creative
    """
    insights = []
    
    roi = row.get("roi", 0)
    ctr = row.get("ctr", 0)
    conv = row.get("conversion_rate", 0)
    creative_type = row.get("creative_type", "Unknown")
    cost = row.get("cost", 0)
    
    # ROI insights
    if roi >= 10:
        insights.append("🎯 ROI sangat baik! Pertimbangkan scale up budget")
    elif roi >= 5:
        insights.append("✅ ROI bagus, performa stabil")
    elif roi >= 2:
        insights.append("⚠️ ROI cukup, ada ruang untuk optimasi")
    elif roi >= 1:
        insights.append("⚡ ROI rendah, perlu evaluasi target audience")
    else:
        insights.append("🚨 ROI negatif, pertimbangkan pause campaign")
    
    # CTR insights
    if ctr < 0.01 and cost > 100000:
        insights.append("📉 CTR rendah (<1%), coba ganti creative/thumbnail")
    elif ctr > 0.05:
        insights.append("👍 CTR tinggi, creative menarik perhatian")
    
    # Conversion insights
    if conv < 0.02 and ctr > 0.02:
        insights.append(
            "🔄 CTR bagus tapi konversi rendah - cek landing page/harga"
        )
    
    # Creative type insights
    if creative_type == "Kartu produk" and roi < 5:
        insights.append("💡 Pertimbangkan gunakan Video untuk ROI lebih baik")
    
    return insights


def score_creatives(tenant_id: str) -> List[Dict[str, Any]]:
    """
    Score all creatives for a tenant
    
    Args:
        tenant_id: Tenant ID
        
    Returns:
        List of scored creatives with recommendations
    """
    df = load_creative_data(tenant_id)
    
    if len(df) == 0:
        return []
    
    df = prepare_features(df)
    df = detect_anomalies(df)
    
    results = []
    
    for _, row in df.iterrows():
        creative_score = calculate_creative_score(row)
        performance_grade = calculate_performance_score(row["roi"])
        insights = generate_insights(row)
        
        results.append({
            "product_id": row["product_id"],
            "campaign_id": row.get("campaign_id", ""),
            "creative_type": row.get("creative_type", "Unknown"),
            "cost": float(row.get("cost", 0)),
            "revenue": float(row.get("revenue", 0)),
            "roi": float(row.get("roi", 0)),
            "ctr": float(row.get("ctr", 0)),
            "conversion_rate": float(row.get("conversion_rate", 0)),
            "creative_score": round(creative_score, 1),
            "performance_grade": performance_grade,
            "insights": insights,
            "is_anomaly": bool(row.get("is_anomaly", False)),
            "anomaly_type": row.get("anomaly_type"),
            "anomaly_description": row.get("anomaly_description"),
        })
    
    # Sort by creative score descending
    results.sort(key=lambda x: x["creative_score"], reverse=True)
    
    return results


def get_creative_type_comparison(tenant_id: str) -> Dict[str, Any]:
    """
    Compare Video vs Kartu Produk performance
    """
    df = load_creative_data(tenant_id)
    
    if len(df) == 0:
        return {"error": "No data"}
    
    df = prepare_features(df)
    
    comparison = {}
    
    for creative_type in ["Video", "Kartu produk"]:
        type_df = df[df["creative_type"] == creative_type]
        
        if len(type_df) == 0:
            continue
        
        comparison[creative_type] = {
            "count": len(type_df),
            "total_cost": float(type_df["cost"].sum()),
            "total_revenue": float(type_df["revenue"].sum()),
            "total_orders": int(type_df["orders"].sum()),
            "avg_roi": float(type_df["roi"].mean()),
            "avg_ctr": float(type_df["ctr"].mean()),
            "avg_conversion": float(type_df["conversion_rate"].mean()),
            "cost_per_order": float(
                type_df["cost"].sum() / max(1, type_df["orders"].sum())
            ),
        }
    
    # Recommendation
    if "Video" in comparison and "Kartu produk" in comparison:
        video_roi = comparison["Video"]["avg_roi"]
        card_roi = comparison["Kartu produk"]["avg_roi"]
        
        if video_roi > card_roi * 1.2:
            recommendation = "Video performa lebih baik, prioritaskan budget untuk Video"
        elif card_roi > video_roi * 1.2:
            recommendation = "Kartu Produk lebih efisien, cocok untuk budget terbatas"
        else:
            recommendation = "Performa seimbang, diversifikasi budget keduanya"
    else:
        recommendation = "Data tidak cukup untuk perbandingan"
    
    return {
        "comparison": comparison,
        "recommendation": recommendation,
        "generated_at": datetime.now().isoformat(),
    }


if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 3:
        print("Usage: python creative_scorer.py <score|compare> <tenant_id>")
        sys.exit(1)
    
    action = sys.argv[1]
    tenant_id = sys.argv[2]
    
    if action == "score":
        results = score_creatives(tenant_id)
        print(json.dumps(results[:20], indent=2, ensure_ascii=False))
    elif action == "compare":
        result = get_creative_type_comparison(tenant_id)
        print(json.dumps(result, indent=2, ensure_ascii=False))
    else:
        print(f"Unknown action: {action}")
        sys.exit(1)
