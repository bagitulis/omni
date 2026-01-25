#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads Analysis - Executive Summary Report
===============================================
Generate concise executive summary for stakeholders.
"""

from datetime import datetime
from typing import List, Dict, Any


def generate_executive_summary(products: List[Dict], evaluation: Dict, 
                               period_info: Dict = None) -> str:
    """Generate executive summary in Markdown format."""
    
    def fmt(n):
        return f"Rp {n:,.0f}".replace(",", ".")
    
    lines = []
    
    # Title
    if period_info and period_info.get("end"):
        end = period_info["end"]
        if "-W" in str(end):
            year = end.split("-")[0]
            week = int(end.split("-W")[1])
            quarter = f"Q{(week-1)//13 + 1}"
            period_str = f"{quarter} {year}"
        else:
            period_str = str(end)
        lines.append(f"# Shopee Ads Executive Summary - {period_str}")
    else:
        lines.append("# Shopee Ads Executive Summary")
    
    lines.append(f"\n**Generated:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
    
    # Status Badge
    lines.append(f"## {evaluation['status_emoji']} Status: {evaluation['status']}\n")
    
    # Key Metrics
    lines.append("## 💰 Key Metrics\n")
    lines.append(f"| Metric | Value |")
    lines.append(f"|--------|-------|")
    lines.append(f"| Total Investment | {fmt(evaluation['total_cost'])} |")
    lines.append(f"| Total Revenue | {fmt(evaluation['total_revenue'])} |")
    lines.append(f"| Total Profit | {fmt(evaluation['total_profit'])} |")
    lines.append(f"| Overall ROI | **{evaluation['roi']:.2f}x** |")
    lines.append(f"| Potential Savings | {fmt(evaluation['potential_savings'])} |\n")
    
    # Distribution
    lines.append("## 📊 Product Distribution\n")
    dist = evaluation["distribution"]
    total = sum(dist.values())
    lines.append(f"- ✅ **LANJUTKAN:** {dist['LANJUTKAN']} ({dist['LANJUTKAN']/total*100:.0f}%)")
    lines.append(f"- ⏸️ **PANTAU:** {dist['PANTAU']} ({dist['PANTAU']/total*100:.0f}%)")
    lines.append(f"- 🛑 **HENTIKAN:** {dist['HENTIKAN']} ({dist['HENTIKAN']/total*100:.0f}%)\n")
    
    # Top 5 Products
    top5 = [p for p in products if "LANJUTKAN" in p["category"]][:5]
    if top5:
        lines.append("## ⭐ Top 5 Products to Scale Up\n")
        lines.append("| Product | ROI | Score | Action |")
        lines.append("|---------|-----|-------|--------|")
        for p in top5:
            lines.append(f"| {p['product_name'][:35]} | {p['roi']:.2f}x | {p['score']:.0f} | Tambah budget 30-50% |")
        lines.append("")
    
    # Top 5 to Stop
    stop5 = [p for p in products if "HENTIKAN" in p["category"]][:5]
    if stop5:
        lines.append("## 🛑 Top 5 Products to Stop\n")
        lines.append("| Product | Loss | ROI | Action |")
        lines.append("|---------|------|-----|--------|")
        for p in stop5:
            loss = abs(p["profit"]) if p["profit"] < 0 else 0
            lines.append(f"| {p['product_name'][:35]} | {fmt(loss)} | {p['roi']:.2f}x | STOP segera |")
        lines.append("")
    
    # Quick Actions
    lines.append("## 🎯 Quick Actions\n")
    lines.append(f"1. **Scale up** {len(top5)} high-performing products (+30-50% budget)")
    lines.append(f"2. **Stop** {len(stop5)} underperforming products (save {fmt(evaluation['potential_savings'])})")
    lines.append(f"3. **Monitor** {dist['PANTAU']} products for optimization opportunities\n")
    
    return "\n".join(lines)


def get_high_impact_products(products: List[Dict], n: int = 5) -> List[Dict]:
    """Get products with highest potential impact for scaling."""
    # Filter profitable products with good momentum
    candidates = [
        p for p in products 
        if p["profit"] > 0 and p["roi"] >= 2.0 and "LANJUTKAN" in p["category"]
    ]
    # Sort by score
    return sorted(candidates, key=lambda x: x["score"], reverse=True)[:n]


def get_all_stop_products_categorized(products: List[Dict]) -> Dict[str, List[Dict]]:
    """Get all products to stop, categorized by reason."""
    stop_products = [p for p in products if "HENTIKAN" in p["category"]]
    
    categories = {
        "high_loss": [],
        "low_roi": [],
        "negative_trend": [],
        "data_limited": [],
    }
    
    for p in stop_products:
        if p["profit"] < -100000:
            categories["high_loss"].append(p)
        elif p["roi"] < 1.0:
            categories["low_roi"].append(p)
        elif "Turun" in p.get("trend", ""):
            categories["negative_trend"].append(p)
        else:
            categories["data_limited"].append(p)
    
    return categories
