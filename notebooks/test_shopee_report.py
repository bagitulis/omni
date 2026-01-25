#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Test Shopee Ads Report Generation
"""
import sys
sys.path.insert(0, ".")

from shopee_ads.reports import run_full_analysis

result = run_full_analysis(verbose=True)
evaluation = result["evaluation"]

print()
print("=" * 60)
print("SHOPEE ADS REPORT SUMMARY")
print("=" * 60)
print(f"Status: {evaluation['status_emoji']} {evaluation['status']}")
print(f"Total Cost: Rp {evaluation['total_cost']:,.0f}".replace(",", "."))
print(f"Total Revenue: Rp {evaluation['total_revenue']:,.0f}".replace(",", "."))
print(f"Total Profit: Rp {evaluation['total_profit']:,.0f}".replace(",", "."))
print(f"ROI: {evaluation['roi']:.2f}x")
print(f"Products: {evaluation['total_products']}")
print()
print("Distribution:")
for k, v in evaluation["distribution"].items():
    print(f"  {k}: {v}")
print()
print(f"Full Report: {result['full_report']}")
print("=" * 60)
