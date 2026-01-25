#!/bin/bash
# Comprehensive Test Script for Go Backend
# Run from inside the backend container or with appropriate network access

BASE_URL="${TEST_BASE_URL:-http://localhost:3000}"
USERNAME="${TEST_USERNAME:-tester}"
PASSWORD="${TEST_PASSWORD:-tester@123}"
TENANT_A="${TEST_TENANT_A:-yumna_bertigamart}"
TENANT_B="${TEST_TENANT_B:-tika_nusseyba}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
SKIPPED=0

# Test function
run_test() {
    local name="$1"
    local expected_status="$2"
    local actual_status="$3"
    local response="$4"

    if [ "$actual_status" == "$expected_status" ]; then
        echo -e "${GREEN}✓ PASS${NC}: $name"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $name (expected: $expected_status, got: $actual_status)"
        echo "  Response: ${response:0:200}"
        ((FAILED++))
        return 1
    fi
}

# Check if response contains string
contains() {
    local response="$1"
    local search="$2"
    echo "$response" | grep -q "$search"
    return $?
}

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}   Omni Backend Integration Test Suite     ${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""
echo "Base URL: $BASE_URL"
echo "Username: $USERNAME"
echo ""

# ===========================================
# SECTION 1: HEALTH CHECK
# ===========================================
echo -e "\n${YELLOW}=== 1. Health Check Tests ===${NC}"

HEALTH_RESPONSE=$(wget -qO- "$BASE_URL/api/health" 2>&1)
HEALTH_STATUS=$?

if [ $HEALTH_STATUS -eq 0 ]; then
    if contains "$HEALTH_RESPONSE" '"status":"healthy"' || contains "$HEALTH_RESPONSE" '"status":"ok"'; then
        run_test "Health endpoint returns healthy" "0" "0" "$HEALTH_RESPONSE"
    else
        run_test "Health endpoint returns healthy" "healthy" "unhealthy" "$HEALTH_RESPONSE"
    fi
else
    echo -e "${RED}✗ FAIL${NC}: Cannot connect to backend at $BASE_URL"
    ((FAILED++))
    echo "Exiting - backend not available"
    exit 1
fi

# ===========================================
# SECTION 2: AUTHENTICATION TESTS
# ===========================================
echo -e "\n${YELLOW}=== 2. Authentication Tests ===${NC}"

# Test 2.1: Login with valid credentials
LOGIN_RESPONSE=$(wget -qO- \
    --header="Content-Type: application/json" \
    --post-data="{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" \
    "$BASE_URL/api/auth/login" 2>&1)

if contains "$LOGIN_RESPONSE" '"success":true'; then
    run_test "Login with valid credentials" "0" "0" "$LOGIN_RESPONSE"
else
    run_test "Login with valid credentials" "success:true" "fail" "$LOGIN_RESPONSE"
fi

# Extract token
TOKEN=$(echo "$LOGIN_RESPONSE" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
TENANT_ID=$(echo "$LOGIN_RESPONSE" | sed -n 's/.*"tenant_id":"\([^"]*\)".*/\1/p')

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Cannot continue without token${NC}"
    exit 1
fi

echo "  Token obtained: ${TOKEN:0:50}..."
echo "  Tenant ID: $TENANT_ID"

# Test 2.2: Login with invalid credentials
INVALID_LOGIN=$(wget -qO- \
    --header="Content-Type: application/json" \
    --post-data='{"username":"nonexistent","password":"wrongpass"}' \
    "$BASE_URL/api/auth/login" 2>&1)

if contains "$INVALID_LOGIN" '"success":false'; then
    run_test "Login with invalid credentials rejected" "0" "0" "$INVALID_LOGIN"
else
    run_test "Login with invalid credentials rejected" "success:false" "unexpected" "$INVALID_LOGIN"
fi

# Test 2.3: Login with wrong password
WRONG_PASS=$(wget -qO- \
    --header="Content-Type: application/json" \
    --post-data="{\"username\":\"$USERNAME\",\"password\":\"wrongpassword123\"}" \
    "$BASE_URL/api/auth/login" 2>&1)

if contains "$WRONG_PASS" '"success":false'; then
    run_test "Login with wrong password rejected" "0" "0" "$WRONG_PASS"
else
    run_test "Login with wrong password rejected" "success:false" "unexpected" "$WRONG_PASS"
fi

# Test 2.4: Verify token
VERIFY_RESPONSE=$(wget -qO- \
    --header="Authorization: Bearer $TOKEN" \
    "$BASE_URL/api/auth/verify" 2>&1)

if contains "$VERIFY_RESPONSE" '"valid":true'; then
    run_test "Token verification works" "0" "0" "$VERIFY_RESPONSE"
else
    run_test "Token verification works" "valid:true" "fail" "$VERIFY_RESPONSE"
fi

# Test 2.5: Verify invalid token rejected
INVALID_TOKEN_RESP=$(wget -qO- \
    --header="Authorization: Bearer invalid.token.here" \
    "$BASE_URL/api/auth/verify" 2>&1)

if contains "$INVALID_TOKEN_RESP" '"valid":false' || contains "$INVALID_TOKEN_RESP" '"success":false'; then
    run_test "Invalid token rejected" "0" "0" "$INVALID_TOKEN_RESP"
else
    run_test "Invalid token rejected" "valid:false" "unexpected" "$INVALID_TOKEN_RESP"
fi

# ===========================================
# SECTION 3: TENANT MANAGEMENT TESTS
# ===========================================
echo -e "\n${YELLOW}=== 3. Tenant Management Tests ===${NC}"

# Test 3.1: Get available tenants
TENANTS_RESPONSE=$(wget -qO- \
    --header="Authorization: Bearer $TOKEN" \
    "$BASE_URL/api/auth/tenants" 2>&1)

if contains "$TENANTS_RESPONSE" '"success":true' && contains "$TENANTS_RESPONSE" '"tenants"'; then
    run_test "Get available tenants" "0" "0" "$TENANTS_RESPONSE"
else
    run_test "Get available tenants" "success:true" "fail" "$TENANTS_RESPONSE"
fi

# Test 3.2: Switch tenant (developer only)
SWITCH_RESPONSE=$(wget -qO- \
    --header="Authorization: Bearer $TOKEN" \
    --header="Content-Type: application/json" \
    --post-data="{\"tenant_id\":\"$TENANT_A\"}" \
    "$BASE_URL/api/auth/switch-tenant" 2>&1)

if contains "$SWITCH_RESPONSE" '"success":true'; then
    run_test "Switch tenant (developer)" "0" "0" "$SWITCH_RESPONSE"
    # Get new token for tenant A
    TOKEN_A=$(echo "$SWITCH_RESPONSE" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
else
    run_test "Switch tenant (developer)" "success:true" "fail" "$SWITCH_RESPONSE"
    TOKEN_A="$TOKEN"
fi

# Switch to tenant B
SWITCH_B_RESPONSE=$(wget -qO- \
    --header="Authorization: Bearer $TOKEN" \
    --header="Content-Type: application/json" \
    --post-data="{\"tenant_id\":\"$TENANT_B\"}" \
    "$BASE_URL/api/auth/switch-tenant" 2>&1)

if contains "$SWITCH_B_RESPONSE" '"success":true'; then
    TOKEN_B=$(echo "$SWITCH_B_RESPONSE" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
    run_test "Switch to different tenant" "0" "0" "$SWITCH_B_RESPONSE"
else
    TOKEN_B="$TOKEN"
    ((SKIPPED++))
    echo -e "${YELLOW}○ SKIP${NC}: Switch to different tenant (may not have access)"
fi

# ===========================================
# SECTION 4: SNAKE_CASE RESPONSE FORMAT TESTS
# ===========================================
echo -e "\n${YELLOW}=== 4. Response Format Tests (AGENTS.MD Compliance) ===${NC}"

# Test 4.1: Response uses snake_case for tenant_id
if contains "$LOGIN_RESPONSE" '"tenant_id"'; then
    run_test "Response uses snake_case (tenant_id)" "0" "0" "found tenant_id"
else
    run_test "Response uses snake_case (tenant_id)" "snake_case" "camelCase" "$LOGIN_RESPONSE"
fi

# Test 4.2: Response uses snake_case for shop_name in tenants
if contains "$TENANTS_RESPONSE" '"shop_name"'; then
    run_test "Tenants use snake_case (shop_name)" "0" "0" "found shop_name"
else
    run_test "Tenants use snake_case (shop_name)" "snake_case" "camelCase" "$TENANTS_RESPONSE"
fi

# Test 4.3: All responses have success field
if contains "$LOGIN_RESPONSE" '"success"'; then
    run_test "Response includes success field" "0" "0" "found success"
else
    run_test "Response includes success field" "success" "missing" "$LOGIN_RESPONSE"
fi

# ===========================================
# SECTION 5: SECURITY TESTS
# ===========================================
echo -e "\n${YELLOW}=== 5. Security Tests ===${NC}"

# Test 5.1: Request without auth rejected
NO_AUTH_RESPONSE=$(wget -qO- "$BASE_URL/api/orders" 2>&1)

if contains "$NO_AUTH_RESPONSE" '"success":false' || contains "$NO_AUTH_RESPONSE" 'unauthorized' || [ -z "$NO_AUTH_RESPONSE" ]; then
    run_test "Request without auth rejected" "0" "0" "${NO_AUTH_RESPONSE:0:100}"
else
    run_test "Request without auth rejected" "rejected" "allowed" "$NO_AUTH_RESPONSE"
fi

# Test 5.2: Password not in response
if ! contains "$LOGIN_RESPONSE" "$PASSWORD" && ! contains "$LOGIN_RESPONSE" '\$2b\$'; then
    run_test "Password not in response" "0" "0" "password hidden"
else
    run_test "Password not in response" "hidden" "exposed" "WARNING: Password may be exposed!"
fi

# Test 5.3: No stack trace in error response
ERROR_RESPONSE=$(wget -qO- \
    --header="Authorization: Bearer $TOKEN" \
    "$BASE_URL/api/orders/invalid-id-test" 2>&1)

if ! contains "$ERROR_RESPONSE" '"stack"' && ! contains "$ERROR_RESPONSE" 'goroutine'; then
    run_test "No stack trace in errors" "0" "0" "stack hidden"
else
    run_test "No stack trace in errors" "hidden" "exposed" "$ERROR_RESPONSE"
fi

# Test 5.4: No SQL errors exposed
if ! contains "$ERROR_RESPONSE" 'SELECT' && ! contains "$ERROR_RESPONSE" 'INSERT' && ! contains "$ERROR_RESPONSE" 'gorm'; then
    run_test "No SQL errors exposed" "0" "0" "sql hidden"
else
    run_test "No SQL errors exposed" "hidden" "exposed" "$ERROR_RESPONSE"
fi

# ===========================================
# SECTION 6: TENANT ISOLATION TESTS
# ===========================================
echo -e "\n${YELLOW}=== 6. Tenant Isolation Tests ===${NC}"

# Test 6.1: Tenant A data request
ORDERS_A=$(wget -qO- \
    --header="Authorization: Bearer $TOKEN_A" \
    "$BASE_URL/api/orders" 2>&1)

# Test 6.2: Tenant B data request
ORDERS_B=$(wget -qO- \
    --header="Authorization: Bearer $TOKEN_B" \
    "$BASE_URL/api/orders" 2>&1)

# Test 6.3: Check no tenant B info in tenant A response
if ! contains "$ORDERS_A" "$TENANT_B"; then
    run_test "Tenant A response has no Tenant B data" "0" "0" "isolated"
else
    run_test "Tenant A response has no Tenant B data" "isolated" "leaked" "$ORDERS_A"
fi

# Test 6.4: Tenant ID injection via query param should be ignored
INJECT_RESPONSE=$(wget -qO- \
    --header="Authorization: Bearer $TOKEN_A" \
    "$BASE_URL/api/orders?tenant_id=$TENANT_B" 2>&1)

# Should not cause 500 error
if ! contains "$INJECT_RESPONSE" '"status":500' && ! contains "$INJECT_RESPONSE" 'Internal Server Error'; then
    run_test "Tenant injection via query param safe" "0" "0" "safe"
else
    run_test "Tenant injection via query param safe" "safe" "vulnerable" "$INJECT_RESPONSE"
fi

# ===========================================
# SUMMARY
# ===========================================
echo -e "\n${BLUE}============================================${NC}"
echo -e "${BLUE}               TEST SUMMARY                 ${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""
echo -e "  ${GREEN}Passed:${NC}  $PASSED"
echo -e "  ${RED}Failed:${NC}  $FAILED"
echo -e "  ${YELLOW}Skipped:${NC} $SKIPPED"
echo ""

TOTAL=$((PASSED + FAILED))
if [ $TOTAL -gt 0 ]; then
    PERCENT=$((PASSED * 100 / TOTAL))
    echo -e "  Success Rate: ${PERCENT}%"
fi

echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed. Please review.${NC}"
    exit 1
fi
