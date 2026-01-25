#!/usr/bin/env python3
"""
Comprehensive API Testing Script
Tests login, tenant isolation, filter preferences, and data loading
"""

import requests
import json
import sqlite3
import os
from pathlib import Path

BASE_URL = "http://localhost:3000/api"
SQLITE_PATH = Path(__file__).parent.parent / "backend-node" / "data" / "yumna_bertigamart.db"

class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    RESET = '\033[0m'

def log_success(msg):
    print(f"{Colors.GREEN}✅ {msg}{Colors.RESET}")

def log_error(msg):
    print(f"{Colors.RED}❌ {msg}{Colors.RESET}")

def log_info(msg):
    print(f"{Colors.BLUE}ℹ️  {msg}{Colors.RESET}")

def log_warning(msg):
    print(f"{Colors.YELLOW}⚠️  {msg}{Colors.RESET}")


def check_sqlite_database():
    """Check SQLite database for existing data"""
    print("\n" + "="*60)
    print("1. CHECKING SQLITE DATABASE")
    print("="*60)
    
    if not SQLITE_PATH.exists():
        log_warning(f"SQLite file not found: {SQLITE_PATH}")
        return {}
    
    file_size = SQLITE_PATH.stat().st_size
    if file_size == 0:
        log_warning(f"SQLite file is empty (0 bytes)")
        return {}
    
    log_info(f"SQLite file size: {file_size} bytes")
    
    try:
        conn = sqlite3.connect(str(SQLITE_PATH))
        cursor = conn.cursor()
        
        # Get all tables
        cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
        tables = [row[0] for row in cursor.fetchall()]
        
        log_info(f"Tables found: {len(tables)}")
        
        counts = {}
        for table in tables:
            try:
                cursor.execute(f"SELECT COUNT(*) FROM [{table}]")
                count = cursor.fetchone()[0]
                counts[table] = count
                if count > 0:
                    log_success(f"  {table}: {count} rows")
                else:
                    print(f"     {table}: 0 rows")
            except Exception as e:
                log_error(f"  Error reading {table}: {e}")
        
        conn.close()
        return counts
    except Exception as e:
        log_error(f"Failed to read SQLite: {e}")
        return {}


def test_login(username: str, password: str) -> dict:
    """Test login and return token"""
    print("\n" + "="*60)
    print("2. TESTING LOGIN")
    print("="*60)
    
    try:
        response = requests.post(
            f"{BASE_URL}/auth/login",
            json={"username": username, "password": password},
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        data = response.json()
        
        if response.status_code == 200 and data.get("token"):
            log_success(f"Login successful!")
            log_info(f"  Username: {username}")
            log_info(f"  Tenant ID: {data.get('tenantId', 'N/A')}")
            log_info(f"  Token: {data['token'][:50]}...")
            return data
        else:
            log_error(f"Login failed: {response.status_code}")
            log_error(f"  Response: {data}")
            return {}
    except Exception as e:
        log_error(f"Login request failed: {e}")
        return {}


def test_api_endpoint(name: str, url: str, token: str, tenant_id: str) -> dict:
    """Test an API endpoint"""
    try:
        response = requests.get(
            url,
            headers={
                "Authorization": f"Bearer {token}",
                "x-tenant-id": tenant_id
            },
            timeout=10
        )
        
        if response.status_code == 200:
            data = response.json()
            return {"success": True, "data": data, "status": 200}
        else:
            return {"success": False, "error": response.text, "status": response.status_code}
    except Exception as e:
        return {"success": False, "error": str(e), "status": 0}


def test_filter_preferences(token: str, tenant_id: str):
    """Test filter preferences API"""
    print("\n" + "="*60)
    print("3. TESTING FILTER PREFERENCES")
    print("="*60)
    
    platforms = ["inventory", "shopee", "lazada", "tiktok"]
    
    for platform in platforms:
        result = test_api_endpoint(
            f"Filter {platform}",
            f"{BASE_URL}/filter-preferences?platform={platform}&page=product",
            token,
            tenant_id
        )
        
        if result["success"]:
            data = result["data"]
            if data.get("success") and data.get("data"):
                filter_data = data["data"]
                visible_cols = len(filter_data.get("visibleColumns", []))
                log_success(f"{platform}: OK (visibleColumns: {visible_cols})")
            else:
                log_success(f"{platform}: OK (empty preferences)")
        else:
            log_error(f"{platform}: Failed - {result['error'][:100]}")


def test_data_loading(token: str, tenant_id: str):
    """Test data loading from various endpoints"""
    print("\n" + "="*60)
    print("4. TESTING DATA LOADING")
    print("="*60)
    
    endpoints = [
        ("Inventory List", f"{BASE_URL}/inventory/list?limit=5"),
        ("Inventory Config", f"{BASE_URL}/inventory/config"),
        ("Inventory Stats", f"{BASE_URL}/inventory/stats"),
        ("Shopee Orders", f"{BASE_URL}/shopee/orders"),
        ("Shopee Products", f"{BASE_URL}/shopee/db/products"),
        ("Lazada Orders", f"{BASE_URL}/lazada/orders"),
        ("Lazada Products", f"{BASE_URL}/lazada/db/products"),
        ("TikTok Orders", f"{BASE_URL}/tiktok/orders"),
        ("TikTok Products", f"{BASE_URL}/tiktok/db/products"),
        ("Products Master", f"{BASE_URL}/products/master"),
        ("Wholesale Settings", f"{BASE_URL}/wholesale/settings"),
    ]
    
    results = {}
    for name, url in endpoints:
        result = test_api_endpoint(name, url, token, tenant_id)
        
        if result["success"]:
            data = result["data"]
            # Try to get count from various response formats
            count = None
            if isinstance(data, dict):
                count = data.get("total") or data.get("count") or len(data.get("data", []) or data.get("products", []) or data.get("orders", []))
            
            if count is not None and count > 0:
                log_success(f"{name}: {count} items")
            else:
                log_info(f"{name}: OK (empty)")
            results[name] = {"status": "ok", "count": count}
        else:
            log_error(f"{name}: {result['status']} - {result['error'][:80]}")
            results[name] = {"status": "error", "error": result["error"]}
    
    return results


def test_tenant_isolation(token: str, correct_tenant: str):
    """Test that tenant isolation works"""
    print("\n" + "="*60)
    print("5. TESTING TENANT ISOLATION")
    print("="*60)
    
    # Test with correct tenant
    result_correct = test_api_endpoint(
        "Correct Tenant",
        f"{BASE_URL}/inventory/list?limit=1",
        token,
        correct_tenant
    )
    
    if result_correct["success"]:
        log_success(f"Correct tenant ({correct_tenant}): Access granted")
    else:
        log_error(f"Correct tenant ({correct_tenant}): Access denied!")
    
    # Test with wrong tenant
    wrong_tenant = "tika_nusseyba" if correct_tenant == "yumna_bertigamart" else "yumna_bertigamart"
    result_wrong = test_api_endpoint(
        "Wrong Tenant",
        f"{BASE_URL}/inventory/list?limit=1",
        token,
        wrong_tenant
    )
    
    # The token should be tied to a specific tenant, so this might still work
    # but return different data
    if result_wrong["success"]:
        log_warning(f"Wrong tenant ({wrong_tenant}): Access granted (check if data is isolated)")
    else:
        log_success(f"Wrong tenant ({wrong_tenant}): Properly rejected")
    
    # Test with no tenant header
    try:
        response = requests.get(
            f"{BASE_URL}/inventory/list?limit=1",
            headers={"Authorization": f"Bearer {token}"},
            timeout=10
        )
        if response.status_code == 401:
            log_success("No tenant header: Properly rejected (401)")
        elif response.status_code == 200:
            log_warning("No tenant header: Access granted (tenant from JWT)")
        else:
            log_info(f"No tenant header: Status {response.status_code}")
    except Exception as e:
        log_error(f"No tenant test failed: {e}")


def test_save_filter_preferences(token: str, tenant_id: str):
    """Test saving filter preferences"""
    print("\n" + "="*60)
    print("6. TESTING SAVE FILTER PREFERENCES")
    print("="*60)
    
    payload = {
        "platform": "shopee",
        "page": "product",
        "visibleColumns": ["sku", "name", "price", "stock"],
        "columnFilters": {"status": "NORMAL"},
        "searchQuery": "",
        "lockedColumns": ["sku"]
    }
    
    try:
        response = requests.post(
            f"{BASE_URL}/filter-preferences",
            json=payload,
            headers={
                "Authorization": f"Bearer {token}",
                "x-tenant-id": tenant_id,
                "Content-Type": "application/json"
            },
            timeout=10
        )
        
        if response.status_code in [200, 201]:
            log_success("Save filter preferences: OK")
            
            # Verify by reading back
            result = test_api_endpoint(
                "Verify saved",
                f"{BASE_URL}/filter-preferences?platform=shopee&page=product",
                token,
                tenant_id
            )
            
            if result["success"]:
                data = result["data"].get("data", {})
                saved_cols = data.get("visibleColumns", [])
                if saved_cols == payload["visibleColumns"]:
                    log_success(f"Verification: Data matches! ({saved_cols})")
                else:
                    log_warning(f"Verification: Data mismatch - got {saved_cols}")
        else:
            log_error(f"Save failed: {response.status_code} - {response.text}")
    except Exception as e:
        log_error(f"Save request failed: {e}")


def main():
    print("\n" + "="*60)
    print("   COMPREHENSIVE API TESTING")
    print("   Go Backend with PostgreSQL")
    print("="*60)
    
    # 1. Check SQLite
    sqlite_data = check_sqlite_database()
    
    # 2. Test Login
    login_result = test_login("yumna", "password123")
    if not login_result.get("token"):
        log_error("Cannot continue without valid login")
        return
    
    token = login_result["token"]
    tenant_id = login_result.get("tenantId", "yumna_bertigamart")
    
    # 3. Test Filter Preferences
    test_filter_preferences(token, tenant_id)
    
    # 4. Test Data Loading
    test_data_loading(token, tenant_id)
    
    # 5. Test Tenant Isolation
    test_tenant_isolation(token, tenant_id)
    
    # 6. Test Save Filter Preferences
    test_save_filter_preferences(token, tenant_id)
    
    print("\n" + "="*60)
    print("   TESTING COMPLETE")
    print("="*60 + "\n")


if __name__ == "__main__":
    main()
