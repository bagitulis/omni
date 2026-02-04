#!/bin/bash

# Configuration
API_URL="http://localhost:8080"
TENANT_ID="1" # Replace with actual tenant ID if different
TOKEN="<YOUR_ACCESS_TOKEN>" # User needs to replace this
PACKAGE_ID="582445147764327664"

# Color codes
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "---------------------------------------------------"
echo "TikTok Shipping Verification Script"
echo "---------------------------------------------------"

# 1. TikTok - Arrange Shipment (Pickup)
echo -e "\n${GREEN}[1] TikTok: Arrange Shipment (Pickup) for Package $PACKAGE_ID${NC}"
curl -X POST "$API_URL/api/tiktok/shipping/arrange" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "package_id": "'"$PACKAGE_ID"'",
    "handover_method": "PICKUP"
}'

# 2. TikTok - Print Label
echo -e "\n\n${GREEN}[2] TikTok: Get Shipping Document (Label) for Package $PACKAGE_ID${NC}"
curl -X GET "$API_URL/api/tiktok/shipping/document/$PACKAGE_ID?document_type=SHIPPING_LABEL" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID"

echo -e "\n\n---------------------------------------------------"
echo "Done."
