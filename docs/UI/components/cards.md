# Cards

## Base Card

```css
.card {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border: 1px solid #e5e7eb;
}
```

```vue
<div class="card">
  <h3>Card Title</h3>
  <p>Card content goes here.</p>
</div>
```

---

## Card Variants

### Interactive Card (Hover Effect)

```css
.card-interactive {
  transition: all 0.2s ease;
  cursor: pointer;
}

.card-interactive:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}
```

```vue
<div class="card card-interactive" @click="navigateTo()">
  <h3>Clickable Card</h3>
</div>
```

---

## Metric Card (Dashboard)

**Best Example:** [ShopeeAdsDashboard.vue](../../../frontend/src/components/analytics/ShopeeAdsDashboard.vue)

```css
.metric-card {
  background: white;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.metric-label {
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 4px;
  display: block;
}

.metric-value {
  font-size: 24px;
  font-weight: 600;
  color: #1f2937;
}

.metric-value.cost {
  color: #dc2626; /* Red for cost */
}

.metric-value.revenue {
  color: #047857; /* Green for revenue */
}
```

```vue
<div class="metric-card">
  <span class="metric-label">Total Cost</span>
  <div class="metric-value cost">$1,234.56</div>
</div>
```

---

## Comparison Card (Features)

```css
.comparison-card {
  background: white;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.type-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e5e7eb;
}

.type-name {
  font-weight: 600;
  color: #1f2937;
  font-size: 14px;
}

.type-metrics {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}
```

```vue
<div class="comparison-card">
  <div class="type-header">
    <span class="type-icon">🎯</span>
    <span class="type-name">GMV Max ROAS</span>
  </div>
  <div class="type-metrics">
    <div class="metric">
      <span class="metric-label">Cost</span>
      <span class="metric-value">$1,234</span>
    </div>
  </div>
</div>
```

---

## Section Card (Settings)

```css
.section-card {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}

.section-header {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e5e7eb;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}
```

```vue
<div class="section-card">
  <div class="section-header">
    <h2 class="section-title">Settings Section</h2>
  </div>
  <div class="section-content">
    <!-- Settings content -->
  </div>
</div>
```

---

## Card Grid Layouts

### Auto-fit Grid

```css
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
}
```

### Summary Cards (4-column)

```css
.summary-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

@media (max-width: 1200px) {
  .summary-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .summary-cards {
    grid-template-columns: 1fr;
  }
}
```

---

## Card with Header & Footer

```css
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e5e7eb;
}

.card-body {
  margin-bottom: 16px;
}

.card-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid #e5e7eb;
}
```

```vue
<div class="card">
  <div class="card-header">
    <h3>Card Title</h3>
    <button class="btn btn-ghost btn-sm">Action</button>
  </div>
  <div class="card-body">
    <p>Card content goes here.</p>
  </div>
  <div class="card-footer">
    <button class="btn btn-secondary">Cancel</button>
    <button class="btn btn-primary">Save</button>
  </div>
</div>
```

---

## Empty Card State

```css
.card-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  color: #6b7280;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
}
```

```vue
<div class="card card-empty">
  <div class="empty-icon">📊</div>
  <p>No data available</p>
</div>
```

---

## Quick Reference

| Property          | Value                         |
| ----------------- | ----------------------------- |
| **Padding**       | `20px` (standard)             |
| **Border Radius** | `8px`                         |
| **Shadow**        | `0 1px 3px rgba(0,0,0,0.1)`   |
| **Border**        | `1px solid #e5e7eb`           |
| **Hover Shadow**  | `0 4px 12px rgba(0,0,0,0.15)` |
| **Grid Gap**      | `16px`                        |

---

**See Also:** [Layout & Spacing](../layout.md) | [Colors](../colors.md)
