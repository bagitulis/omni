# Frontend Development Context

> Auto-injected when working in `frontend/` or `frontend-react/` directory.
> **Parent rules:** See root `AGENTS.md` for critical rules.

---

## Active Stacks

| Directory         | Stack                          | Status     |
| ----------------- | ------------------------------ | ---------- |
| `frontend/`       | Vue 3 + Composition API + Vite | Legacy     |
| `frontend-react/` | React 19 + Ant Design + Vite   | **ACTIVE** |

**IMPORTANT:** New development should target `frontend-react/`. Vue app is reference only.

---

## React Stack (frontend-react/) - PRIMARY

- React 19
- TypeScript 5.7+
- Vite 6
- Ant Design 5
- Zustand (client state)
- TanStack Query (server state)
- React Router 7

### React Structure

```
frontend-react/
├── src/
│   ├── api/             # API client functions
│   ├── components/      # Reusable components
│   │   ├── layout/      # AppLayout, Sidebar, Header, MobileNav
│   │   ├── ui/          # Extended Ant Design
│   │   ├── tables/      # Data tables
│   │   ├── forms/       # Form components
│   │   └── modals/      # Modal components
│   ├── pages/           # Route pages
│   ├── hooks/           # Custom hooks
│   ├── stores/          # Zustand stores
│   ├── types/           # TypeScript interfaces (snake_case!)
│   ├── lib/             # Utilities
│   └── styles/          # Theme and global CSS
└── public/
```

### React Design System

| Property          | Value                    |
| ----------------- | ------------------------ |
| Primary color     | `#0369a1` (sky-700)      |
| Base font size    | 12px                     |
| Border radius     | 3px (sharp, Ginee-style) |
| Control height    | 32px                     |
| Sidebar expanded  | 220px                    |
| Sidebar collapsed | 80px                     |
| Header height     | 48px                     |
| Mobile nav height | 56px                     |

### React Anti-Patterns

| Forbidden              | Do Instead                    |
| ---------------------- | ----------------------------- |
| Google Fonts           | System fonts only             |
| Nested submenus        | Flat navigation (Ginee-style) |
| camelCase API types    | snake_case to match backend   |
| Hardcoded colors       | Use theme tokens              |
| Border radius > 6px    | Use 3px (sharp corners)       |
| `as any`, `@ts-ignore` | Fix types properly            |

---

## Vue Stack (frontend/) - LEGACY REFERENCE

- Vue 3 + Composition API
- TypeScript
- Vite
- TailwindCSS
- Pinia (state management)

### Vue Structure

```
frontend/
├── src/
│   ├── components/      # Reusable Vue components
│   ├── views/           # Page views
│   ├── api/             # API client functions
│   ├── stores/          # Pinia stores
│   ├── composables/     # Reusable composition functions
│   └── types/           # TypeScript type definitions
└── public/
```

---

## Naming Conventions (BOTH STACKS)

| Type              | Convention        | Example                     |
| ----------------- | ----------------- | --------------------------- |
| Components        | PascalCase        | `OrderTable.vue/.tsx`       |
| Composables/Hooks | camelCase + `use` | `useOrders.ts`              |
| API types         | **snake_case**    | `order_sn` (match backend!) |
| Local vars        | camelCase         | `orderList`                 |
| Props (Vue)       | kebab-case        | `:order-id="id"`            |

---

## CRITICAL: API Type Matching

Backend returns snake_case JSON. Frontend types **MUST** match exactly:

```typescript
// ✅ CORRECT - matches backend response
interface Order {
  order_sn: string;
  created_at: string;
  total_amount: number;
  buyer_username: string;
}

// ❌ WRONG - mismatched naming (will break!)
interface Order {
  orderSn: string; // Backend sends order_sn
  createdAt: string; // Backend sends created_at
}
```

---

## API Calls

Location: `src/api/`

```typescript
// Always type API functions
export async function getOrders(tenantId: string): Promise<Order[]> {
  const response = await api.get("/orders", {
    headers: { "X-Tenant-ID": tenantId },
  });
  return response.data.data;
}
```

---

## Error Handling

```typescript
try {
  const data = await fetchOrders();
} catch (error) {
  // Show user-friendly message
  toast.error("Failed to load orders");
  // Log for debugging
  console.error("Order fetch failed:", error);
}

// ❌ NEVER empty catch
// catch (e) { } ← FORBIDDEN
```

---

## Backend API Response Format

Backend always returns:

```json
{
  "success": true,
  "data": { ... }
}
```

Or on error:

```json
{
  "success": false,
  "error": "error message"
}
```

Handle both cases:

```typescript
const response = await api.get("/orders");
if (!response.data.success) {
  throw new Error(response.data.error);
}
return response.data.data;
```

---

## Migration Reference

When implementing React pages, reference Vue equivalents:

| React Page          | Vue Reference                         |
| ------------------- | ------------------------------------- |
| DashboardPage.tsx   | `views/Dashboard.vue`                 |
| OrdersPage.tsx      | `components/OrderManager/*.vue`       |
| ProductListPage.tsx | `views/MasterProduct/ProductList.vue` |
| ProductEditPage.tsx | `views/MasterProduct/ProductEdit.vue` |
| InventoryPage.tsx   | `views/Inventory.vue`                 |

---

## Anti-Patterns (BOTH STACKS)

| Forbidden             | Do Instead                          |
| --------------------- | ----------------------------------- |
| camelCase API types   | Use snake_case to match backend     |
| Untyped API calls     | Always define TypeScript interfaces |
| Empty catch blocks    | Show error to user + log            |
| Props without type    | Use `defineProps<T>()` / interface  |
| Direct store mutation | Use actions/composables             |
