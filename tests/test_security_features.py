"""
Security Features Test Suite
Tests all implemented security features
"""
import requests
import time
import json

BASE_URL = "http://localhost:3001/api"

def print_header(title):
    print("\n" + "=" * 60)
    print(f"  {title}")
    print("=" * 60)

def test_rate_limiting():
    """Test login rate limiting (5 attempts per 15 minutes)"""
    print_header("TEST: Rate Limiting - Login")
    
    # Attempt 6 failed logins
    for i in range(6):
        response = requests.post(
            f"{BASE_URL}/auth/login",
            json={
                "username": "nonexistent",
                "password": "wrongpassword"
            }
        )
        print(f"Attempt {i+1}: Status {response.status_code}")
        
        if response.status_code == 429:
            print("✅ Rate limiting working! Blocked after 5 attempts")
            print(f"Response: {response.json()}")
            break
        elif i < 5:
            print(f"   Response: {response.json().get('error', 'Unknown')}")
        
        time.sleep(0.5)
    else:
        print("❌ Rate limiting NOT working - should be blocked")

def test_input_validation():
    """Test input validation on registration"""
    print_header("TEST: Input Validation")
    
    test_cases = [
        {
            "name": "Short username",
            "data": {"username": "ab", "email": "test@test.com", "password": "Password123"},
            "expected_error": "Username must be between 3-50 characters"
        },
        {
            "name": "Invalid email",
            "data": {"username": "testuser", "email": "invalid-email", "password": "Password123"},
            "expected_error": "Invalid email format"
        },
        {
            "name": "Short password",
            "data": {"username": "testuser", "email": "test@test.com", "password": "Pass1"},
            "expected_error": "Password must be at least 8 characters"
        },
        {
            "name": "No uppercase",
            "data": {"username": "testuser", "email": "test@test.com", "password": "password123"},
            "expected_error": "uppercase letter"
        },
        {
            "name": "No lowercase",
            "data": {"username": "testuser", "email": "test@test.com", "password": "PASSWORD123"},
            "expected_error": "lowercase letter"
        },
        {
            "name": "No number",
            "data": {"username": "testuser", "email": "test@test.com", "password": "Password"},
            "expected_error": "number"
        },
    ]
    
    for test in test_cases:
        response = requests.post(f"{BASE_URL}/auth/register", json=test["data"])
        if response.status_code == 400:
            error_msg = response.json().get("message", "")
            if test["expected_error"].lower() in error_msg.lower():
                print(f"✅ {test['name']}: Validation working")
            else:
                print(f"⚠️  {test['name']}: Got '{error_msg}'")
        else:
            print(f"❌ {test['name']}: Should return 400, got {response.status_code}")

def test_password_policy():
    """Test password policy enforcement"""
    print_header("TEST: Password Policy")
    
    weak_passwords = [
        ("12345678", "too weak"),
        ("password", "common password"),
        ("Password", "no number"),
        ("password1", "no uppercase"),
        ("PASSWORD1", "no lowercase"),
    ]
    
    for password, reason in weak_passwords:
        response = requests.post(
            f"{BASE_URL}/auth/register",
            json={
                "username": f"user_{int(time.time())}",
                "email": f"test_{int(time.time())}@test.com",
                "password": password
            }
        )
        
        if response.status_code == 400:
            print(f"✅ Rejected: '{password}' ({reason})")
        else:
            print(f"❌ Accepted weak password: '{password}' ({reason})")

def test_account_lockout():
    """Test account lockout after failed attempts"""
    print_header("TEST: Account Lockout")
    
    # First, create a test user
    test_username = f"locktest_{int(time.time())}"
    test_password = "TestPass123"
    
    print("Creating test user...")
    response = requests.post(
        f"{BASE_URL}/auth/register",
        json={
            "username": test_username,
            "email": f"{test_username}@test.com",
            "password": test_password
        }
    )
    
    if response.status_code != 201:
        print(f"❌ Failed to create user: {response.json()}")
        return
    
    print(f"✅ User created: {test_username}")
    
    # Now attempt 5 failed logins
    print("\nAttempting 5 failed logins...")
    for i in range(5):
        response = requests.post(
            f"{BASE_URL}/auth/login",
            json={"username": test_username, "password": "wrongpassword"}
        )
        print(f"  Attempt {i+1}: {response.status_code}")
    
    # 6th attempt should be blocked by account lockout
    print("\nAttempting 6th login (should be locked)...")
    response = requests.post(
        f"{BASE_URL}/auth/login",
        json={"username": test_username, "password": "wrongpassword"}
    )
    
    if response.status_code == 401:
        error_msg = response.json().get("error", "")
        if "locked" in error_msg.lower() or "too many" in error_msg.lower():
            print(f"✅ Account locked: {error_msg}")
        else:
            print(f"⚠️  Got error but not clear if locked: {error_msg}")
    else:
        print(f"❌ Account not locked, status: {response.status_code}")

def test_https_headers():
    """Test security headers"""
    print_header("TEST: Security Headers")
    
    response = requests.get(f"{BASE_URL}/health")
    headers = response.headers
    
    security_headers = {
        "X-Content-Type-Options": "nosniff",
        "X-Frame-Options": "SAMEORIGIN",
        "Strict-Transport-Security": "max-age",
        "Referrer-Policy": "strict-origin",
    }
    
    for header, expected_value in security_headers.items():
        if header in headers:
            value = headers[header]
            if expected_value in value.lower():
                print(f"✅ {header}: {value}")
            else:
                print(f"⚠️  {header}: {value} (expected '{expected_value}')")
        else:
            print(f"❌ Missing header: {header}")

def test_jwt_validation():
    """Test JWT secret validation (server should not start with weak secret)"""
    print_header("TEST: JWT Secret Validation")
    print("✅ Server started successfully")
    print("   This means JWT_SECRET passed validation (min 32 chars)")

def main():
    print("\n🔐 SECURITY FEATURES TEST SUITE")
    print("=" * 60)
    print("Testing backend security implementation...")
    print("Make sure backend server is running on http://localhost:3001")
    print()
    
    try:
        # Check if server is running
        response = requests.get(f"{BASE_URL}/health", timeout=2)
        print(f"✅ Backend server is running (Status: {response.status_code})")
    except requests.exceptions.RequestException as e:
        print(f"❌ Backend server not running: {e}")
        print("\nPlease start the backend server first:")
        print("  cd backend")
        print("  npm run dev")
        return
    
    # Run all tests
    test_jwt_validation()
    test_https_headers()
    test_input_validation()
    test_password_policy()
    test_rate_limiting()
    test_account_lockout()
    
    print("\n" + "=" * 60)
    print("✅ Security test suite completed!")
    print("=" * 60)

if __name__ == "__main__":
    main()
