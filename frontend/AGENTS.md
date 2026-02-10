# React Frontend Context

> Auto-injected when working in `frontend/` directory.
> **Parent rules:** See root `AGENTS.md` for critical rules.
> **Detailed specs:** See skill `react-frontend-rules` for full design system.

---

## Stack

- React 19 + TypeScript 5.7+ + Vite 6
- Ant Design 5 (UI framework)
- Zustand (client state)
- TanStack Query (server state)
- React Router 7

---

## Structure

```
frontend/
├── src/
│   ├── api/             # API client functions
│   ├── components/      # Reusable components
│   │   ├── layout/      # AppLayout, Sidebar, Header, MobileNav
│   │   ├── ui/          # Extended Ant Design components
│   │   ├── tables/      # Data tables
│   │   ├── forms/       # Form components
│   │   ├── modals/      # Modal components  ← LOWERCASE!
│   │   ├── orders/      # Order-specific components
│   │   └── charts/      # Chart components
│   ├── pages/           # Route pages
│   ├── hooks/           # Custom hooks (useOrders, etc.)
│   ├── stores/          # Zustand stores
│   ├── types/           # TypeScript interfaces (snake_case!)
│   ├── lib/             # Utilities
│   └── styles/          # Theme and global CSS
└── public/
```

---

## ⚠️ CRITICAL: Directory Casing (Windows Bug)

**RECURRING ISSUE**: `components/modals/` keeps flipping to `components/Modals/` on Windows, causing TS1261 build failures.

### Root Cause

- Vue legacy (`frontend-vue/`) uses `Modals/` (uppercase M) — **legitimate, don't change**
- React (`frontend/`) uses `modals/` (lowercase m) — **correct per convention**
- Windows `core.ignorecase=true` means git doesn't track directory casing changes
- When files are written, Windows may resolve to uppercase from Vue's pattern

### Rules

| Rule                                | Detail                                                          |
| ----------------------------------- | --------------------------------------------------------------- |
| **ALL React component directories** | **lowercase**: `modals/`, `tables/`, `forms/`, `layout/`, `ui/` |
| **ALL imports**                     | `@/components/modals/SomeModal` (lowercase)                     |
| **NEVER**                           | `@/components/Modals/` (uppercase) in React code                |

### If Build Fails with TS1261

```bash
# Two-step rename on Windows (can't rename same-casing directly)
mv frontend/src/components/Modals frontend/src/components/_modals_temp
mv frontend/src/components/_modals_temp frontend/src/components/modals
```

---

## API Type Convention

All API types use **snake_case** to match Go backend:

```typescript
// ✅ CORRECT
interface Order {
  order_sn: string;
  created_at: string;
  total_amount: number;
}

// ❌ WRONG
interface Order {
  orderSn: string; // Backend sends order_sn
  createdAt: string; // Backend sends created_at
}
```

---

## Anti-Patterns

| Forbidden              | Do Instead                  |
| ---------------------- | --------------------------- |
| `@/components/Modals/` | `@/components/modals/`      |
| Google Fonts           | System fonts only           |
| camelCase API types    | snake_case to match backend |
| Hardcoded colors       | Use Ant Design theme tokens |
| Border radius > 6px    | 3px sharp corners           |
| `as any`, `@ts-ignore` | Fix types properly          |
