# Typography

## Font Families

```css
:root {
  --font-sans:
    -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  --font-mono: "SF Mono", Monaco, Consolas, monospace;
}
```

---

## Font Scale

| Token     | Size     | Px   | Use Case            |
| --------- | -------- | ---- | ------------------- |
| text-xs   | 0.75rem  | 12px | Helper text, badges |
| text-sm   | 0.875rem | 14px | Labels, small text  |
| text-base | 1rem     | 16px | Body text           |
| text-lg   | 1.125rem | 18px | Emphasis            |
| text-xl   | 1.25rem  | 20px | H4                  |
| text-2xl  | 1.5rem   | 24px | H3, metrics         |
| text-3xl  | 1.875rem | 30px | H2                  |
| text-4xl  | 2.25rem  | 36px | H1 (mobile)         |
| text-5xl  | 3rem     | 48px | H1 (desktop)        |

---

## Font Weights

| Weight   | Value | Use Case                       |
| -------- | ----- | ------------------------------ |
| normal   | 400   | Body text                      |
| medium   | 500   | Emphasis                       |
| semibold | 600   | Buttons, labels, table headers |
| bold     | 700   | Headings                       |

---

## Heading Styles

| Heading | Desktop | Mobile | Weight |
| ------- | ------- | ------ | ------ |
| H1      | 40px    | 32px   | 700    |
| H2      | 30px    | 24px   | 700    |
| H3      | 24px    | 20px   | 600    |
| H4      | 20px    | 18px   | 600    |
| H5      | 18px    | 16px   | 600    |
| H6      | 16px    | 16px   | 600    |

```css
h1 {
  font-size: 2.5rem;
  font-weight: 700;
  line-height: 1.2;
}
h2 {
  font-size: 1.875rem;
  font-weight: 700;
  line-height: 1.3;
}
h3 {
  font-size: 1.5rem;
  font-weight: 600;
  line-height: 1.4;
}
```

---

## Text Colors

| Color    | Hex       | Use Case       |
| -------- | --------- | -------------- |
| Gray-800 | `#1f2937` | Headings       |
| Gray-700 | `#374151` | Body text      |
| Gray-600 | `#4b5563` | Secondary text |
| Gray-500 | `#6b7280` | Muted text     |
| Blue-600 | `#2563eb` | Links          |

---

## Text Utilities

```css
/* Truncation */
.truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
```

---

## Special Styles

| Style  | Size | Weight | Color    |
| ------ | ---- | ------ | -------- |
| Label  | 14px | 600    | Gray-700 |
| Helper | 12px | 400    | Gray-500 |
| Badge  | 12px | 600    | Varies   |
| Metric | 24px | 600    | Gray-800 |

---

## Links

```css
a {
  color: #3b82f6;
  text-decoration: none;
}
a:hover {
  color: #2563eb;
  text-decoration: underline;
}
a:focus {
  outline: 2px solid #3b82f6;
  outline-offset: 2px;
}
```

---

## Code

```css
code {
  font-family: var(--font-mono);
  font-size: 0.875em;
  background: #f3f4f6;
  padding: 2px 6px;
  border-radius: 4px;
}

pre {
  font-family: var(--font-mono);
  background: #1f2937;
  color: #f3f4f6;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
}
```

---

## Accessibility

- ✅ Base font size: 16px (1rem)
- ✅ Minimum text contrast: 4.5:1
- ✅ Line height: 1.5 for body text
- ✅ Max line length: 65-75 characters

---

**See Also:** [Colors](./colors.md) | [Accessibility](./accessibility.md)
