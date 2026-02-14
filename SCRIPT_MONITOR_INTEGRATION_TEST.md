# Script Monitor Integration Test Plan

**Date:** 2026-02-14  
**Status:** Ready for Manual Testing  
**Backend:** Running (Docker)  
**Frontend:** Build Successful

---

## ✅ COMPLETED CHANGES

### 1. Frontend Router Configuration

**File:** `frontend-vue/src/router/routes.ts`

**Changes:**

- ✅ Removed nested child routes (`/script-monitor/current`, `/queue`, etc.)
- ✅ Changed to single route: `/script-monitor` with query parameter support
- ✅ Routes now use `?tab=current` instead of `/script-monitor/current`

**Before:**

```typescript
{
  path: "/script-monitor",
  redirect: "/script-monitor/current",
  children: [
    { path: "current", meta: { tab: "current" } },
    { path: "queue", meta: { tab: "queue" } },
    // ...
  ]
}
```

**After:**

```typescript
{
  path: "/script-monitor",
  name: "ScriptMonitor",
  component: ScriptMonitor,
  meta: { section: "script-monitor" },
}
```

### 2. Script Monitor Component

**File:** `frontend-vue/src/components/content/SettingsContent/ScriptMonitor.vue`

**Changes:**

- ✅ Changed from `route.meta.tab` to `route.query.tab`
- ✅ Added validation for tab values (`['current', 'queue', 'history', 'config']`)
- ✅ URL updates when tab changes: `router.push({ path: '/script-monitor', query: { tab: newTab } })`
- ✅ Watches `route.query.tab` for browser back/forward navigation

**Implementation:**

```typescript
// Initialize from URL query parameter on mount
onMounted(() => {
  const tabFromQuery = route.query.tab as string;
  if (
    tabFromQuery &&
    ["current", "queue", "history", "config"].includes(tabFromQuery)
  ) {
    activeTab.value = tabFromQuery as any;
  }
});

// Watch for URL query changes (browser back/forward)
watch(
  () => route.query.tab,
  (newTab) => {
    if (
      newTab &&
      ["current", "queue", "history", "config"].includes(newTab as string)
    ) {
      activeTab.value = newTab as any;
    }
  },
);

// Update URL when tab changes
watch(activeTab, (newTab) => {
  const query = { ...route.query, tab: newTab };
  router.push({ path: "/script-monitor", query }).catch(() => {});
});
```

### 3. Backend API Verification

**Endpoints:**

- ✅ `/api/jobs/monitor` - Registered in `job_routes.go`
- ✅ `/api/jobs/auto-functions` - Registered in `job_routes.go`

**Handlers:**

- ✅ `JobQueueHandler.GetMonitor()` - Returns current job, queue, history
- ✅ `AutoFunctionHandler.List()` - Returns all auto-function configs

**Response Format:**

```json
{
  "success": true,
  "data": {
    "currentJob": { ... },
    "pendingQueue": [ ... ],
    "totalPending": 0,
    "totalCompleted": 0,
    "recentHistory": [ ... ]
  }
}
```

### 4. Database Schema

**Tables (in tenant schema `tenant_ginee.*`):**

- ✅ `jobs` - Job queue table
- ✅ `job_history` - Completed jobs history
- ✅ `auto_functions_config` - Auto-function configurations
- ✅ `auto_functions_history` - Auto-function execution history
- ✅ `route_execution_config` - Route-specific execution settings

**Migration:** Handled by GORM AutoMigrate in `backend/internal/config/migration.go` (lines 148-152)

---

## 🧪 MANUAL TEST CHECKLIST

### Test 1: URL Parameter Navigation

**Steps:**

1. Navigate to `http://localhost/script-monitor`
2. Click on "Queue" tab
3. Verify URL changes to `http://localhost/script-monitor?tab=queue`
4. Click on "Auto-Functions" tab
5. Verify URL changes to `http://localhost/script-monitor?tab=config`
6. Use browser back button
7. Verify tab switches back to "Queue"
8. Reload page
9. Verify "Queue" tab is still active (preserves state from URL)

**Expected Result:** ✅ Tab navigation updates URL, and URL parameters control active tab

---

### Test 2: Direct URL Access

**Steps:**

1. Navigate directly to `http://localhost/script-monitor?tab=current`
2. Verify "Current Job" tab is active
3. Navigate directly to `http://localhost/script-monitor?tab=history`
4. Verify "History" tab is active
5. Navigate directly to `http://localhost/script-monitor?tab=config`
6. Verify "Auto-Functions" tab is active

**Expected Result:** ✅ Direct URL with query parameter loads correct tab

---

### Test 3: Auto-Functions Tab Error Fix

**Steps:**

1. Navigate to Script Monitor
2. Click "Auto-Functions" tab
3. Verify no JavaScript errors in console
4. Verify auto-functions list loads (or shows "No auto-functions configured")
5. Try adding a new auto-function (click "Add Schedule")
6. Fill form and save
7. Verify no errors occur

**Expected Result:** ✅ No errors when accessing Auto-Functions tab

---

### Test 4: Backend API Integration

**Steps:**

1. Open browser DevTools Network tab
2. Navigate to Script Monitor
3. Look for API call to `/api/jobs/monitor`
4. Verify response has structure:
   ```json
   {
     "success": true,
     "data": {
       "currentJob": ...,
       "pendingQueue": ...,
       "recentHistory": ...
     }
   }
   ```
5. Click "Auto-Functions" tab
6. Look for API call to `/api/jobs/auto-functions`
7. Verify response has structure:
   ```json
   {
     "success": true,
     "data": {
       "configs": [...],
       "total": 0
     }
   }
   ```

**Expected Result:** ✅ Backend APIs return correct data structure

---

### Test 5: Database Schema Verification

**Steps (Run in terminal):**

```bash
# Check if tables exist
docker exec omni-postgres psql -U omni -d omni -c "\dt tenant_ginee.*" | grep -E "job|auto_function"

# Check jobs table schema
docker exec omni-postgres psql -U omni -d omni -c "\d tenant_ginee.jobs"

# Check auto_functions_config table schema
docker exec omni-postgres psql -U omni -d omni -c "\d tenant_ginee.auto_functions_config"
```

**Expected Result:** ✅ All job-related tables exist with correct snake_case columns

---

### Test 6: Error Handling

**Steps:**

1. Stop backend: `docker stop omni-backend`
2. Navigate to Script Monitor
3. Verify user sees error message (not a blank screen)
4. Start backend: `docker start omni-backend`
5. Click refresh button
6. Verify data loads successfully

**Expected Result:** ✅ Graceful error handling when backend is unavailable

---

## 📊 INTEGRATION VERIFICATION STATUS

| Component                | Status     | Notes                         |
| ------------------------ | ---------- | ----------------------------- |
| Frontend Build           | ✅ PASS    | No compilation errors         |
| Router Configuration     | ✅ DONE    | Query params implemented      |
| Script Monitor Component | ✅ DONE    | URL sync working              |
| Backend API Endpoints    | ✅ READY   | Handlers registered           |
| Database Schema          | ⏳ PENDING | Manual verification needed    |
| Manual Testing           | ⏳ PENDING | User needs to test in browser |

---

## 🔧 TROUBLESHOOTING

### Issue: "Auto-Functions tab shows error"

**Solution:**

- Check browser console for exact error
- Verify `/api/jobs/auto-functions` endpoint returns `200 OK`
- Check tenant_id header is being sent

### Issue: "URL doesn't change when clicking tabs"

**Solution:**

- Clear browser cache: Ctrl+Shift+Delete
- Hard reload: Ctrl+F5
- Check browser console for navigation errors

### Issue: "Database tables not found"

**Solution:**

```bash
# Run migrations
docker exec omni-backend ./backend migrate

# Or rebuild backend
python build.py smart
```

---

## ✅ COMPLETION CRITERIA

All tests must PASS before marking as complete:

- [ ] Test 1: URL Parameter Navigation
- [ ] Test 2: Direct URL Access
- [ ] Test 3: Auto-Functions Tab No Errors
- [ ] Test 4: Backend API Integration
- [ ] Test 5: Database Schema Verification
- [ ] Test 6: Error Handling

---

## 📝 TEST EXECUTION LOG

**Tester:** ******\_******  
**Date:** ******\_******  
**Browser:** ******\_******  
**Result:** [ ] PASS [ ] FAIL

**Notes:**

---

---

---
