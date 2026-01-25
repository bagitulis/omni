#!/usr/bin/env python3
"""
Test Go backend login with existing users
Users are stored in PostgreSQL with tenant schemas:
- yumna (tenant: yumna_bertigamart) - password: yumna123
- tika (tenant: tika_nusseyba) - password: tika123
"""

import requests
import json

# Go backend URL (running in Docker on port 3000)
GO_BACKEND_URL = "http://localhost:3000"

# Known user credentials (verified from Node.js SQLite)
USERS = {
    "yumna": {"password": "password123", "expected_tenant": "yumna_bertigamart"},
    "tika": {"password": "password123", "expected_tenant": "tika_nusseyba"}
}

def test_login(username: str, password: str):
    """Test login with Go backend"""
    url = f"{GO_BACKEND_URL}/api/auth/login"
    payload = {
        "username": username,
        "password": password
    }
    
    print(f"\n{'='*60}")
    print(f"Testing login for user: {username}")
    print(f"URL: {url}")
    print(f"Payload: {json.dumps(payload)}")
    print("-"*60)
    
    try:
        response = requests.post(url, json=payload, timeout=10)
        print(f"Status Code: {response.status_code}")
        print(f"Response: {json.dumps(response.json(), indent=2)}")
        
        if response.status_code == 200:
            data = response.json()
            print(f"\n✅ LOGIN SUCCESS!")
            print(f"   User: {data.get('user', {}).get('username')}")
            print(f"   TenantID: {data.get('tenantId')}")
            print(f"   Token: {data.get('token', '')[:50]}...")
            return data
        else:
            print(f"\n❌ LOGIN FAILED!")
            return None
    except requests.exceptions.ConnectionError:
        print(f"❌ Cannot connect to {GO_BACKEND_URL}")
        print("   Make sure Go backend is running!")
        return None
    except Exception as e:
        print(f"❌ Error: {e}")
        return None

def test_token_refresh(token: str, tenant_id: str, platform: str = "shopee"):
    """Test token refresh with Go backend"""
    url = f"{GO_BACKEND_URL}/api/tokens/{platform}/refresh"
    headers = {
        "Authorization": f"Bearer {token}",
        "x-tenant-id": tenant_id
    }
    
    print(f"\n{'='*60}")
    print(f"Testing token refresh for platform: {platform}")
    print(f"URL: {url}")
    print(f"TenantID: {tenant_id}")
    print("-"*60)
    
    try:
        response = requests.post(url, headers=headers, timeout=30)
        print(f"Status Code: {response.status_code}")
        print(f"Response: {json.dumps(response.json(), indent=2)}")
        
        if response.status_code == 200:
            print(f"\n✅ TOKEN REFRESH SUCCESS!")
            return response.json()
        else:
            print(f"\n❌ TOKEN REFRESH FAILED!")
            return None
    except Exception as e:
        print(f"❌ Error: {e}")
        return None

def get_token_status(token: str, tenant_id: str, platform: str = "shopee"):
    """Get token status"""
    url = f"{GO_BACKEND_URL}/api/tokens/{platform}/status"
    headers = {
        "Authorization": f"Bearer {token}",
        "x-tenant-id": tenant_id
    }
    
    print(f"\n{'='*60}")
    print(f"Getting token status for platform: {platform}")
    print(f"URL: {url}")
    print("-"*60)
    
    try:
        response = requests.get(url, headers=headers, timeout=10)
        print(f"Status Code: {response.status_code}")
        print(f"Response: {json.dumps(response.json(), indent=2)}")
        return response.json()
    except Exception as e:
        print(f"❌ Error: {e}")
        return None

def health_check():
    """Check if backend is running"""
    url = f"{GO_BACKEND_URL}/api/health"
    try:
        response = requests.get(url, timeout=5)
        print(f"Health check: {response.status_code}")
        return response.status_code == 200
    except:
        return False

if __name__ == "__main__":
    print("="*60)
    print("Go Backend Login Test")
    print("="*60)
    
    # First check if backend is running
    if not health_check():
        print("\n❌ Go backend is not running!")
        print("   Start it with: cd backend && go run cmd/server/main.go")
        exit(1)
    
    print("\n✅ Go backend is running!")
    
    # Test login for each known user
    for username, user_info in USERS.items():
        print(f"\n{'='*60}")
        print(f"Testing login for: {username}")
        print(f"Expected tenant: {user_info['expected_tenant']}")
        print("="*60)
        
        result = test_login(username, user_info['password'])
        if result:
            print(f"\n🎉 Login SUCCESS for {username}!")
            
            # Verify tenant
            actual_tenant = result.get("tenantId", "")
            if actual_tenant == user_info['expected_tenant']:
                print(f"   ✅ Tenant matches: {actual_tenant}")
            else:
                print(f"   ❌ Tenant mismatch! Expected: {user_info['expected_tenant']}, Got: {actual_tenant}")
            
            # Test token status
            token = result.get("token")
            tenant_id = result.get("tenantId")
            
            if token and tenant_id:
                print(f"\n--- Token Status for {username} ---")
                get_token_status(token, tenant_id, "shopee")
                
                # Test token refresh
                print(f"\n--- Token Refresh for {username} ---")
                test_token_refresh(token, tenant_id, "shopee")
        else:
            print(f"\n❌ Login FAILED for {username}!")
    
    print("\n" + "="*60)
    print("Test Complete!")
    print("="*60)
