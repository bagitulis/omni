# 🚀 HYBRID ANALYTICS IMPLEMENTATION PLAN

**Date:** 2026-01-27  
**Scope:** Shopee/TikTok Ads Analytics - Full Stack Revamp  
**Strategy:** Hybrid (Python ML + Go Backend + Modern Frontend)

---

## ✅ COMPLETED

### Phase 1A: Backend Upload - Shopee Ads

- [x] Import dependencies (io, strings, ads service)
- [x] Implement Upload() handler untuk Shopee CSV
- [x] File validation (CSV only, max 50MB)
- [x] ParseCSV integration
- [x] SaveBatch dengan transaction

---

## 🔄 IN PROGRESS

### Phase 1B: Backend Upload - TikTok Ads

- [ ] Edit `backend/internal/handlers/analytics/tiktok_ads.go`
- [ ] Implement Excel upload (.xlsx support)
- [ ] Use excelize library untuk parsing
- [ ] Period extraction dari filename
- [ ] Save creative data + batches

### Phase 1C: Backend Upload - Helper Functions

- [ ] Create `backend/internal/services/ads/helpers.go`
  - buildColumnIndex()
  - parseShopeeRow()
  - splitLines()
  - cleanCurrency()
  - parseFloat()
  - parseInt()

---

## 📝 TODO: BACKEND

### Phase 2: ML Report Service

**File:** `backend/internal/services/ml/report_service.go`

```go
type MLReportService struct {
    db         *gorm.DB
    pythonPath string
    scriptsDir string
}

// GenerateReport triggers Python ML script
func (s *MLReportService) GenerateReport(ctx context.Context,
    tenantID, platform string) (*MLReport, error)

// GetLatestReport retrieves most recent HTML report
func (s *MLReportService) GetLatestReport(ctx context.Context,
    tenantID, platform string) (*MLReport, error)

// ListReports returns all available reports
func (s *MLReportService) ListReports(ctx context.Context,
    tenantID, platform string) ([]*MLReport, error)
```

**Dependencies:**

- `os/exec` for subprocess
- `backend/internal/models/ml_report.go` (new model)

### Phase 3: Python Executor

**File:** `backend/internal/services/ml/python_executor.go`

```go
type PythonExecutor struct {
    pythonPath string
    timeout    time.Duration
}

// Execute runs Python script and captures output
func (e *PythonExecutor) Execute(ctx context.Context,
    scriptPath string, args []string) (*ExecResult, error)
```

### Phase 4: Report Endpoints

**File:** `backend/internal/handlers/ml/report_handler.go`

```
POST   /api/ml/reports/generate
GET    /api/ml/reports/:platform/list
GET    /api/ml/reports/:platform/latest
GET    /api/ml/reports/:platform/:filename
```

### Phase 5: Cursor Pagination

**File:** `backend/internal/services/ads/shopee_ads_analytics.go`

```go
// GetDataWithCursor returns paginated data for virtual scrolling
func (s *ShopeeAdsService) GetDataWithCursor(ctx context.Context,
    cursor string, limit int) (*CursorResult, error)
```

### Phase 6: In-Memory Cache

**File:** `backend/internal/cache/memory_cache.go`

```go
type MemoryCache struct {
    store sync.Map
    ttl   map[string]time.Time
    mu    sync.RWMutex
}

func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration)
func (c *MemoryCache) Get(key string) (interface{}, bool)
func (c *MemoryCache) Delete(key string)
```

---

## 📝 TODO: FRONTEND

### Phase 7: Install Dependencies

```bash
cd frontend
npm install --save \
  apexcharts \
  vue3-apexcharts \
  @tanstack/vue-virtual@^3.0.0 \
  @tanstack/vue-query@^5.0.0 \
  @vueuse/core@^11.0.0
```

### Phase 8: Chart Components

**File:** `frontend/src/components/analytics/charts/RevenueAreaChart.vue`

- ApexCharts area chart
- Time series revenue vs cost
- Zoom & pan support

**File:** `frontend/src/components/analytics/charts/CreativePerformanceChart.vue`

- Column chart untuk creative comparison
- ROI color coding

**File:** `frontend/src/components/analytics/charts/RoiDistributionChart.vue`

- Donut chart untuk ROI categories

### Phase 9: Virtual Data Table

**File:** `frontend/src/components/analytics/TiktokAdsVirtualTable.vue`

```vue
<template>
  <div ref="parentRef" class="virtual-table-container">
    <div :style="{ height: `${totalSize}px` }">
      <div v-for="virtualRow in virtualItems" :key="virtualRow.index">
        <!-- Row rendering -->
      </div>
    </div>
  </div>
</template>
```

**Composable:** `frontend/src/composables/useVirtualScroll.ts`
**Composable:** `frontend/src/composables/useInfiniteQuery.ts`

### Phase 10: AI Report Gallery

**File:** `frontend/src/views/analytics/AIReportGallery.vue`

- Grid layout dengan report cards
- Generate new report modal
- Progress tracking

### Phase 11: Dashboard Refactor

**Files:**

- `frontend/src/views/analytics/TiktokAdsAnalytics.vue` (refactor layout)
- `frontend/src/views/analytics/ShopeeAdsAnalytics.vue` (refactor layout)
- New dashboard grid components

---

## 📝 TODO: PYTHON ML

### Phase 12: FastAPI Wrapper (Future)

**File:** `notebooks/api/main.py`

```python
from fastapi import FastAPI
from tiktok_ads.intelligence import ComprehensiveAnalysisEngine

app = FastAPI()

@app.post("/api/ml/tiktok-ads/generate-report")
async def generate_report(tenant_id: str):
    engine = ComprehensiveAnalysisEngine(...)
    report = engine.analyzeAll()
    return HTMLReportGenerator().generate(report)
```

---

## 🧪 TESTING CHECKLIST

### Backend Tests

- [ ] Upload Shopee CSV - success case
- [ ] Upload Shopee CSV - invalid file
- [ ] Upload Shopee CSV - large file (>50MB)
- [ ] Upload TikTok Excel - success case
- [ ] Parse Indonesian CSV encoding (UTF-8 BOM)
- [ ] ML report generation via subprocess
- [ ] Cursor pagination - 10k rows
- [ ] Cache TTL expiration

### Frontend Tests

- [ ] Virtual scroll - smooth 60fps dengan 10k rows
- [ ] Chart rendering - all chart types
- [ ] Infinite scroll - load more data
- [ ] Report gallery - list dan download
- [ ] Upload progress - real-time feedback

### Integration Tests

- [ ] End-to-end upload flow
- [ ] Dashboard metrics calculation
- [ ] ML report display
- [ ] Multi-tenant isolation

---

## 🐛 CLEANUP CHECKLIST

- [ ] Remove duplicate code
- [ ] All files < 300 lines
- [ ] No unused imports
- [ ] No console.log in production
- [ ] snake_case for JSON responses
- [ ] Proper error handling
- [ ] Structured logging (no fmt.Printf)

---

## 🚀 DEPLOYMENT

### Docker Build

```bash
# Backend
docker compose up -d --build backend-go

# Frontend
docker compose up -d --build frontend

# Python ML (cron job)
docker compose up -d python-ml-cron
```

### Verification (2 Bukti Wajib)

**Bukti 1: Docker Logs**

```bash
docker logs backend-go --tail 100 | grep "Uploaded"
# Expected: "Uploaded 247 products from shopee_ads_2026-01-15.csv"
```

**Bukti 2: API Test**

```bash
curl -X POST http://localhost:3000/api/analytics/shopee-ads/upload \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-Tenant-ID: yumna_bertigamart" \
  -F "file=@test_data.csv"

# Expected: {"success":true,"data":{"total_rows":247,...}}
```

---

## 📅 TIMELINE ESTIMATE

- **Week 1:** Backend Upload + ML Service (Phase 1-4)
- **Week 2:** Backend Optimization + Frontend Setup (Phase 5-7)
- **Week 3:** Frontend Charts + Virtual Table (Phase 8-9)
- **Week 4:** AI Gallery + Testing + Cleanup (Phase 10-11 + Testing)
- **Week 5:** Deployment + Documentation

---

## 🔗 REFERENCES

- AGENTS.md: Clean Code guidelines
- README.md: Workflow requirements
- Frontend composables: useShopeeAdsAnalytics.ts, useTiktokAdsAnalytics.ts
- Backend models: backend/internal/models/ads.go
- Python ML: notebooks/tiktok_ads/intelligence/
