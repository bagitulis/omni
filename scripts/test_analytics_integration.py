#!/usr/bin/env python3
"""
Analytics Integration Test Script
Tests all API endpoints for the Analytics redesign Phase 3
"""
import requests
import json
import sys
import re
import os

# Set environment for UTF-8
os.environ["PYTHONIOENCODING"] = "utf-8"

BASE_URL = "http://localhost:3000"


def clean_text(text):
    """Remove emojis and non-ASCII characters for safe printing"""
    if not text:
        return "N/A"
    # Remove emojis and special unicode characters
    return re.sub(r'[^\x00-\x7F]+', '', str(text)).strip() or "(emoji)"

def login():
    """Login and get auth token"""
    print("=" * 60)
    print("STEP 1: LOGIN")
    print("=" * 60)
    
    resp = requests.post(f"{BASE_URL}/api/auth/login", json={
        "username": "yumna",
        "password": "password123"
    })
    
    if resp.status_code != 200:
        print(f"FAIL: Login failed with status {resp.status_code}")
        return None
    
    data = resp.json()
    if not data.get("success"):
        print(f"FAIL: Login unsuccessful - {data}")
        return None
    
    token = data.get("token")
    print(f"OK: Login successful, tenant: {data.get('tenant_id')}")
    return token


def test_unified_summary(token):
    """Test unified summary API"""
    print("\n" + "=" * 60)
    print("STEP 2: TEST UNIFIED SUMMARY")
    print("=" * 60)
    
    headers = {"Authorization": f"Bearer {token}"}
    resp = requests.get(f"{BASE_URL}/api/analytics/unified/summary", headers=headers)
    
    if resp.status_code != 200:
        print(f"FAIL: Status {resp.status_code}")
        return False
    
    data = resp.json()
    if not data.get("success"):
        print(f"FAIL: {data}")
        return False
    
    summary = data.get("data", {})
    combined = summary.get("combined", {})
    
    print(f"OK: Combined ROAS: {combined.get('overall_roas', 0):.2f}x")
    print(f"    Total Revenue: Rp {combined.get('total_revenue', 0):,.0f}")
    print(f"    Total Cost: Rp {combined.get('total_cost', 0):,.0f}")
    print(f"    Total Orders: {combined.get('total_orders', 0):,}")
    
    return True


def test_unified_kpi(token):
    """Test unified KPI API"""
    print("\n" + "=" * 60)
    print("STEP 3: TEST UNIFIED KPI")
    print("=" * 60)
    
    headers = {"Authorization": f"Bearer {token}"}
    resp = requests.get(f"{BASE_URL}/api/analytics/unified/kpi", headers=headers)
    
    if resp.status_code != 200:
        print(f"FAIL: Status {resp.status_code}")
        return False
    
    data = resp.json()
    kpi = data.get("data", {})
    actions = kpi.get("actions", {})
    
    print(f"OK: Total Products: {kpi.get('total_products', 0)}")
    print(f"    Average ROAS: {kpi.get('avg_roas', 0):.2f}x")
    print(f"    Actions:")
    print(f"      Scale Up: {actions.get('scale_up', 0)}")
    print(f"      Maintain: {actions.get('maintain', 0)}")
    print(f"      Reduce: {actions.get('reduce', 0)}")
    print(f"      Stop: {actions.get('stop', 0)}")
    
    return True


def test_products_from_ads(token):
    """Test products from ads API"""
    print("\n" + "=" * 60)
    print("STEP 4: TEST PRODUCTS FROM ADS")
    print("=" * 60)
    
    headers = {"Authorization": f"Bearer {token}"}
    resp = requests.get(f"{BASE_URL}/api/analytics/products/from-ads", headers=headers)
    
    if resp.status_code != 200:
        print(f"FAIL: Status {resp.status_code}")
        return False, None
    
    data = resp.json()
    products = data.get("data", [])
    count = data.get("count", len(products))
    
    print(f"OK: Found {count} products from ads database")
    
    if products:
        p = products[0]
        print(f"    Sample: {clean_text(p.get('product_name', 'N/A'))[:50]}...")
        print(f"    ROAS: {p.get('avg_roas', 0):.2f}x, Source: {p.get('source')}")
        return True, p.get('product_id')
    
    return True, None


def test_products_classified(token):
    """Test classified products API"""
    print("\n" + "=" * 60)
    print("STEP 5: TEST CLASSIFIED PRODUCTS")
    print("=" * 60)
    
    headers = {"Authorization": f"Bearer {token}"}
    resp = requests.get(f"{BASE_URL}/api/analytics/products/classified", headers=headers)
    
    if resp.status_code != 200:
        print(f"FAIL: Status {resp.status_code}")
        return False
    
    data = resp.json()
    products = data.get("data", {})
    counts = data.get("counts", {})
    
    print(f"OK: Classification counts:")
    print(f"    Scale Up: {counts.get('scale_up', 0)}")
    print(f"    Maintain: {counts.get('maintain', 0)}")
    print(f"    Reduce: {counts.get('reduce', 0)}")
    print(f"    Stop: {counts.get('stop', 0)}")
    
    # Check if we have products with recommendations
    scale_up = products.get("scale_up", [])
    if scale_up:
        p = scale_up[0]
        print(f"\n    Sample Scale Up Product:")
        print(f"      Name: {clean_text(p.get('product_name', 'N/A'))[:40]}...")
        print(f"      ROAS: {p.get('roas', 0):.2f}x")
        print(f"      Recommendation: {clean_text(p.get('recommendation', 'N/A'))[:50]}...")
    
    return True


def test_simulation(token, product_id):
    """Test budget simulation API"""
    print("\n" + "=" * 60)
    print("STEP 6: TEST BUDGET SIMULATION")
    print("=" * 60)
    
    if not product_id:
        print("SKIP: No product_id available")
        return True
    
    headers = {"Authorization": f"Bearer {token}"}
    payload = {
        "product_id": product_id,
        "target_roas": 8.0,
        "budget_per_day": 200000,
        "period_days": 7
    }
    
    resp = requests.post(
        f"{BASE_URL}/api/analytics/simulation/calculate",
        headers=headers,
        json=payload
    )
    
    if resp.status_code != 200:
        print(f"FAIL: Status {resp.status_code}")
        return False
    
    data = resp.json()
    result = data.get("data", {})
    
    print(f"OK: Simulation Result:")
    print(f"    Feasibility: {result.get('feasibility')}")
    print(f"    Confidence: {result.get('confidence_percent')}%")
    print(f"    Recommendation: {clean_text(result.get('recommendation', 'N/A'))[:60]}...")
    
    return True


def test_calendar(token):
    """Test calendar events API"""
    print("\n" + "=" * 60)
    print("STEP 7: TEST CALENDAR EVENTS")
    print("=" * 60)
    
    headers = {"Authorization": f"Bearer {token}"}
    resp = requests.get(
        f"{BASE_URL}/api/analytics/intelligence/calendar?days=14",
        headers=headers
    )
    
    if resp.status_code != 200:
        print(f"FAIL: Status {resp.status_code}")
        return False
    
    data = resp.json()
    events_data = data.get("data", {})
    events = events_data.get("events", [])
    
    print(f"OK: Found {len(events)} calendar events")
    for e in events[:3]:
        print(f"    {e.get('date')}: {e.get('description')} (x{e.get('multiplier')})")
    
    return True


def test_cache_status(token):
    """Test cache status API"""
    print("\n" + "=" * 60)
    print("STEP 8: TEST CACHE STATUS")
    print("=" * 60)
    
    headers = {"Authorization": f"Bearer {token}"}
    resp = requests.get(f"{BASE_URL}/api/analytics/cache/status", headers=headers)
    
    if resp.status_code != 200:
        print(f"FAIL: Status {resp.status_code}")
        return False
    
    data = resp.json()
    print(f"OK: Cache status retrieved")
    return True


def main():
    print("\n" + "=" * 60)
    print("ANALYTICS INTEGRATION TEST")
    print("=" * 60)
    
    token = login()
    if not token:
        print("\nFAIL: Cannot proceed without auth token")
        sys.exit(1)
    
    results = []
    
    results.append(("Unified Summary", test_unified_summary(token)))
    results.append(("Unified KPI", test_unified_kpi(token)))
    
    success, product_id = test_products_from_ads(token)
    results.append(("Products from Ads", success))
    
    results.append(("Classified Products", test_products_classified(token)))
    results.append(("Budget Simulation", test_simulation(token, product_id)))
    results.append(("Calendar Events", test_calendar(token)))
    results.append(("Cache Status", test_cache_status(token)))
    
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    
    passed = sum(1 for _, r in results if r)
    total = len(results)
    
    for name, result in results:
        status = "PASS" if result else "FAIL"
        print(f"  [{status}] {name}")
    
    print(f"\nTotal: {passed}/{total} tests passed")
    
    if passed == total:
        print("\n*** ALL TESTS PASSED - SUCCESS = TRUE ***\n")
        return 0
    else:
        print("\n*** SOME TESTS FAILED ***\n")
        return 1


if __name__ == "__main__":
    sys.exit(main())
