---
globs: ["**/*.vue", "**/*.ts", "**/*.tsx", "frontend/**"]
description: Vue/TypeScript frontend development rules for OMNI project
---

# Frontend Development Rules

## Stack

- Vue 3 + Composition API
- TypeScript
- Vite
- TailwindCSS
- Pinia (state management)

---

## Naming Conventions

| Type             | Convention                 | Example          |
| ---------------- | -------------------------- | ---------------- |
| Components       | PascalCase                 | `OrderTable.vue` |
| Composables      | camelCase with `use`       | `useOrders.ts`   |
| API types        | snake_case (match backend) | `order_sn`       |
| Local vars       | camelCase                  | `orderList`      |
| Props (template) | kebab-case                 | `:order-id="id"` |

---

## API Type Matching

Backend returns snake_case, frontend should match:

```typescript
// ✅ CORRECT - matches backend
interface Order {
  order_sn: string;
  created_at: string;
  total_amount: number;
}

// ❌ WRONG - mismatched naming
interface Order {
  orderSn: string;
  createdAt: string;
}
```

---

## Component Structure

```vue
<script setup lang="ts">
// 1. Imports
// 2. Props/Emits
// 3. Composables
// 4. Reactive state
// 5. Computed
// 6. Methods
// 7. Lifecycle
</script>

<template>
  <!-- Template -->
</template>

<style scoped>
/* Scoped styles */
</style>
```

---

## API Calls

Location: `src/api/`

```typescript
// Use typed API functions
export async function getOrders(tenantId: string): Promise<Order[]> {
  const response = await api.get(`/orders`, {
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
```
