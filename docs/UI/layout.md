# Layout & Spacing

## Spacing Scale

Based on 4px base unit:

| Token   | Value | Use Case              |
| ------- | ----- | --------------------- |
| **xs**  | 4px   | Tight spacing         |
| **sm**  | 8px   | Small gaps            |
| **md**  | 12px  | Default component gap |
| **lg**  | 16px  | Standard spacing      |
| **xl**  | 20px  | Large spacing         |
| **2xl** | 24px  | Section padding       |
| **3xl** | 32px  | Section margin        |
| **4xl** | 40px  | Large sections        |

```css
:root {
  --space-xs: 4px;
  --space-sm: 8px;
  --space-md: 12px;
  --space-lg: 16px;
  --space-xl: 20px;
  --space-2xl: 24px;
  --space-3xl: 32px;
}
```

---

## Container Widths

| Size | Width  | Use Case               |
| ---- | ------ | ---------------------- |
| sm   | 640px  | Mobile expanded        |
| md   | 768px  | Tablet                 |
| lg   | 1024px | Small desktop          |
| xl   | 1280px | Default page container |
| 2xl  | 1536px | Large screens          |

```css
.page-container {
  padding: var(--space-2xl);
  max-width: 1280px;
  margin: 0 auto;
}
```

---

## Layout Patterns

### Main Layout (Sidebar + Content)

```css
.main-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.sidebar {
  width: 200px;
}
.sidebar.collapsed {
  width: 56px;
}
.main-content {
  flex: 1;
  overflow-y: auto;
}
```

### Grid Layouts

```css
/* Auto-fit responsive grid */
.grid-auto {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--space-lg);
}

/* Fixed columns (use Tailwind: grid-cols-2, grid-cols-3, etc) */
```

### Flex Layouts

```css
.flex-row {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}
.flex-between {
  display: flex;
  justify-content: space-between;
}
.flex-col {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}
```

---

## Border Radius

| Token  | Value | Use Case    |
| ------ | ----- | ----------- |
| **sm** | 4px   | Small items |
| **md** | 6px   | Buttons     |
| **lg** | 8px   | Cards       |
| **xl** | 12px  | Modals      |

```css
:root {
  --radius-sm: 4px;
  --radius-md: 6px;
  --radius-lg: 8px;
  --radius-xl: 12px;
}
```

---

## Shadows

```css
:root {
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.1);
}
```

---

## Responsive Breakpoints

| Breakpoint | Width   | Tailwind |
| ---------- | ------- | -------- |
| Mobile     | < 640px | default  |
| sm         | 640px   | sm:      |
| md         | 768px   | md:      |
| lg         | 1024px  | lg:      |
| xl         | 1280px  | xl:      |
| 2xl        | 1536px  | 2xl:     |

```css
/* Mobile-first approach */
.card-grid {
  grid-template-columns: 1fr; /* Mobile: 1 col */
}
@media (min-width: 768px) {
  .card-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (min-width: 1024px) {
  .card-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}
```

---

## Component Spacing Standards

| Component      | Padding     | Gap    | Margin Bottom |
| -------------- | ----------- | ------ | ------------- |
| **Card**       | `20px`      | -      | -             |
| **Section**    | -           | `20px` | `32px`        |
| **Form Group** | -           | -      | `16px`        |
| **Button**     | `10px 16px` | `8px`  | -             |
| **Table Cell** | `12px 16px` | -      | -             |
| **Modal**      | `24px`      | -      | -             |

---

## Mobile Considerations

```css
/* Landscape mode adjustments */
@media (max-width: 768px) and (orientation: landscape) {
  .header {
    height: 48px;
  }
  .sidebar {
    display: none;
  }
}

/* iOS Safe Area */
.header {
  padding-top: max(16px, env(safe-area-inset-top));
}
.bottom-nav {
  padding-bottom: max(16px, env(safe-area-inset-bottom));
}
```

---

**See Also:** [Components](./components/) | [Colors](./colors.md)
