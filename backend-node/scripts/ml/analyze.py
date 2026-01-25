"""
TikTok Ads ML - Main Entry Point
Unified CLI for all ML operations
"""

import json
import sys
import argparse
from datetime import datetime
from typing import Optional

from roi_predictor import train_roi_model, predict_roi
from creative_scorer import score_creatives, get_creative_type_comparison


def run_full_analysis(tenant_id: str) -> dict:
    """
    Run complete ML analysis for a tenant
    
    Returns comprehensive analysis including:
    - ROI predictions
    - Creative scores
    - Type comparison
    - Top recommendations
    """
    results = {
        "tenant_id": tenant_id,
        "generated_at": datetime.now().isoformat(),
        "analyses": {},
    }
    
    # 1. Train/update ROI model
    print(f"Training ROI model for {tenant_id}...")
    roi_training = train_roi_model(tenant_id)
    results["analyses"]["roi_model"] = roi_training
    
    # 2. Get ROI predictions
    if roi_training.get("success"):
        print("Generating ROI predictions...")
        predictions = predict_roi(tenant_id)
        results["analyses"]["roi_predictions"] = predictions[:20]  # Top 20
    
    # 3. Score creatives
    print("Scoring creatives...")
    scores = score_creatives(tenant_id)
    results["analyses"]["creative_scores"] = {
        "top_performers": scores[:10],
        "underperformers": scores[-10:] if len(scores) > 10 else [],
        "total_analyzed": len(scores),
    }
    
    # 4. Creative type comparison
    print("Comparing creative types...")
    comparison = get_creative_type_comparison(tenant_id)
    results["analyses"]["creative_comparison"] = comparison
    
    # 5. Generate summary recommendations
    recommendations = generate_summary_recommendations(scores, comparison)
    results["recommendations"] = recommendations
    
    return results


def generate_summary_recommendations(
    scores: list,
    comparison: dict
) -> list:
    """
    Generate top-level recommendations based on analysis
    """
    recommendations = []
    
    # Analyze scores
    if scores:
        avg_score = sum(s["creative_score"] for s in scores) / len(scores)
        
        if avg_score < 30:
            recommendations.append({
                "priority": "high",
                "category": "overall_performance",
                "message": "Performa keseluruhan perlu perbaikan signifikan",
                "action": "Review strategi targeting dan creative"
            })
        
        # Count anomalies
        anomalies = [s for s in scores if s.get("is_anomaly")]
        if len(anomalies) > 5:
            recommendations.append({
                "priority": "medium",
                "category": "anomaly",
                "message": f"Ditemukan {len(anomalies)} anomali yang perlu dicek",
                "action": "Review data dengan anomali untuk validasi"
            })
    
    # Analyze creative type comparison
    comp = comparison.get("comparison", {})
    if "Video" in comp and "Kartu produk" in comp:
        video_roi = comp["Video"]["avg_roi"]
        card_roi = comp["Kartu produk"]["avg_roi"]
        
        if video_roi > card_roi * 2:
            recommendations.append({
                "priority": "high",
                "category": "creative_type",
                "message": f"Video 2x lebih efektif (ROI {video_roi:.1f} vs {card_roi:.1f})",
                "action": "Alokasikan lebih banyak budget ke Video"
            })
    
    return recommendations


def main():
    parser = argparse.ArgumentParser(
        description="TikTok Ads ML Analysis Tool"
    )
    parser.add_argument(
        "action",
        choices=["train", "predict", "score", "compare", "full"],
        help="Action to perform"
    )
    parser.add_argument(
        "tenant_id",
        help="Tenant ID to analyze"
    )
    parser.add_argument(
        "--product-ids",
        nargs="*",
        help="Specific product IDs for prediction"
    )
    parser.add_argument(
        "--output",
        "-o",
        help="Output file path (default: stdout)"
    )
    
    args = parser.parse_args()
    
    result = None
    
    if args.action == "train":
        result = train_roi_model(args.tenant_id)
    elif args.action == "predict":
        result = predict_roi(args.tenant_id, args.product_ids)
    elif args.action == "score":
        result = score_creatives(args.tenant_id)
    elif args.action == "compare":
        result = get_creative_type_comparison(args.tenant_id)
    elif args.action == "full":
        result = run_full_analysis(args.tenant_id)
    
    output = json.dumps(result, indent=2, ensure_ascii=False)
    
    if args.output:
        with open(args.output, "w", encoding="utf-8") as f:
            f.write(output)
        print(f"Output written to {args.output}")
    else:
        print(output)


if __name__ == "__main__":
    main()
