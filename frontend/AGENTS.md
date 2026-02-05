# Frontend Development Context

> Auto-injected when working in `frontend/` directory.
> **Parent rules:** See root `AGENTS.md` for critical rules.

---

## Stack

- Vue 3 + Composition API
- TypeScript
- Vite
- TailwindCSS
- Pinia (state management)

---

## Structure

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

## Naming Conventions

| Type             | Convention        | Example                     |
| ---------------- | ----------------- | --------------------------- |
| Components       | PascalCase        | `OrderTable.vue`            |
| Composables      | camelCase + `use` | `useOrders.ts`              |
| API types        | **snake_case**    | `order_sn` (match backend!) |
| Local vars       | camelCase         | `orderList`                 |
| Props (template) | kebab-case        | `:order-id="id"`            |

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

## Component Structure

```vue
<script setup lang="ts">
// 1. Imports
import { ref, computed, onMounted } from 'vue'
import { useOrders } from '@/composables/useOrders'

// 2. Props/Emits
const props = defineProps<{ orderId: string }>()
const emit = defineEmits<{ (e: 'update', id: string): void }>()

// 3. Composables
const { orders, fetchOrders } = useOrders()

// 4. Reactive state
const loading = ref(false)

// 5. Computed
const filteredOrders = computed(() => orders.value.filter(...))

// 6. Methods
async function handleSubmit() { ... }

// 7. Lifecycle
onMounted(() => fetchOrders())
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
// catch (e) { } ← DILARANG
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

## Anti-Patterns

| Forbidden             | Do Instead                          |
| --------------------- | ----------------------------------- |
| camelCase API types   | Use snake_case to match backend     |
| Untyped API calls     | Always define TypeScript interfaces |
| Empty catch blocks    | Show error to user + log            |
| Props without type    | Use `defineProps<T>()`              |
| Direct store mutation | Use actions/composables             |
