#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""Compare TikTok and Shopee Ads Reports - for analysis."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

print("="*60)
print("GENERATING SHOPEE ADS REPORT")
print("="*60)
from shopee_ads.reports import run_quarterly_analysis as shopee_run
shopee_result = shopee_run(verbose=True)

print("\n" + "="*60)
print("GENERATING TIKTOK ADS REPORT")
print("="*60)
from tiktok_ads.reports import run_quarterly_analysis as tiktok_run
tiktok_result = tiktok_run(verbose=True)

print("\n" + "="*60)
print("COMPARISON SUMMARY")
print("="*60)

# Shopee
print("\n📊 SHOPEE ADS:")
print(f"   Full Report: {shopee_result.get('full_report', 'N/A')}")
print(f"   Executive:   {shopee_result.get('executive_summary', 'N/A')}")
shopee_eval = shopee_result.get('evaluation', {})
print(f"   Health Score: {shopee_eval.get('health', {}).get('score', 0)}/100")
print(f"   Risk Level:   {shopee_eval.get('risk', {}).get('level', 'N/A')}")
print(f"   ROI:          {shopee_eval.get('roi', 0):.2f}x")

# TikTok
print("\n🎵 TIKTOK ADS:")
print(f"   Full Report: {tiktok_result.get('html_path', 'N/A')}")
print(f"   Executive:   {tiktok_result.get('executive_html_path', 'N/A')}")

# Read both HTML files and count lines
shopee_html_path = shopee_result.get('full_report')
tiktok_html_path = tiktok_result.get('html_path')

if shopee_html_path:
    with open(shopee_html_path, 'r', encoding='utf-8') as f:
        shopee_lines = len(f.readlines())
    print(f"\n📄 Shopee HTML Lines: {shopee_lines}")
    
if tiktok_html_path:
    with open(tiktok_html_path, 'r', encoding='utf-8') as f:
        tiktok_lines = len(f.readlines())
    print(f"📄 TikTok HTML Lines: {tiktok_lines}")

print("\n✅ Reports generated successfully!")
print(f"   Open the HTML files to compare visually.")
