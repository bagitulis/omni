# Badges & Status Indicators

## Base Badge

```css
.badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  line-height: 1;
}
```

```vue
<span class="badge badge-primary">New</span>
```

---

## Badge Variants

### Primary

```css
.badge-primary {
  background: #dbeafe;
  color: #1e40af;
}
```

### Success

```css
.badge-success {
  background: #d1fae5;
  color: #065f46;
}
```

### Warning

```css
.badge-warning {
  background: #fef3c7;
  color: #92400e;
}
```

### Danger

```css
.badge-danger {
  background: #fee2e2;
  color: #991b1b;
}
```

### Neutral

```css
.badge-neutral {
  background: #f3f4f6;
  color: #374151;
}
```

---

## Role Badges

**Best Example:** [AdminUsersPage.vue](../../../frontend/src/pages/AdminUsersPage.vue)

```css
.role-badge {
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.role-badge.owner {
  background: #ffe0b2;
  color: #e65100;
}

.role-badge.admin {
  background: #c8e6c9;
  color: #1b5e20;
}

.role-badge.user {
  background: #bbdefb;
  color: #0d47a1;
}
```

```vue
<span :class="['role-badge', user.role]">{{ user.role }}</span>
```

---

## Mode Badges (Analytics)

```css
.mode-badge {
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 500;
}

.mode-gmv {
  background: #dbeafe;
  color: #1e40af;
}

.mode-auto {
  background: #fef3c7;
  color: #92400e;
}

.mode-manual {
  background: #e0e7ff;
  color: #3730a3;
}
```

```vue
<span class="mode-badge" :class="getBadgeClass(row.biddingMode)">
  {{ row.biddingMode }}
</span>
```

---

## Status Indicators (ROAS/ROI)

```css
.status-excellent {
  color: #166534;
  font-weight: 600;
}

.status-good {
  color: #1d4ed8;
  font-weight: 600;
}

.status-ok {
  color: #92400e;
  font-weight: 500;
}

.status-poor {
  color: #991b1b;
  font-weight: 600;
}
```

```vue
<span :class="getStatusClass(row.roas)">
  {{ formatROAS(row.roas) }}
</span>
```

---

## Badge Sizes

### Small

```css
.badge-sm {
  padding: 2px 8px;
  font-size: 11px;
}
```

### Default

```css
.badge {
  padding: 4px 10px;
  font-size: 12px;
}
```

### Large

```css
.badge-lg {
  padding: 6px 14px;
  font-size: 14px;
}
```

---

## Badge with Dot Indicator

```css
.badge-dot::before {
  content: "";
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: 6px;
  background: currentColor;
}
```

```vue
<span class="badge badge-success badge-dot">Active</span>
```

---

## Badge with Icon

```vue
<span class="badge badge-warning">
  <i class="pi pi-exclamation-triangle"></i>
  Pending
</span>
```

```css
.badge i {
  margin-right: 4px;
  font-size: 10px;
}
```

---

## Removable Badge (Tag)

```vue
<span class="badge badge-primary">
  JavaScript
  <button class="badge-remove" aria-label="Remove" @click="removeTag">
    <i class="pi pi-times"></i>
  </button>
</span>
```

```css
.badge-remove {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  margin-left: 6px;
  color: currentColor;
  opacity: 0.7;
  font-size: 10px;
}

.badge-remove:hover {
  opacity: 1;
}
```

---

## Outlined Badges

```css
.badge-outlined {
  background: transparent;
  border: 1px solid currentColor;
}

.badge-outlined.badge-primary {
  color: #1e40af;
  border-color: #1e40af;
}

.badge-outlined.badge-success {
  color: #065f46;
  border-color: #065f46;
}
```

---

## Badge in Context

### In Table Cell

```vue
<td>
  <span class="badge badge-success">Active</span>
</td>
```

### In Button

```vue
<button class="btn btn-primary">
  Messages
  <span class="badge badge-danger" style="margin-left: 8px;">3</span>
</button>
```

### In Card Header

```vue
<div class="card-header">
  <h3>Product Name</h3>
  <span class="badge badge-warning">Low Stock</span>
</div>
```

---

## Quick Reference

| Variant     | Background | Text Color | Use Case          |
| ----------- | ---------- | ---------- | ----------------- |
| **Primary** | `#dbeafe`  | `#1e40af`  | General info      |
| **Success** | `#d1fae5`  | `#065f46`  | Success, active   |
| **Warning** | `#fef3c7`  | `#92400e`  | Warning, pending  |
| **Danger**  | `#fee2e2`  | `#991b1b`  | Error, critical   |
| **Neutral** | `#f3f4f6`  | `#374151`  | Default, inactive |

| Size     | Padding    | Font Size |
| -------- | ---------- | --------- |
| **sm**   | `2px 8px`  | `11px`    |
| **Base** | `4px 10px` | `12px`    |
| **lg**   | `6px 14px` | `14px`    |

---

**See Also:** [Colors](../colors.md) | [Typography](../typography.md)
