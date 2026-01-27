#!/bin/bash

# Test Wholesale Endpoints
# RESPONSIBILITY: Test batch MPQ and wholesale reset endpoints

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:3000}"
TENANT_ID="${TENANT_ID:-yumna_bertigamart}"
JWT_TOKEN="${JWT_TOKEN}"

if [ -z "$JWT_TOKEN" ]; then
    echo -e "${RED}ERROR: JWT_TOKEN environment variable is required${NC}"
    echo "Usage: JWT_TOKEN=your_token_here ./test_wholesale_endpoints.sh"
    exit 1
fi

echo -e "${YELLOW}Testing Wholesale Endpoints${NC}"
echo "API URL: $API_URL"
echo "Tenant: $TENANT_ID"
echo ""

# Test 1: GET /api/wholesale/settings
echo -e "${YELLOW}Test 1: GET /api/wholesale/settings${NC}"
curl -s -X GET \
  "$API_URL/api/wholesale/settings" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID" \
  -H "Content-Type: application/json" | jq '.'
echo ""

# Test 2: POST /api/wholesale/shopee/batch-mpq
echo -e "${YELLOW}Test 2: POST /api/wholesale/shopee/batch-mpq${NC}"
curl -s -X POST \
  "$API_URL/api/wholesale/shopee/batch-mpq" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"sku": "TEST-SKU-001", "price": 50000}
    ],
    "mpq": 5
  }' | jq '.'
echo ""

# Test 3: POST /api/wholesale/shopee/batch-delete-skus
echo -e "${YELLOW}Test 3: POST /api/wholesale/shopee/batch-delete-skus${NC}"
curl -s -X POST \
  "$API_URL/api/wholesale/shopee/batch-delete-skus" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "skus": ["TEST-SKU-001"]
  }' | jq '.'
echo ""

# Test 4: POST /api/wholesale/shopee/batch-wholesale-reset
echo -e "${YELLOW}Test 4: POST /api/wholesale/shopee/batch-wholesale-reset${NC}"
curl -s -X POST \
  "$API_URL/api/wholesale/shopee/batch-wholesale-reset" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"sku": "TEST-SKU-001", "price": 50000}
    ]
  }' | jq '.'
echo ""

echo -e "${GREEN}Testing completed!${NC}"
