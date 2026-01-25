"""
Test script for API sync endpoints
Tests the product manager sync buttons and analytics endpoints

Run with: python tests/test_api_sync.py
"""
import requests
import json
import os
from datetime import datetime

# Configuration
BASE_URL = os.getenv("API_BASE_URL", "https://yndigital.my.id/api")
# For local testing:
# BASE_URL = "http://localhost:3000/api"

# Test credentials
USERNAME = os.getenv("TEST_USERNAME", "yumna")
PASSWORD = os.getenv("TEST_PASSWORD", "password123")
TENANT_ID = os.getenv("TEST_TENANT_ID", "yumna_bertigamart")

def get_auth_token():
    """Login and get JWT token"""
    print("\n🔐 Logging in...")
    try:
        resp = requests.post(
            f"{BASE_URL}/auth/login",
            json={"username": USERNAME, "password": PASSWORD},
            timeout=30
        )
        if resp.status_code == 200:
            data = resp.json()
            # Handle different response formats
            token = (
                data.get("token") or  # Direct token field
                data.get("data", {}).get("accessToken") or  # Nested in data.accessToken
                data.get("accessToken")  # Direct accessToken field
            )
            if token:
                print(f"   ✅ Login successful")
                return token
            print(f"   ❌ Login succeeded but no token found in response")
            print(f"   Response keys: {list(data.keys())}")
        else:
            print(f"   ❌ Login failed: {resp.status_code} - {resp.text[:200]}")
        return None
    except Exception as e:
        print(f"   ❌ Login error: {e}")
        return None

def get_auth_headers(token):
    """Create headers with auth token"""
    return {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {token}",
        "x-tenant-id": TENANT_ID
    }

def test_health():
    """Test health endpoint"""
    print("\n📋 Testing /api/health...")
    try:
        resp = requests.get(f"{BASE_URL}/health", timeout=10)
        print(f"   Status: {resp.status_code}")
        if resp.status_code == 200:
            print("   ✅ Health check passed")
            return True
        else:
            print(f"   ❌ Health check failed: {resp.text}")
            return False
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False

def test_shopee_products(headers):
    """Test Shopee products sync endpoint"""
    print("\n📋 Testing GET /api/shopee/products...")
    try:
        resp = requests.get(
            f"{BASE_URL}/shopee/products",
            headers=headers,
            params={"offset": 0, "limit": 10},
            timeout=60
        )
        print(f"   Status: {resp.status_code}")
        if resp.status_code == 200:
            data = resp.json()
            if data.get("success"):
                print(f"   ✅ Retrieved {len(data.get('products', []))} products")
                return True
            else:
                print(f"   ❌ API error: {data.get('error')}")
        else:
            print(f"   ❌ Error: {resp.text[:200]}")
        return False
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False

def test_lazada_products(headers):
    """Test Lazada products sync endpoint"""
    print("\n📋 Testing GET /api/lazada/products...")
    try:
        resp = requests.get(
            f"{BASE_URL}/lazada/products",
            headers=headers,
            params={"offset": 0, "limit": 10},
            timeout=60
        )
        print(f"   Status: {resp.status_code}")
        if resp.status_code == 200:
            data = resp.json()
            if data.get("success"):
                print(f"   ✅ Retrieved {len(data.get('products', []))} products")
                return True
            else:
                print(f"   ❌ API error: {data.get('error')}")
        else:
            print(f"   ❌ Error: {resp.text[:200]}")
        return False
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False

def test_tiktok_products_search(headers):
    """Test TikTok products search/sync endpoint"""
    print("\n📋 Testing POST /api/tiktok/products/search...")
    try:
        resp = requests.post(
            f"{BASE_URL}/tiktok/products/search",
            headers=headers,
            json={"status": "ACTIVATE", "page_size": 10, "sync_to_db": True},
            timeout=60
        )
        print(f"   Status: {resp.status_code}")
        if resp.status_code == 200:
            data = resp.json()
            if data.get("success"):
                print(f"   ✅ Retrieved {data.get('total', 0)} products")
                return True
            else:
                print(f"   ❌ API error: {data.get('error')}")
        else:
            print(f"   ❌ Error: {resp.text[:200]}")
        return False
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False

def test_analytics_shopee_sync(headers):
    """Test Shopee analytics sync endpoint"""
    print("\n📋 Testing POST /api/analytics/shopee/sync...")
    try:
        resp = requests.post(
            f"{BASE_URL}/analytics/shopee/sync",
            headers=headers,
            json={"month": 12, "year": 2025, "forceResync": False},
            timeout=60
        )
        print(f"   Status: {resp.status_code}")
        if resp.status_code == 200:
            data = resp.json()
            print(f"   ✅ Response: {data.get('message', 'OK')}")
            return True
        elif resp.status_code == 404:
            print("   ❌ Endpoint not found - route not registered")
        else:
            print(f"   ❌ Error: {resp.text[:200]}")
        return False
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False

def test_analytics_tiktok_sync(headers):
    """Test TikTok analytics sync endpoint"""
    print("\n📋 Testing POST /api/analytics/tiktok/sync...")
    try:
        resp = requests.post(
            f"{BASE_URL}/analytics/tiktok/sync",
            headers=headers,
            json={"month": 12, "year": 2025, "forceResync": False},
            timeout=60
        )
        print(f"   Status: {resp.status_code}")
        if resp.status_code == 200:
            data = resp.json()
            print(f"   ✅ Response: {data.get('message', 'OK')}")
            return True
        elif resp.status_code == 404:
            print("   ❌ Endpoint not found - route not registered")
        else:
            print(f"   ❌ Error: {resp.text[:200]}")
        return False
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False

def test_analytics_sync_status(headers):
    """Test analytics sync status endpoints"""
    print("\n📋 Testing GET /api/analytics/shopee/sync-status...")
    try:
        resp = requests.get(
            f"{BASE_URL}/analytics/shopee/sync-status",
            headers=headers,
            timeout=30
        )
        print(f"   Shopee Status: {resp.status_code}")
        
        resp2 = requests.get(
            f"{BASE_URL}/analytics/tiktok/sync-status",
            headers=headers,
            timeout=30
        )
        print(f"   TikTok Status: {resp2.status_code}")
        
        if resp.status_code == 200 and resp2.status_code == 200:
            print("   ✅ Both sync-status endpoints working")
            return True
        return False
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False

def main():
    """Run all tests"""
    print("=" * 60)
    print(f" API Sync Test - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f" Base URL: {BASE_URL}")
    print("=" * 60)
    
    # First, get auth token
    token = get_auth_token()
    if not token:
        print("\n❌ Failed to authenticate. Cannot proceed with tests.")
        return 1
    
    headers = get_auth_headers(token)
    
    results = {
        "health": test_health(),
        "shopee_products": test_shopee_products(headers),
        "lazada_products": test_lazada_products(headers),
        "tiktok_products": test_tiktok_products_search(headers),
        "analytics_shopee_sync": test_analytics_shopee_sync(headers),
        "analytics_tiktok_sync": test_analytics_tiktok_sync(headers),
        "analytics_sync_status": test_analytics_sync_status(headers),
    }
    
    print("\n" + "=" * 60)
    print(" SUMMARY")
    print("=" * 60)
    
    passed = sum(1 for v in results.values() if v)
    total = len(results)
    
    for name, result in results.items():
        status = "✅ PASS" if result else "❌ FAIL"
        print(f"   {name}: {status}")
    
    print(f"\n   Total: {passed}/{total} tests passed")
    print("=" * 60)
    
    return 0 if passed == total else 1

if __name__ == "__main__":
    exit(main())
