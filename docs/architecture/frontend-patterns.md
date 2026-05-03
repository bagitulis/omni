---
last_updated: 2026-05-03
updated_by: agent
relates_to: frontend/src/
stale_if_changed:
  - frontend/src/api/client.ts
  - frontend/src/stores/
  - frontend/src/hooks/
  - frontend/src/App.tsx
---

# Frontend Architecture Patterns

> **Stack**: React 19 + Vite 6 + Ant Design 5 + TypeScript + Zustand + TanStack Query
> **Location**: `frontend/src/`

---

## Quick Decision Tree

| Need | Use | Location |
|------|-----|----------|
| Server data (API responses) | TanStack Query | `hooks/useX.ts` |
| Global UI state (sidebar, tabs) | Zustand store | `stores/xStore.ts` |
| Local component state | `useState` | Component itself |
| API calls | Axios client | `api/x.ts` |
| Types | TypeScript interfaces | `types/x.ts` |

---

## Data Flow

```
User Action → Component → Hook (TanStack Query) → API Client → Backend
                                    ↓
                              Cache (auto-managed)
                                    ↓
                              Zustand Store (only if global UI state needed)
```

---

## 1. API Client (`api/client.ts`)

Singleton Axios instance with:
- **JWT injection**: Reads token from `authStore`, adds `Authorization: Bearer` header
- **Tenant header**: Adds `x-tenant-id` from auth state
- **CSRF token**: Handles double-submit cookie for mutations
- **Error handling**: Centralized via `clientErrorHandler.ts`
- **Dev logging**: Request/response logging in development

```typescript
// Usage in API functions
const response = await apiClient.get<OrderListResponse>("/orders", { params });
const response = await apiClient.post<CreateResult>("/products", data);
```

**Response format** (matches backend):
```typescript
interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}
```

---

## 2. API Functions (`api/*.ts`)

One file per domain. Each function:
1. Calls `apiClient` with typed response
2. Checks `success` field
3. Throws on error
4. Returns typed data

```typescript
// api/orders.ts
export async function getOrders(params: GetOrdersParams): Promise<OrderListResponse> {
  const response = await apiClient.get<OrderListResponse>("/orders", { params });
  if (!response.success) throw new Error(response.error);
  return response.data!;
}
```

**Files**: `orders.ts`, `products.ts`, `inventory.ts`, `auth.ts`, `settings.ts`, etc.

---

## 3. TanStack Query Hooks (`hooks/*.ts`)

Server state management with automatic caching, refetching, and invalidation.

**Query pattern** (read data):
```typescript
export function useOrders(params: GetOrdersParams) {
  return useQuery({
    queryKey: ["orders", params],  // Cache key includes params
    queryFn: () => getOrders(params),
    staleTime: 0,
  });
}
```

**Mutation pattern** (write data):
```typescript
export function useOrderActions() {
  const queryClient = useQueryClient();
  
  const shipMutation = useMutation({
    mutationFn: (orderSns: string[]) => bulkShipOrders(orderSns),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["orders"] });
      message.success("Orders shipped");
    },
  });
  
  return { shipOrders: shipMutation.mutateAsync, isShipping: shipMutation.isPending };
}
```

**Query client config** (`api/queryClient.ts`):
- `staleTime: 5 minutes`
- `retry: 1`
- `refetchOnWindowFocus: false`

---

## 4. Zustand Stores (`stores/*.ts`)

Client-only state (UI preferences, not server data).

```typescript
// stores/appStore.ts
import { create } from "zustand";

export const useAppStore = create<AppState>((set) => ({
  activeTab: "order-management",
  sidebarCollapsed: false,
  setActiveTab: (tab) => set({ activeTab: tab }),
  toggleSidebar: () => set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),
}));
```

**Key stores**:
- `appStore.ts` — Active tab, sidebar, platform selection
- `authStore.ts` — User, token, tenant_id, login/logout
- `notificationStore.ts` — Toast notifications
- `modalsStore.ts` — Modal visibility

---

## 5. Types (`types/*.ts`)

**CRITICAL: All API types use `snake_case`** to match Go backend JSON.

```typescript
// ✅ CORRECT
interface Order {
  order_sn: string;
  created_at: string;
  total_amount: number;
  buyer_username: string;
}

// ❌ WRONG — will break at runtime
interface Order {
  orderSn: string;    // Backend sends order_sn
  createdAt: string;  // Backend sends created_at
}
```

---

## 6. Router (`App.tsx`)

React Router with lazy-loaded pages:

```typescript
const OrdersPage = React.lazy(() => import("./pages/orders/OrdersPage"));

<Routes>
  <Route path="/login" element={<LoginPage />} />
  <Route element={<ProtectedRoute><AppLayout /></ProtectedRoute>}>
    <Route path="/" element={<DashboardPage />} />
    <Route path="/order-manager" element={<OrdersPage />} />
    <Route path="/products" element={<UnifiedProductsPage />} />
    <Route path="/inventory" element={<SimplifiedInventoryPage />} />
    <Route path="/settings" element={<SettingsPage />} />
  </Route>
</Routes>
```

---

## 7. Page Component Pattern

```typescript
export default function OrdersPage() {
  // 1. Custom hook for all logic
  const { state, handlers } = useOrdersLogic();
  
  // 2. Responsive breakpoints
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;
  
  // 3. Render with Ant Design
  return (
    <div style={{ padding: isMobile ? 12 : 24 }}>
      <OrderHeader {...state} />
      <OrderTable data={state.data} onShip={handlers.handleShip} />
    </div>
  );
}
```

**Rules**:
- Pages are thin — logic lives in hooks
- Use Ant Design Grid breakpoints for responsive
- Lazy-load all pages

---

## 8. Adding a New Feature (Step-by-Step)

### Step 1: Types
```
frontend/src/types/myfeature.ts
```

### Step 2: API Functions
```
frontend/src/api/myfeature.ts
```

### Step 3: Query Hook
```
frontend/src/hooks/useMyFeature.ts
```

### Step 4: Page Component
```
frontend/src/pages/myfeature/MyFeaturePage.tsx
```

### Step 5: Register Route
```
frontend/src/App.tsx → add <Route path="/my-feature" ... />
```

### Step 6: (Optional) Zustand Store
Only if you need global UI state that persists across pages.

---

## 9. Directory Conventions

| Directory | Casing | Contents |
|-----------|--------|----------|
| `components/modals/` | **lowercase** | Modal components |
| `components/tables/` | **lowercase** | Table components |
| `components/forms/` | **lowercase** | Form components |
| `components/layout/` | **lowercase** | Layout components |
| `components/ui/` | **lowercase** | Extended Ant Design |
| `pages/orders/` | **lowercase** | Page components (PascalCase files) |

**CRITICAL (Windows)**: Always use lowercase directory names. Uppercase causes TS1261 build errors.

---

## 10. Key Conventions Summary

| Convention | Rule |
|-----------|------|
| API types | `snake_case` (match backend) |
| Component files | `PascalCase.tsx` |
| Hook files | `camelCase.ts` (prefix `use`) |
| Store files | `camelCase.ts` (suffix `Store`) |
| Directories | **lowercase** always |
| Colors | Theme tokens only (no hardcoded) |
| Font | System fonts, 12px base |
| Border radius | 3px (sharp, Ginee-style) |
| Spacing | 4px grid |
