#!/bin/bash

# Test script untuk price update dengan SKU FG1592K1110005A
# Sesuai dengan README.md - Bukti 2: Output script dengan keterangan sukses eksplisit

echo "========================================="
echo "TEST PRICE UPDATE BATCH"
echo "SKU: FG1592K1110005A"
echo "========================================="

# 1. Login untuk mendapatkan token
echo ""
echo "[STEP 1] Login dengan credentials yumna/password123..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "yumna",
    "password": "password123"
  }')

echo "Login Response: $LOGIN_RESPONSE"

# Extract token
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | sed 's/"token":"//')

if [ -z "$TOKEN" ]; then
  echo "❌ GAGAL: Tidak bisa login - token kosong"
  exit 1
fi

echo "✅ Login berhasil! Token: ${TOKEN:0:50}..."

# 2. Test update price batch dengan SKU FG1592K1110005A
echo ""
echo "[STEP 2] Test update price batch dengan SKU FG1592K1110005A..."
echo "Request: Update harga dari 25700 -> 30000"

PRICE_UPDATE_RESPONSE=$(curl -s -X POST http://localhost:3000/api/inventory/update-price-batch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "items": [
      {
        "sku": "FG1592K1110005A",
        "price": 30000,
        "platforms": ["shopee", "lazada", "tiktok"]
      }
    ]
  }')

echo "Price Update Response: $PRICE_UPDATE_RESPONSE"

# Check if successful
if echo "$PRICE_UPDATE_RESPONSE" | grep -q '"success":true'; then
  echo ""
  echo "✅ ============================================"
  echo "✅ SUKSES! Price update batch berhasil!"
  echo "✅ SKU: FG1592K1110005A"
  echo "✅ Harga baru: 30000"
  echo "✅ Platforms: shopee, lazada, tiktok"
  echo "✅ ============================================"
  exit 0
else
  echo ""
  echo "❌ ============================================"
  echo "❌ GAGAL! Price update batch error"
  echo "❌ Response: $PRICE_UPDATE_RESPONSE"
  echo "❌ ============================================"
  exit 1
fi
