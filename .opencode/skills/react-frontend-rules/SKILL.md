---
description: React frontend rules for OMNI React migration - Ant Design + Ginee-style UI
---

# React Frontend Rules

> **Role:** Rules for agents implementing React frontend (frontend/).
> **Stack:** React 19 + Vite 6 + Ant Design 5 + TypeScript + Zustand + TanStack Query

---

## ⚠️ CRITICAL PRINCIPLES (READ FIRST)

### 1. Backend Integration is MANDATORY

- Frontend MUST integrate with existing Go backend API
- API endpoints: `http://localhost:3000/api/*`
- Response format: `{ success: true, data: {...} }` or `{ success: false, error: "..." }`
- All API types use **snake_case** to match backend exactly
- Test API integration before marking task complete

### 2. Layout Freedom with Consistency

- Layout can be **completely redesigned** - AI recommendations welcome
- DO NOT copy Vue layout - create fresh, modern UI
- **BUT** must maintain consistency:
  - Same fonts throughout (system fonts, 12px base)
  - Same color tokens (primary #0369a1)
  - Same spacing system (4px grid)
  - Same border radius (3px sharp corners)
  - Same control heights (32px)

### 3. Directory Casing (CRITICAL - Windows)

> **RECURRING BUG**: `components/modals/` vs `components/Modals/` causes TS1261 build failures.

- React frontend uses **lowercase** directory names: `modals/`, `tables/`, `forms/`, `layout/`, `ui/`
- Vue legacy uses **uppercase** `Modals/` — this is intentional and must NOT be changed
- On Windows (`core.ignorecase=true`), the OS may silently flip `modals/` → `Modals/`
- **ALWAYS use lowercase** when creating/importing files: `@/components/modals/SomeModal`
- **NEVER** use `@/components/Modals/` in React frontend imports
- If build fails with TS1261 casing error: rename directory via temp name (`modals` → `_temp` → `modals`)

### 4. Font Consistency (CRITICAL)

```css
/* ONLY use this font stack - NO Google Fonts */
font-family:
  system-ui,
  -apple-system,
  BlinkMacSystemFont,
  "Segoe UI",
  Roboto,
  sans-serif;

/* Font sizes must be consistent */
--font-xs: 10px; /* Badges, captions */
--font-sm: 12px; /* DEFAULT - body, tables */
--font-base: 14px; /* Emphasized text */
--font-lg: 16px; /* Subheadings */
--font-xl: 18px; /* Section titles */
--font-2xl: 20px; /* Page titles */
```

**NEVER mix font sizes randomly. Use the scale above.**

---

## Project Structure

```
frontend/
├── src/
│   ├── api/                    # API client functions
│   │   ├── client.ts           # Axios instance
│   │   ├── auth.ts
│   │   ├── orders.ts
│   │   ├── products.ts
│   │   └── inventory.ts
│   │
│   ├── components/             # Reusable components
│   │   ├── layout/             # AppLayout, Sidebar, Header, MobileNav
│   │   ├── ui/                 # Extended Ant Design components
│   │   ├── tables/             # Data tables
│   │   ├── forms/              # Form components
│   │   └── modals/             # Modal components
│   │
│   ├── pages/                  # Route pages
│   │   ├── auth/
│   │   ├── dashboard/
│   │   ├── orders/
│   │   ├── products/
│   │   ├── inventory/
│   │   ├── analytics/
│   │   └── settings/
│   │
│   ├── hooks/                  # Custom hooks
│   ├── stores/                 # Zustand stores
│   ├── types/                  # TypeScript interfaces (snake_case!)
│   ├── lib/                    # Utilities
│   └── styles/                 # Theme and global CSS
```

---

## Design System (CRITICAL)

### Primary Color: Sky Blue

| Token     | Hex       | Usage                 |
| --------- | --------- | --------------------- |
| `sky-700` | `#0369a1` | **PRIMARY** - Buttons |
| `sky-600` | `#0284c7` | Hover states          |
| `sky-800` | `#075985` | Active/pressed states |
| `sky-100` | `#e0f2fe` | Selected backgrounds  |

### Layout Dimensions

| Element           | Value |
| ----------------- | ----- |
| Sidebar expanded  | 220px |
| Sidebar collapsed | 80px  |
| Header height     | 48px  |
| Mobile nav height | 56px  |
| Content padding   | 24px  |
| Border radius     | 3px   |
| Control height    | 32px  |

### Typography

- **Font:** System fonts only (NO Google Fonts)
- **Base size:** 12px (data-dense OMS)
- **Font stack:** `system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif`

### Typography Scale (MUST USE - from Design System)

| Token       | Size | Line Height | Weight | Usage                    |
| ----------- | ---- | ----------- | ------ | ------------------------ |
| `text-xs`   | 10px | 14px        | 400    | Badges, captions         |
| `text-sm`   | 12px | 16px        | 400    | **DEFAULT** body, tables |
| `text-base` | 14px | 20px        | 400    | Emphasized text          |
| `text-lg`   | 16px | 24px        | 500    | Subheadings              |
| `text-xl`   | 18px | 28px        | 600    | Section titles           |
| `text-2xl`  | 20px | 28px        | 600    | Page titles              |
| `text-3xl`  | 24px | 32px        | 600    | Large headings           |

### Font Weights

| Token           | Weight | Usage                  |
| --------------- | ------ | ---------------------- |
| `font-normal`   | 400    | Body text, table cells |
| `font-medium`   | 500    | Labels, menu items     |
| `font-semibold` | 600    | Headings, buttons      |

**CRITICAL: Only use sizes from the scale above. No 13px, 15px, 17px, etc.**

---

## Naming Conventions

| Type       | Convention     | Example                |
| ---------- | -------------- | ---------------------- |
| Components | PascalCase     | `OrderTable.tsx`       |
| Hooks      | camelCase+use  | `useOrders.ts`         |
| API types  | **snake_case** | `order_sn` (CRITICAL!) |
| Local vars | camelCase      | `orderList`            |
| Files      | PascalCase     | `DashboardPage.tsx`    |
| Stores     | camelCase      | `authStore.ts`         |

---

## API Type Matching (CRITICAL!)

Backend returns snake_case JSON. Types **MUST** match exactly:

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

## Component Structure

```tsx
// 1. Imports
import { useState, useEffect } from 'react';
import { Card, Table, Button } from 'antd';
import { useOrders } from '@/hooks/useOrders';

// 2. Types
interface Props {
  orderId: string;
}

// 3. Component
export function OrderCard({ orderId }: Props) {
  // 4. Hooks
  const { data, isLoading } = useOrders();

  // 5. State
  const [selected, setSelected] = useState<string[]>([]);

  // 6. Handlers
  const handleClick = () => { ... };

  // 7. Render
  return (
    <Card>...</Card>
  );
}
```

---

## Ant Design Theme Override

```typescript
// src/styles/theme.ts
export const antdTheme: ThemeConfig = {
  token: {
    colorPrimary: "#0369a1", // sky-700
    colorSuccess: "#16a34a",
    colorWarning: "#d97706",
    colorError: "#dc2626",
    fontSize: 12, // Data-dense
    borderRadius: 3, // Sharp corners (Ginee-style)
    controlHeight: 32,
  },
  components: {
    Table: {
      fontSize: 12,
      cellPaddingBlock: 8,
      headerBg: "#f8fafc",
    },
  },
};
```

---

## Localhost Bypass

Dev mode bypasses authentication on localhost:

```typescript
// In auth guard
const isLocalhost = window.location.hostname === "localhost";
if (isLocalhost) {
  // Skip auth check for AI layout review
  return true;
}
```

---

## API Response Format

Backend always returns:

```json
{ "success": true, "data": { ... } }
// or
{ "success": false, "error": "message" }
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

## Anti-Patterns (FORBIDDEN)

| Forbidden              | Do Instead                      |
| ---------------------- | ------------------------------- |
| camelCase API types    | Use snake_case to match backend |
| `as any`, `@ts-ignore` | Fix types properly              |
| Google Fonts           | Use system fonts only           |
| Hardcoded colors       | Use theme tokens                |
| Next.js / SSR          | SPA only (Vite + React)         |
| Nested submenus        | Flat navigation (Ginee-style)   |
| Empty catch blocks     | Handle or log error             |
| Rounded corners > 6px  | Use 3px (sharp, Ginee-style)    |

---

## File Size Quality Signal (~300 Lines)

> **~300 lines is NOT a hard limit.** It's a quality signal for SRP/DRY/OOP compliance.
> If a code file exceeds ~300 lines, review for violations. Clean code slightly exceeding is OK.

| Type      | Guideline  | Action if Exceeded                 |
| --------- | ---------- | ---------------------------------- |
| Component | ~300 lines | Review — split if SRP/DRY violated |
| Page      | ~300 lines | Review — extract sections/hooks    |
| Hook      | ~200 lines | Review — split complex logic       |
| Store     | ~200 lines | Review — separate concerns         |

---

## References

- Design System: Defined in this document (see Design System section above)
- Vue Legacy Reference: `frontend-vue/src/` (for logic patterns only, NOT layout)
- Ginee Style: Task-first dashboard, flat navigation, data-dense tables

---

## Backend Integration Guide

### API Base Configuration

```typescript
// src/api/client.ts
import axios from "axios";

const api = axios.create({
  baseURL: "/api", // Proxied to localhost:3000
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor - add auth token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem("auth_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor - handle errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized
      localStorage.removeItem("auth_token");
      window.location.href = "/login";
    }
    return Promise.reject(error);
  },
);

export default api;
```

### API Call Pattern

```typescript
// src/api/orders.ts
import api from "./client";
import type { Order, OrderListResponse } from "@/types/order";

export async function getOrders(params?: {
  status?: string;
  platform?: string;
  page?: number;
}): Promise<Order[]> {
  const response = await api.get<OrderListResponse>("/orders", { params });

  if (!response.data.success) {
    throw new Error(response.data.error);
  }

  return response.data.data;
}
```

### TanStack Query Integration

```typescript
// src/hooks/useOrders.ts
import { useQuery } from "@tanstack/react-query";
import { getOrders } from "@/api/orders";

export function useOrders(filters?: { status?: string }) {
  return useQuery({
    queryKey: ["orders", filters],
    queryFn: () => getOrders(filters),
    staleTime: 30_000, // 30 seconds
  });
}
```

### Backend API Endpoints Reference

| Endpoint            | Method  | Description              |
| ------------------- | ------- | ------------------------ |
| `/api/orders`       | GET     | List orders with filters |
| `/api/orders/:id`   | GET     | Get order detail         |
| `/api/products`     | GET     | List products            |
| `/api/products/:id` | GET/PUT | Get/update product       |
| `/api/inventory`    | GET     | List inventory           |
| `/api/analytics/*`  | GET     | Analytics data           |
| `/api/health`       | GET     | Health check             |

---

## Consistency Enforcement

### Color Tokens (USE ONLY THESE)

```typescript
// ✅ CORRECT - use tokens
<Button type="primary">Submit</Button>  // Uses colorPrimary from theme
<Tag color="success">Done</Tag>         // Uses colorSuccess from theme

// ❌ WRONG - hardcoded colors
<Button style={{ backgroundColor: '#0369a1' }}>Submit</Button>
<div style={{ color: 'blue' }}>Text</div>
```

### Spacing Tokens (4px Grid)

```typescript
// ✅ CORRECT - use Ant Design spacing or consistent values
<Space size="middle">          // 16px
<Card style={{ padding: 24 }}> // Uses 4px grid
<Row gutter={[16, 16]}>        // Consistent gutters

// ❌ WRONG - random values
<div style={{ margin: 13 }}>   // Not on 4px grid
<div style={{ padding: 17 }}>  // Not on 4px grid
```

### Typography Consistency

```typescript
// ✅ CORRECT - use Ant Design Typography or theme
<Typography.Title level={4}>Page Title</Typography.Title>
<Typography.Text>Body text</Typography.Text>

// Or with consistent custom styles
<h1 className="text-2xl font-semibold">Title</h1>  // 20px
<p className="text-sm">Body</p>                     // 12px

// ❌ WRONG - inconsistent font sizes
<p style={{ fontSize: 13 }}>Text</p>  // Not in scale
<p style={{ fontSize: 15 }}>Text</p>  // Not in scale
```

### Border Radius Consistency

```typescript
// ✅ CORRECT - 3px everywhere (Ginee sharp style)
borderRadius: 3,      // Default
borderRadiusSM: 2,    // Small elements
borderRadiusLG: 6,    // Large cards/modals only

// ❌ WRONG - rounded/pill styles
borderRadius: 8,      // Too rounded
borderRadius: 16,     // Way too rounded
borderRadius: 9999,   // Pill (only for avatars/badges)
```

---

## Pre-Implementation Checklist

Before implementing any component/page:

- [ ] Backend API endpoint exists and works (`curl` test)
- [ ] TypeScript types match backend response (snake_case)
- [ ] Using theme tokens (not hardcoded colors)
- [ ] Using 4px spacing grid
- [ ] Using 12px base font size
- [ ] Using 3px border radius
- [ ] Using system fonts only

## Post-Implementation Checklist

After implementing any component/page:

- [ ] API integration works (data loads correctly)
- [ ] No TypeScript errors
- [ ] Responsive at 375px, 768px, 1200px
- [ ] No horizontal scroll
- [ ] Fonts consistent (check DevTools)
- [ ] Colors match Design System
- [ ] Layout not messy/broken at any viewport
