# InventoryContent - Development Rules

> **Parent:** `frontend/AGENTS.md` | **Files:** 60+ Vue/TypeScript files

---

## Structure

```
InventoryContent/
├── composables/           # 38 composables for inventory operations
│   ├── useCellEdit*.ts    # Inline cell editing
│   ├── useFilter*.ts      # Filter state management
│   ├── useBatch*.ts       # Batch operations
│   ├── useLock*.ts        # Lock management
│   └── usePagination.ts   # Pagination logic
├── InventoryContent.vue   # Main component
├── InventoryTable.vue     # Table component
├── InventoryFilters.vue   # Filter UI
└── *.vue                  # Other partials
```

---

## This Module Is Complex Because

1. **Inline cell editing** - Direct edit on table cells
2. **Multi-filter state** - Filters persist across navigation
3. **Batch operations** - Select multiple, apply actions
4. **Lock management** - Prevent concurrent edits
5. **Cross-platform** - Shopee, TikTok, Lazada inventory

---

## Composable Patterns (MANDATORY)

### Cell Editing

```typescript
// ✅ CORRECT - Cell edit composable pattern
export function useCellEdit(productId: string) {
  const isEditing = ref(false);
  const originalValue = ref<string>("");
  const currentValue = ref<string>("");

  function startEdit(value: string) {
    originalValue.value = value;
    currentValue.value = value;
    isEditing.value = true;
  }

  function cancelEdit() {
    currentValue.value = originalValue.value;
    isEditing.value = false;
  }

  async function saveEdit() {
    await api.updateProduct(productId, { value: currentValue.value });
    isEditing.value = false;
  }

  return { isEditing, currentValue, startEdit, cancelEdit, saveEdit };
}
```

### Filter Persistence

```typescript
// ✅ CORRECT - Filters persist to URL/localStorage
export function useInventoryFilters() {
  const filters = ref<FilterState>(loadFromStorage());

  watch(
    filters,
    (newFilters) => {
      saveToStorage(newFilters);
      updateURLParams(newFilters);
    },
    { deep: true },
  );

  return { filters, resetFilters, applyFilters };
}
```

### Batch Operations

```typescript
// ✅ CORRECT - Batch selection pattern
export function useBatchSelection() {
  const selectedIds = ref<Set<string>>(new Set());

  function toggleSelection(id: string) {
    if (selectedIds.value.has(id)) {
      selectedIds.value.delete(id);
    } else {
      selectedIds.value.add(id);
    }
  }

  function selectAll(ids: string[]) {
    selectedIds.value = new Set(ids);
  }

  function clearSelection() {
    selectedIds.value.clear();
  }

  return { selectedIds, toggleSelection, selectAll, clearSelection };
}
```

---

## Lock Management

```typescript
// ✅ CORRECT - Optimistic locking
export function useLock(resourceId: string) {
  const lockOwner = ref<string | null>(null);
  const isLocked = computed(() => lockOwner.value !== null);

  async function acquireLock() {
    const result = await api.acquireLock(resourceId);
    if (result.success) {
      lockOwner.value = result.owner;
    }
    return result.success;
  }

  async function releaseLock() {
    await api.releaseLock(resourceId);
    lockOwner.value = null;
  }

  return { isLocked, lockOwner, acquireLock, releaseLock };
}
```

---

## Platform-Specific Handling

```typescript
// ✅ CORRECT - Platform-aware inventory
interface InventoryItem {
  product_id: string;
  platform: "shopee" | "tiktok" | "lazada";
  stock: number;
  price: number;
}

function getPlatformService(platform: string) {
  switch (platform) {
    case "shopee":
      return shopeeService;
    case "tiktok":
      return tiktokService;
    case "lazada":
      return lazadaService;
    default:
      throw new Error(`Unknown platform: ${platform}`);
  }
}
```

---

## Anti-Patterns (FORBIDDEN)

```typescript
// ❌ WRONG - Edit state in component, not composable
// <script setup>
const isEditing = ref(false); // Move to composable!

// ❌ WRONG - Filters lost on navigation
// Filter state must persist!

// ❌ WRONG - No lock check before edit
async function saveProduct() {
  await api.save(product); // Check lock first!
}

// ❌ WRONG - Batch operation without confirmation
async function deleteSelected() {
  await api.deleteMany(selectedIds); // Confirm first!
}
```

---

## Max Lines Per File

| Type           | Limit     |
| -------------- | --------- |
| Vue components | 300 lines |
| Composables    | 300 lines |

**If exceeding:** Extract logic to composables, split component.
