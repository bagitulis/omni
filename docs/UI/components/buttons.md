# Buttons

## Base Button

```css
.btn {
  padding: 10px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  font-size: 14px;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
```

---

## Button Variants

### Primary (Main Actions)

```css
.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
}
```

```vue
<button class="btn btn-primary">Save Changes</button>
```

### Secondary (Cancel, Alternative)

```css
.btn-secondary {
  background: #e5e7eb;
  color: #374151;
}

.btn-secondary:hover:not(:disabled) {
  background: #d1d5db;
}
```

```vue
<button class="btn btn-secondary">Cancel</button>
```

### Danger (Delete, Destructive)

```css
.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
}
```

```vue
<button class="btn btn-danger">Delete</button>
```

### Ghost (Minimal, Text-like)

```css
.btn-ghost {
  background: transparent;
  color: #6b7280;
}

.btn-ghost:hover:not(:disabled) {
  background: #f3f4f6;
  color: #374151;
}
```

```vue
<button class="btn btn-ghost">Learn More</button>
```

---

## Platform-Branded Buttons

### Shopee

```css
.btn-shopee {
  background: #f53d2d;
  color: white;
}

.btn-shopee:hover:not(:disabled) {
  background: #d93025;
}
```

### TikTok

```css
.btn-tiktok {
  background: #1d4ed8;
  color: white;
}

.btn-tiktok:hover:not(:disabled) {
  background: #1e40af;
}
```

### Lazada

```css
.btn-lazada {
  background: #0f1689;
  color: white;
}

.btn-lazada:hover:not(:disabled) {
  background: #0a0f5c;
}
```

---

## Button Sizes

### Small

```css
.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}
```

```vue
<button class="btn btn-primary btn-sm">Small Button</button>
```

### Large

```css
.btn-lg {
  padding: 12px 24px;
  font-size: 16px;
}
```

```vue
<button class="btn btn-primary btn-lg">Large Button</button>
```

---

## Icon Buttons

### Icon Only

```css
.btn-icon {
  padding: 10px;
  width: 40px;
  height: 40px;
  justify-content: center;
}
```

```vue
<button class="btn btn-icon btn-ghost" aria-label="Close">
  <i class="pi pi-times"></i>
</button>
```

### Icon with Text

```vue
<button class="btn btn-primary">
  <i class="pi pi-save"></i>
  Save
</button>
```

---

## Action Buttons (Table)

```css
.action-btn {
  padding: 6px 10px;
  margin: 0 4px;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 16px;
  color: #6b7280;
  transition: all 0.2s;
}

.action-btn:hover {
  color: #3b82f6;
  background: #f3f4f6;
  border-radius: 4px;
}
```

```vue
<button class="action-btn" aria-label="Edit" @click="editItem(item)">
  ✏️
</button>
<button class="action-btn" aria-label="Delete" @click="deleteItem(item)">
  🗑️
</button>
```

---

## Loading State

```css
.btn-loading {
  position: relative;
  pointer-events: none;
}

.btn-loading::after {
  content: "";
  position: absolute;
  width: 16px;
  height: 16px;
  top: 50%;
  left: 50%;
  margin-left: -8px;
  margin-top: -8px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
```

```vue
<button
  class="btn btn-primary"
  :class="{ 'btn-loading': isLoading }"
  :disabled="isLoading"
>
  Submit
</button>
```

---

## Button Groups

```css
.btn-group {
  display: flex;
  gap: 8px;
}
```

```vue
<div class="btn-group">
  <button class="btn btn-primary">Save</button>
  <button class="btn btn-secondary">Cancel</button>
</div>
```

---

## Accessibility

- ✅ Use `type="button"` for non-submit buttons
- ✅ Use `aria-label` for icon-only buttons
- ✅ Disable with `:disabled` attribute
- ✅ Provide visual focus states
- ✅ Ensure 44px minimum touch target on mobile

---

## Quick Reference

| Variant       | Background    | Text Color | Use Case       |
| ------------- | ------------- | ---------- | -------------- |
| **Primary**   | `#3b82f6`     | `white`    | Main actions   |
| **Secondary** | `#e5e7eb`     | `#374151`  | Cancel, back   |
| **Danger**    | `#ef4444`     | `white`    | Delete, remove |
| **Ghost**     | `transparent` | `#6b7280`  | Minimal UI     |
| **Shopee**    | `#f53d2d`     | `white`    | Shopee brand   |
| **TikTok**    | `#1d4ed8`     | `white`    | TikTok brand   |
| **Lazada**    | `#0f1689`     | `white`    | Lazada brand   |

---

**See Also:** [Forms](./forms.md) | [Colors](../colors.md) | [Accessibility](../accessibility.md)
