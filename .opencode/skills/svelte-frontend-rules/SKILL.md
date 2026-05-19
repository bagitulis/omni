---
description: Svelte/SvelteKit frontend rules - TailwindCSS v4, Svelte 5, performance-first UI. Use when working on Svelte frontend code (web/, web/svelte/).
---

# Svelte Frontend Rules

> **Role:** Rules for agents implementing SvelteKit frontend.
> **Stack:** SvelteKit 5 + Svelte 5 + TailwindCSS v4 + TypeScript + Vite

---

## CRITICAL PRINCIPLES

### 1. Backend Integration

- Frontend MUST integrate with existing Go backend API
- Response format: `{ success: true, data: {...} }` or `{ success: false, error: "..." }`
- All API types use **snake_case** to match backend exactly
- Use `fetch` with proper error handling

### 2. Performance First

| Rule | Implementation |
| ---- | -------------- |
| No heavy UI libs | TailwindCSS only (no Ant Design, Material, Bootstrap) |
| Lazy load charts | Dynamic import Chart.js on demand |
| Server-side pagination | Max 50 rows per request |
| Skeleton loading | Show skeleton UI, not spinners |
| Virtual scroll | For tables > 100 rows |
| Debounce search | 300ms debounce on input |
| System fonts only | No Google Fonts |

### 3. Svelte 5 Runes (MANDATORY)

- Use `$state()` for reactive state (NOT `let x = 0`)
- Use `$derived()` for computed values (NOT `$:`)
- Use `$effect()` for side effects (NOT `$:` with side effects)
- Use `$props()` for component props
- Use `$bindable()` for two-way binding props

**FORBIDDEN (Svelte 4 syntax):**
- `$:` reactive declarations
- `export let prop`
- `import { writable }` for local state (stores only for shared state)

### 4. File Naming

- Components: `PascalCase.svelte` (e.g., `DataTable.svelte`)
- Routes: `kebab-case` directories (e.g., `/campaign-details/`)
- Stores/utils: `camelCase.ts` (e.g., `dashboardStore.ts`)

### 5. TailwindCSS v4

- Use Tailwind utility classes for ALL styling
- Custom CSS only when Tailwind cannot express it
- No `@apply` in component styles
- Responsive: `sm:`, `md:`, `lg:` breakpoints

---

## Anti-Patterns (FORBIDDEN)

| Forbidden | Do Instead |
| --------- | ---------- |
| `$:` reactive declarations | `$state()`, `$derived()`, `$effect()` |
| `export let prop` | `$props()` |
| Google Fonts | System fonts only |
| Ant Design / Material UI | TailwindCSS utilities |
| `fetch` without error handling | API client with try/catch |
| Client-side sorting large data | Server-side sort + paginate |
| Loading all data on mount | Paginate, max 50 rows |
| `<style>` for Tailwind-doable | Use utility classes |
| `any` type | Define TypeScript interfaces |

---

## Build & Verification

```bash
cd web/svelte && npm run build    # MUST pass
cd web/svelte && npm run check    # Type check
```

**Evidence required:**
```
FRONTEND DOD COMPLETE:
- Build: npm run build exit 0
- Types: npm run check clean
- Screenshots: .sisyphus/evidence/[page]-{desktop,mobile}.png
- Console: No JS errors
```
