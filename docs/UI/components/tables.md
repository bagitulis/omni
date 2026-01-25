# Tables

## Pattern Terbaik

**Best Example:** [ShopeeAdsDataTable.vue](../../../frontend/src/components/analytics/ShopeeAdsDataTable.vue) & [TiktokAdsDataTable.vue](../../../frontend/src/components/analytics/TiktokAdsDataTable.vue)

### Structure Pattern

```vue
<div class="table-wrapper">
  <table class="product-table" aria-label="Product data table">
    <thead>
      <tr>
        <th scope="col">Column Name</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="row in data" :key="row.id">
        <td>{{ row.value }}</td>
      </tr>
    </tbody>
  </table>
</div>
```

---

## Styling Patterns

### General Tables (Admin, Settings)

```css
.data-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.data-table th,
.data-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.data-table th {
  background: #f9fafb;
  color: #374151;
  font-weight: 600;
  white-space: nowrap;
}

.data-table tbody tr:hover {
  background: #f9fafb;
}

.table-wrapper {
  overflow-x: auto;
  border-radius: 8px;
}
```

### Analytics Tables (Platform Branding)

**Shopee Style:**

```css
.product-table th {
  background: #f53d2d; /* Shopee red */
  color: white;
  font-weight: 600;
}
```

**TikTok Style:**

```css
.creative-table th {
  background: #1d4ed8; /* TikTok blue */
  color: white;
  font-weight: 600;
}
```

**Lazada Style:**

```css
.product-table th {
  background: #0f1689; /* Lazada dark blue */
  color: white;
  font-weight: 600;
}
```

---

## Accessibility Features

- ✅ `aria-label` on table
- ✅ `scope="col"` on headers
- ✅ Semantic `<thead>`, `<tbody>`
- ✅ Proper column headers

---

## Responsive Handling

```css
.table-wrapper {
  overflow-x: auto; /* Horizontal scroll on mobile */
}

@media (max-width: 768px) {
  .data-table th,
  .data-table td {
    padding: 8px 12px;
    font-size: 14px;
  }
}
```

---

## Empty & Loading States

```vue
<div v-if="loading" class="loading-state">
  <span aria-hidden="true">⏳</span> Loading data...
</div>

<div v-else-if="data.length === 0" class="empty-state">
  <p>No data found</p>
</div>

<table v-else class="data-table">
  <!-- table content -->
</table>
```

```css
.loading-state,
.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: #6b7280;
  text-align: center;
}
```

---

## Sorting & Pagination

**Sortable Headers:**

```vue
<th scope="col" @click="sortBy('column')" class="sortable">
  Column Name
  <span v-if="sortColumn === 'column'">
    {{ sortDirection === 'asc' ? '↑' : '↓' }}
  </span>
</th>
```

```css
.sortable {
  cursor: pointer;
  user-select: none;
}

.sortable:hover {
  background: #e5e7eb;
}
```

---

## Quick Reference

| Property          | Value                       |
| ----------------- | --------------------------- |
| **Padding**       | `12px 16px`                 |
| **Header BG**     | `#f9fafb` (general)         |
| **Border**        | `1px solid #e5e7eb`         |
| **Border Radius** | `8px`                       |
| **Shadow**        | `0 1px 3px rgba(0,0,0,0.1)` |
| **Hover BG**      | `#f9fafb`                   |

---

**See Also:** [Layout & Spacing](../layout.md) | [Colors](../colors.md)
