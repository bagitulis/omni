#!/bin/bash

# Configuration
API_URL="http://localhost:8080"
TENANT_ID="1" # Replace with actual tenant ID if different
TOKEN="<YOUR_ACCESS_TOKEN>" # User needs to replace this

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "---------------------------------------------------"
echo "Shipping Verification Script"
echo "---------------------------------------------------"

if [ "$TOKEN" == "<YOUR_ACCESS_TOKEN>" ]; then
    echo -e "${RED}Please update the TOKEN variable in this script with your access token.${NC}"
    exit 1
fi

# 1. Shopee - Arrange Pickup (Order: 260203Q98DEKHK)
echo -e "\n${GREEN}[1] Shopee: Arrange Pickup (Order 260203Q98DEKHK)${NC}"
echo "Note: This will automatically select tomorrow's date for pickup if not specified."
curl -X POST "$API_URL/api/shopee/shipping/arrange" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "order_sn": "260203Q98DEKHK",
    "pickup": {
        "address_id": 12345
    }
}'
# Note: address_id 12345 is a placeholder. User might need to fetch options first.

# 2. Shopee - Print Label
echo -e "\n\n${GREEN}[2] Shopee: Get Shipping Label${NC}"
curl -X GET "$API_URL/api/shopee/shipping/label/260203Q98DEKHK?document_type=THERMAL_AIR_WAYBILL" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID"

# 3. TikTok - Arrange Shipment (Package: 582445147764327664)
echo -e "\n\n${GREEN}[3] TikTok: Arrange Shipment (Package 582445147764327664)${NC}"
curl -X POST "$API_URL/api/tiktok/shipping/arrange" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "package_id": "582445147764327664",
    "handover_method": "PICKUP"
}'

# 4. TikTok - Print Label
echo -e "\n\n${GREEN}[4] TikTok: Get Shipping Document${NC}"
curl -X GET "$API_URL/api/tiktok/shipping/document/582445147764327664?document_type=SHIPPING_LABEL" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID"

echo -e "\n\n---------------------------------------------------"
echo "Done."
