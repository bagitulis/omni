# Analysis: Draft Status Behavior for Imported Master Products

## Executive Summary

Imported Master Products are intentionally set to **"draft" status** as part of a review workflow, not due to a bug. This analysis explains why, what options exist, and provides a recommendation.

---

## Current Behavior

### What Happens When You Import a Product

1. **User imports** a Shopee product via `POST /api/master-products/import`
2. **Backend fetches** product data from Shopee (title, description, images, SKUs)
3. **Master Product is created** with:
   - ✅ Title, description, images from Shopee
   - ✅ All SKU variants created
   - ✅ Platform link to Shopee created
   - ✅ Sync status marked as "synced"
   - ❌ **Product status set to "draft"** (not "active")
4. **Result**: Product appears in system but needs activation

### Code Location

**File:** `backend/internal/services/master_product/import_service.go`  
**Line:** 171

```go
masterProduct := &models.MasterProduct{
    // ... other fields ...
    Status: models.MasterProductStatusDraft,  // ← Always set to draft
}
```

### Available Status Options

From `backend/internal/models/master_product.go`:

```go
const (
    MasterProductStatusDraft    = "draft"      // Review state
    MasterProductStatusActive   = "active"     // Published state
    MasterProductStatusArchived = "archived"   // Deleted/hidden state
)
```

---

## Why It's Set to Draft (INTENTIONAL DESIGN)

This is **not a bug** — it's an intentional design pattern:

### 1. Review Workflow
- Allows users to inspect imported data before activation
- Users can verify title, description, images are correct
- Prevents bad data from being published to all platforms

### 2. Data Validation Safety
During import, certain data gets transformed:
- SKUs without `seller_sku` are skipped (data quality issue)
- Images are truncated to 8 (Lazada platform limit)
- Descriptions truncated to 5000 characters (Shopee platform limit)

**User should verify** these transformations didn't lose critical data.

### 3. Testing Ground for Sync
- Product is "draft" BUT can still be synced to TikTok/Lazada
- Allows testing sync before public activation
- Users can verify sync works correctly on other platforms first

### 4. Aligns with Best Practices
- **Shopify**: Import → Review → Publish
- **TikTok Seller Center**: Draft → Preview → Activate
- **Amazon Seller Central**: Create → Save as draft → Publish

---

## Available Options

### Option A: Keep As-Is (Current Implementation)

**Behavior:** Products stay in draft state indefinitely until user manually changes status

**Pros:**
- ✅ Mature workflow, used by all major e-commerce platforms
- ✅ Allows thorough review before activation
- ✅ Prevents accidental publication of bad data
- ✅ Can test sync on other platforms first
- ✅ No implementation work needed

**Cons:**
- ❌ Products don't appear in main product list
- ❌ Requires manual status change to activate
- ❌ May confuse users expecting immediate visibility

**Best For:** Users who want complete control and data verification

---

### Option B: Auto-Activate on Import

**Behavior:** Imported products immediately go to "active" status

**Implementation:** Change line 171 from `models.MasterProductStatusDraft` to `models.MasterProductStatusActive`

**Pros:**
- ✅ Products immediately visible and usable
- ✅ Faster workflow, no extra steps
- ✅ Simpler mental model for users

**Cons:**
- ❌ Bypasses review step — risk of publishing bad data
- ❌ Skipped SKUs still get imported (no chance to fix)
- ❌ Hard to rollback if data is wrong
- ❌ Goes against industry best practices
- ❌ No opportunity to test sync first

**Best For:** Users who trust the import process is always correct

---

### Option C: Add "Activate" Button in UI (RECOMMENDED)

**Behavior:** Products import as draft, but user can activate with one click

**Implementation:**
1. Add API endpoint: `POST /api/master-products/:id/activate`
2. Endpoint changes status: "draft" → "active"
3. Add "Activate" button in UI product list/detail
4. Show count: "3 draft products ready to activate"

**Pros:**
- ✅ Combines best of both: review + fast activation
- ✅ Matches user expectations (import shows up, then activate)
- ✅ One-click activation when ready
- ✅ Professional, enterprise-grade UX
- ✅ Provides audit trail (when imported vs. when activated)

**Cons:**
- ❌ Requires API endpoint development
- ❌ Requires UI changes
- ❌ More code to maintain

**Best For:** Professional, user-friendly experience (Recommended)

---

## Recommendation: Use Option C

### Why This Is Best

1. **User Expectations**
   - People expect imported products to "show up" in the system
   - They also want a review step before it goes live
   - One "Activate" button satisfies both

2. **Data Quality**
   - Review workflow catches data issues
   - Users can edit before activation if needed
   - Prevents bad data syncing to all platforms

3. **Industry Standard**
   - Every major e-commerce platform uses this pattern
   - Users are already familiar with it
   - Shows Omni is professionally designed

4. **Implementation Effort**
   - Small addition to existing code
   - One new API endpoint
   - One UI button
   - Well worth the user experience improvement

### Implementation Roadmap

**Phase 1: Documentation (Now)**
- Document the current draft workflow
- Explain why it exists

**Phase 2: API Endpoint (Next Sprint)**
- Add `ActivateProduct()` handler in `import_handler.go`
- Route: `POST /api/master-products/:id/activate`
- Changes status: "draft" → "active"
- Logs activation timestamp

**Phase 3: UI Enhancement (Following Sprint)**
- Show "Draft" badge on products in list
- Add "Activate" button on draft products
- Show count in sidebar: "3 draft products"
- Show toast notification on successful activation

---

## Code Locations Reference

| Item | Location |
|------|----------|
| Status constants | `backend/internal/models/master_product.go:72-74` |
| Import service | `backend/internal/services/master_product/import_service.go:171` |
| Import handler | `backend/internal/handlers/master_product/import_handler.go` |
| Frontend service | `frontend/src/services/masterProductService.ts` |
| Product list view | `frontend/src/views/MasterProduct/ProductList.vue` |

---

## Summary Table

| Aspect | Option A (As-Is) | Option B (Auto-Active) | Option C (Activate Button) |
|--------|-----------------|----------------------|---------------------------|
| **Review Workflow** | ✅ Full review possible | ❌ No review | ✅ Full review possible |
| **Immediate Visibility** | ❌ Must activate first | ✅ Instant | ✅ Optional activate |
| **Data Quality** | ✅ High (user reviews) | ❌ Risk of bad data | ✅ High (user reviews) |
| **User Experience** | 🟡 Professional but manual | ✅ Fast but risky | ✅✅ Professional + Fast |
| **Implementation Effort** | ✅ None (0 hours) | ✅ Minimal (15 min) | 🟡 Medium (4-6 hours) |
| **Recommendation** | 🟡 If no dev time | ❌ Not recommended | ✅✅ Recommended |

---

## Questions for User

Before implementation, clarify:

1. **User expectation**: Do users expect imported products to be immediately visible or require review?
2. **Data trust**: How much do we trust the Shopee import data? Do users typically review before activating?
3. **Activation friction**: Is the manual activation step a pain point for users?
4. **Timeline**: When is this feature needed? (Helps prioritize Option C implementation)

---

## Next Steps

1. ✅ **Complete:** Analysis done and documented
2. **TODO:** Review with product team and get decision on Option A/B/C
3. **TODO:** If Option C chosen, create implementation task in backlog
4. **TODO:** Add status filtering in product list (show only active, or show drafts separately)

