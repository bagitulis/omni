# Frontend Composables - Development Rules

> **Parent:** `frontend/AGENTS.md` | **Files:** 55 TypeScript files

---

## Structure (Current: Flat)

```
composables/
├── Dashboard/         # useDashboardHandlers, useDashboardNavigation, useDashboardPlatformOps
├── Analytics/         # useAnalytics, useMLAnalytics, useMLReports, useAdsReports
├── Product/           # useProductCreate, useProductManager, useProductSync
├── Routing/           # useRouteManager, useRouteFlowController, useRouteState
├── Sheets/            # useGoogleSheetsLinks, useGoogleSheetsSettings
├── UI/                # useRightSidebarStatus, usePageHeader, useTheme
└── *.ts               # All 55 composables currently flat
```

---

## Composable Pattern (MANDATORY)

```typescript
// ✅ CORRECT - Vue 3 Composition API
import { ref, computed, onMounted, onUnmounted } from "vue";

export function useOrders(tenantId: string) {
  const orders = ref<Order[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const pendingOrders = computed(() =>
    orders.value.filter((o) => o.status === "pending"),
  );

  async function fetchOrders() {
    loading.value = true;
    try {
      orders.value = await api.getOrders(tenantId);
    } catch (e) {
      error.value = e.message;
    } finally {
      loading.value = false;
    }
  }

  onMounted(() => fetchOrders());

  return {
    orders,
    loading,
    error,
    pendingOrders,
    fetchOrders,
  };
}

// ❌ WRONG - Options API style
export default {
  data() {
    return { orders: [] };
  },
};
```

---

## Naming Convention (MANDATORY)

```typescript
// ✅ CORRECT - Always prefix with "use"
export function useOrders() {}
export function useProductSync() {}
export function useDashboardHandlers() {}

// ❌ WRONG - Missing "use" prefix
export function getOrders() {}
export function orderHelpers() {}
export const orders = {};
```

---

## API Type Matching (MANDATORY)

```typescript
// ✅ CORRECT - Match backend snake_case
interface Order {
  order_sn: string;
  created_at: string;
  total_amount: number;
  tenant_id: string;
}

// ❌ WRONG - camelCase (doesn't match backend)
interface Order {
  orderSn: string;
  createdAt: string;
}
```

---

## Composable Categories

| Category              | Pattern                  | Files                                            |
| --------------------- | ------------------------ | ------------------------------------------------ |
| **Data Fetching**     | `use{Resource}`          | `useOrders`, `useProducts`                       |
| **State Management**  | `use{Feature}State`      | `useRouteState`, `useFilterState`                |
| **Operations**        | `use{Feature}Actions`    | `useBatchSkuCheck`, `usePriceUpdate`             |
| **UI Logic**          | `use{Component}Logic`    | `useScriptMonitorLogic`, `usePageHeader`         |
| **Platform-Specific** | `use{Platform}{Feature}` | `useShopeeAdsAnalytics`, `useTiktokAdsAnalytics` |

---

## Return Value Convention

```typescript
// ✅ CORRECT - Always return object with named exports
export function useOrders() {
  return {
    // State
    orders,
    loading,
    error,
    // Computed
    pendingOrders,
    // Methods
    fetchOrders,
    refreshOrders,
  };
}

// ❌ WRONG - Return array (hard to destructure)
export function useOrders() {
  return [orders, loading, fetchOrders];
}
```

---

## Error Handling

```typescript
// ✅ CORRECT - User-friendly + debug log
try {
  const data = await fetchOrders()
} catch (error) {
  toast.error('Failed to load orders')
  console.error('Order fetch failed:', error)
  errorState.value = error.message
}

// ❌ WRONG - Silent failure
try {
  await fetchOrders()
} catch (e) {}

// ❌ WRONG - Raw error to user
catch (error) {
  toast.error(error.message)  // May expose internal details
}
```

---

## Cleanup Pattern

```typescript
// ✅ CORRECT - Clean up subscriptions/timers
export function usePolling() {
  const intervalId = ref<number | null>(null);

  onMounted(() => {
    intervalId.value = setInterval(poll, 5000);
  });

  onUnmounted(() => {
    if (intervalId.value) {
      clearInterval(intervalId.value);
    }
  });
}
```

---

## Anti-Patterns (FORBIDDEN)

```typescript
// ❌ WRONG - Direct API calls in components
// Put in composable instead!

// ❌ WRONG - Mutating props
props.orders.push(newOrder);

// ❌ WRONG - Global mutable state
let globalOrders = [];

// ❌ WRONG - Mixing camelCase and snake_case
interface Order {
  order_sn: string;
  totalAmount: number; // Pick one!
}

// ❌ WRONG - Heavy logic in computed
const processedOrders = computed(() => {
  // 50 lines of complex transformation...
});
// Extract to method or separate composable
```

---

## Max Lines Per File

| Type             | Limit     |
| ---------------- | --------- |
| Composable files | 300 lines |

**If exceeding:** Split into smaller composables by concern.
