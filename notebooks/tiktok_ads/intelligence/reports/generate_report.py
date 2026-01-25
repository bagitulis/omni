#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Report Generator
================
Generates comprehensive TikTok Ads Intelligence Report.
"""

import sys
from pathlib import Path
from datetime import datetime

# Add parent to path
sys.path.insert(0, str(Path(__file__).parent.parent))

from intelligence import IntelligenceEngine


def generate_report(db_path: str, min_cost: float = 50000):
    """Generate full intelligence report."""
    engine = IntelligenceEngine(dbPath=db_path)
    
    results = engine.analyzeAll(minCost=min_cost, verbose=True)
    df = engine.toDataFrame(results)
    df = df.sort_values('compositeScore', ascending=False)
    
    now = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    
    print()
    print('=' * 80)
    print('  TIKTOK ADS INTELLIGENCE REPORT')
    print(f'  Generated: {now}')
    print('=' * 80)
    
    # Summary Statistics
    print()
    print('📊 SUMMARY STATISTICS')
    print('-' * 40)
    print(f'  Total Products Analyzed: {len(df)}')
    print(f'  Total Cost: Rp {df["totalCost"].sum():,.0f}')
    print(f'  Total Revenue: Rp {df["totalRevenue"].sum():,.0f}')
    print(f'  Total Profit: Rp {df["totalProfit"].sum():,.0f}')
    total_cost = max(df["totalCost"].sum(), 1)
    overall_roas = df["totalRevenue"].sum() / total_cost
    print(f'  Overall ROAS: {overall_roas:.2f}x')
    
    # Action Distribution
    print()
    print('📌 ACTION DISTRIBUTION')
    print('-' * 40)
    action_counts = df['action'].value_counts()
    for action, count in action_counts.items():
        print(f'  {action}: {count} products')
    
    # Lifecycle Distribution
    print()
    print('🔄 LIFECYCLE STAGE DISTRIBUTION')
    print('-' * 40)
    lifecycle_counts = df['lifecycleStage'].value_counts()
    for stage, count in lifecycle_counts.items():
        print(f'  {stage}: {count} products')
    
    # Fatigue Distribution
    print()
    print('😴 FATIGUE STATUS DISTRIBUTION')
    print('-' * 40)
    fatigue_counts = df['fatigueStatus'].value_counts()
    for status, count in fatigue_counts.items():
        print(f'  {status}: {count} products')
    
    # Top Performers
    print()
    print('🌟 TOP 5 PERFORMERS (Highest Composite Score)')
    print('-' * 40)
    top5 = df.nlargest(5, 'compositeScore')
    for i, (_, row) in enumerate(top5.iterrows(), 1):
        cost = max(row['totalCost'], 1)
        roas = row['totalRevenue'] / cost
        name = row['productName'][:50]
        print(f'{i}. {name}')
        print(f'   Score: {row["compositeScore"]:.1f} | ROAS: {roas:.2f}x | {row["action"]}')
    
    # Products to Scale
    print()
    print('🚀 PRODUCTS TO SCALE (SCALE_UP actions)')
    print('-' * 40)
    scale_up = df[df['action'].str.contains('SCALE_UP', na=False)]
    if len(scale_up) > 0:
        for i, (_, row) in enumerate(scale_up.iterrows(), 1):
            name = row['productName'][:50]
            print(f'{i}. {name}')
            print(f'   Budget: Rp {row["currentBudget"]:,.0f} -> Rp {row["recommendedBudget"]:,.0f}')
    else:
        print('   No products recommended for scale up at this time')
    
    # Products to Stop
    print()
    print('🛑 PRODUCTS TO STOP (Score < 30 or ROAS < 1)')
    print('-' * 40)
    df_with_roas = df.copy()
    df_with_roas['roas'] = df_with_roas['totalRevenue'] / df_with_roas['totalCost'].clip(lower=1)
    stop_candidates = df_with_roas[
        (df_with_roas['compositeScore'] < 30) | (df_with_roas['roas'] < 1)
    ]
    if len(stop_candidates) > 0:
        for i, (_, row) in enumerate(stop_candidates.head(10).iterrows(), 1):
            name = row['productName'][:50]
            print(f'{i}. {name}')
            print(f'   Score: {row["compositeScore"]:.1f} | ROAS: {row["roas"]:.2f}x | Cost: Rp {row["totalCost"]:,.0f}')
    else:
        print('   No products recommended for stopping')
    
    # Fatigued Products
    print()
    print('😵 FATIGUED PRODUCTS (Need Creative Refresh)')
    print('-' * 40)
    fatigued = df[df['fatigueStatus'] == 'FATIGUED']
    if len(fatigued) > 0:
        for i, (_, row) in enumerate(fatigued.iterrows(), 1):
            name = row['productName'][:50]
            print(f'{i}. {name}')
            print(f'   Fatigue Score: {row["fatigueScore"]:.1f}% | Stage: {row["lifecycleStage"]}')
    else:
        print('   No fatigued products found')
    
    # Budget Recommendations Summary
    print()
    print('💰 BUDGET RECOMMENDATIONS SUMMARY')
    print('-' * 40)
    total_current = df['currentBudget'].sum()
    total_recommended = df['recommendedBudget'].sum()
    change = total_recommended - total_current
    change_pct = (change / max(total_current, 1)) * 100
    print(f'  Current Total Budget: Rp {total_current:,.0f}')
    print(f'  Recommended Total Budget: Rp {total_recommended:,.0f}')
    print(f'  Net Change: Rp {change:+,.0f} ({change_pct:+.1f}%)')
    
    print()
    print('=' * 80)
    print('  END OF REPORT')
    print('=' * 80)
    
    return df


if __name__ == '__main__':
    # Default database path
    db_path = r'c:\Users\yumna\Desktop\Project\omni\backend\config\databases\yumna_bertigamart.db'
    
    # Allow custom path from command line
    if len(sys.argv) > 1:
        db_path = sys.argv[1]
    
    df = generate_report(db_path)
