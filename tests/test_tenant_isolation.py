#!/usr/bin/env python3
"""Test tenant isolation"""
import requests

BASE_URL = 'http://localhost:3000/api'

# Login as yumna
resp = requests.post(f'{BASE_URL}/auth/login', json={'username': 'yumna', 'password': 'password123'})
yumna_token = resp.json()['token']
yumna_tenant = resp.json()['tenantId']

print(f'Yumna tenant from JWT: {yumna_tenant}')

# Test 1: Request with correct tenant header
r1 = requests.get(f'{BASE_URL}/inventory/list?limit=3', headers={
    'Authorization': f'Bearer {yumna_token}',
    'x-tenant-id': 'yumna_bertigamart'
})
data1 = r1.json()
print(f'\n1. Correct tenant: {r1.status_code} - {data1.get("total", 0)} items')

# Test 2: Request with wrong tenant header
r2 = requests.get(f'{BASE_URL}/inventory/list?limit=3', headers={
    'Authorization': f'Bearer {yumna_token}',
    'x-tenant-id': 'tika_nusseyba'
})
data2 = r2.json()
print(f'2. Wrong tenant header (tika): {r2.status_code} - {data2.get("total", 0)} items')

# Test 3: Request with no tenant header
r3 = requests.get(f'{BASE_URL}/inventory/list?limit=3', headers={
    'Authorization': f'Bearer {yumna_token}'
})
data3 = r3.json()
print(f'3. No tenant header: {r3.status_code} - {data3}')

# Summary
print('\n--- TENANT ISOLATION SUMMARY ---')
if data1.get('total', 0) == 139 and data2.get('total', 0) == 0:
    print('✅ TENANT ISOLATION WORKING: yumna has 139, tika has 0')
elif data1.get('total', 0) == data2.get('total', 0):
    print('⚠️  WARNING: Same data count - check if tenant header is being ignored')
else:
    print(f'Data differs: yumna={data1.get("total")}, tika={data2.get("total")}')
