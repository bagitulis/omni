# 🚀 Lightshot E-Commerce System - E2E Testing Guide

## Overview

Comprehensive end-to-end testing suite untuk semua fitur Lightshot E-Commerce System, mencakup semua tab di left sidebar dengan sub-menuya.

## Test Coverage

### Tabs Ditest:

1. **Operation**

   - Health Check
   - Shopee Operations
   - Job Queue Status
   - Order Sync
   - Platform Auth Status

2. **Product Management**

   - Product List
   - Product Detail
   - Product Search
   - Product Categories
   - SKU Check
   - Product Pricing

3. **Order Manager**

   - Orders List
   - Shopee Orders
   - TikTok Orders
   - Order Detail
   - Order Tracking
   - Order Filters
   - Order Statistics

4. **Inventory**

   - Inventory List
   - Inventory Detail
   - Stock Levels
   - Sync Status
   - Inventory Columns
   - Configuration
   - Alerts

5. **Route Mapper**

   - Route List
   - Route Detail
   - Route Configuration
   - Route Execution Config
   - Route Status
   - Execution Logs

6. **Script Monitor**

   - **Current Running**
     - Current Running Jobs
     - Job Progress
   - **Queue**
     - Queue Status
     - Queue Jobs
     - Retry Jobs
   - **History**
     - Job History
     - Completed Jobs
     - Failed Jobs
   - **Auto-Functions**
     - Auto-Functions List
     - Auto-Function Detail
     - Auto-Function Status

7. **Analytics**

   - Analytics Dashboard
   - Order Analytics
   - Revenue Analytics
   - Product Analytics
   - Inventory Analytics
   - Platform Metrics
   - Performance Metrics
   - Reports

8. **Shopee**

   - Connection Status
   - Orders
   - Products
   - Shipping
   - Wallet
   - Analytics
   - Sync

9. **TikTok**
   - Connection Status
   - Orders
   - Products
   - Shipping
   - Analytics
   - Campaigns
   - Sync

## Setup

### Prerequisites

- Node.js v16+
- npm/yarn
- Backend running on http://localhost:3000
- Valid credentials: yumna / password123

### Installation

```bash
cd backend
npm install
```

## Running Tests

### Run All Tests (Master Suite)

```bash
npm run test:e2e
```

This will:

1. Run all test suites sequentially
2. Save individual JSON results for each tab
3. Generate master report with overall statistics
4. Output results to `backend/test-results/` folder

### Run Individual Tests

```bash
# Operation Tab
npm run test:e2e:operation

# Product Management
npm run test:e2e:products

# Order Manager
npm run test:e2e:orders

# Inventory
npm run test:e2e:inventory

# Route Mapper
npm run test:e2e:routes

# Script Monitor
npm run test:e2e:scripts

# Analytics
npm run test:e2e:analytics

# Shopee Platform
npm run test:e2e:shopee

# TikTok Platform
npm run test:e2e:tiktok
```

## Test Results

All results are automatically saved to:

```
backend/test-results/
├── operation.json
├── product-management.json
├── order-manager.json
├── inventory.json
├── route-mapper.json
├── script-monitor.json
├── analytics.json
├── shopee.json
├── tiktok.json
└── master-report.json
```

### Result Format

Each JSON file contains:

```json
{
  "tab": "Tab Name",
  "timestamp": "2026-01-09T12:00:00.000Z",
  "summary": {
    "totalTests": 10,
    "passed": 8,
    "failed": 2,
    "totalDuration": 5000,
    "successRate": "80.00%"
  },
  "tests": [
    {
      "name": "Test Name",
      "status": "PASS|FAIL",
      "duration": 500,
      "error": "Error message if FAIL",
      "data": {}
    }
  ]
}
```

### Master Report Format

```json
{
  "timestamp": "2026-01-09T12:00:00.000Z",
  "summary": {
    "totalTests": 80,
    "passed": 72,
    "failed": 8,
    "totalDuration": 45000,
    "successRate": "90.00%"
  },
  "testSuites": [...]
}
```

## Review Results

Open any JSON file with JSON viewer or use command line:

```bash
# View master report
cat backend/test-results/master-report.json

# View specific tab results
cat backend/test-results/operation.json

# Pretty print (Windows PowerShell)
Get-Content backend/test-results/master-report.json | ConvertFrom-Json | ConvertTo-Json
```

## Authentication

Tests use these credentials:

- **Email**: yumna@example.com
- **Password**: password123
- **Tenant ID**: yumna

## Troubleshooting

### Tests Fail - Backend Not Running

```bash
# Make sure backend is running
npm run dev

# Or use Docker
docker-compose up
```

### Tests Timeout

Increase timeout in individual test files:

```typescript
const api = axios.create({
  baseURL: BASE_URL,
  timeout: 20000, // Increase from 10000
});
```

### Results Directory Not Created

Results directory will be created automatically, but if permission issues:

```bash
mkdir -p backend/test-results
```

## Performance Benchmarks

Expected test execution times:

- Individual test: 50-500ms
- Tab (5-10 tests): 500-3000ms
- Full suite (80+ tests): 30-60 seconds

## Notes

- Tests are read-only (no data modification)
- All tests use HTTP calls to backend API
- Results are JSON format for easy automation/parsing
- Each tab has own JSON file for modular review
- Master report provides overall statistics

## Support

For issues or questions about tests:

1. Check backend logs: `backend/logs/`
2. Verify API endpoints are accessible
3. Confirm authentication credentials
4. Check test-results JSON for detailed error messages
