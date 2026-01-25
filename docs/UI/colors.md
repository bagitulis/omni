# Color System

## Color Palette

### Primary Colors

```css
:root {
  --color-primary-50: #eff6ff;
  --color-primary-100: #dbeafe;
  --color-primary-200: #bfdbfe;
  --color-primary-300: #93c5fd;
  --color-primary-400: #60a5fa;
  --color-primary-500: #3b82f6; /* Primary */
  --color-primary-600: #2563eb;
  --color-primary-700: #1d4ed8;
  --color-primary-800: #1e40af;
  --color-primary-900: #1e3a8a;
}
```

**Usage:**

- Primary-500: Main buttons, links, brand color
- Primary-600: Hover states for primary buttons
- Primary-700: Active/pressed states

---

## Semantic Colors

### Success (Green)

```css
:root {
  --color-success-50: #f0fdf4;
  --color-success-100: #dcfce7;
  --color-success-500: #10b981; /* Success */
  --color-success-600: #059669;
  --color-success-700: #047857; /* Revenue, profit */
  --color-success-800: #065f46;
  --color-success-900: #166534;
}
```

**Usage:**

- Success messages
- Revenue/profit indicators
- Active/enabled status
- Positive trends

### Error/Danger (Red)

```css
:root {
  --color-error-50: #fef2f2;
  --color-error-100: #fee2e2;
  --color-error-500: #ef4444; /* Error */
  --color-error-600: #dc2626; /* Cost, danger */
  --color-error-700: #b91c1c;
  --color-error-800: #991b1b;
  --color-error-900: #7f1d1d;
}
```

**Usage:**

- Error messages
- Cost/expense indicators
- Delete/destructive actions
- Negative trends

### Warning (Yellow)

```css
:root {
  --color-warning-50: #fffbeb;
  --color-warning-100: #fef3c7;
  --color-warning-500: #f59e0b; /* Warning */
  --color-warning-600: #d97706;
  --color-warning-700: #b45309;
  --color-warning-800: #92400e;
}
```

**Usage:**

- Warning messages
- Pending status
- Caution indicators

---

## Neutral Colors (Gray)

```css
:root {
  --color-gray-50: #f9fafb; /* Light backgrounds */
  --color-gray-100: #f3f4f6; /* Backgrounds */
  --color-gray-200: #e5e7eb; /* Light borders */
  --color-gray-300: #d1d5db; /* Borders */
  --color-gray-400: #9ca3af; /* Disabled text */
  --color-gray-500: #6b7280; /* Muted text */
  --color-gray-600: #4b5563; /* Secondary text */
  --color-gray-700: #374151; /* Body text */
  --color-gray-800: #1f2937; /* Headings */
  --color-gray-900: #111827; /* Heavy text */
}
```

**Usage:**

- Gray-50/100: Light backgrounds, table headers
- Gray-200/300: Borders, dividers
- Gray-500: Helper text, placeholders
- Gray-700/800: Main text content
- Gray-900: Headings, emphasis

---

## Platform Brand Colors

### Shopee

```css
:root {
  --color-shopee: #f53d2d; /* Main red */
  --color-shopee-hover: #d93025; /* Hover state */
  --color-shopee-light: #fee2e2; /* Light background */
}
```

### TikTok

```css
:root {
  --color-tiktok: #1d4ed8; /* Main blue */
  --color-tiktok-hover: #1e40af; /* Hover state */
  --color-tiktok-light: #dbeafe; /* Light background */
}
```

### Lazada

```css
:root {
  --color-lazada: #0f1689; /* Dark blue */
  --color-lazada-hover: #0a0f5c; /* Hover state */
  --color-lazada-light: #e0e7ff; /* Light background */
}
```

---

## Text Colors

```css
:root {
  --color-text-primary: #1f2937; /* Gray-800 */
  --color-text-secondary: #6b7280; /* Gray-500 */
  --color-text-tertiary: #9ca3af; /* Gray-400 */
  --color-text-inverse: #ffffff; /* White on dark bg */
}
```

**Usage:**

```css
.heading {
  color: var(--color-text-primary);
}

.description {
  color: var(--color-text-secondary);
}

.disabled {
  color: var(--color-text-tertiary);
}
```

---

## Background Colors

```css
:root {
  --color-bg-primary: #ffffff; /* White */
  --color-bg-secondary: #f9fafb; /* Gray-50 */
  --color-bg-tertiary: #f3f4f6; /* Gray-100 */
  --color-bg-overlay: rgba(0, 0, 0, 0.5); /* Modal backdrop */
}
```

---

## Border Colors

```css
:root {
  --color-border-light: #e5e7eb; /* Gray-200 */
  --color-border-medium: #d1d5db; /* Gray-300 */
  --color-border-dark: #9ca3af; /* Gray-400 */
}
```

---

## Status Colors (ROAS/ROI)

```css
.status-excellent {
  color: #166534; /* Deep green */
}

.status-good {
  color: #1d4ed8; /* Blue */
}

.status-ok {
  color: #92400e; /* Dark yellow */
}

.status-poor {
  color: #991b1b; /* Deep red */
}
```

---

## Accessibility & Contrast

### Minimum Contrast Ratios (WCAG AA)

| Text Size       | Contrast Ratio | Example                       |
| --------------- | -------------- | ----------------------------- |
| **Normal text** | 4.5:1          | `#6b7280` on `#ffffff`        |
| **Large text**  | 3:1            | `#9ca3af` on `#ffffff`        |
| **UI elements** | 3:1            | Border `#d1d5db` on `#ffffff` |

### Safe Color Combinations

✅ **Good Contrast:**

- Gray-800 (`#1f2937`) on White (`#ffffff`) - 14.6:1
- Gray-700 (`#374151`) on White - 10.5:1
- Gray-600 (`#4b5563`) on White - 7.5:1
- Gray-500 (`#6b7280`) on White - 4.9:1 ✅ (meets 4.5:1)

⚠️ **Borderline:**

- Gray-400 (`#9ca3af`) on White - 2.8:1 ❌ (use for large text only)

---

## Usage Guidelines

### Do's ✅

- Use semantic colors for their intended purpose
- Maintain consistent contrast ratios
- Use brand colors sparingly for emphasis
- Test color combinations for accessibility

### Don'ts ❌

- Don't use red for success or green for errors
- Don't rely on color alone (add icons/text)
- Don't use low-contrast combinations
- Don't override platform brand colors

---

## Dark Mode (Future)

```css
:root {
  /* Light mode (default) */
  --color-text: #1f2937;
  --color-bg: #ffffff;
}

@media (prefers-color-scheme: dark) {
  :root {
    --color-text: #f3f4f6;
    --color-bg: #1f2937;
  }
}
```

---

## Quick Reference

| Color          | Hex       | Use Case          |
| -------------- | --------- | ----------------- |
| **Primary**    | `#3b82f6` | Buttons, links    |
| **Success**    | `#10b981` | Success states    |
| **Error**      | `#ef4444` | Error states      |
| **Warning**    | `#f59e0b` | Warning states    |
| **Shopee**     | `#f53d2d` | Shopee branding   |
| **TikTok**     | `#1d4ed8` | TikTok branding   |
| **Lazada**     | `#0f1689` | Lazada branding   |
| **Text**       | `#1f2937` | Primary text      |
| **Text Muted** | `#6b7280` | Secondary text    |
| **Border**     | `#e5e7eb` | Borders, dividers |
| **Background** | `#f9fafb` | Light backgrounds |

---

**See Also:** [Accessibility](./accessibility.md) | [Typography](./typography.md)
