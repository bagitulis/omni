# 🎨 UI/UX Design System - Omni Platform

> **Comprehensive design system for consistent, accessible, and maintainable frontend development**

---

## 📚 Documentation Structure

```
docs/UI/
├── README.MD              # This file - Overview & Quick Start
├── colors.md              # Color palette, semantic colors, contrast
├── typography.md          # Fonts, scale, hierarchy
├── layout.md              # Spacing, grid, breakpoints
├── accessibility.md       # WCAG compliance, a11y guidelines
└── components/
    ├── buttons.md         # Button variants & states
    ├── forms.md           # Input, select, validation
    ├── tables.md          # ⭐ Table patterns (sorting, pagination)
    ├── cards.md           # Card layouts & variants
    ├── modals.md          # Modal structure & transitions
    └── badges.md          # Status indicators & badges
```

---

## 🚀 Quick Start

### 1. **Pick Your Component**

| Need to build...      | Read...                                          |
| --------------------- | ------------------------------------------------ |
| **Button**            | [components/buttons.md](./components/buttons.md) |
| **Form**              | [components/forms.md](./components/forms.md)     |
| **Table**             | [components/tables.md](./components/tables.md)   |
| **Card**              | [components/cards.md](./components/cards.md)     |
| **Modal**             | [components/modals.md](./components/modals.md)   |
| **Badge/Status**      | [components/badges.md](./components/badges.md)   |
| **Layout & Spacing**  | [layout.md](./layout.md)                         |
| **Colors & Contrast** | [colors.md](./colors.md)                         |
| **Typography**        | [typography.md](./typography.md)                 |
| **Accessibility**     | [accessibility.md](./accessibility.md)           |

### 2. **Use Design Tokens**

```css
/* ❌ Don't hardcode values */
padding: 16px;
border-radius: 8px;
color: #667eea;

/* ✅ Use CSS variables */
padding: var(--space-lg);
border-radius: var(--radius-lg);
color: var(--color-primary-500);
```

### 3. **Follow Patterns**

Copy-paste pattern dari dokumentasi, lalu sesuaikan:

```vue
<!-- From components/buttons.md -->
<button class="btn btn-primary">Save Changes</button>

<!-- From components/forms.md -->
<div class="form-group">
  <label for="email">Email</label>
  <input id="email" type="email" v-model="form.email" />
</div>
```

### 4. **Test Accessibility**

```bash
# After frontend changes
cd frontend && npm run build
cd .. && docker compose -f docker-compose.n8n.production-full.yml up -d --build frontend
Start-Sleep -Seconds 6
cd backend && node tests/e2e/lighthouse-quick-test.js [groups...]
```

---

## 🎯 Quality Standards

| Metric                | Target    | How to Check                 |
| --------------------- | --------- | ---------------------------- |
| **Accessibility**     | ≥ 95%     | Lighthouse CI                |
| **Color Contrast**    | ≥ 4.5:1   | [colors.md](./colors.md)     |
| **Mobile Responsive** | All pages | Test at 375px, 768px, 1024px |
| **Load Time**         | < 3s      | Lighthouse Performance       |

---

## 📖 Core Concepts

### Design Tokens

All spacing, colors, and typography use CSS variables:

```css
/* Spacing (4px base) */
--space-xs: 4px;
--space-sm: 8px;
--space-md: 12px;
--space-lg: 16px;
--space-xl: 20px;
--space-2xl: 24px;

/* Colors */
--color-primary-500: #3b82f6;
--color-success-500: #10b981;
--color-error-500: #ef4444;

/* Border Radius */
--radius-sm: 4px;
--radius-md: 6px;
--radius-lg: 8px;
```

**See:** [layout.md](./layout.md) | [colors.md](./colors.md)

### Component Patterns

Reusable class names for consistency:

```css
.btn, .btn-primary, .btn-secondary       /* Buttons */
.card, .metric-card                       /* Cards */
.form-group, .form-actions                /* Forms */
.data-table, .table-wrapper               /* Tables */
.badge, .badge-success                    /* Badges */
```

**See:** [components/](./components/)

### Accessibility First

- ✅ All buttons have accessible names
- ✅ All inputs have labels
- ✅ Color contrast ≥ 4.5:1
- ✅ Keyboard navigation works
- ✅ Screen reader tested

**See:** [accessibility.md](./accessibility.md)

---

## 🎨 Design Tokens Reference

### Colors

| Category    | Token                 | Value     | Usage           |
| ----------- | --------------------- | --------- | --------------- |
| **Primary** | `--color-primary-500` | `#3b82f6` | Main actions    |
| **Success** | `--color-success-500` | `#10b981` | Success states  |
| **Error**   | `--color-error-500`   | `#ef4444` | Error states    |
| **Warning** | `--color-warning-500` | `#f59e0b` | Warning states  |
| **Shopee**  | `--color-shopee`      | `#f53d2d` | Shopee branding |
| **TikTok**  | `--color-tiktok`      | `#1d4ed8` | TikTok branding |
| **Lazada**  | `--color-lazada`      | `#0f1689` | Lazada branding |

**Full palette:** [colors.md](./colors.md)

### Spacing

| Token         | Value | Usage                 |
| ------------- | ----- | --------------------- |
| `--space-xs`  | 4px   | Tight spacing         |
| `--space-sm`  | 8px   | Small gaps            |
| `--space-md`  | 12px  | Default component gap |
| `--space-lg`  | 16px  | Standard spacing      |
| `--space-xl`  | 20px  | Large spacing         |
| `--space-2xl` | 24px  | Section padding       |
| `--space-3xl` | 32px  | Section margin        |

**Full scale:** [layout.md](./layout.md)

### Typography

| Token         | Value | Weight | Usage       |
| ------------- | ----- | ------ | ----------- |
| `--text-xs`   | 12px  | 400    | Helper text |
| `--text-sm`   | 14px  | 400    | Small text  |
| `--text-base` | 16px  | 400    | Body text   |
| `--text-lg`   | 18px  | 400    | Large text  |
| `--text-xl`   | 20px  | 600    | H4 headings |
| `--text-2xl`  | 24px  | 600    | H3 headings |
| `--text-3xl`  | 30px  | 700    | H2 headings |
| `--text-4xl`  | 36px  | 700    | H1 headings |

**Full scale:** [typography.md](./typography.md)

---

## 🧩 Component Quick Reference

### Buttons

```vue
<!-- Primary action -->
<button class="btn btn-primary">Save</button>

<!-- Secondary/cancel -->
<button class="btn btn-secondary">Cancel</button>

<!-- Danger/delete -->
<button class="btn btn-danger">Delete</button>

<!-- Platform branded -->
<button class="btn btn-shopee">Refresh</button>
```

**Full patterns:** [components/buttons.md](./components/buttons.md)

### Forms

```vue
<div class="form-group">
  <label for="email">Email</label>
  <input id="email" type="email" v-model="form.email" required />
  <small>Helper text goes here</small>
</div>
```

**Full patterns:** [components/forms.md](./components/forms.md)

### Tables

```vue
<div class="table-wrapper">
  <table class="data-table" aria-label="Product data">
    <thead>
      <tr>
        <th scope="col">Name</th>
        <th scope="col">Price</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="item in items" :key="item.id">
        <td>{{ item.name }}</td>
        <td>{{ item.price }}</td>
      </tr>
    </tbody>
  </table>
</div>
```

**Full patterns:** [components/tables.md](./components/tables.md)

### Cards

```vue
<div class="card">
  <h3>Card Title</h3>
  <p>Card content</p>
</div>
```

**Full patterns:** [components/cards.md](./components/cards.md)

### Badges

```vue
<span class="badge badge-success">Active</span>
<span class="badge badge-warning">Pending</span>
<span class="badge badge-danger">Error</span>
```

**Full patterns:** [components/badges.md](./components/badges.md)

---

## ✅ Pre-Commit Checklist

Before committing frontend changes:

- [ ] **Accessibility:** All buttons have accessible names
- [ ] **Forms:** All inputs have associated `<label for="id">`
- [ ] **Headings:** Hierarchy is correct (h1 → h2 → h3)
- [ ] **Colors:** Contrast ≥ 4.5:1 for text
- [ ] **Keyboard:** Tab navigation works
- [ ] **Responsive:** Test at 375px, 768px, 1024px
- [ ] **Lighthouse:** Accessibility ≥ 95% (test relevant groups)

---

## 🧪 Testing Workflow

### 1. Local Development

```bash
cd frontend
npm run dev
# Test at http://localhost:5173
```

### 2. Production Build

```bash
cd frontend
npm run build

cd ..
docker compose -f docker-compose.n8n.production-full.yml up -d --build frontend
```

### 3. Lighthouse Testing

Test **ONLY** groups you changed:

```bash
cd backend
Start-Sleep -Seconds 6

# Test specific groups
node tests/e2e/lighthouse-quick-test.js analytics       # Shopee/TikTok Ads
node tests/e2e/lighthouse-quick-test.js operation       # Orders
node tests/e2e/lighthouse-quick-test.js inventory       # Inventory
node tests/e2e/lighthouse-quick-test.js settings        # Settings
```

| Group           | Pages Tested                      |
| --------------- | --------------------------------- |
| `analytics`     | Shopee Ads, TikTok Ads dashboards |
| `operation`     | Shopee/Lazada/TikTok orders       |
| `inventory`     | Inventory management              |
| `settings`      | Settings pages                    |
| `scriptmonitor` | Script monitor                    |
| `admin`         | Admin pages                       |
| `routemapping`  | Route mapping                     |

---

## 📚 Additional Resources

### Internal Docs

- [STANDARDS.MD](./STANDARDS.MD) - Old monolithic doc (being phased out)
- [FRONTEND_PATTERN_ANALYSIS.MD](./FRONTEND_PATTERN_ANALYSIS.MD) - Analysis doc (archived)

### External Resources

- [Tailwind CSS](https://tailwindcss.com/docs)
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [WebAIM Contrast Checker](https://webaim.org/resources/contrastchecker/)
- [Lighthouse CI](https://github.com/GoogleChrome/lighthouse-ci)

---

## 🤝 Contributing

When adding new patterns:

1. **Document first** - Add to relevant `.md` file
2. **Follow existing patterns** - Don't introduce new naming
3. **Test accessibility** - Lighthouse score ≥ 95%
4. **Keep it DRY** - Max 300 lines per file

---

**Questions?** Check specific documentation files linked above, or review existing components in [frontend/src/components/](../../frontend/src/components/).
