# Frontend Development Context

> Auto-injected when reading files in `frontend/` directory.

## Stack

- Vue 3 + Composition API
- TypeScript
- Vite
- TailwindCSS
- Pinia (state management)

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

## Naming Conventions

| Type             | Convention        | Example                     |
| ---------------- | ----------------- | --------------------------- |
| Components       | PascalCase        | `OrderTable.vue`            |
| Composables      | camelCase + `use` | `useOrders.ts`              |
| API types        | snake_case        | `order_sn` (match backend!) |
| Props (template) | kebab-case        | `:order-id="id"`            |

## Critical Rules

1. **API types = snake_case** - Must match backend JSON
2. **Typed API calls** - All API functions should be typed
3. **Error handling** - Show user-friendly messages

## API Integration

Backend returns:

```json
{
  "success": true,
  "data": { "order_sn": "123", "total_amount": 100 }
}
```

Frontend types must match:

```typescript
interface Order {
  order_sn: string; // ✅ matches backend
  total_amount: number; // ✅ matches backend
}
```
