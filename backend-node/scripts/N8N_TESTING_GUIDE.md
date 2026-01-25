# 🚀 N8N Dummy Data Testing - Quick Reference

## ✅ Test Status: PASSED

All 3 platforms successfully exported dummy data to N8N webhook.

### 📊 Test Results Summary

- **Total Orders Exported**: 9
- **Total Export Value**: Rp 6.275.000
- **Platforms Tested**: 3 (Shopee, Lazada, TikTok)
- **Success Rate**: 100%
- **Timestamp**: 2026-01-13T12:34:21.588Z

### Platform Breakdown

| Platform  | Orders | Total Value  | Status     |
| --------- | ------ | ------------ | ---------- |
| 🛍️ Shopee | 3      | Rp 1.975.000 | ✅ SUCCESS |
| 🛒 Lazada | 3      | Rp 2.275.000 | ✅ SUCCESS |
| 🎵 TikTok | 3      | Rp 2.025.000 | ✅ SUCCESS |

## 🔗 Frontend URLs

### Order Manager

- **URL**: https://yndigital.my.id/order-manager?type=today
- **Features**:
  - View orders by platform (Shopee, Lazada, TikTok)
  - Export to N8N functionality
  - Real-time order monitoring

### N8N Workflow Editor

- **URL**: https://n8n.yndigital.my.id
- **Status**: ✅ Active (Workflow: "My workflow" activated)
- **Webhook Endpoint**: https://n8n.yndigital.my.id/webhook/export-orders

### Backend API

- **URL**: http://localhost:3000
- **Health Check**: http://localhost:3000/api/health
- **N8N Export Endpoint**: POST /api/n8n/export-orders

## 🧪 Running Tests

### Option 1: Full Test (All Platforms)

```bash
cd backend
npx ts-node scripts/testN8nDummyExportFull.ts
```

### Option 2: Test Menu

```bash
npx ts-node scripts/testN8nMenu.ts --help
```

### Option 3: Quick Test

```bash
npx ts-node scripts/testDummyDataExport.ts
```

### Option 4: Check N8N Status

```bash
npx ts-node scripts/checkN8nStatus.ts
```

## 📁 Test Artifacts

### Results File Location

```
backend/test-results/n8n-export-dummy-data.json
```

### View Results

```bash
cat backend/test-results/n8n-export-dummy-data.json | jq
```

## 🎯 Dummy Data Overview

### Sample Data per Platform

Each platform has 3 test orders with:

- Realistic buyer information
- Multiple items per order
- Various product types
- Prices ranging from Rp 100K - Rp 800K per item

### Shopee Orders (3)

1. **Order 20260113-001**: Rudi Hartono - Rp 450.000 (3 items)
2. **Order 20260113-002**: Siti Nurhaliza - Rp 850.000 (4 items)
3. **Order 20260113-003**: Budi Santosa - Rp 675.000 (2 items)

### Lazada Orders (3)

1. **Order 980001**: Ahmad Wijaya - Rp 1.200.000 (2 items)
2. **Order 980002**: Dwi Cahyanto - Rp 650.000 (3 items)
3. **Order 980003**: Eka Putri Ananda - Rp 425.000 (2 items)

### TikTok Orders (3)

1. **Order 540001**: Eka Putri Ananda - Rp 350.000 (2 items)
2. **Order 540002**: Riya Subramanian - Rp 1.050.000 (3 items)
3. **Order 540003**: Chen Wei - Rp 625.000 (2 items)

## 🔄 Export Flow

```
Frontend Export Request
        ↓
Backend /api/n8n/export-orders
        ↓
Validate & Prepare Data
        ↓
Send to N8N Webhook
        ↓
N8N Workflow Execution
        ↓
Google Sheets Export (or configured destination)
```

## 📋 Test Configuration

- **Tenant ID**: yumna_bertigamart
- **Export Type**: today
- **Test Run**: true
- **Batch ID**: Unique per test run
- **Total Records**: 9 orders, 23 items

## ✨ Next Steps

1. **View Orders in UI**

   - Go to https://yndigital.my.id/order-manager?type=today
   - Switch between Shopee, Lazada, TikTok tabs
   - See dummy orders listed

2. **Export to N8N**

   - Click "Export" button for each platform
   - Monitor workflow at https://n8n.yndigital.my.id
   - Check Google Sheets destination

3. **Monitor Executions**

   - Check N8N workflow executions
   - Verify data transformation
   - Confirm export to destination

4. **Debugging**
   - Check backend logs: `docker logs omni-backend`
   - Check N8N logs: `docker logs omni-n8n`
   - Check test results: `cat backend/test-results/n8n-export-dummy-data.json`

## 🐛 Troubleshooting

### Issue: N8N Connection Failed

- Check N8N container: `docker ps | grep n8n`
- Check N8N logs: `docker logs omni-n8n --tail 20`
- Verify URL: https://n8n.yndigital.my.id is accessible

### Issue: Orders Not Appearing

- Check tenant ID in request headers
- Verify database is populated
- Check backend logs: `docker logs omni-backend --tail 20`

### Issue: Export Failed

- Verify N8N webhook is active
- Check webhook URL in backend config
- Verify tenant configuration

## 📝 Test Scripts Location

```
backend/scripts/
├── testN8nDummyExportFull.ts    ← Full test (all platforms)
├── testDummyDataExport.ts        ← Quick test
├── checkN8nStatus.ts             ← N8N status check
└── testN8nMenu.ts                ← Interactive menu
```

## 🎓 Key Endpoints

| Method | Endpoint                 | Purpose              |
| ------ | ------------------------ | -------------------- |
| POST   | `/api/n8n/export-orders` | Export orders to N8N |
| GET    | `/api/health`            | Backend health check |
| POST   | `/webhook/export-orders` | N8N webhook receiver |

## 📈 Metrics

- **Average Response Time**: < 1 second
- **Success Rate**: 100%
- **Data Integrity**: ✅ All fields validated
- **N8N Integration**: ✅ Active and working

---

**Last Updated**: 2026-01-13T12:34:21Z
**Test Runner**: Automated N8N Export Test Suite
**Status**: ✅ ALL SYSTEMS OPERATIONAL
