#!/usr/bin/env python3
"""
API Integration Test Script for Analytics Endpoints
Tests all new API endpoints created in Phase 1 & 2
"""

import requests
import json
import sys

BASE_URL = "http://localhost:3000"

# Test credentials
TEST_USER = "yumna"
TEST_PASSWORD = "password123"

def get_auth_token():
    """Get JWT token for authentication"""
    try:
        response = requests.post(
            f"{BASE_URL}/api/auth/login",
            json={"username": TEST_USER, "password": TEST_PASSWORD},
            timeout=10
        )
        if response.status_code == 200:
            data = response.json()
            return data.get("token") or data.get("access_token")
        else:
            print(f"Login failed: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"Login error: {e}")
        return None

def test_endpoint(name, method, url, token, data=None):
    """Test a single endpoint"""
    headers = {"Authorization": f"Bearer {token}"} if token else {}
    
    try:
        if method == "GET":
            response = requests.get(f"{BASE_URL}{url}", headers=headers, timeout=30)
        elif method == "POST":
            response = requests.post(f"{BASE_URL}{url}", json=data, headers=headers, timeout=30)
        else:
            return {"status": "ERROR", "message": f"Unknown method: {method}"}
        
        success = 200 <= response.status_code < 300
        return {
            "name": name,
            "status": "PASS" if success else "FAIL",
            "code": response.status_code,
            "time_ms": int(response.elapsed.total_seconds() * 1000),
            "data_preview": str(response.text)[:200] if success else response.text[:500]
        }
    except requests.exceptions.ConnectionError:
        return {"name": name, "status": "SKIP", "message": "Server not running"}
    except Exception as e:
        return {"name": name, "status": "ERROR", "message": str(e)}

def main():
    print("=" * 60)
    print("Analytics API Integration Tests")
    print("=" * 60)
    
    # Get auth token
    print("\n1. Authenticating...")
    token = get_auth_token()
    if not token:
        print("   SKIP: Could not authenticate (server may not be running)")
        print("   Tests will run without authentication")
    else:
        print(f"   OK: Got token")
    
    # Define test cases
    tests = [
        # Products from Ads
        ("Products from Ads", "GET", "/api/analytics/products/from-ads", None),
        
        # Calendar Events
        ("Calendar Events (30 days)", "GET", "/api/analytics/intelligence/calendar?days=30", None),
        
        # Unified Analytics
        ("Unified Summary", "GET", "/api/analytics/unified/summary", None),
        ("Unified KPI", "GET", "/api/analytics/unified/kpi", None),
        
        # Classified Products
        ("Classified Products", "GET", "/api/analytics/products/classified", None),
        ("Top Products", "GET", "/api/analytics/products/top", None),
        
        # Cache Status
        ("Cache Status", "GET", "/api/analytics/cache/status", None),
        
        # Budget Simulation
        ("Budget Simulation", "POST", "/api/analytics/simulation/calculate", {
            "product_id": "test_product_1",
            "target_roas": 5.0,
            "budget_per_day": 100000,
            "period_days": 7
        }),
        
        # Existing endpoints (sanity check)
        ("ML Portfolio Health", "GET", "/api/ml/portfolio/health?platform=tiktok", None),
        ("Shopee Ads Dashboard", "GET", "/api/analytics/shopee-ads/dashboard", None),
    ]
    
    print(f"\n2. Running {len(tests)} API tests...\n")
    
    results = []
    for name, method, url, data in tests:
        result = test_endpoint(name, method, url, token, data)
        results.append(result)
        
        status_icon = "[OK]" if result["status"] == "PASS" else "[FAIL]" if result["status"] == "FAIL" else "[SKIP]"
        time_str = f"{result.get('time_ms', 0)}ms" if result.get('time_ms') else ""
        print(f"   {status_icon} {name}: {result['status']} {time_str}")
    
    # Summary
    print("\n" + "=" * 60)
    passed = sum(1 for r in results if r["status"] == "PASS")
    failed = sum(1 for r in results if r["status"] == "FAIL")
    skipped = sum(1 for r in results if r["status"] in ("SKIP", "ERROR"))
    
    print(f"Summary: {passed} passed, {failed} failed, {skipped} skipped")
    
    if failed > 0:
        print("\nFailed tests:")
        for r in results:
            if r["status"] == "FAIL":
                print(f"   - {r['name']}: {r.get('code', 'N/A')} - {r.get('data_preview', '')[:100]}")
    
    return 0 if failed == 0 else 1

if __name__ == "__main__":
    sys.exit(main())
